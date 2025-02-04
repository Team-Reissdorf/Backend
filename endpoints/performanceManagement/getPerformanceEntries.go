package performanceManagement

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/Team-Reissdorf/Backend/formatHelper"
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
// @Param since query string false "Date in YYYY-MM-DD format to get all entries since then"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} PerformanceResponse "Request successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Athlete of performance entry not found"
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
	athleteId, err1 := strconv.ParseUint(athleteIdString, 10, 32)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to parse athlete ID")
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid athlete ID"})
		return
	}

	// Get the since query parameter from the context
	since := c.Query("since")
	if since != "" {
		err := formatHelper.IsDate(since)
		if err != nil {
			err = errors.Wrap(err, "Invalid 'since' query parameter")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid 'since' query parameter"})
			return
		}
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Check if the athlete exists for the given trainer
	exists, err2 := athleteManagement.AthleteExistsForTrainer(ctx, uint(athleteId), trainerEmail)
	if err2 != nil {
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check if the athlete exists"})
		return
	}
	if !exists {
		endpoints.Logger.Debug(ctx, "Athlete does not exist")
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete not found"})
		return
	}

	// Get the latest performance entry from the database
	performanceEntry, err3 := getLatestPerformanceEntry(ctx, uint(athleteId))
	if err3 != nil {
		endpoints.Logger.Error(ctx, err3)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the latest performance entry"})
		return
	}
	performanceEntries := make([]databaseUtils.Performance, 1)
	performanceEntries[0] = *performanceEntry

	// Translate performance to response type
	performanceBody, err4 := translatePerformanceToResponse(ctx, *performanceEntry)
	if err4 != nil {
		err4 = errors.Wrap(err4, "Failed to translate the performance entry")
		endpoints.Logger.Error(ctx, err4)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		return
	}
	performanceBodies := make([]PerformanceBodyWithId, 1)
	performanceBodies[0] = *performanceBody

	c.JSON(
		http.StatusOK,
		PerformanceResponse{
			Message:            "Request successful",
			PerformanceEntries: performanceBodies,
		},
	)
}
