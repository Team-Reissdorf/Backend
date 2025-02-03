package performanceManagement

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

type PerformanceResponse struct {
	Message            string                  `json:"message" example:"Request successful"`
	PerformanceEntries []PerformanceBodyWithId `json:"performance_entries"`
}

// GetPerformanceEntries returns performance entries or the latest, if no 'since' parameter is given
// @Summary Returns one or more performance entries
// @Description Returns the latest performance entry or all entries until the 'since' parameter
// @Tags Performance Management
// @Produce json
// @Param AthleteId path int true "Get performance entries of the given athlete_id"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} PerformanceResponse "Request successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Athlete not found"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/get/{AthleteId} [get]
func GetPerformanceEntries(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "GetPerformanceEntries")
	defer span.End()

	// Get the athlete id from the context
	athleteIdString := c.Param("AthleteId")
	if athleteIdString == "" {
		endpoints.Logger.Debug(ctx, "Missing or invalid athlete ID")
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Missing or invalid athlete ID"})
		return
	}
	athleteId, err1 := strconv.Atoi(athleteIdString)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to parse athlete ID")
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid athlete ID"})
		return
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

}
