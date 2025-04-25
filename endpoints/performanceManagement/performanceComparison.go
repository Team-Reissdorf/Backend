package performanceManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

// isSmallerBetter checks if a smaller points value is better than a bigger value
func isSmallerBetter(bronze, gold uint64) bool {
	return bronze > gold
}

// isLeftBetter compares two points based on the smallerIsBetter boolean
func isLeftBetter(points1, points2 uint64, smallerIsBetter bool) bool {
	if smallerIsBetter {
		return points1 <= points2
	} else {
		return points1 >= points2
	}
}

// getMedalStatus checks the medal status of the athletes performance entry
func getMedalStatus(ctx context.Context, exerciseGoal databaseUtils.ExerciseGoal, points uint64) string {
	ctx, span := endpoints.Tracer.Start(ctx, "GetMedalStatus")
	defer span.End()

	smallerIsBetter := isSmallerBetter(exerciseGoal.Bronze, exerciseGoal.Gold)

	switch {
	case isLeftBetter(points, exerciseGoal.Gold, smallerIsBetter):
		return GoldStatus
	case isLeftBetter(points, exerciseGoal.Silver, smallerIsBetter):
		return SilverStatus
	case isLeftBetter(points, exerciseGoal.Bronze, smallerIsBetter):
		return BronzeStatus
	default:
		return ""
	}
}

// getPerformanceYear parses the performance date and returns the year as int
func getPerformanceYear(ctx context.Context, performanceDate string) (int, error) {
	_, span := endpoints.Tracer.Start(ctx, "GetPerformanceYear")
	defer span.End()

	t, err1 := time.Parse(time.DateOnly, performanceDate)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to parse date: "+performanceDate)
		return -1, err1
	}
	performanceYear := t.Year()

	return performanceYear, nil
}

// getExerciseGoal gets the exercise goal from the database based on the given parameters
func getExerciseGoal(ctx context.Context, exerciseId uint, performanceYear int, age int, sex string) (databaseUtils.ExerciseGoal, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetExerciseGoal")
	defer span.End()

	var exerciseGoal databaseUtils.ExerciseGoal

	err := DatabaseFlow.GetDB(ctx).Model(&databaseUtils.ExerciseGoal{}).
		Joins("JOIN exercise_rulesets ON exercise_rulesets.id = exercise_goals.ruleset_id").
		Where("exercise_rulesets.ruleset_year = ? AND exercise_rulesets.exercise_id = ? AND "+
			"exercise_goals.from_age <= ? AND exercise_goals.to_age >= ? AND exercise_goals.sex = ?",
			strconv.Itoa(performanceYear), exerciseId, age, age, sex).
		First(&exerciseGoal).
		Error

	return exerciseGoal, err
}

// getBestPerformanceEntry returns the best performance entry of the given list.
// This function requires that all performance entries of the list are of the same exercise, athlete and performance year!
func getBestPerformanceEntry(ctx context.Context, performances *[]PerformanceBodyWithId) (*PerformanceBodyWithId, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetBestPerformanceEntry")
	defer span.End()

	// Check how many entries are in the list
	if len(*performances) == 1 {
		performance := (*performances)[0]
		return &performance, nil
	} else if len(*performances) < 1 {
		return nil, errors.New("No performance entries found")
	}

	// Get the athlete
	athlete, err0 := athleteManagement.GetAthleteDirectly(ctx, (*performances)[0].AthleteId)
	if err0 != nil {
		err0 = errors.Wrap(err0, "Failed to get the athlete")
		return nil, err0
	}

	// Get the athlete's age
	if len((*athlete).BirthDate) < 10 {
		return nil, errors.New("Invalid BirthDate: must be at least 10 characters long")
	}
	age, err1 := athleteManagement.CalculateAge(ctx, (*athlete).BirthDate[:10])
	if err1 != nil {
		err1 = errors.New("Failed to calculate age for best performance entry")
		return nil, err1
	}

	// Parse the performance year
	performanceYear, err2 := getPerformanceYear(ctx, (*performances)[0].Date)
	if err2 != nil {
		return nil, err2
	}

	// Get the exercise goal
	exerciseGoal, err3 := getExerciseGoal(ctx, (*performances)[0].ExerciseId, performanceYear, age, (*athlete).Sex)
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to get the exercise goal")
		return nil, err3
	}

	// Check if smaller is better
	smallerBetter := isSmallerBetter(exerciseGoal.Bronze, exerciseGoal.Gold)

	// Get the best performance entry
	bestPerformanceEntry := (*performances)[0]
	for idx, performanceEntry := range *performances {
		if idx == 0 {
			continue
		}

		// Check if this performance entry is better than the last best
		better := isLeftBetter(performanceEntry.Points, bestPerformanceEntry.Points, smallerBetter)
		if better {
			bestPerformanceEntry = performanceEntry
		}
	}

	return &bestPerformanceEntry, nil
}
