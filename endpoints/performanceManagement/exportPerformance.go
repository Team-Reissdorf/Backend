package performanceManagement

import (
	"encoding/csv"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type ExportRequest struct {
	AthleteIDs []int `json:"athlete_ids"`
}

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

// ExportPerformances exports all performance entries of the given athletes as a csv-file
// @Summary Exports all performance entries of the given athletes as a csv-file
// @Description Exports all performance entries of the given athletes as a csv-file
// @Tags Performance Management
// @Produce text/csv
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} PerformanceCSV "Request successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "One or more athletes do not exist"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/export [post]
func ExportPerformances(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "GetOneAthlete")
	defer span.End()

	// JSON-Body einlesen
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Falls keine IDs übergeben wurden
	if len(req.AthleteIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No athlete IDs provided"})
		return
	}

	// CSV-Header setzen
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=performances.csv")
	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	// CSV-Header schreiben
	//_ = w.Write([]string{"Nachname", "Vorname", "Geschlecht", "Geburtsjahr", "Geburtstag", "Übung", "Disziplin", "Datum", "Medaille", "Punkte"})

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Über alle Athlete-IDs iterieren und Performances abrufen
	for _, athleteID := range req.AthleteIDs {
		// Athleteninformationen abrufen
		athlete, err := athleteManagement.GetAthlete(ctx, uint(athleteID), trainerEmail)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Athlete not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch athlete data"})
			return
		}

		// Performances abrufen
		performances, err := getAllPerformanceBodies(ctx, uint(athleteID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch performances"})
			return
		}

		// Performances in CSV schreiben mit Athleten-Infos
		for _, p := range *performances {
			birthday := athlete.BirthDate
			birthyear := birthday[:4]
			birthdate := birthday[8:10] + "." + birthday[5:7] + "." + birthday[:4]
			date := p.Date[8:10] + "." + p.Date[5:7] + "." + p.Date[:4]
			sex := athlete.Sex
			if sex == "f" {
				sex = "w"
			}
			_ = w.Write([]string{
				athlete.LastName,
				athlete.FirstName,
				sex,
				birthyear,
				birthdate,
				strconv.Itoa(int(p.ExerciseId)),
				strconv.Itoa(int(p.PerformanceId)),
				date,
				p.Medal,
				strconv.FormatUint(p.Points, 10),
			})
		}
	}
}
