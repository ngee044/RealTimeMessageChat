package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	apperrors "github.com/hyunkyulee/RealTimeMessageChat/RestAPI/pkg/errors"
)

// Response represents a standard API response
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Code      string      `json:"code,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// PaginatedData represents paginated response data
type PaginatedData struct {
	Items  interface{} `json:"items"`
	Total  int64       `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

// OK sends a success response with optional data
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// OKWithMessage sends a success response with a message
func OKWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, items interface{}, total int64, limit, offset int) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: PaginatedData{
			Items:  items,
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
		Timestamp: time.Now().Unix(),
	})
}

// Error sends an error response based on the error type
func Error(c *gin.Context, err error) {
	if appErr := apperrors.GetAppError(err); appErr != nil {
		c.JSON(appErr.StatusCode, Response{
			Success:   false,
			Error:     appErr.Message,
			Code:      appErr.Code,
			Timestamp: time.Now().Unix(),
		})
		return
	}

	// Default internal error
	InternalError(c, "Internal server error")
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success:   false,
		Error:     message,
		Code:      apperrors.ErrCodeBadRequest,
		Timestamp: time.Now().Unix(),
	})
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success:   false,
		Error:     message,
		Code:      apperrors.ErrCodeValidation,
		Timestamp: time.Now().Unix(),
	})
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success:   false,
		Error:     message,
		Code:      apperrors.ErrCodeNotFound,
		Timestamp: time.Now().Unix(),
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success:   false,
		Error:     message,
		Code:      apperrors.ErrCodeUnauthorized,
		Timestamp: time.Now().Unix(),
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success:   false,
		Error:     message,
		Code:      apperrors.ErrCodeForbidden,
		Timestamp: time.Now().Unix(),
	})
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Success:   false,
		Error:     message,
		Code:      apperrors.ErrCodeConflict,
		Timestamp: time.Now().Unix(),
	})
}

// TooManyRequests sends a 429 Too Many Requests response
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, Response{
		Success:   false,
		Error:     message,
		Code:      apperrors.ErrCodeTooManyRequests,
		Timestamp: time.Now().Unix(),
	})
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success:   false,
		Error:     message,
		Code:      apperrors.ErrCodeInternal,
		Timestamp: time.Now().Unix(),
	})
}

// ServiceUnavailable sends a 503 Service Unavailable response
func ServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, Response{
		Success:   false,
		Error:     message,
		Code:      apperrors.ErrCodeServiceUnavail,
		Timestamp: time.Now().Unix(),
	})
}
