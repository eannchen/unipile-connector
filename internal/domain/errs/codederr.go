package errs

import "fmt"

// Kind represents the kind of error
type Kind string

// ErrorKinds represents the kind of errors
const (
	ValidationErrorKind Kind = "VALIDATION ERROR" // Handler validation errors
	BusinessErrorKind   Kind = "BUSINESS ERROR"   // Usecase business logic errors
	SystemErrorKind     Kind = "SYSTEM ERROR"     // System/infrastructure errors
)

// CodedError represents a coded error with code, message, and error
type CodedError struct {
	Kind    Kind   `json:"kind"` // Identifier
	Message string `json:"message"`
	Err     error  `json:"error"` // original error
}

func (e *CodedError) Error() string {
	if e.Message != "" {
		if e.Err != nil {
			return fmt.Sprintf("%s: %s (%s)", e.Kind, e.Message, e.Err.Error())
		}
		return fmt.Sprintf("%s: %s", e.Kind, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Kind, e.Err.Error())
	}
	return string(e.Kind)
}

// Unwrap returns the original error
func (e *CodedError) Unwrap() error {
	return e.Err
}

// WrapValidationError wraps an error with a validation error
func WrapValidationError(err error, msg string) error {
	return &CodedError{
		Err:     err,
		Kind:    ValidationErrorKind,
		Message: msg,
	}
}

// WrapBusinessError wraps an error with a business error
func WrapBusinessError(err error, msg string) error {
	return &CodedError{
		Err:     err,
		Kind:    BusinessErrorKind,
		Message: msg,
	}
}

// WrapInternalError wraps an error with a system error
func WrapInternalError(err error, msg string) error {
	return &CodedError{
		Err:     err,
		Kind:    SystemErrorKind,
		Message: msg,
	}
}
