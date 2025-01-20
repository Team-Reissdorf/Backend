package athleteManagement

import (
	"context"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

type AthleteBody struct {
	FirstName string `json:"firstName" example:"Bob"`
	LastName  string `json:"lastName" example:"Alice"`
	Email     string `json:"email" example:"bob.alice@example.com"`
	BirthDate string `json:"birthDate" example:"DD.MM.YYYY"`
	Sex       string `json:"sex" example:"<m|w|d>"`
}

// CreateAthlete creates a new athlete profile
// @Summary Creates a new athlete profile
// @Description Creates a new athlete profile with the given data. Duplicate email and birthdate combinations are not allowed.
// @Tags Athlete Management
// @Accept json
// @Produce json
// @Param Athlete body AthleteBody true "Details of an athlete to create a profile"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 201 {object} endpoints.SuccessResponse "Creation successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 409 {object} endpoints.ErrorResponse "Athlete already exists"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/athlete/create [post]
func CreateAthlete(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "CreateAthlete")
	defer span.End()

	// Bind JSON body to struct
	var body AthleteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		endpoints.Logger.Debug(ctx, err)
		c.JSON(
			http.StatusBadRequest,
			endpoints.ErrorResponse{
				Error: "Invalid request body",
			},
		)
		c.Abort()
		return
	}

	// Get the user id from the context
	// userId := authHelper.GetUserIdFromContext(ctx, c)
	// ToDo: Verify that the user is a trainer

	// Check formats
	email := body.Email
	if err := formatHelper.IsEmail(email); err != nil {
		err = errors.Wrap(err, "Invalid email address")
		endpoints.Logger.Debug(ctx, err)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid email address"})
		c.Abort()
		return
	}

	birthDate := body.BirthDate
	if err := formatHelper.IsDate(birthDate); err != nil {
		err = errors.Wrap(err, "Invalid date")
		endpoints.Logger.Debug(ctx, err)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid birth date"})
		c.Abort()
		return
	}

	sex := strings.ToLower(string(body.Sex[0]))

	// Create the athlete
	athletes := make([]databaseModels.Athlete, 1)
	athletes[0] = databaseModels.Athlete{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     email,
		BirthDate: birthDate,
		Sex:       sex,
	}
	err1, alreadyExistingAthletes := createNewAthletes(ctx, athletes)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to create the athlete")
		endpoints.Logger.Error(ctx, err1)
		c.JSON(
			http.StatusInternalServerError,
			endpoints.ErrorResponse{
				Error: "Internal server error",
			},
		)
		c.Abort()
		return
	}

	// Check if the athlete already exists
	if len(alreadyExistingAthletes) > 0 {
		err := errors.New("Athlete already exists")
		endpoints.Logger.Debug(ctx, err)
		c.JSON(
			http.StatusConflict,
			endpoints.ErrorResponse{
				Error: err.Error(),
			},
		)
		c.Abort()
		return
	}

	c.JSON(
		http.StatusCreated,
		endpoints.SuccessResponse{
			Message: "Creation successful",
		},
	)
}

// createNewAthletes creates new athletes in the database and returns the athletes that already exist
func createNewAthletes(ctx context.Context, athletes []databaseModels.Athlete) (error, []databaseModels.Athlete) {
	ctx, span := endpoints.Tracer.Start(ctx, "CreateNewAthletes")
	defer span.End()

	// Check if an athlete already exists in the database
	var alreadyExistingAthletes []databaseModels.Athlete
	var newAthletes []databaseModels.Athlete
	for _, athlete := range athletes {
		// ToDo: Check if the athlete already exists
		if true {
			alreadyExistingAthletes = append(alreadyExistingAthletes, athlete)
		} else {
			newAthletes = append(newAthletes, athlete)
		}
	}

	// ToDo: Write the new athletes to the database

	return nil, alreadyExistingAthletes
}
