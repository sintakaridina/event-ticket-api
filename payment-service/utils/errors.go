package utils

import (
	"errors"
	"net/http"
)

// Common error types
var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInternal      = errors.New("internal server error")
)

// AppError represents an application error with status code and message
type AppError struct {
	Err        error
	StatusCode int
	Message    string
}

// Error returns the error message
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Err:        ErrNotFound,
		StatusCode: http.StatusNotFound,
		Message:    message,
	}
}

// NewAlreadyExistsError creates a new already exists error
func NewAlreadyExistsError(message string) *AppError {
	return &AppError{
		Err:        ErrAlreadyExists,
		StatusCode: http.StatusConflict,
		Message:    message,
	}
}

// NewInvalidInputError creates a new invalid input error
func NewInvalidInputError(message string) *AppError {
	return &AppError{
		Err:        ErrInvalidInput,
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Err:        ErrUnauthorized,
		StatusCode: http.StatusUnauthorized,
		Message:    message,
	}
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Err:        ErrForbidden,
		StatusCode: http.StatusForbidden,
		Message:    message,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(message string) *AppError {
	return &AppError{
		Err:        ErrInternal,
		StatusCode: http.StatusInternalServerError,
		Message:    message,
	}
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExistsError checks if the error is an already exists error
func IsAlreadyExistsError(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidInputError checks if the error is an invalid input error
func IsInvalidInputError(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorizedError checks if the error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbiddenError checks if the error is a forbidden error
func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsInternalError checks if the error is an internal server error
func IsInternalError(err error) bool {
	return errors.Is(err, ErrInternal)
}

// GetStatusCode returns the status code for the error
func GetStatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}

	switch {
	case IsNotFoundError(err):
		return http.StatusNotFound
	case IsAlreadyExistsError(err):
		return http.StatusConflict
	case IsInvalidInputError(err):
		return http.StatusBadRequest
	case IsUnauthorizedError(err):
		return http.StatusUnauthorized
	case IsForbiddenError(err):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// GetErrorMessage returns a user-friendly error message
func GetErrorMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}

	return err.Error()
}