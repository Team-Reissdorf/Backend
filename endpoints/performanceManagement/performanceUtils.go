package performanceManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
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
		err := tx.Model(&databaseUtils.Exercise{}).Where("id = ?", exerciseId).Count(&exerciseCount).Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to check if the exercise exists")
		return false, err1
	}

	return exerciseCount > 0, nil
}

// CreateNewPerformances creates new performances in the database
func CreateNewPerformances(ctx context.Context, performanceEntries []databaseUtils.Performance) error {
	ctx, span := endpoints.Tracer.Start(ctx, "createNewPerformances")
	defer span.End()

	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Create(&performanceEntries).Error
		return err
	})
	err1 = databaseUtils.TranslatePostgresError(err1)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to write the performance entry to the database")
		return err1
	}

	return nil
}

// TranslatePerformanceBody translates the performance body to a performance db entry
func TranslatePerformanceBody(ctx context.Context, performanceBodies []PerformanceBody) []databaseUtils.Performance {
	ctx, span := endpoints.Tracer.Start(ctx, "TranslatePerformanceBody")
	defer span.End()

	performances := make([]databaseUtils.Performance, len(performanceBodies))
	for idx, performance := range performanceBodies {
		performances[idx] = databaseUtils.Performance{
			Points:     performance.Points,
			Date:       performance.Date,
			ExerciseId: performance.ExerciseId,
			AthleteId:  performance.AthleteId,
		}
	}

	return performances
}
