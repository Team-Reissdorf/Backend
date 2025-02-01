package performanceManagement

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

// CreatePerformance creates a new performance entry
// @Summary Creates a new performance entry
// @Description Creates a new performance entry with the given data. Maximum 3 entries/discipline/athlete/day are allowed (performance limit).
// @Tags Performance Management
// @Accept json
// @Produce json
// @Param Performance body PerformanceBody true "Details of a performance"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 201 {object} endpoints.SuccessResponse "Creation successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Athlete or exercise does not exist"
// @Failure 409 {object} endpoints.ErrorResponse "Performance limit reached"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/create [post]
func CreatePerformance(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "CreatePerformance")
	defer span.End()

	// Bind JSON body to struct
	var body PerformanceBody
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Validate the date format
	err1 := formatHelper.IsDate(body.Date)
	if err1 != nil {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid date format"})
		return
	}

	// Check if the athlete exists for this trainer
	athleteExists, err2 := athleteManagement.AthleteExistsForTrainer(ctx, body.AthleteId, trainerEmail)
	if err2 != nil {
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check if the athlete exists"})
		return
	}
	if !athleteExists {
		endpoints.Logger.Debug(ctx, "Athlete does not exist")
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete does not exist"})
		return
	}

	// Check if the given exercise exists
	exerciseExists, err3 := CheckIfExerciseExists(ctx, body.ExerciseId)
	if err3 != nil {
		endpoints.Logger.Debug(ctx, err3)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check if the exercise exists"})
		return
	}
	if !exerciseExists {
		endpoints.Logger.Debug(ctx, "Exercise does not exist")
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Exercise does not exist"})
		return
	}

	// Create performance entry in the database
	performanceBodies := make([]PerformanceBody, 1)
	performanceBodies[0] = body
	err4 := createNewPerformances(ctx, translatePerformanceBody(ctx, performanceBodies))
	if err4 != nil {
		err4 = errors.Wrap(err4, "Failed to create the performance entry")
		endpoints.Logger.Error(ctx, err4)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to create the performance entry"})
		return
	}

	c.JSON(
		http.StatusCreated,
		endpoints.SuccessResponse{
			Message: "Creation successful",
		},
	)
}
