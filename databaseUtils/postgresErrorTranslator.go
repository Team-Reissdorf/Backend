package databaseUtils

import (
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
)

var (
	ErrForeignKeyViolation = errors.New("Foreign key constraint violation")
)

func TranslatePostgresError(err error) error {
	if err == nil {
		return err
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return ErrForeignKeyViolation
		}
	}

	return err
}
