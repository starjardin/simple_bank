package db

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrorRecordNotFound = pgx.ErrNoRows

const (
	ForeignKeyViolation = "23503"
	UniqueViolation     = "23505"
)

func ErrorCode(err error) string {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
