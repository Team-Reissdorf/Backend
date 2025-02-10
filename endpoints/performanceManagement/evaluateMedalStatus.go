package performanceManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// evaluateMedalStatus checks which result a performance entry achieved
func evaluateMedalStatus(ctx context.Context, exerciseId uint, age int, sex string, points uint64) (string, error) {
	// Get the exercise goal to check whether the athlete has reached a medal or not, and if so, which one
	var exerciseGoal databaseUtils.ExerciseGoal
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.ExerciseGoal{}).
			Where("exercise_id = ? AND from_age <= ? AND to_age >= ? AND sex = ?", exerciseId, age, age, sex).
			First(&exerciseGoal).
			Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to evaluate the medal status")
		return "", err1
	}

	// Check if a smaller or a bigger value is better
	smallerIsBetter := exerciseGoal.Bronze > exerciseGoal.Gold

	// Create the compare function based on the smallerIsBetter variable
	compare := func(p, g uint64) bool {
		if smallerIsBetter {
			return p <= g
		} else {
			return p >= g
		}
	}

	// Check the medal status of the athletes performance entry
	switch {
	case compare(points, exerciseGoal.Gold):
		return GoldStatus, nil
	case compare(points, exerciseGoal.Silver):
		return SilverStatus, nil
	case compare(points, exerciseGoal.Bronze):
		return BronzeStatus, nil
	default:
		return "", nil
	}
}
