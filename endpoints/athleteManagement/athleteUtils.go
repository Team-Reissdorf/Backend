package athleteManagement

import (
	"context"
	"strings"
	"time"

	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	NoNewAthletesError = errors.New("No new Athletes found")
)

// translateAthleteBodies translates the athlete body to an athlete db entry.
func translateAthleteBodies(ctx context.Context, athleteBodies []AthleteBody, trainerEmail string) []databaseUtils.Athlete {
	_, span := endpoints.Tracer.Start(ctx, "TranslateAthleteBodies")
	defer span.End()

	athletes := make([]databaseUtils.Athlete, len(athleteBodies))
	for idx, athlete := range athleteBodies {
		athletes[idx] = databaseUtils.Athlete{
			FirstName:    athlete.FirstName,
			LastName:     athlete.LastName,
			BirthDate:    athlete.BirthDate,
			Sex:          athlete.Sex,
			Email:        athlete.Email,
			TrainerEmail: trainerEmail,
		}
	}

	return athletes
}

// translateAthleteToResponse converts an athlete database object to response type
func translateAthleteToResponse(ctx context.Context, athlete databaseUtils.Athlete, swimcert bool) (*AthleteBodyWithId, error) {
	_, span := endpoints.Tracer.Start(ctx, "TranslateAthleteToResponse")
	defer span.End()

	// Reformat the date to the correct format
	birthDate, err := formatHelper.FormatDate(athlete.BirthDate)
	if err != nil {
		return nil, err
	}

	athleteResponse := AthleteBodyWithId{
		AthleteId: athlete.ID,
		FirstName: athlete.FirstName,
		LastName:  athlete.LastName,
		Email:     athlete.Email,
		BirthDate: birthDate,
		Sex:       athlete.Sex,
		SwimCert:  swimcert,
	}

	return &athleteResponse, nil
}

// validateAthlete checks if all values of an athlete are valid
// Throws: Forwards errors of the formatHelper
func validateAthlete(ctx context.Context, athlete *databaseUtils.Athlete) error {
	_, span := endpoints.Tracer.Start(ctx, "ValidateAthlete")
	defer span.End()
	if err := formatHelper.IsEmpty(athlete.FirstName); err != nil {
		return errors.Wrap(err, "First Name")
	}

	if err := formatHelper.IsEmpty(athlete.LastName); err != nil {
		return errors.Wrap(err, "Last Name")
	}

	// Capitalize the first letter of the name
	athlete.FirstName = strings.ToUpper(string(athlete.FirstName[0])) + athlete.FirstName[1:]
	athlete.LastName = strings.ToUpper(string(athlete.LastName[0])) + athlete.LastName[1:]

	athlete.Email = strings.ToLower(athlete.Email)
	if err := formatHelper.IsEmpty(athlete.Email); err != nil {
		return errors.Wrap(err, "Email address")
	} else if err = formatHelper.IsEmail(athlete.Email); err != nil {
		err = errors.Wrap(err, "Invalid email address")
		return err
	}

	if err := formatHelper.IsEmpty(athlete.BirthDate); err != nil {
		return errors.Wrap(err, "Birthdate")
	} else if err = formatHelper.IsDate(athlete.BirthDate); err != nil {
		err = errors.Wrap(err, "Invalid date")
		return err
	}

	if err := formatHelper.IsEmpty(athlete.Sex); err != nil {
		return errors.Wrap(err, "Sex")
	}
	if err := formatHelper.IsFuture(athlete.BirthDate); err != nil {
		err = errors.Wrap(err, "Date is in the future")
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
			return err
		})
		if err1 != nil {
			err1 = errors.Wrap(err1, "Failed to create new athletes")
			return err1, alreadyExistingAthletes
		}
	} else {
		return NoNewAthletesError, alreadyExistingAthletes
	}

	return nil, alreadyExistingAthletes
}

// updateAthlete updates the given athlete in the database
func updateAthlete(ctx context.Context, athlete databaseUtils.Athlete) error {
	ctx, span := endpoints.Tracer.Start(ctx, "EditAthlete")
	defer span.End()

	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(databaseUtils.Athlete{}).Where("id = ?", athlete.ID).Updates(athlete).Error
		return err
	})
	err1 = databaseUtils.TranslatePostgresError(err1)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to update the athlete")
	}
	return err1
}

// athleteExists checks if the given athlete is already in the database (email and birth_date combination).
// Note: If the error is not nil, the bool is invalid.
func athleteExists(ctx context.Context, athlete *databaseUtils.Athlete, checkWithId bool) (bool, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "CheckAthleteExists")
	defer span.End()

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

// GetAthlete returns the athlete of the given id
func GetAthlete(ctx context.Context, athleteId uint, trainerEmail string) (*databaseUtils.Athlete, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetAthleteFromDB")
	defer span.End()

	var athlete databaseUtils.Athlete
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Athlete{}).Where("trainer_email = ? AND id = ?", strings.ToLower(trainerEmail), athleteId).First(&athlete).Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get the athlete")
		return nil, err1
	}

	return &athlete, nil
}

// GetAthleteDirectly returns the athlete of the given id
func GetAthleteDirectly(ctx context.Context, athleteId uint) (*databaseUtils.Athlete, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetAthleteFromDB")
	defer span.End()

	var athlete databaseUtils.Athlete
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Athlete{}).
			Where("id = ?", athleteId).
			First(&athlete).
			Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get the athlete")
		return nil, err1
	}

	return &athlete, nil
}

// GetAthleteFromPerformanceId returns the athlete of the given performance entry
func GetAthleteFromPerformanceId(ctx context.Context, performanceId uint, trainerEmail string) (*databaseUtils.Athlete, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "GetAthleteFromPerformanceEntryFromDB")
	defer span.End()

	var athlete databaseUtils.Athlete
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Athlete{}).
			Joins("LEFT JOIN performances ON performances.athlete_id = athletes.id").
			Where("trainer_email = ? AND performances.id = ?", strings.ToLower(trainerEmail), performanceId).
			First(&athlete).
			Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to get the athlete")
		return nil, err1
	}

	return &athlete, nil
}

// CalculateAge parses the birthDate string and returns the age
func CalculateAge(ctx context.Context, birthDate string) (int, error) {
	_, span := endpoints.Tracer.Start(ctx, "CalculateAge")
	defer span.End()

	birthDay, err1 := time.Parse("2006-01-02", birthDate)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to parse the birth date")
		return -1, err1
	}
	age := time.Now().Year() - birthDay.Year()
	if time.Now().Before(birthDay.AddDate(age, 0, 0)) {
		age--
	}
	return age, nil
}
