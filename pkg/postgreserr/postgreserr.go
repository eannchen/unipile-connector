package postgreserr

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// ParseErrorCode parses the error code from the error
func ParseErrorCode(err error) (string, error) {
	if err == nil {
		return "", nil
	}
	var pgErr *pgconn.PgError
	if ok := errors.As(err, &pgErr); ok {
		return pgErr.Code, nil
	}
	return "", errors.New("parse error code failed")
}

// ErrDuplicateKey is the error code for duplicate key
var ErrDuplicateKey = "23505"
