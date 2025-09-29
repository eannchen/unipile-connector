package postgreserr

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestParseErrorCode_Success(t *testing.T) {
	pgErr := &pgconn.PgError{Code: ErrDuplicateKey}
	code, err := ParseErrorCode(pgErr)
	require.NoError(t, err)
	require.Equal(t, ErrDuplicateKey, code)
}

func TestParseErrorCode_Failure(t *testing.T) {
	_, err := ParseErrorCode(errors.New("not a pg error"))
	require.Error(t, err)
	require.Equal(t, "parse error code failed", err.Error())
}

func TestIs(t *testing.T) {
	pgErr := &pgconn.PgError{Code: ErrDuplicateKey}
	require.True(t, Is(pgErr, ErrDuplicateKey))
	require.False(t, Is(pgErr, "12345"))
	require.False(t, Is(nil, ErrDuplicateKey))
	require.False(t, Is(errors.New("not pg"), ErrDuplicateKey))
}
