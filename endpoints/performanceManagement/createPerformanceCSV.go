package performanceManagement

import (
	"encoding/csv"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/Team-Reissdorf/Backend/endpoints/exerciseManagement"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// BulkCreatePerformanceResponse defines the response structure for the bulk performance creation endpoint
type BulkCreatePerformanceResponse struct {
	Message       string                   `json:"message" example:"Bulk Creation successful"`
	FailedEntries []FailedPerformanceEntry `json:"failed_entries,omitempty"`
}

// FailedPerformanceEntry represents a failed CSV entry with the row number and reason for the error
type FailedPerformanceEntry struct {
	Row    int    `json:"row" example:"2"`
	Reason string `json:"reason" example:"Invalid athlete ID"`
}

// BulkCreatePerformanceEntries allows bulk creation of performance entries from a CSV file
// @Summary      Bulk create performance entries from CSV
// @Description  Upload a .csv file (semicolon-delimited, German Excel export) to bulk-create performance entries.
//
//	The first row is treated as header and skipped. Returns a list of rows that failed validation.
//
// @Tags         Performance
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "CSV file; must have extension .csv; columns: Nachname;Vorname;Geschlecht;Geburtsjahr;Geburtsdatum;Übung;Kategorie;Datum;Ergebnis;Punkte"
// @Param        Authorization  header  string  false  "Bearer JWT token"
// @Success      201  {object}  BulkCreatePerformanceResponse  "Bulk creation successful"
// @Failure      400  {object}  endpoints.ErrorResponse  "Bad request: missing file / invalid CSV / wrong extension"
// @Failure      401  {object}  endpoints.ErrorResponse  "Unauthorized: invalid or missing token"
// @Failure      409  {object}  endpoints.ErrorResponse  "Conflict: all entries failed, none created"
// @Failure      500  {object}  endpoints.ErrorResponse  "Internal server error (DB failure or file read error)"
// @Router       /v1/performance/bulk-create [post]
func BulkCreatePerformanceEntries(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "BulkCreatePerformanceEntries")
	defer span.End()

	// File aus dem multipart/form-data
	f, err1 := c.FormFile("file")
	if err1 != nil {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "File Field `file` in Request is missing"})
		return
	}
	// nur .csv zulassen
	if ext := strings.ToLower(filepath.Ext(f.Filename)); ext != ".csv" {
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Only .csv-files are allowed"})
		return
	}

	infile, err2 := f.Open()
	if err2 != nil {
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Could not open file"})
		return
	}
	defer infile.Close()

	// CSV-Reader mit Semikolon-Delimiter
	reader := csv.NewReader(infile)
	reader.Comma = ';'

	records, err3 := reader.ReadAll()
	if err3 != nil {
		endpoints.Logger.Warn(ctx, errors.Wrap(err3, "Failed to read CSV"))
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid CSV-Format"})
		return
	}

	var performanceEntries []databaseUtils.Performance
	var failedEntries []FailedPerformanceEntry

	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Zeile für Zeile, Header überspringen
	for i, rec := range records {
		if i == 0 {
			// Header
			continue
		}
		rowNum := i + 1

		// Spaltenanzahl
		if len(rec) < 10 {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid column count"})
			continue
		}

		// alle Felder einlesen
		lastName := strings.TrimSpace(rec[0])
		firstName := strings.TrimSpace(rec[1])
		gender := strings.TrimSpace(rec[2])
		birthYearStr := strings.TrimSpace(rec[3])
		birthDateRaw := strings.TrimSpace(rec[4])
		exerciseName := strings.TrimSpace(rec[5])
		category := strings.TrimSpace(rec[6])
		performanceDate := strings.TrimSpace(rec[7])
		resultRaw := strings.TrimSpace(rec[8])
		pointsStr := strings.TrimSpace(rec[9])

		// Geburtsjahr parsen
		_, err4 := strconv.Atoi(birthYearStr)
		if err4 != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid birthyear"})
			continue
		}

		// Exercise finden
		exercise, err5 := exerciseManagement.GetExerciseByNameAndDiscipline(ctx, exerciseName, category)
		if err5 != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Exercise not found"})
			continue
		}

		// Datum validieren
		if err6 := formatHelper.IsDate(performanceDate); err6 != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid date"})
			continue
		}
		if err7 := formatHelper.IsFuture(performanceDate); err7 != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid date. Date is in the future"})
			continue
		}

		// Geschlecht validieren
		if err8 := formatHelper.IsSex(gender); err8 != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{
				Row:    rowNum,
				Reason: "Invalid gender",
			})
			continue
		}

		// Ergebnis/Dauer validieren
		if err := formatHelper.IsDuration(resultRaw); err != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{
				Row:    rowNum,
				Reason: "Ungültiges Ergebnisformat",
			})
			continue
		}

		// Punkte parsen
		points, err9 := strconv.Atoi(pointsStr)
		if err9 != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid points"})
			continue
		}

		// Athlet finden (inkl. Geburtsjahr und Geschlecht)
		athlete, err10 := athleteManagement.GetAthleteByDetails(ctx, firstName, lastName, birthDateRaw, trainerEmail)
		if err10 != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Athlete not found"})
			continue
		}

		// Alter kalkulieren
		if _, err11 := athleteManagement.CalculateAge(ctx, birthDateRaw); err11 != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Age could not be calculated"})
			continue
		}

		// alles valid – Eintrag zum Bulk-Push vormerken
		performanceEntries = append(performanceEntries, databaseUtils.Performance{
			AthleteId:  athlete.ID,
			ExerciseId: exercise.ID,
			Date:       performanceDate,
			//Result:     resultRaw,      // neu im Struct
			Points: uint64(points),
		})
	}

	// 6) Fehlerfall: Keine Einträge
	if len(performanceEntries) == 0 {
		c.AbortWithStatusJSON(http.StatusConflict, endpoints.ErrorResponse{Error: "All entries failed"})
		return
	}

	// 7) Bulk-Insert
	if err12 := createNewPerformances(ctx, performanceEntries); err12 != nil {
		endpoints.Logger.Error(ctx, errors.Wrap(err12, "Failed to create performance entries"))
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Error while saving"})
		return
	}

	// 8) Antwort mit Liste der fehlerhaften Zeilen
	c.JSON(http.StatusCreated, BulkCreatePerformanceResponse{
		Message:       "Bulk creation successful",
		FailedEntries: failedEntries,
	})
}
