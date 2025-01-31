package performanceManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// CheckIfExerciseExists checks if an exercise with the given id exists
func CheckIfExerciseExists(ctx context.Context, exerciseId uint) (bool, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "ExerciseExistsCheck")
	defer span.End()

	var exerciseCount int64
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseModels.Athlete{}).Where("exercise_id = ?", exerciseId).Count(&exerciseCount).Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to check if the exercise exists")
		return false, err1
	}

	return exerciseCount > 0, nil
}

func createNewPerformance(ctx context.Context, body []PerformanceBody) error {
	return nil
}
