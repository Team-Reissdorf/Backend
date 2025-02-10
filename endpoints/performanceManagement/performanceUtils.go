package performanceManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const (
	GoldStatus   = "gold"
	SilverStatus = "silver"
	BronzeStatus = "bronze"
)

// createNewPerformances creates new performances in the database
func createNewPerformances(ctx context.Context, performanceEntries []databaseUtils.Performance) error {
	ctx, span := endpoints.Tracer.Start(ctx, "CreateNewPerformanceEntriesInDB")
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

// translatePerformanceBodies translates the performance body to a performance db entry
func translatePerformanceBodies(ctx context.Context, performanceBodies []PerformanceBody, age int, sex string) ([]databaseUtils.Performance, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "TranslatePerformanceBodies")
	defer span.End()

	performances := make([]databaseUtils.Performance, len(performanceBodies))
	for idx, performance := range performanceBodies {
		// Get the correct medal status for the performance entry
		medalStatus, err := evaluateMedalStatus(ctx, performance.ExerciseId, age, sex, performance.Points)
		if err != nil {
			return nil, err
		}

		performances[idx] = databaseUtils.Performance{
			Points:     performance.Points,
			Medal:      medalStatus,
			Date:       performance.Date,
			ExerciseId: performance.ExerciseId,
			AthleteId:  performance.AthleteId,
		}
	}

	return performances, nil
}

// getLatestPerformanceBody gets the latest performance body of an athlete
func getLatestPerformanceBody(ctx context.Context, athleteId uint) (*PerformanceBodyWithId, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetLatestPerformanceBodyFromDB")
	defer span.End()

	var performanceBody PerformanceBodyWithId
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Performance{}).
			Select("performances.id AS performance_id, points, exercises.unit AS unit, medal, date, exercise_id, athlete_id").
			Joins("LEFT JOIN exercises ON performances.exercise_id = exercises.id").
			Where("athlete_id = ?", athleteId).
			Order("date DESC").
			First(&performanceBody).
			Error

		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get the latest performance entry")
		return nil, err1
	}

	// Format the date field of the performance body
	var err2 error
	performanceBody.Date, err2 = formatHelper.FormatDate(performanceBody.Date)
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to format the date of the performance entry")
		return nil, err2
	}

	return &performanceBody, nil
}

// getPerformanceBodiesSince gets all performance entries of an athlete since the given date
func getPerformanceBodiesSince(ctx context.Context, athleteId uint, sinceDate string) (*[]PerformanceBodyWithId, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetPerformanceBodiesSinceFromDB")
	defer span.End()

	var performanceBodies []PerformanceBodyWithId
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Performance{}).
			Select("performances.id AS performance_id, points, exercises.unit AS unit, medal, date, exercise_id, athlete_id").
			Joins("LEFT JOIN exercises ON performances.exercise_id = exercises.id").
			Where("athlete_id = ? AND date >= ?", athleteId, sinceDate).
			Order("date DESC").
			Find(&performanceBodies).
			Error

		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get the performance entries since "+sinceDate)
		return nil, err1
	}

	// Format the date fields of the performance bodies
	for idx, performanceBody := range performanceBodies {
		var err2 error
		performanceBodies[idx].Date, err2 = formatHelper.FormatDate(performanceBody.Date)
		if err2 != nil {
			err2 = errors.Wrap(err2, "Failed to format the date of a performance entry")
			return nil, err2
		}
	}

	return &performanceBodies, nil
}

// getAllPerformanceBodies gets all performance bodies of an athlete
func getAllPerformanceBodies(ctx context.Context, athleteId uint) (*[]PerformanceBodyWithId, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetAllPerformanceBodiesFromDB")
	defer span.End()

	var performanceBodies []PerformanceBodyWithId
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Performance{}).
			Select("performances.id AS performance_id, points, exercises.unit AS unit, medal, date, exercise_id, athlete_id").
			Joins("LEFT JOIN exercises ON performances.exercise_id = exercises.id").
			Where("athlete_id = ?", athleteId).
			Order("date DESC").
			Find(&performanceBodies).
			Error

		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get all performance bodies")
		return nil, err1
	}

	// Format the date fields of the performance bodies
	for idx, performanceBody := range performanceBodies {
		var err2 error
		performanceBodies[idx].Date, err2 = formatHelper.FormatDate(performanceBody.Date)
		if err2 != nil {
			err2 = errors.Wrap(err2, "Failed to format the date of a performance entry")
			return nil, err2
		}
	}

	return &performanceBodies, nil
}

func editPerformanceEntry(ctx context.Context, performanceBody PerformanceBodyEdit) error {
	ctx, span := endpoints.Tracer.Start(ctx, "EditPerformanceEntryInDB")
	defer span.End()

	// ToDo
}

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
