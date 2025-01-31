package athleteManagement

import (
	"fmt"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

// EditAthlete edits an existing athlete profile
// @Summary Edits an existing athlete profile
// @Description Edits an existing athlete profile with the given details. Duplicate email and birthdate combinations are not allowed.
// @Tags Athlete Management
// @Accept json
// @Produce json
// @Param Athlete body AthleteBodyWithId true "Edited details of an athlete"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} endpoints.SuccessResponse "Creation successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Athlete could not be found for this trainer"
// @Failure 409 {object} endpoints.ErrorResponse "Athlete already exists"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/athlete/edit [put]
func EditAthlete(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "EditAthlete")
	defer span.End()

	// Bind JSON body to struct
	var body AthleteBodyWithId
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Translate into a database object
	athlete := databaseModels.Athlete{
		AthleteId:    body.AthleteId,
		FirstName:    body.FirstName,
		LastName:     body.LastName,
		BirthDate:    body.BirthDate,
		Sex:          body.Sex,
		Email:        body.Email,
		TrainerEmail: trainerEmail,
	}

	// Check if the user exists and is assigned to the given trainer
	var athleteCount int64
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(databaseModels.Athlete{}).
			Where("athlete_id = ? AND trainer_email = ?", body.AthleteId, trainerEmail).Count(&athleteCount).Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to check whether the athlete exists and is assigned to the trainer")
		endpoints.Logger.Error(ctx, err1)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check if the athlete exists"})
		return
	}
	if athleteCount < 1 {
		endpoints.Logger.Debug(ctx, fmt.Sprintf("Athlete with id %d does not exist", body.AthleteId))
		c.AbortWithStatusJSON(http.StatusNotFound, "Athlete does not exist")
		return
	} else if athleteCount == 1 {
		endpoints.Logger.Debug(ctx, fmt.Sprintf("Athlete with id %d exists and is assigned to the given trainer", body.AthleteId))
	} else if athleteCount > 1 { // Should never happen if the database works correct
		endpoints.Logger.Error(ctx, fmt.Sprintf("Athlete with id %d exists %d times", body.AthleteId, athleteCount))
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Athlete exists multiple times. Please consult the database engineer!"})
		return
	}

	// Validate all values of the athlete and check if another athlete with the given unique combo exists
	exists, err2 := athleteExists(ctx, &athlete, true)
	if errors.Is(err2, formatHelper.InvalidSexLengthError) || errors.Is(err2, formatHelper.InvalidSexValue) {
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Sex needs to be <m|f|d>, but is " + athlete.Sex})
		return
	} else if errors.Is(err2, formatHelper.DateFormatInvalidError) {
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid date format"})
		return
	} else if errors.Is(err2, formatHelper.InvalidEmailAddressFormatError) || errors.Is(err2, formatHelper.EmailAddressContainsNameError) || errors.Is(err2, formatHelper.EmailAddressInvalidTldError) {
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid email address format"})
		return
	} else if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to validate the athlete")
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to validate the athlete"})
		return
	}
	if exists {
		err := errors.New("Another athlete with the same personal information already exists")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusConflict, endpoints.ErrorResponse{Error: err.Error()})
		return
	}

	err3 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(databaseModels.Athlete{}).Where("athlete_id = ?", athlete.AthleteId).Updates(athlete).Error
		return err
	})
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to update the athlete")
		endpoints.Logger.Error(ctx, err3)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to update the athlete"})
		return
	}

	c.JSON(
		http.StatusOK,
		endpoints.SuccessResponse{
			Message: "Edited successful",
		},
	)
}
