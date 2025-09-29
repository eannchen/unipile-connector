package postgreserr

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// ParseErrorCode parses the error code from the error
func ParseErrorCode(err error) (string, error) {
	var pgErr *pgconn.PgError
	if ok := errors.As(err, &pgErr); ok {
		return pgErr.Code, nil
	}
	return "", errors.New("parse error code failed")
}

// Is checks if the error is a specific error code
func Is(err error, code string) bool {
	if err == nil {
		return false
	}
	if code, err := ParseErrorCode(err); err == nil && code == code {
		return true
	}
	return false
}

// ErrDuplicateKey is the error code for duplicate key
var ErrDuplicateKey = "23505"
