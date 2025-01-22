package athleteManagement

import (
	"github.com/Team-Reissdorf/Backend/endpoints"
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
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid request body"})
		c.Abort()
		return
	}

	// Get the user id from the context
	// userId := authHelper.GetUserIdFromContext(ctx, c)
	// ToDo: Verify that the user is a trainer
	// trainerEmail := "blabla@test.com"

	// ToDo: Check if the email-birthdate-firstname combo exists more than once if information about it has changed
}
