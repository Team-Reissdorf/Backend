package endpoints

import (
	"context"
	"github.com/Team-Reissdorf/Backend/database_models"
	"github.com/Team-Reissdorf/Backend/endpoints/standardJsonAnswers"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
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
// @Success 200 {object} standardJsonAnswers.SuccessResponse "Creation successful"
// @Failure 400 {object} standardJsonAnswers.ErrorResponse "Invalid request body"
// @Failure 401 {object} standardJsonAnswers.ErrorResponse "The token is invalid"
// @Failure 409 {object} standardJsonAnswers.ErrorResponse "Athlete already exists"
// @Failure 500 {object} standardJsonAnswers.ErrorResponse "Internal server error"
// @Router /v1/athlete/create [post]
func CreateAthlete(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "CreateAthlete")
	defer span.End()

	// Bind JSON body to struct
	var body AthleteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		logger.Debug(ctx, err)
		c.JSON(
			http.StatusBadRequest,
			standardJsonAnswers.ErrorResponse{
				Error: "Invalid request body",
			},
		)
		c.Abort()
		return
	}

	// Get the user id from the context
	// userId := authHelper.GetUserIdFromContext(ctx, c)
	// ToDo: Verify that the user has the coach role

	// Create the athlete
	athletes := make([]database_models.Athlete, 1)
	athletes[0] = database_models.Athlete{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
		BirthDate: body.BirthDate,
		Sex:       body.Sex,
	}
	err1, alreadyExistingAthletes := createNewAthletes(ctx, athletes)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to create the athlete")
		logger.Error(ctx, err1)
		c.JSON(
			http.StatusInternalServerError,
			standardJsonAnswers.ErrorResponse{
				Error: "Internal server error",
			},
		)
		c.Abort()
		return
	}

	// Check if the athlete already exists
	if len(alreadyExistingAthletes) > 0 {
		err := errors.New("Athlete already exists")
		logger.Debug(ctx, err)
		c.JSON(
			http.StatusConflict,
			standardJsonAnswers.ErrorResponse{
				Error: err.Error(),
			},
		)
		c.Abort()
		return
	}

	c.JSON(
		http.StatusOK,
		standardJsonAnswers.SuccessResponse{
			Message: "Creation successful",
		},
	)
}

// createNewAthletes creates new athletes in the database and returns the athletes that already exist
func createNewAthletes(ctx context.Context, athletes []database_models.Athlete) (error, []database_models.Athlete) {
	ctx, span := tracer.Start(ctx, "CreateNewAthletes")
	defer span.End()

	// Check if an athlete already exists in the database
	var alreadyExistingAthletes []database_models.Athlete
	var newAthletes []database_models.Athlete
	for _, athlete := range athletes {
		// ToDo: Check if the email and birthdate combination already exists
		if true {
			alreadyExistingAthletes = append(alreadyExistingAthletes, athlete)
		} else {
			newAthletes = append(newAthletes, athlete)
		}
	}

	// ToDo: Write the new athletes to the database

	return nil, alreadyExistingAthletes
}
