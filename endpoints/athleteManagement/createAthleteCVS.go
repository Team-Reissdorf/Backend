package athleteManagement

import (
	"encoding/csv"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"mime/multipart"
	"net/http"
	"strings"
)

type AlreadyExistingAthletesResponse struct {
	Message                 string                   `json:"message" example:"Creation successful"`
	AlreadyExistingAthletes []databaseModels.Athlete `json:"alreadyExistingAthletes"`
}

var csvColumnCount = 5

// CreateAthleteCVS bulk creates new athletes in the db from a cvs file
// @Summary Bulk creates new athletes from cvs file
// @Description Upload a CSV file to create multiple athlete profiles. If an athlete already exists, the process will continue, and the response will indicate which athletes already exist.
// @Tags Athlete Management
// @Accept multipart/form-data
// @Produce json
// @Param Athletes formData file true "CSV file containing details of multiple athletes to create profiles"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 201 {object} endpoints.SuccessResponse "Creation successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 409 {object} endpoints.ErrorResponse "One or more athlete(s) already exist"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/athlete/bulk-create [post]
func CreateAthleteCVS(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "CreateMultipleAthletes")
	defer span.End()

	// Bind body to csv file
	file, err1 := c.FormFile("Athletes")
	if err1 != nil || file == nil {
		err1 = errors.Wrap(err1, "Failed to get the file")
		endpoints.Logger.Debug(ctx, err1)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "File is missing or invalid"})
		c.Abort()
		return
	}

	// Check MIME type
	fileHeader := file.Header.Get("Content-Type")
	if !strings.HasPrefix(fileHeader, "text/csv") && !strings.HasPrefix(fileHeader, "application/vnd.ms-excel") {
		err := errors.New("Invalid file type, only CSV files are allowed")
		endpoints.Logger.Debug(ctx, err)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err.Error()})
		c.Abort()
		return
	}

	// Get the user id from the context
	// userId := authHelper.GetUserIdFromContext(ctx, c)
	// ToDo: Verify that the user is a trainer
	trainerEmail := "blabla@test.com"

	// Open the CSV file
	fileContent, err2 := file.Open()
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to open file")
		endpoints.Logger.Debug(ctx, err2)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Could not open file"})
		c.Abort()
		return
	}
	defer func(fileContent multipart.File) {
		err := fileContent.Close()
		if err != nil {
			err = errors.Wrap(err, "Failed to close file")
			endpoints.Logger.Error(ctx, err)
		}
	}(fileContent)

	// Read CSV file
	reader := csv.NewReader(fileContent)
	records, err3 := reader.ReadAll()
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to read csv")
		endpoints.Logger.Debug(ctx, err3)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "File could not be read. Invalid CSV format?"})
		c.Abort()
		return
	}

	// Parse data
	var athletes []databaseModels.Athlete
	for _, record := range records {
		// Ensure the column count is correct
		if len(record) != csvColumnCount {
			err := errors.New("Inconsistent number of columns in the CSV file")
			endpoints.Logger.Debug(ctx, err)
			c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err.Error()})
			c.Abort()
			return
		}

		// Map CSV data to an athlete object
		athlete := databaseModels.Athlete{
			FirstName:    record[0],
			LastName:     record[1],
			BirthDate:    record[3],
			Sex:          record[4],
			Email:        record[2],
			TrainerEmail: trainerEmail,
		}
		athletes = append(athletes, athlete)
	}

	// Write athletes to the db
	err4, alreadyExistingAthletes := createNewAthletes(ctx, athletes)
	if errors.Is(err4, formatHelper.InvalidSexLengthError) || errors.Is(err4, formatHelper.InvalidSexValue) {
		endpoints.Logger.Debug(ctx, err4)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Sex needs to be <m|f|d>"})
		c.Abort()
		return
	} else if errors.Is(err1, NoNewAthletesError) {
		endpoints.Logger.Debug(ctx, err1)
		c.JSON(http.StatusConflict, endpoints.ErrorResponse{Error: "No new Athletes"})
		c.Abort()
		return
	} else if err4 != nil {
		err4 = errors.Wrap(err4, "Failed to create the athletes")
		endpoints.Logger.Error(ctx, err4)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		c.Abort()
		return
	}

	c.JSON(
		http.StatusCreated,
		AlreadyExistingAthletesResponse{
			Message:                 "Creation successful",
			AlreadyExistingAthletes: alreadyExistingAthletes,
		},
	)
}
