package athleteManagement

import (
	"encoding/csv"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type AlreadyExistingAthletesResponse struct {
	Message                 string                  `json:"message" example:"Creation successful"`
	AlreadyExistingAthletes []databaseUtils.Athlete `json:"already_existing_athletes"`
}

var csvColumnCount = 5

// CreateAthleteCSV bulk creates new athletes in the db from a csv file
// @Summary Bulk creates new athletes from csv file
// @Description Upload a CSV file to create multiple athlete profiles. If an athlete already exists, the process will continue, and the response will indicate which athletes already exist.
// @Tags Athlete Management
// @Accept multipart/form-data
// @Produce json
// @Param Athletes formData file true "CSV file containing details of multiple athletes to create profiles"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 201 {object} AlreadyExistingAthletesResponse "Creation successful"
// @Success 202 {object} AlreadyExistingAthletesResponse "Athletes already exist"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/athlete/bulk-create [post]
func CreateAthleteCSV(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "CreateMultipleAthletes")
	defer span.End()

	// Bind body to csv file
	file, err1 := c.FormFile("Athletes")
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
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

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
	reader.Comma = ';'
	records, err3 := reader.ReadAll()
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to read csv. Invalid CSV format?")
		endpoints.Logger.Warn(ctx, err3)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "File could not be read. Invalid CSV format?"})
		return
	}

	// Parse data
	var athleteEntries []databaseUtils.Athlete
	for _, record := range records {
		// Ensure the column count is correct
		if len(record) != csvColumnCount {
			err := errors.New("Inconsistent number of columns in the CSV file")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err.Error()})
			return
		}

		// Map CSV data to an athlete object
		athlete := databaseUtils.Athlete{
			FirstName:    record[0],
			LastName:     record[1],
			BirthDate:    record[3],
			Sex:          record[4],
			Email:        record[2],
			TrainerEmail: trainerEmail,
		}
		// This is for a design issue revolving the date format in the csv file
		// The date format in the csv file is dd.mm.yyyy
		// The date format in the db is yyyy-mm-dd
		// So we need to convert the date format from dd.mm.yyyy to yyyy-mm-dd
		// Instead of using strings, we should use a date format library like time which we already use in the rest of the code
		if formatHelper.IsDate(athlete.BirthDate) == formatHelper.DateFormatInvalidError {
			if len(athlete.BirthDate) == 10 {
				athlete.BirthDate = athlete.BirthDate[6:10] + "-" + athlete.BirthDate[3:5] + "-" + athlete.BirthDate[0:2]
			}
		}

		athleteEntries = append(athleteEntries, athlete)

		// Validate the athlete body
		err1 := validateAthlete(ctx, &athlete)
		if errors.Is(err1, formatHelper.InvalidSexLengthError) || errors.Is(err1, formatHelper.InvalidSexValue) {
			endpoints.Logger.Debug(ctx, err1)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Sex needs to be <m|f|d>"})
			return
		} else if errors.Is(err1, formatHelper.DateFormatInvalidError) {
			endpoints.Logger.Debug(ctx, err1)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid date format"})
			return
		} else if errors.Is(err1, formatHelper.DateInFutureError) {
			endpoints.Logger.Debug(ctx, err1)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Date is in the Future"})
			return
		} else if errors.Is(err1, formatHelper.InvalidEmailAddressFormatError) || errors.Is(err1, formatHelper.EmailAddressContainsNameError) || errors.Is(err1, formatHelper.EmailAddressInvalidTldError) {
			endpoints.Logger.Debug(ctx, err1)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid email address format"})
			return
		} else if err1 != nil {
			err1 = errors.Wrap(err1, "Failed to validate the athlete body")
			endpoints.Logger.Error(ctx, err1)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
			return
		}
	}

	// Write athletes to the db
	err4, alreadyExistingAthletes := createNewAthletes(ctx, athleteEntries)
	if errors.Is(err4, NoNewAthletesError) {
		endpoints.Logger.Debug(ctx, err4)
		c.JSON(http.StatusAccepted, AlreadyExistingAthletesResponse{Message: "No new Athletes.", AlreadyExistingAthletes: nil})
		return
	} else if err4 != nil {
		err4 = errors.Wrap(err4, "Failed to create the athletes")
		endpoints.Logger.Error(ctx, err4)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
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
