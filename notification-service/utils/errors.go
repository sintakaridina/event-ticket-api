package utils

import (
	"errors"
	"fmt"
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
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Err:        ErrNotFound,
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
	}
}

// NewAlreadyExistsError creates a new already exists error
func NewAlreadyExistsError(resource string) *AppError {
	return &AppError{
		Err:        ErrAlreadyExists,
		StatusCode: http.StatusConflict,
		Message:    fmt.Sprintf("%s already exists", resource),
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
func NewInternalError(err error) *AppError {
	return &AppError{
		Err:        ErrInternal,
		StatusCode: http.StatusInternalServerError,
		Message:    fmt.Sprintf("internal server error: %v", err),
	}
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if an error is an already exists error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidInput checks if an error is an invalid input error
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if an error is a forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsInternal checks if an error is an internal server error
func IsInternal(err error) bool {
	return errors.Is(err, ErrInternal)
}

// GetStatusCode gets the HTTP status code for an error
func GetStatusCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.StatusCode
	}

	switch {
	case IsNotFound(err):
		return http.StatusNotFound
	case IsAlreadyExists(err):
		return http.StatusConflict
	case IsInvalidInput(err):
		return http.StatusBadRequest
	case IsUnauthorized(err):
		return http.StatusUnauthorized
	case IsForbidden(err):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// GetErrorMessage gets a user-friendly error message
func GetErrorMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return err.Error()
}