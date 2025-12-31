package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application error with code and HTTP status
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap implements the errors.Unwrap interface
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap wraps an existing error with app error context
func Wrap(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from error
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// Common error codes
const (
	ErrCodeValidation       = "VALIDATION_ERROR"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeInternal         = "INTERNAL_ERROR"
	ErrCodeBadRequest       = "BAD_REQUEST"
	ErrCodeServiceUnavail   = "SERVICE_UNAVAILABLE"
	ErrCodeTooManyRequests  = "TOO_MANY_REQUESTS"
	ErrCodeDatabaseError    = "DATABASE_ERROR"
	ErrCodeCacheError       = "CACHE_ERROR"
	ErrCodeQueueError       = "QUEUE_ERROR"
	ErrCodeInvalidPayload   = "INVALID_PAYLOAD"
	ErrCodeDuplicateKey     = "DUPLICATE_KEY"
	ErrCodeInvalidOperation = "INVALID_OPERATION"
)

// Pre-defined errors
var (
	ErrValidation       = New(ErrCodeValidation, "Validation failed", http.StatusBadRequest)
	ErrNotFound         = New(ErrCodeNotFound, "Resource not found", http.StatusNotFound)
	ErrUnauthorized     = New(ErrCodeUnauthorized, "Unauthorized", http.StatusUnauthorized)
	ErrForbidden        = New(ErrCodeForbidden, "Forbidden", http.StatusForbidden)
	ErrConflict         = New(ErrCodeConflict, "Resource conflict", http.StatusConflict)
	ErrInternal         = New(ErrCodeInternal, "Internal server error", http.StatusInternalServerError)
	ErrBadRequest       = New(ErrCodeBadRequest, "Bad request", http.StatusBadRequest)
	ErrServiceUnavail   = New(ErrCodeServiceUnavail, "Service unavailable", http.StatusServiceUnavailable)
	ErrTooManyRequests  = New(ErrCodeTooManyRequests, "Too many requests", http.StatusTooManyRequests)
	ErrDatabaseError    = New(ErrCodeDatabaseError, "Database error", http.StatusInternalServerError)
	ErrCacheError       = New(ErrCodeCacheError, "Cache error", http.StatusInternalServerError)
	ErrQueueError       = New(ErrCodeQueueError, "Queue error", http.StatusInternalServerError)
	ErrInvalidPayload   = New(ErrCodeInvalidPayload, "Invalid payload", http.StatusBadRequest)
	ErrDuplicateKey     = New(ErrCodeDuplicateKey, "Duplicate key", http.StatusConflict)
	ErrInvalidOperation = New(ErrCodeInvalidOperation, "Invalid operation", http.StatusBadRequest)
)
