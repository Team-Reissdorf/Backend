package performanceManagement

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"

	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
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
// @Summary Bulk create performance entries from CSV file
// @Description Upload a CSV file to bulk create performance entries. If some entries fail, they will be returned in response.
// @Tags Performance Management
// @Accept multipart/form-data
// @Produce json
// @Param Performances formData file true "CSV file containing performance entries"
// @Param Authorization header string false "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 201 {object} BulkCreatePerformanceResponse "Bulk creation successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 409 {object} endpoints.ErrorResponse "All entries failed, none have been created"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/bulk-create [post]
func BulkCreatePerformanceEntries(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "BulkCreatePerformanceEntries")
	defer span.End()

	file, err1 := c.FormFile("Performances")
	if err1 != nil || file == nil {
		err1 = errors.Wrap(err1, "Failed to get the file")
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "File is missing or invalid"})
		return
	}

	fileHeader := file.Header.Get("Content-Type")
	if !strings.HasPrefix(fileHeader, "text/csv") && !strings.HasPrefix(fileHeader, "application/vnd.ms-excel") {
		err := errors.New("Invalid file type, only CSV files are allowed")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err.Error()})
		return
	}

	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	fileContent, err2 := file.Open()
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to open file")
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Could not open file"})
		return
	}
	defer fileContent.Close()

	reader := csv.NewReader(fileContent)
	records, err3 := reader.ReadAll()
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to read CSV. Invalid format?")
		endpoints.Logger.Warn(ctx, err3)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid CSV format"})
		return
	}

	var performanceBodies []PerformanceBody
	var failedEntries []FailedPerformanceEntry
	var age int
	var errF error

	for i, record := range records {
		// Ignore rows with fewer than 5 columns (first + three valid + last)
		if len(record) < 5 {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: i + 1, Reason: "Invalid column count"})
			continue
		}

		// Trim whitespace and skip first and last columns
		trimmed := []string{}
		for _, cell := range record {
			trimmed = append(trimmed, strings.TrimSpace(cell))
		}
		athleteIDStr := trimmed[1]
		exerciseIDStr := trimmed[2]
		pointsStr := trimmed[3]
		date := trimmed[4]

		athleteID, errA := strconv.Atoi(athleteIDStr)
		exerciseID, errB := strconv.Atoi(exerciseIDStr)
		points, errC := strconv.Atoi(pointsStr)

		if errA != nil || errB != nil || errC != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: i + 1, Reason: "Invalid numeric values"})
			continue
		}

		if err := formatHelper.IsDate(date); err != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: i + 1, Reason: "Invalid date format"})
			continue
		}

		athlete, errD := athleteManagement.GetAthlete(ctx, uint(athleteID), trainerEmail)
		if errD != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: i + 1, Reason: "Athlete does not exist"})
			continue
		}

		birthDate, errE := formatHelper.FormatDate(athlete.BirthDate)
		if errE != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: i + 1, Reason: "Failed to parse birth date"})
			continue
		}
		age, errF = athleteManagement.CalculateAge(ctx, birthDate)
		if errF != nil {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: i + 1, Reason: "Failed to calculate age"})
			continue
		}

		performanceBodies = append(performanceBodies, PerformanceBody{
			Points:     uint64(points),
			Date:       date,
			ExerciseId: uint(exerciseID),
			AthleteId:  uint(athleteID),
		})
	}

	if len(performanceBodies) == 0 {
		c.AbortWithStatusJSON(http.StatusConflict, endpoints.ErrorResponse{Error: "All entries failed, none have been created"})
		return
	}

	performanceEntries, err6 := translatePerformanceBodies(ctx, performanceBodies, age, "")
	if err6 != nil {
		endpoints.Logger.Error(ctx, err6)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to process performance entries"})
		return
	}

	err7 := createNewPerformances(ctx, performanceEntries)
	if err7 != nil {
		err7 = errors.Wrap(err7, "Failed to create performance entries")
		endpoints.Logger.Error(ctx, err7)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to store performance entries"})
		return
	}

	c.JSON(http.StatusCreated, BulkCreatePerformanceResponse{
		Message:       "Bulk creation successful",
		FailedEntries: failedEntries,
	})
}
