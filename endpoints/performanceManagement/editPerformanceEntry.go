package performanceManagement

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

// EditPerformanceEntry edits a performance entry
// @Summary Edits an existing performance entry
// @Description Edits an existing performance entry with the given details.
// @Tags Performance Management
// @Produce json
// @Param Performance body PerformanceBodyEdit true "Edited details of a performance entry"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} endpoints.SuccessResponse "Edited successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Performance entry not found"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/edit [put]
func EditPerformanceEntry(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "EditPerformanceEntry")
	defer span.End()

	// Bind JSON body to struct
	var body PerformanceBodyEdit
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Check if the given performance entry is for an athlete of the given trainer
	// ToDo

	// Translate to database entry
	performanceEntry := databaseUtils.Performance{
		ID:         body.PerformanceId,
		Points:     body.Points,
		Date:       body.Date,
		ExerciseId: body.ExerciseId,
	}

}
