package athleteManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"strings"
)

var (
	NoNewAthletesError = errors.New("No new Athletes found")
)

// createNewAthletes creates new athletes in the database and returns the athletes that already exist.
// Note: If the error is not nil, the returned list is invalid.
func createNewAthletes(ctx context.Context, athletes []databaseModels.Athlete) (error, []databaseModels.Athlete) {
	ctx, span := endpoints.Tracer.Start(ctx, "CreateNewAthletes")
	defer span.End()

	// Check if an athlete already exists in the database
	var alreadyExistingAthletes []databaseModels.Athlete
	var newAthletes []databaseModels.Athlete
	for _, athlete := range athletes {
		exists, err := athleteExists(ctx, &athlete)
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
func athleteExists(ctx context.Context, athlete *databaseModels.Athlete) (bool, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "CheckAthleteExists")
	defer span.End()

	// Validate all values of the athlete
	if err := validateAthlete(ctx, athlete); err != nil {
		return false, err
	}

	// Check if the email and birth_date combo already exists
	var athleteCount int64
	err2 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseModels.Athlete{}).
			Where("email LIKE ? AND birth_date = ? AND first_name LIKE ?",
				strings.ToLower(athlete.Email), athlete.BirthDate, athlete.FirstName).
			Count(&athleteCount).Error
		err = errors.Wrap(err, "Failed to check if the athlete exists")
		return err
	})
	if err2 != nil {
		return false, err2
	}

	return athleteCount > 0, nil
}

// validateAthlete checks if all values of an athlete are valid
func validateAthlete(ctx context.Context, athlete *databaseModels.Athlete) error {
	ctx, span := endpoints.Tracer.Start(ctx, "ValidateAthlete")
	defer span.End()

	// Capitalize the first letter of the name
	athlete.FirstName = strings.ToUpper(string(athlete.FirstName[0])) + athlete.FirstName[1:]
	athlete.LastName = strings.ToUpper(string(athlete.LastName[0])) + athlete.LastName[1:]
	endpoints.Logger.Debug(ctx, athlete.FirstName)

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
