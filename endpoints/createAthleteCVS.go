package endpoints

import (
	"encoding/csv"
	"github.com/Team-Reissdorf/Backend/database_models"
	"github.com/Team-Reissdorf/Backend/endpoints/standardJsonAnswers"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"mime/multipart"
	"net/http"
	"strings"
)

type AlreadyExistingAthletesResponse struct {
	Message                 string                    `json:"message" example:"Creation successful"`
	AlreadyExistingAthletes []database_models.Athlete `json:"alreadyExistingAthletes"`
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
// @Success 201 {object} standardJsonAnswers.SuccessResponse "Creation successful"
// @Failure 400 {object} standardJsonAnswers.ErrorResponse "Invalid request body"
// @Failure 409 {object} standardJsonAnswers.ErrorResponse "One or more athlete(s) already exist"
// @Failure 500 {object} standardJsonAnswers.ErrorResponse "Internal server error"
// @Router /v1/athlete/bulk-create [post]
func CreateAthleteCVS(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "CreateMultipleAthletes")
	defer span.End()

	// Bind body to csv file
	file, err1 := c.FormFile("Athletes")
	if err1 != nil || file == nil {
		err1 = errors.Wrap(err1, "Failed to get the file")
		logger.Debug(ctx, err1)
		c.JSON(http.StatusBadRequest, standardJsonAnswers.ErrorResponse{Error: "File is missing or invalid"})
		c.Abort()
		return
	}

	// Check MIME type
	fileHeader := file.Header.Get("Content-Type")
	if !strings.HasPrefix(fileHeader, "text/csv") && !strings.HasPrefix(fileHeader, "application/vnd.ms-excel") {
		err := errors.New("Invalid file type, only CSV files are allowed")
		logger.Debug(ctx, err)
		c.JSON(http.StatusBadRequest, standardJsonAnswers.ErrorResponse{Error: err.Error()})
		c.Abort()
		return
	}

	// Get the user id from the context
	// userId := authHelper.GetUserIdFromContext(ctx, c)
	// ToDo: Verify that the user is a trainer

	// Open the CSV file
	fileContent, err2 := file.Open()
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to open file")
		logger.Debug(ctx, err2)
		c.JSON(http.StatusBadRequest, standardJsonAnswers.ErrorResponse{Error: "Could not open file"})
		c.Abort()
		return
	}
	defer func(fileContent multipart.File) {
		err := fileContent.Close()
		if err != nil {
			err = errors.Wrap(err, "Failed to close file")
			logger.Error(ctx, err)
		}
	}(fileContent)

	// Read CSV file
	reader := csv.NewReader(fileContent)
	records, err3 := reader.ReadAll()
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to read csv")
		logger.Debug(ctx, err3)
		c.JSON(http.StatusBadRequest, standardJsonAnswers.ErrorResponse{Error: "File could not be read. Invalid CSV format?"})
		c.Abort()
		return
	}

	// Parse data
	var athletes []database_models.Athlete
	for _, record := range records {
		// Ensure the column count is correct
		if len(record) != csvColumnCount {
			err := errors.New("Inconsistent number of columns in the CSV file")
			logger.Debug(ctx, err)
			c.JSON(http.StatusBadRequest, standardJsonAnswers.ErrorResponse{Error: err.Error()})
			c.Abort()
			return
		}

		// Check formats
		email := record[2]
		if err := formatHelper.IsEmail(email); err != nil {
			err = errors.Wrap(err, "Invalid email address")
			logger.Debug(ctx, err)
			c.JSON(http.StatusBadRequest, standardJsonAnswers.ErrorResponse{Error: "Invalid email address"})
			c.Abort()
			return
		}

		birthDate := record[3]
		if err := formatHelper.IsDate(birthDate); err != nil {
			err = errors.Wrap(err, "Invalid date")
			logger.Debug(ctx, err)
			c.JSON(http.StatusBadRequest, standardJsonAnswers.ErrorResponse{Error: "Invalid birth date"})
			c.Abort()
			return
		}

		sex := strings.ToLower(string(record[4][0]))

		// Map CSV data to an athlete object
		athlete := database_models.Athlete{
			Person: database_models.Person{
				FirstName: record[0],
				LastName:  record[1],
				Email:     email,
			},
			BirthDate: birthDate,
			Sex:       sex,
		}
		athletes = append(athletes, athlete)
	}

	// Write athletes to the db
	err4, alreadyExistingAthletes := createNewAthletes(ctx, athletes)
	if err4 != nil {
		err4 = errors.Wrap(err4, "Failed to create the athletes")
		logger.Error(ctx, err4)
		c.JSON(http.StatusInternalServerError, standardJsonAnswers.ErrorResponse{Error: "Internal server error"})
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
