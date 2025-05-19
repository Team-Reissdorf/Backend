package performanceManagement

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/LucaSchmitz2003/FlowWatch"

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

var csvColumnCount = 10

// BulkCreatePerformanceEntries allows bulk creation of performance entries from a CSV file
// @Summary      Bulk create performance entries from CSV
// @Description  Upload a .csv file to bulk-create performance entries.
// @Tags         Performance Management
// @Accept       multipart/form-data
// @Produce      json
// @Param        Performances  formData  file  true  "CSV file; must have extension .csv; columns: lastName;firstName;gender;birthYear;birthDate;exercise;category;date;result;points"
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

	// File from multipart/form-data
	f, err1 := c.FormFile("Performances")
	if err1 != nil {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "File Field `Performances` in Request is missing"})
		return
	}
	// Check MIME type
	fileHeader := f.Header.Get("Content-Type")
	if !strings.HasPrefix(fileHeader, "text/csv") && !strings.HasPrefix(fileHeader, "application/vnd.ms-excel") {
		err := errors.New("Invalid file type, only CSV files are allowed")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err.Error()})
		return
	}

	infile, err2 := f.Open()
	if err2 != nil {
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Could not open file"})
		return
	}
	defer infile.Close()

	// CSV-Reader with comma.delimeter
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

	for i, rec := range records {
		rowNum := i + 1

		// Spaltenanzahl
		if len(rec) < csvColumnCount {
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid column count"})
			continue
		}

		// read all fields
		lastName := strings.TrimSpace(rec[0])
		firstName := strings.TrimSpace(rec[1])
		gender := strings.ToLower(rec[2])
		birthYearStr := strings.TrimSpace(rec[3])
		birthDateRaw := strings.TrimSpace(rec[4])
		if t, err := time.Parse("02.01.2006", birthDateRaw); err == nil {
			birthDateRaw = t.Format("2006-01-02")
		}
		exerciseName := strings.TrimSpace(rec[5])
		category := strings.TrimSpace(rec[6])
		performanceDate := strings.TrimSpace(rec[7])
		resultRaw := strings.TrimSpace(rec[8])
		pointsStr := strings.TrimSpace(rec[9])

		FlowWatch.GetLogHelper().Debug(ctx, "Results are ", resultRaw)

		// parse birthYear
		_, err4 := strconv.Atoi(birthYearStr)
		if err4 != nil {
			FlowWatch.GetLogHelper().Debug(ctx, "Failed to parse birth year", err4)
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid birthyear"})
			continue
		}

		// find exercise
		exercise, err5 := exerciseManagement.GetExerciseByNameAndDiscipline(ctx, exerciseName, category)
		if err5 != nil {
			FlowWatch.GetLogHelper().Debug(ctx, "Failed to get exercise", err5)
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Exercise not found"})
			continue
		}

		// validate date
		// This is for a design issue revolving the date format in the csv file
		// The date format in the csv file is dd.mm.yyyy
		// The date format in the db is yyyy-mm-dd
		// So we need to convert the date format from dd.mm.yyyy to yyyy-mm-dd
		// Instead of using strings, we should use a date format library like time which we already use in the rest of the code
		if errors.Is(formatHelper.IsDate(performanceDate), formatHelper.DateFormatInvalidError) {
			if len(performanceDate) == 10 {
				performanceDate = performanceDate[6:10] + "-" + performanceDate[3:5] + "-" + performanceDate[0:2]
			}
		}

		if err6 := formatHelper.IsDate(performanceDate); err6 != nil {
			FlowWatch.GetLogHelper().Debug(ctx, "Failed to parse date", err6)
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid date"})
			continue
		}

		if err7 := formatHelper.IsFuture(performanceDate); err7 != nil {
			FlowWatch.GetLogHelper().Debug(ctx, "Failed to check future", err7)
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid date. Date is in the future"})
			continue
		}

		gender = strings.TrimSpace(gender)

		if len(gender) == 0 {
			FlowWatch.GetLogHelper().Debug(ctx, "Empty sex attribute in record: ", rec)
			c.AbortWithStatusJSON(http.StatusBadRequest,
				endpoints.ErrorResponse{Error: "Sex attribute cannot be empty"})
			return
		}
		gender = gender[:1]

		// Normalize the sex attribute
		switch gender {
		case "m", "f", "d":

		case "w":
			gender = "f"
		default:
			FlowWatch.GetLogHelper().Debug(ctx, "Invalid sex attribute: ", gender)
			c.AbortWithStatusJSON(http.StatusBadRequest,
				endpoints.ErrorResponse{Error: fmt.Sprintf("Invalid sex attribute: %s", gender)})
			return
		}
		
		// validate gender
		if err8 := formatHelper.IsSex(gender); err8 != nil {
			FlowWatch.GetLogHelper().Debug(ctx, "Failed to parse gender", err8)
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid gender"})
			continue
		}

		// TODO: Check
		// validate result   -> ...
		time_units := []string{"second", "minute"}
		if slices.Contains(time_units, exercise.Unit) {
			if err9 := formatHelper.IsDuration(resultRaw); err9 != nil {
				FlowWatch.GetLogHelper().Debug(ctx, "Failed to validate result", err9)
				failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid result format"})
				continue
			}
		}

		// normalize units
		normalizedResult, err10 := formatHelper.NormalizeResult(resultRaw, exercise.Unit)
		if err10 != nil {
			FlowWatch.GetLogHelper().Debug(ctx, "Failed to normalize result", err10)
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Failed to normalize result"})
			continue
		}

		// parse points
		_, err11 := strconv.Atoi(pointsStr)
		if err11 != nil {
			FlowWatch.GetLogHelper().Debug(ctx, "Failed to parse points", err11)
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Invalid points"})
			continue
		}

		// find athlete
		// TODO: check if the athlete exists bzw. if the function works
		athlete, err12 := athleteManagement.GetAthleteByDetails(ctx, firstName, lastName, birthDateRaw, trainerEmail)
		if err12 != nil {
			FlowWatch.GetLogHelper().Debug(ctx, "Failed to get athlete by details", err12)
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Athlete not found"})
			continue
		}

		// validate date
		// This is for a design issue revolving the date format in the csv file
		// The date format in the csv file is dd.mm.yyyy
		// The date format in the db is yyyy-mm-dd
		// So we need to convert the date format from dd.mm.yyyy to yyyy-mm-dd
		// Instead of using strings, we should use a date format library like time which we already use in the rest of the code
		if errors.Is(formatHelper.IsDate(birthDateRaw), formatHelper.DateFormatInvalidError) {
			if len(birthDateRaw) == 10 {
				birthDateRaw = birthDateRaw[6:10] + "-" + birthDateRaw[3:5] + "-" + birthDateRaw[0:2]
			}
		}

		// calculate age
		age, err13 := athleteManagement.CalculateAge(ctx, birthDateRaw)
		if err13 != nil {
			FlowWatch.GetLogHelper().Debug(ctx, "Failed to calculate age", err13)
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Age could not be calculated"})
			continue
		}

		medalStatus, err14 := evaluateMedalStatus(ctx, exercise.ID, performanceDate, age, athlete.Sex, uint64(normalizedResult))
		if err14 != nil {
			FlowWatch.GetLogHelper().Debug(ctx, "Failed to evaluate result", err14)
			failedEntries = append(failedEntries, FailedPerformanceEntry{Row: rowNum, Reason: "Could not evaluate medal status"})
			continue
		}

		performanceEntries = append(performanceEntries, databaseUtils.Performance{
			AthleteId:  athlete.ID,
			ExerciseId: exercise.ID,
			Date:       performanceDate,
			Points:     uint64(normalizedResult),
			Medal:      medalStatus,
		})

		FlowWatch.GetLogHelper().Debug(ctx, "Performance entry created", performanceEntries)
	}

	//  error: no entries
	if len(performanceEntries) == 0 {
		c.AbortWithStatusJSON(http.StatusConflict, endpoints.ErrorResponse{Error: "No entries found"})
		return
	}

	// Bulk-Insert
	if err15 := createNewPerformances(ctx, performanceEntries); err15 != nil {
		endpoints.Logger.Error(ctx, errors.Wrap(err15, "Failed to create performance entries"))
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Error while saving"})
		return
	}

	// list with wrong rows as response
	c.JSON(http.StatusCreated, BulkCreatePerformanceResponse{
		Message:       "Bulk creation successful",
		FailedEntries: failedEntries,
	})
}
