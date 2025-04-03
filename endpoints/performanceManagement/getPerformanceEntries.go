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
// @Failure 400 {object} endpoints.ErrorResponse "Date parameter is before the since parameter"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Athlete does not exist"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/get/{AthleteId} [get]
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
		//Check if the date is in the correct format
		err2 := formatHelper.IsDate(since)
		if err2 != nil {
			err2 = errors.Wrap(err2, "Invalid 'since' query parameter")
			endpoints.Logger.Debug(ctx, err2)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid 'since' query parameter"})
			return
		}
		//Check if date is in the past
		err3 := formatHelper.IsFuture(since)
		if err3 != nil {
			err3 = errors.Wrap(err3, "Date is in the future")
			endpoints.Logger.Debug(ctx, err3)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Date is in the Future"})
			return
		}
	}

	// Get the date query parameter from the context
	date := c.Query("date")
	dateIsSet := date != ""
	if dateIsSet {
		//Check if the date is in the correct format
		err2 := formatHelper.IsDate(date)
		if err2 != nil {
			err2 = errors.Wrap(err2, "Invalid 'date' query parameter")
			endpoints.Logger.Debug(ctx, err2)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid 'date' query parameter"})
			return
		}
		//Check if date is in the past
		err3 := formatHelper.IsFuture(date)
		if err3 != nil {
			err3 = errors.Wrap(err3, "Date is in the future")
			endpoints.Logger.Debug(ctx, err3)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Date is in the future"})
			return
		}
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Check if the athlete exists for the given trainer
	exists, err4 := athleteManagement.AthleteExistsForTrainer(ctx, uint(athleteId), trainerEmail)
	if err4 != nil {
		endpoints.Logger.Error(ctx, err4)
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
	if since != "" && date != "" {
		if err := formatHelper.IsBefore(date, since); err != nil {
			endpoints.Logger.Debug(ctx, "Date parameter is before the since parameter")
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Date parameter is before the since parameter"})
			return
		}
	}
	if dateIsSet {
		// Get all performance bodies date the specified date from the database
		performanceBodiesDate, err := getPerformanceBodiesDate(ctx, uint(athleteId), date)
		if err != nil {
			endpoints.Logger.Error(ctx, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get all performance bodies from " + date})
			return
		}
		if performanceBodiesDate != nil {
			performanceBodies = *performanceBodiesDate
		} else {
			performanceBodies = []PerformanceBodyWithId{}
		}
	} else if sinceIsSet {
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
