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
func translatePerformanceBodies(ctx context.Context, performanceBodies []PerformanceBody) []databaseUtils.Performance {
	ctx, span := endpoints.Tracer.Start(ctx, "TranslatePerformanceBodies")
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

// translatePerformanceToResponse converts a performance database object to response type
func translatePerformanceToResponse(ctx context.Context, performance databaseUtils.Performance) (*PerformanceBodyWithId, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "TranslatePerformanceToResponse")
	defer span.End()

	// Reformat the date to the correct format
	date, err := formatHelper.FormatDate(performance.Date)
	if err != nil {
		return nil, err
	}

	performanceResponse := PerformanceBodyWithId{
		PerformanceId: performance.ID,
		Points:        performance.Points,
		Date:          date,
		ExerciseId:    performance.ExerciseId,
		AthleteId:     performance.AthleteId,
	}

	return &performanceResponse, nil
}

// getLatestPerformanceEntry gets the latest performance entry of an athlete
func getLatestPerformanceEntry(ctx context.Context, athleteId uint) (*databaseUtils.Performance, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetLatestPerformanceEntryFromDB")
	defer span.End()

	var performanceEntry databaseUtils.Performance
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Performance{}).Where("athlete_id = ?", athleteId).Order("date DESC").First(&performanceEntry).Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get the latest performance entry")
		return nil, err1
	}

	return &performanceEntry, nil
}

// getLatestPerformanceEntriesSince gets all performance entries of an athlete since the given date
func getPerformanceEntriesSince(ctx context.Context, athleteId uint, sinceDate string) (*[]databaseUtils.Performance, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetPerformanceEntriesSinceFromDB")
	defer span.End()

	var performanceEntries []databaseUtils.Performance
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Performance{}).Where("athlete_id = ? AND date >= ?", athleteId, sinceDate).Order("date DESC").Find(&performanceEntries).Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get the performance entries since "+sinceDate)
		return nil, err1
	}

	return &performanceEntries, nil
}

// getAllPerformanceEntries gets all performance entries of an athlete
func getAllPerformanceEntries(ctx context.Context, athleteId uint) (*[]databaseUtils.Performance, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetAllPerformanceEntriesFromDB")
	defer span.End()

	var performanceEntries []databaseUtils.Performance
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Performance{}).Where("athlete_id = ?", athleteId).Order("date DESC").Find(&performanceEntries).Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get all performance entries")
		return nil, err1
	}

	return &performanceEntries, nil
}
