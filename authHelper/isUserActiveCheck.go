package authHelper

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func isUserActive(ctx context.Context, userId string) (bool, error) {
	ctx, span := tracer.Start(ctx, "isUserActive")
	defer span.End()

	// ToDo: Handle the admin user for backendSettings endpoints

	var count int64
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseModels.Trainer{}).Where("email = ?", userId).Count(&count).Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to check if the user exists")
		return false, err1
	}
	if count == 0 {
		return false, nil
	} else if count > 1 {
		endpoints.Logger.Warn(ctx, "Found more than 1 user with email address ", userId)
	}

	return true, nil
}
