package userManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var (
	TrainerAlreadyExistsErr = errors.New("Trainer already exists")
)

// createTrainer checks if the email address is already in use, otherwise the new trainer will be created in the database.
// Throws: TrainerAlreadyExistsErr and other
func createTrainer(ctx context.Context, trainer databaseModels.Trainer) error {
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		// Check if the trainer already exists
		var count int64
		errA := tx.Model(&databaseModels.Trainer{}).Where("email = ?", trainer.Email).Count(&count).Error
		if errA != nil {
			errA = errors.Wrap(errA, "Failed to count trainers")
			return errA
		}
		if count > 0 {
			return TrainerAlreadyExistsErr
		}

		// Create the trainer in the database
		errB := tx.Create(&trainer).Error
		return errB
	})
	return err1
}
