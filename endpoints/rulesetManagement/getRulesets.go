package rulesetManagement

import (
	"fmt"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type RulesetResponse struct {
	Message      string                       `json:"message" example:"Request successful"`
	RulesetGoals []databaseUtils.ExerciseGoal `json:"rulesets"`
}

// GetRulesets returns the rulesets of a specific exercise with the given exercise_id and year.
// @Summary Returns the rulesets
// @Description Retrieves the rulesets of a specific exercise with the given exercise_id and year.
// @Tags Ruleset Management
// @Produce json
// @Param year query uint true "Year of the ruleset"
// @Param exercise_id query uint true "ID of the exercise"
// @Success 200 {object} RulesetResponse "Request successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid query parameters"
// @Failure 404 {object} endpoints.ErrorResponse "Rulesets not found"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/ruleset/get [get]
func GetRulesets(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "GetRulesets")
	defer span.End()

	// Get the year query parameter
	yearString := c.Query("year")
	if yearString == "" {
		endpoints.Logger.Debug(ctx, "Missing or invalid year")
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Missing or invalid year"})
		return
	}

	year, err := strconv.ParseUint(yearString, 10, 32)
	if err != nil {
		err = errors.Wrap(err, "Failed to parse 'year' query parameter")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid 'year' query parameter"})
		return
	}

	// Get the exercise_id query parameter
	exerciseIdString := c.Query("exercise_id")
	if exerciseIdString == "" {
		endpoints.Logger.Debug(ctx, "Missing or invalid exercise_id")
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Missing or invalid exercise_id"})
		return
	}
	exerciseId, err := strconv.ParseUint(exerciseIdString, 10, 32)
	if err != nil {
		err = errors.Wrap(err, "Failed to parse 'exercise_id' query parameter")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid 'exercise_id' query parameter"})
		return
	}

	// Query the database for the exercise goals
	var rulesets []databaseUtils.ExerciseGoal
	err = DatabaseFlow.GetDB(ctx).
		Model(&databaseUtils.ExerciseGoal{}).
		Joins("JOIN exercise_rulesets ON exercise_goals.ruleset_id = exercise_rulesets.id").
		Joins("JOIN exercises ON exercise_rulesets.exercise_id = exercises.id").
		Where("exercise_rulesets.ruleset_year = ? AND exercise_rulesets.exercise_id = ?", strconv.Itoa(int(year)), strconv.Itoa(int(exerciseId))).
		Find(&rulesets).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) || len(rulesets) == 0 {
		endpoints.Logger.Debug(ctx, "Rulesets not found")
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: fmt.Sprintf("Ruleset not found for year: %d and exercise Id: %d", year, exerciseId)})
		return
	} else if err != nil {
		err = errors.Wrap(err, "Failed to retrieve rulesets")
		endpoints.Logger.Error(ctx, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to retrieve rulesets"})
		return
	}

	c.JSON(
		http.StatusOK,
		RulesetResponse{
			Message:      "Request successful",
			RulesetGoals: rulesets,
		},
	)
}
