package performanceManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strconv"
	"time"
)

// evaluateMedalStatus checks which result a performance entry achieved
func evaluateMedalStatus(ctx context.Context, exerciseId uint, performanceDateString string, age int, sex string, points uint64) (string, error) {
	// Get the exercise goal to check whether the athlete has reached a medal or not, and if so, which one
	var exerciseGoal databaseUtils.ExerciseGoal
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		// Parse the performance date
		t, err1 := time.Parse(time.DateOnly, performanceDateString)
		if err1 != nil {
			err1 = errors.Wrap(err1, "Failed to parse date: "+performanceDateString)
			endpoints.Logger.Debug(ctx, err1)
			return err1
		}
		performanceYear := t.Year()

		// Get the exercise goal
		err := tx.Model(&databaseUtils.ExerciseGoal{}).
			Joins("JOIN exercise_rulesets ON exercise_rulesets.id = exercise_goals.ruleset_id").
			Where("exercise_rulesets.ruleset_year = ? AND exercise_rulesets.exercise_id = ? AND "+
				"exercise_goals.from_age <= ? AND exercise_goals.to_age >= ? AND exercise_goals.sex = ?",
				strconv.Itoa(performanceYear), exerciseId, age, age, sex).
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
