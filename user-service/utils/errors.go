package utils

import (
	"errors"
	"net/http"
	"strings"
)

// Common error types
var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrUnauthorized  = errors.New("unauthorized access")
	ErrForbidden     = errors.New("forbidden access")
	ErrBadRequest    = errors.New("bad request")
	ErrInternal      = errors.New("internal server error")
	ErrValidation    = errors.New("validation error")
)

// AppError represents an application error
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

// NewAppError creates a new application error
func NewAppError(err error, statusCode int, message string) *AppError {
	return &AppError{
		Err:        err,
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Err:        ErrNotFound,
		StatusCode: http.StatusNotFound,
		Message:    resource + " not found",
	}
}

// NewAlreadyExistsError creates a new already exists error
func NewAlreadyExistsError(resource string) *AppError {
	return &AppError{
		Err:        ErrAlreadyExists,
		StatusCode: http.StatusConflict,
		Message:    resource + " already exists",
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "unauthorized access"
	}
	return &AppError{
		Err:        ErrUnauthorized,
		StatusCode: http.StatusUnauthorized,
		Message:    message,
	}
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "forbidden access"
	}
	return &AppError{
		Err:        ErrForbidden,
		StatusCode: http.StatusForbidden,
		Message:    message,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string) *AppError {
	if message == "" {
		message = "bad request"
	}
	return &AppError{
		Err:        ErrBadRequest,
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(err error) *AppError {
	return &AppError{
		Err:        err,
		StatusCode: http.StatusInternalServerError,
		Message:    "internal server error",
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Err:        ErrValidation,
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, ErrNotFound)
	}
	return errors.Is(err, ErrNotFound) || strings.Contains(err.Error(), "not found")
}

// IsAlreadyExistsError checks if an error is an already exists error
func IsAlreadyExistsError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, ErrAlreadyExists)
	}
	return errors.Is(err, ErrAlreadyExists) || strings.Contains(err.Error(), "already exists")
}

// IsUnauthorizedError checks if an error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, ErrUnauthorized)
	}
	return errors.Is(err, ErrUnauthorized) || strings.Contains(err.Error(), "unauthorized")
}

// IsForbiddenError checks if an error is a forbidden error
func IsForbiddenError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, ErrForbidden)
	}
	return errors.Is(err, ErrForbidden) || strings.Contains(err.Error(), "forbidden")
}

// IsBadRequestError checks if an error is a bad request error
func IsBadRequestError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, ErrBadRequest)
	}
	return errors.Is(err, ErrBadRequest) || strings.Contains(err.Error(), "bad request")
}

// IsInternalError checks if an error is an internal server error
func IsInternalError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, ErrInternal)
	}
	return errors.Is(err, ErrInternal) || strings.Contains(err.Error(), "internal server error")
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, ErrValidation)
	}
	return errors.Is(err, ErrValidation) || strings.Contains(err.Error(), "validation")
}