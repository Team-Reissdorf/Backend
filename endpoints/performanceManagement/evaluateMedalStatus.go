package performanceManagement

import (
	"context"

	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/pkg/errors"
)

// evaluateMedalStatus checks which result a performance entry achieved
func evaluateMedalStatus(ctx context.Context, exerciseId uint, performanceDateString string, age int, sex string, points uint64) (string, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "EvaluateMedalStatus")
	defer span.End()

	if sex == "d" {
		sex = "m"
	}

	performanceYear, err1 := getPerformanceYear(ctx, performanceDateString)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Error parsing performance year")
		return "", err1
	}

	// Get the exercise goal to check whether the athlete has reached a medal or not, and if so, which one
	exerciseGoal, err2 := getExerciseGoal(ctx, exerciseId, performanceYear, age, sex)
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to get the exercise goal")
		return "", err2
	}

	// Get the medal status
	medalStatus := getMedalStatus(ctx, exerciseGoal, points)

	return medalStatus, nil
}
