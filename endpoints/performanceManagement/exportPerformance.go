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

	// Help function for rating the medals (gold > silver > bronze)
	medalRank := func(medal string) int {
		switch medal {
		case "gold":
			return 3
		case "silver":
			return 2
		case "bronze":
			return 1
		default:
			return 0
		}
	}

	// Iterate over all athlete IDs
	for _, athleteID := range req.AthleteIDs {
		// Retrieve athlete information
		athlete, err := athleteManagement.GetAthlete(ctx, uint(athleteID), trainerEmail)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, "Could not find athlete")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Could not find athlete"})
			return
		} else if err != nil {
			err = errors.Wrap(err, "Failed to fetch athlete data")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to fetch athlete data"})
			return
		}

		// Retrieve all performance entries
		performances, err := getAllPerformanceBodies(ctx, uint(athleteID))
		if err != nil {
			err = errors.Wrap(err, "Failed to fetch performances")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to fetch performances"})
			return
		}

		//Group the performances per day (assuming p.Date has the format “YYYY-MM-DD”) only
		//the entry with the best medal (based on medalRank) is saved per day.
		bestPerformanceByDate := make(map[string]PerformanceBodyWithId)
		for _, p := range *performances {
			day := p.Date[:10]
			if existing, ok := bestPerformanceByDate[day]; !ok {
				bestPerformanceByDate[day] = p
			} else {
				if medalRank(p.Medal) > medalRank(existing.Medal) {
					bestPerformanceByDate[day] = p
				}
			}
		}

		// The days are sorted for consistent output
		var days []string
		for day := range bestPerformanceByDate {
			days = append(days, day)
		}
		sort.Strings(days)

		// Write the best entry for each day in the CSV
		for _, day := range days {
			p := bestPerformanceByDate[day]

			// Retrieve exercise for the respective performance entry
			var exercise databaseUtils.Exercise
			err := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
				return tx.First(&exercise, p.ExerciseId).Error
			})
			if err != nil {
				err = errors.Wrap(err, "Failed to get the exercise")
				endpoints.Logger.Debug(ctx, err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the exercise"})
				return
			}

			// Formatting of date and birthday
			birthday := athlete.BirthDate
			birthyear := birthday[:4]
			birthdate := birthday[8:10] + "." + birthday[5:7] + "." + birthday[:4]
			formattedDate := p.Date[8:10] + "." + p.Date[5:7] + "." + p.Date[:4]

			sex := athlete.Sex
			if sex == "f" {
				sex = "w"
			}

			// Convert Points to right format
			var formattedPoints string
			switch exercise.Unit {
			case "second":
				secs := p.Points / 1_000
				formattedPoints = strconv.FormatUint(secs, 10)
			case "minute":
				totalSec := p.Points / 1_000
				mins := totalSec / 60
				secs := totalSec % 60
				formattedPoints = fmt.Sprintf("%d:%02d", mins, secs)
			case "meter":
				meters := float64(p.Points) / 100
				formattedPoints = fmt.Sprintf("%.2f", meters)
			default:
				formattedPoints = strconv.FormatUint(p.Points, 10)
			}

			// Write a CSV line
			_ = w.Write([]string{
				athlete.LastName,
				athlete.FirstName,
				sex,
				birthyear,
				birthdate,
				exercise.Name,
				exercise.DisciplineName,
				formattedDate,
				formattedPoints,
				p.Medal,
			})
		}
	}
}
