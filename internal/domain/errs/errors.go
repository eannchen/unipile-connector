package errs

import "errors"

// Validation errors
var (
	ErrUserNotAuthenticated = WrapValidationError(errors.New("user not authenticated"), "User not authenticated")
	ErrInvalidUserID        = WrapValidationError(errors.New("invalid user ID"), "Invalid user ID")
)
