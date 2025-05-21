package performanceManagement

import (
	"encoding/csv"
	"fmt"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"sort"
	"strconv"
)

// ExportRequest defines the athlete IDs to be exported.
type ExportRequest struct {
	AthleteIDs []int `json:"athlete_ids" example:"1"`
}

// PerformanceCSV defines the CSV format of the exported performance data.
type PerformanceCSV struct {
	AthleteLastName  string
	AthleteFirstName string
	AthleteGender    string
	AthleteBirthYear string
	AthleteBirthday  string
	Exercise         string
	Discipline       string
	Date             string
	Medal            string
	Points           uint64
}

// ExportPerformances exports all performance entries of the specified athletes as a csv file.
// Only the entry with the best medal is exported per day (gold > silver > bronze).
// @Summary Exports all performance entries of the specified athletes as a csv file
// @Description Exports all performance entries of the specified athletes as a csv file
// @Tags Performance Management
// @Produce text/csv
// @Param json body ExportRequest true "JSON payload in the format: "athlete_ids": [] "
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {file} file "CSV file"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "One or more athletes do not exist"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/export [post]
func ExportPerformances(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "ExportPerformances")
	defer span.End()

	// Read in JSON body
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid request body"})
		return
	}

	// If no IDs were transferred
	if len(req.AthleteIDs) == 0 {
		var err error
		err = errors.Wrap(err, "No Athlete IDs provided")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "No athlete IDs provided"})
		return
	}

	// Set CSV header
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=performances.csv")
	w := csv.NewWriter(c.Writer)
	w.Comma = ';'
	defer w.Flush()

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// iterate over each athlete ID
	for _, athleteID := range req.AthleteIDs {
		// fetch athlete information
		athlete, err := athleteManagement.GetAthlete(ctx, uint(athleteID), trainerEmail)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			endpoints.Logger.Debug(ctx, errors.Wrap(err, "athlete not found"))
			c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete not found"})
			return
		} else if err != nil {
			endpoints.Logger.Debug(ctx, errors.Wrap(err, "failed to fetch athlete data"))
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to fetch athlete data"})
			return
		}

		// retrieve all performance entries for this athlete
		performances, err := getAllPerformanceBodies(ctx, uint(athleteID))
		if err != nil {
			endpoints.Logger.Debug(ctx, errors.Wrap(err, "failed to fetch performances"))
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to fetch performances"})
			return
		}

		// group performances by exercise ID and day
		type groupKey struct {
			ExerciseID uint
			Day        string
		}
		grouped := make(map[groupKey][]PerformanceBodyWithId)
		for _, p := range *performances {
			day := p.Date[:10]
			key := groupKey{ExerciseID: p.ExerciseId, Day: day}
			grouped[key] = append(grouped[key], p)
		}

		// extract and sort group keys for stable output
		var keys []groupKey
		for k := range grouped {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].ExerciseID != keys[j].ExerciseID {
				return keys[i].ExerciseID < keys[j].ExerciseID
			}
			return keys[i].Day < keys[j].Day
		})

		// process each group: select best performance and write to CSV
		for _, k := range keys {
			entries := grouped[k]

			// determine the best performance entry using goal-based logic
			bestEntry, err := getBestPerformanceEntry(ctx, &entries)
			if err != nil {
				endpoints.Logger.Debug(ctx, errors.Wrap(err, "failed to determine best performance"))
				c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to determine best performance"})
				return
			}

			// fetch exercise metadata
			var exercise databaseUtils.Exercise
			if err := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
				return tx.First(&exercise, k.ExerciseID).Error
			}); err != nil {
				endpoints.Logger.Debug(ctx, errors.Wrap(err, "failed to fetch exercise metadata"))
				c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to fetch exercise metadata"})
				return
			}

			// prepare athlete and date fields
			birthRaw := athlete.BirthDate // "YYYY-MM-DD"
			birthYear := birthRaw[:4]
			birthDate := birthRaw[8:10] + "." + birthRaw[5:7] + "." + birthRaw[:4]

			sex := athlete.Sex
			if sex == "f" {
				sex = "w"
			}

			formattedDate := k.Day[8:10] + "." + k.Day[5:7] + "." + k.Day[:4]

			// format points according to exercise unit
			var formattedPoints string
			switch exercise.Unit {
			case "second":
				secs := bestEntry.Points / 1_000
				formattedPoints = strconv.FormatUint(secs, 10)
			case "minute":
				totalSec := bestEntry.Points / 1_000
				mins := totalSec / 60
				secs := totalSec % 60
				formattedPoints = fmt.Sprintf("%d:%02d", mins, secs)
			case "meter":
				meters := float64(bestEntry.Points) / 100
				formattedPoints = fmt.Sprintf("%.2f", meters)
			default:
				formattedPoints = strconv.FormatUint(bestEntry.Points, 10)
			}

			var medal string
			switch bestEntry.Medal {
			case "gold":
				medal = "3"
			case "silver":
				medal = "2"
			case "bronze":
				medal = "1"
			default:
				medal = "0"
			}

			// write CSV record for this best performance
			_ = w.Write([]string{
				athlete.LastName,
				athlete.FirstName,
				sex,
				birthYear,
				birthDate,
				exercise.Name,
				exercise.DisciplineName,
				formattedDate,
				formattedPoints,
				medal,
			})
		}
	}
}
