package athleteManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strings"
)

var (
	NoNewAthletesError = errors.New("No new Athletes found")
)

// validateAthlete checks if all values of an athlete are valid
// Throws: Forwards errors of the formatHelper
func validateAthlete(ctx context.Context, athlete *databaseUtils.Athlete) error {
	ctx, span := endpoints.Tracer.Start(ctx, "ValidateAthlete")
	defer span.End()

	// Capitalize the first letter of the name
	athlete.FirstName = strings.ToUpper(string(athlete.FirstName[0])) + athlete.FirstName[1:]
	athlete.LastName = strings.ToUpper(string(athlete.LastName[0])) + athlete.LastName[1:]

	athlete.Email = strings.ToLower(athlete.Email)
	if err := formatHelper.IsEmail(athlete.Email); err != nil {
		err = errors.Wrap(err, "Invalid email address")
		return err
	}

	if err := formatHelper.IsDate(athlete.BirthDate); err != nil {
		err = errors.Wrap(err, "Invalid date")
		return err
	}

	athlete.Sex = strings.ToLower(string(athlete.Sex[0]))
	if err := formatHelper.IsSex(athlete.Sex); err != nil {
		err = errors.Wrap(err, "Invalid sex")
		return err
	}

	return nil
}

// createNewAthletes creates new athletes in the database and returns the athletes that already exist.
// Note: If the error is not nil, the returned list is invalid.
func createNewAthletes(ctx context.Context, athletes []databaseUtils.Athlete) (error, []databaseUtils.Athlete) {
	ctx, span := endpoints.Tracer.Start(ctx, "CreateNewAthletes")
	defer span.End()

	// Check if an athlete already exists in the database
	var alreadyExistingAthletes []databaseUtils.Athlete
	var newAthletes []databaseUtils.Athlete
	for _, athlete := range athletes {
		exists, err := athleteExists(ctx, &athlete, false)
		if err != nil {
			return err, nil
		}

		if exists {
			alreadyExistingAthletes = append(alreadyExistingAthletes, athlete)
		} else {
			newAthletes = append(newAthletes, athlete)
		}
	}

	// Write the new athletes to the database if not empty
	if len(newAthletes) > 0 {
		err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
			err := tx.Create(&newAthletes).Error
			err = errors.Wrap(err, "Failed to create new athletes")
			return err
		})
		if err1 != nil {
			return err1, alreadyExistingAthletes
		}
	} else {
		return NoNewAthletesError, alreadyExistingAthletes
	}

	return nil, alreadyExistingAthletes
}

// athleteExists checks if the given athlete is already in the database (email and birth_date combination).
// Note: If the error is not nil, the bool is invalid.
func athleteExists(ctx context.Context, athlete *databaseUtils.Athlete, checkWithId bool) (bool, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "CheckAthleteExists")
	defer span.End()

	// Validate all values of the athlete
	if err := validateAthlete(ctx, athlete); err != nil {
		return false, err
	}

	// Check if the email and birth_date combo already exists
	var athleteCount int64
	err2 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		var err error
		if checkWithId {
			err = tx.Model(&databaseUtils.Athlete{}).
				Where("email ILIKE ? AND birth_date = ? AND first_name ILIKE ? AND id != ?",
					strings.ToLower(athlete.Email), athlete.BirthDate, athlete.FirstName, athlete.ID).
				Count(&athleteCount).Error
		} else {
			err = tx.Model(&databaseUtils.Athlete{}).
				Where("email ILIKE ? AND birth_date = ? AND first_name ILIKE ?",
					strings.ToLower(athlete.Email), athlete.BirthDate, athlete.FirstName).
				Count(&athleteCount).Error
		}
		return err
	})
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to check if the athlete exists")
		return false, err2
	}

	return athleteCount > 0, nil
}

// validateAthlete checks if all values of an athlete are valid
func validateAthlete(ctx context.Context, athlete *databaseUtils.Athlete) error {
	ctx, span := endpoints.Tracer.Start(ctx, "ValidateAthlete")
	defer span.End()

	// Capitalize the first letter of the name
	athlete.FirstName = strings.ToUpper(string(athlete.FirstName[0])) + athlete.FirstName[1:]
	athlete.LastName = strings.ToUpper(string(athlete.LastName[0])) + athlete.LastName[1:]

	athlete.Email = strings.ToLower(athlete.Email)
	if err := formatHelper.IsEmail(athlete.Email); err != nil {
		err = errors.Wrap(err, "Invalid email address")
		return err
	}

	if err := formatHelper.IsDate(athlete.BirthDate); err != nil {
		err = errors.Wrap(err, "Invalid date")
		return err
	}

	athlete.Sex = strings.ToLower(string(athlete.Sex[0]))
	if err := formatHelper.IsSex(athlete.Sex); err != nil {
		err = errors.Wrap(err, "Invalid sex")
		return err
	}

	return nil
}

// AthleteExistsForTrainer checks if an athlete with the given id exists for the given trainer
func AthleteExistsForTrainer(ctx context.Context, athleteId uint, trainerEmail string) (bool, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "AthleteExistsCheck")
	defer span.End()

	var athleteCount int64
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Athlete{}).Where("id = ? AND trainer_email = ?", athleteId, strings.ToLower(trainerEmail)).Count(&athleteCount).Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to check if the athlete exists")
		return false, err1
	}

	return athleteCount > 0, nil
}
