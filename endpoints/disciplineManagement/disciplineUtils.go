package disciplineManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// DisciplineExists checks if the given discipline exists in the database
func DisciplineExists(ctx context.Context, disciplineName string) (bool, error) {
	ctx, span := endpoints.Tracer.Start(ctx, "DisciplineExistsCheck")
	defer span.End()

	var disciplineCount int64
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Discipline{}).Where("name = ?", disciplineName).Count(&disciplineCount).Error
		return err
	})
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to check if the discipline exists")
		return false, err1
	}

	return disciplineCount > 0, nil
}
