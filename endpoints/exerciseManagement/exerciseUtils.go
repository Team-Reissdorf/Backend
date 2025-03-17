package exerciseManagement

import (
	"context"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
)

// translateExerciseToResponse converts an exercise database object to response type
func translateExerciseToResponse(ctx context.Context, exercise databaseUtils.Exercise) (*ExerciseBodyWithId, error) {
	_, span := endpoints.Tracer.Start(ctx, "TranslateExerciseToResponse")
	defer span.End()

	exerciseResponse := ExerciseBodyWithId{
		ExerciseId:     exercise.ID,
		Name:           exercise.Name,
		Unit:           exercise.Unit,
		DisciplineName: exercise.DisciplineName,
	}

	return &exerciseResponse, nil
}
