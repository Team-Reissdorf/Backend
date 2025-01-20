package athleteManagement

import (
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetAllAthletes returns all athletes
// @Summary Returns all athlete profiles
// @Description All athlete profiles of the given trainer are returned
// @Tags Athlete Management
// @Produce json
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} endpoints.SuccessResponse "Request successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/athlete/get-all [get]
func GetAllAthletes(c *gin.Context) {
	_, span := endpoints.Tracer.Start(c.Request.Context(), "GetAllAthletes")
	defer span.End()

	// Get the user id from the context
	// userId := authHelper.GetUserIdFromContext(ctx, c)
	// ToDo: Verify that the user is a trainer

	// ToDo: Get all athletes of the trainer

	// Create test data // ToDo: Remove
	var athletes []databaseModels.Athlete
	athletes = append(athletes,
		databaseModels.Athlete{
			AthleteId: 0,
			FirstName: "John",
			LastName:  "Lennon",
			Email:     "john@lennon.com",
			BirthDate: "1940-10-09",
			Sex:       "m",
		},
		databaseModels.Athlete{
			AthleteId: 1,
			FirstName: "Julio",
			LastName:  "Iglesias",
			Email:     "julio@iglesias.com",
			BirthDate: "1943-06-23",
			Sex:       "m",
		},
	)

	c.JSON(
		http.StatusOK,
		endpoints.SuccessResponse{
			Message: "Creation successful",
		},
	)
}
