package performanceManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strconv"
	"time"
)

const (
	GoldStatus   = "gold"
	SilverStatus = "silver"
	BronzeStatus = "bronze"
)

// Checks if an exercise goal exists for a given athlete's age and exercise in a specific year
func exerciseGoalExistsForAge(ctx context.Context, exerciseId uint, age int, year string) (bool, error) {
	// Call Singleton-Database
	db := DatabaseFlow.GetDB(ctx)

	var count int64
	err := db.Model(&databaseUtils.ExerciseGoal{}).
		Joins("JOIN exercise_rulesets ON exercise_goals.ruleset_id = exercise_rulesets.id").
		Where("exercise_rulesets.exercise_id = ? AND exercise_rulesets.ruleset_year = ? AND exercise_goals.from_age <= ? AND exercise_goals.to_age >= ?", exerciseId, year, age, age).
		Count(&count).
		Error
	if err != nil {
		return false, errors.Wrap(err, "Failed to check if exercise goal exists for age")
	}
	return count > 0, nil
}

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
		medalStatus, err := evaluateMedalStatus(ctx, performance.ExerciseId, performance.Date, age, sex, performance.Points)
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

// getPerformanceBodiesDate gets all performance entries of an athlete from the given date
func getPerformanceBodiesDate(ctx context.Context, athleteId uint, date string) (*[]PerformanceBodyWithId, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetPerformanceBodiesSinceFromDB")
	defer span.End()

	var performanceBodies []PerformanceBodyWithId
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Performance{}).
			Select("performances.id AS performance_id, points, exercises.unit AS unit, medal, date, exercise_id, athlete_id").
			Joins("LEFT JOIN exercises ON performances.exercise_id = exercises.id").
			Where("athlete_id = ? AND date = ?", athleteId, date).
			Order("date DESC").
			Find(&performanceBodies).
			Error

		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get the performance entries from "+date)
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

// performanceExistsForTrainer checks if a performance entry with the given id exists for the given trainer
func performanceExistsForTrainer(ctx context.Context, performanceId uint, trainerEmail string) (bool, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "PerformanceExistsForTrainer")
	defer span.End()

	var performanceCount int64
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Performance{}).
			Joins("INNER JOIN athletes ON performances.athlete_id = athletes.id").
			Where("performances.id = ? AND athletes.trainer_email = ?", performanceId, trainerEmail).
			Count(&performanceCount).
			Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to check if the performance entry exists")
		return false, err1
	}

	return performanceCount > 0, nil
}

// updatePerformanceEntry updates the given performance entry in the database
func updatePerformanceEntry(ctx context.Context, performanceEntry databaseUtils.Performance) error {
	ctx, span := endpoints.Tracer.Start(ctx, "EditPerformanceEntryInDB")
	defer span.End()

	// Calculate the age of the athlete
	birthDate, err3 := formatHelper.FormatDate(performanceEntry.Athlete.BirthDate)
	if err3 != nil {
		endpoints.Logger.Debug(ctx, "Date Formatter not working..."+err3.Error())
		err3 = errors.Wrap(err3, "Failed to parse the birth date")
		return err3
	}

	age, err := athleteManagement.CalculateAge(ctx, birthDate)
	if err != nil {
		endpoints.Logger.Debug(ctx, "Age Calculator not working...")
		err = errors.Wrap(err, "Failed to calculate the age of the athlete")
		return err
	}

	// Extract the date from the performance entry
	performanceDate, err := time.Parse("2006-01-02", performanceEntry.Date)
	if err != nil {
		err = errors.Wrap(err, "Failed to parse the date")
		return err
	}

	performanceYear := strconv.Itoa(performanceDate.Year())
	// Check if the exercise goal exists for the athlete's age
	exists, err := exerciseGoalExistsForAge(ctx, performanceEntry.ExerciseId, age, performanceYear)
	if err != nil {
		return errors.Wrap(err, "Failed to check exercise goal")
	}
	if !exists {
		return errors.New("Failed to find the exercise goal for the age of the athlete. Possible reasons: no matching goal exists, data inconsistency, or invalid Input.")
	}

	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(databaseUtils.Performance{}).Where("id = ?", performanceEntry.ID).Updates(performanceEntry).Error
		return err
	})
	err1 = databaseUtils.TranslatePostgresError(err1)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to update the performance entry")
	}
	return err1
}

// countPerformanceEntriesPerDisciplinePerDay counts all performance entries per discipline per day of the given athlete
func countPerformanceEntriesPerDisciplinePerDay(ctx context.Context, athleteId uint, exerciseId uint, date string) (int64, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "CountPerformanceEntriesPerDisciplinePerDayInDB")
	defer span.End()

	var count int64
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Performance{}).
			Joins("LEFT JOIN exercises ON performances.exercise_id = exercises.id").
			Where("athlete_id = ? AND exercises.id = ? AND date = ?", athleteId, exerciseId, date).
			Count(&count).
			Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to count the already existing performance entries")
	}
	return count, err1
}

// countPerformanceEntriesPerDisciplinePerDayEditMode counts all performance entries per discipline (besides its old one) per day of the given athlete
func countPerformanceEntriesPerDisciplinePerDayEditMode(ctx context.Context, athleteId uint, exerciseId uint, performanceId uint, date string) (int64, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "CountPerformanceEntriesPerDisciplinePerDayEditModeInDB")
	defer span.End()

	var count int64
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		var disciplineName string
		errA := tx.Model(&databaseUtils.Performance{}).
			Joins("LEFT JOIN exercises ON performances.exercise_id = exercises.id").
			Select("discipline_name").
			Where("performances.id = ?", performanceId).
			Find(&disciplineName).
			Error
		if errA != nil {
			errA = errors.Wrap(errA, "Failed to get the discipline_name")
			return errA
		}

		errB := tx.Model(&databaseUtils.Performance{}).
			Joins("LEFT JOIN exercises ON performances.exercise_id = exercises.id").
			Where("athlete_id = ? AND exercises.id = ? AND date = ? AND exercises.discipline_name != ?", athleteId, exerciseId, date, disciplineName).
			Count(&count).
			Error
		if errB != nil {
			errB = errors.Wrap(errB, "Failed to count the already existing performance entries")
			return errB
		}
		return nil
	})
	return count, err1
}
