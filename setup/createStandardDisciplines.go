package setup

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/LucaSchmitz2003/FlowWatch"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// CreateStandardDisciplines creates all discipline entries if they don't already exist
func CreateStandardDisciplines(ctx context.Context) {
	ctx, span := endpoints.Tracer.Start(ctx, "Create standard disciplines")
	defer span.End()

	// Define standard disciplines
	disciplines := []databaseUtils.Discipline{
		{Name: "Ausdauer"},
		{Name: "Kraft"},
		{Name: "Schnelligkeit"},
		{Name: "Koordination"},
	}

	// Write disciplines to the database
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(&databaseUtils.Discipline{}).
			Create(&disciplines).
			Error
		err = errors.Wrap(err, "Failed to create standard disciplines")
		return err
	})
	err1 = databaseUtils.TranslatePostgresError(err1)
	if errors.Is(err1, databaseUtils.ErrForeignKeyViolation) {
		FlowWatch.GetLogHelper().Info(ctx, "Standard disciplines already exist")
	} else if err1 != nil {
		FlowWatch.GetLogHelper().Fatal(ctx, err1)
	}
}
