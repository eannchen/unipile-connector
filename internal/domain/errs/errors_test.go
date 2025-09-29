package errs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSentinelErrors(t *testing.T) {
	var ce *CodedError

	require.True(t, errors.As(ErrUserNotAuthenticated, &ce))
	require.Equal(t, ValidationErrorKind, ce.Kind)
	require.Equal(t, "User not authenticated", ce.Message)

	require.True(t, errors.As(ErrInvalidUserID, &ce))
	require.Equal(t, ValidationErrorKind, ce.Kind)
	require.Equal(t, "Invalid user ID", ce.Message)

	require.True(t, errors.As(ErrInvalidCodeOrExpiredCheckpoint, &ce))
	require.Equal(t, ValidationErrorKind, ce.Kind)
	require.Equal(t, "Invalid code or expired checkpoint", ce.Message)
}
