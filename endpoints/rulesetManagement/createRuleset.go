package rulesetManagement

import (
	"encoding/csv"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"mime/multipart"
	"net/http"
	"strings"
)

var csvColumnCount = 5

// CreateRuleset creates new ruleset entries in the db from a csv file
// @Summary Creates new ruleset entries from csv file
// @Description Upload a CSV file to create multiple ruleset entries.
// @Tags Ruleset Management
// @Accept multipart/form-data
// @Produce json
// @Param RulesetEntries formData file true "CSV file containing details of the ruleset"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} endpoints.SuccessResponse "Creation successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 409 {object} endpoints.ErrorResponse "All ruleset entries already exist; none have been created"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/ruleset/create [post]
func CreateRuleset(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "CreateRulesetEntries")
	defer span.End()

	// Bind body to csv file
	file, err1 := c.FormFile("RulesetEntries")
	if err1 != nil || file == nil {
		err1 = errors.Wrap(err1, "Failed to get the file")
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "File is missing or invalid"})
		return
	}

	// Check MIME type
	fileHeader := file.Header.Get("Content-Type")
	if !strings.HasPrefix(fileHeader, "text/csv") && !strings.HasPrefix(fileHeader, "application/vnd.ms-excel") {
		err := errors.New("Invalid file type, only CSV files are allowed")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err.Error()})
		return
	}

	// Get the user id from the context
	// trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Open the CSV file
	fileContent, err2 := file.Open()
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to open file")
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Could not open file"})
		return
	}
	defer func(fileContent multipart.File) {
		err := fileContent.Close()
		if err != nil {
			err = errors.Wrap(err, "Failed to close file")
			endpoints.Logger.Error(ctx, err)
		}
	}(fileContent)

	// Read the CSV file
	reader := csv.NewReader(fileContent)
	records, err3 := reader.ReadAll()
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to read csv. Invalid CSV format?")
		endpoints.Logger.Warn(ctx, err3)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "File could not be read. Invalid CSV format?"})
		return
	}

	// Parse data
	for _, record := range records {
		// Ensure the column count is correct
		if len(record) != csvColumnCount {
			err := errors.New("Inconsistent number of columns in the CSV file")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err.Error()})
			return
		}
	}
}
