package performanceManagement

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

type PerformancesResponse struct {
	Message            string                  `json:"message" example:"Request successful"`
	PerformanceEntries []PerformanceBodyWithId `json:"performance_entries"`
}

// GetPerformanceEntries returns all performance entries
// @Summary Returns all performance entries of the given athlete
// @Description Returns all performance entries of the given athlete and can be filtered using the 'since' query parameter
// @Tags Performance Management
// @Produce json
// @Param AthleteId path int true "Get all performance entries of the given athlete"
// @Param since query string false "Date in YYYY-MM-DD format to get all entries since then (including the entries from that day)"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} PerformancesResponse "Request successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Athlete does not exist"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/get-all/{AthleteId} [get]
func GetPerformanceEntries(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "GetAllPerformanceEntries")
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
	sinceIsSet := since != ""
	if sinceIsSet {
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
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete does not exist"})
		return
	}

	// Get the performance body/bodies
	var performanceBodies []PerformanceBodyWithId
	if sinceIsSet {
		// Get all performance bodies since the specified date from the database
		performanceBodiesSince, err := getPerformanceBodiesSince(ctx, uint(athleteId), since)
		if err != nil {
			endpoints.Logger.Error(ctx, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get all performance bodies since " + since})
			return
		}
		if performanceBodiesSince != nil {
			performanceBodies = *performanceBodiesSince
		} else {
			performanceBodies = []PerformanceBodyWithId{}
		}
	} else {
		// Get all performance bodies from the database
		allPerformanceBodies, err := getAllPerformanceBodies(ctx, uint(athleteId))
		if err != nil {
			endpoints.Logger.Error(ctx, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get all performance bodies"})
			return
		}
		if allPerformanceBodies != nil {
			performanceBodies = *allPerformanceBodies
		} else {
			performanceBodies = []PerformanceBodyWithId{}
		}
	}

	c.JSON(
		http.StatusOK,
		PerformancesResponse{
			Message:            "Request successful",
			PerformanceEntries: performanceBodies,
		},
	)
}
