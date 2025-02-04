package performanceManagement

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type PerformanceResponse struct {
	Message          string                `json:"message" example:"Request successful"`
	PerformanceEntry PerformanceBodyWithId `json:"performance_entry"`
}

// GetLatestPerformanceEntry returns the latest performance entry
// @Summary Returns the latest performance entry
// @Description Returns the latest performance entry from the database
// @Tags Performance Management
// @Produce json
// @Param AthleteId path int true "Get the latest performance entry of the given athlete_id"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} PerformanceResponse "Request successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Athlete or performance entry not found"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/get-latest/{AthleteId} [get]
func GetLatestPerformanceEntry(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "GetLatestPerformanceEntry")
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
	// Check if a performance entry could be found
	if errors.Is(err3, gorm.ErrRecordNotFound) {
		err := errors.New("No performance entry exists for this athlete")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: err.Error()})
		return
	} else if err3 != nil {
		endpoints.Logger.Error(ctx, err3)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the latest performance entry"})
		return
	}

	// Translate performance entries to response type
	performanceBody, err4 := translatePerformanceToResponse(ctx, *performanceEntry)
	if err4 != nil {
		err4 = errors.Wrap(err4, "Failed to translate the performance")
		endpoints.Logger.Error(ctx, err4)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		return
	}

	c.JSON(
		http.StatusOK,
		PerformanceResponse{
			Message:          "Request successful",
			PerformanceEntry: *performanceBody,
		},
	)
}
