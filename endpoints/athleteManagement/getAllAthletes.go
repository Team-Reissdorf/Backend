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

	c.JSON(
		http.StatusOK,
		endpoints.SuccessResponse{
			Message: "Creation successful",
		},
	)
}
