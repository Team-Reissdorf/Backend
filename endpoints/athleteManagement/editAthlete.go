package athleteManagement

import (
	"fmt"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
// @Success 200 {object} endpoints.SuccessResponse "Edited successful"
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
	athleteEntry := databaseUtils.Athlete{
		ID:           body.AthleteId,
		FirstName:    body.FirstName,
		LastName:     body.LastName,
		BirthDate:    body.BirthDate,
		Sex:          body.Sex,
		Email:        body.Email,
		TrainerEmail: trainerEmail,
	}

	// Validate the athlete body
	err1 := validateAthlete(ctx, &athleteEntry)
	if errors.Is(err1, formatHelper.EmptyStringError) {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err1.Error()})
		return
	} else if errors.Is(err1, formatHelper.InvalidSexLengthError) || errors.Is(err1, formatHelper.InvalidSexValue) {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Sex needs to be <m|f|d>"})
		return
	} else if errors.Is(err1, formatHelper.DateFormatInvalidError) {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid date format"})
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

	// Check if the user exists and is assigned to the given trainer
	exists, err2 := AthleteExistsForTrainer(ctx, athleteEntry.ID, trainerEmail)
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to check if the athlete exists and is assigned to the trainer")
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check if the athlete exists"})
		return
	}
	if !exists {
		endpoints.Logger.Debug(ctx, fmt.Sprintf("Athlete with id %d does not exist", body.AthleteId))
		c.AbortWithStatusJSON(http.StatusNotFound, "Athlete does not exist")
		return
	} else {
		endpoints.Logger.Debug(ctx, fmt.Sprintf("Athlete with id %d exists and is assigned to the given trainer", body.AthleteId))
	}

	// Update the athlete in the database
	err3 := updateAthlete(ctx, athleteEntry)
	if errors.Is(err3, databaseUtils.ErrForeignKeyViolation) {
		err3 = errors.Wrap(err3, "Another athlete with the same personal information already exists")
		endpoints.Logger.Debug(ctx, err3)
		c.AbortWithStatusJSON(http.StatusConflict, endpoints.ErrorResponse{Error: "Another athlete with the same personal information already exists"})
		return
	} else if err3 != nil {
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
