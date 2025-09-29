package errs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodedErrorErrorFormatting(t *testing.T) {
	baseErr := errors.New("root cause")
	ce := &CodedError{Kind: ValidationErrorKind, Message: "Invalid input", Err: baseErr}
	ce.setDetail()

	require.Equal(t, "root cause", ce.Detail)
	require.Equal(t, "VALIDATION ERROR: Invalid input (root cause)", ce.Error())
	require.Equal(t, baseErr, ce.Unwrap())
}

func TestCodedErrorWithoutMessage(t *testing.T) {
	baseErr := errors.New("boom")
	ce := &CodedError{Kind: BusinessErrorKind, Err: baseErr}

	require.Equal(t, "BUSINESS ERROR: boom", ce.Error())
}

func TestCodedErrorWithoutErr(t *testing.T) {
	ce := &CodedError{Kind: SystemErrorKind, Message: "Something happened"}
	require.Equal(t, "SYSTEM ERROR: Something happened", ce.Error())
}

func TestWrapValidationError(t *testing.T) {
	baseErr := fmt.Errorf("validation failed")
	err := WrapValidationError(baseErr, "Invalid data supplied")

	var ce *CodedError
	require.True(t, errors.As(err, &ce))
	require.Equal(t, ValidationErrorKind, ce.Kind)
	require.Equal(t, "Invalid data supplied", ce.Message)
	require.Equal(t, baseErr.Error(), ce.Detail)
}

func TestWrapBusinessError(t *testing.T) {
	baseErr := fmt.Errorf("business rule broken")
	err := WrapBusinessError(baseErr, "Rule violation")

	var ce *CodedError
	require.True(t, errors.As(err, &ce))
	require.Equal(t, BusinessErrorKind, ce.Kind)
	require.Equal(t, "Rule violation", ce.Message)
}

func TestWrapInternalError(t *testing.T) {
	baseErr := fmt.Errorf("io failure")
	err := WrapInternalError(baseErr, "Write failed")

	var ce *CodedError
	require.True(t, errors.As(err, &ce))
	require.Equal(t, SystemErrorKind, ce.Kind)
	require.Equal(t, "Write failed", ce.Message)
}
