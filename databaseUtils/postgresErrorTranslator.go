package databaseUtils

import (
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
)

const (
	CodeForeignKeyViolation = "23503"
	CodeDuplicateKeyValue   = "23505"
)

var (
	ErrForeignKeyViolation = errors.New("Foreign key constraint violation")
)

// TranslatePostgresError translates a postgres error to a more readable error
func TranslatePostgresError(err error) error {
	if err == nil {
		return err
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case CodeForeignKeyViolation:
			return ErrForeignKeyViolation
		case CodeDuplicateKeyValue:
			return ErrForeignKeyViolation
		}
	}

	return err
}
