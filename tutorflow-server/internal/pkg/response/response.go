package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Response is a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string            `json:"code,omitempty"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Meta contains pagination info
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Success returns a success response
func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessage returns a success response with message
func SuccessWithMessage(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created returns a created response
func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// Paginated returns a paginated response
func Paginated(c echo.Context, data interface{}, page, perPage int, total int64) error {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Error returns an error response
func Error(c echo.Context, status int, message string) error {
	return c.JSON(status, Response{
		Success: false,
		Error: &ErrorInfo{
			Message: message,
		},
	})
}

// ErrorWithCode returns an error response with error code
func ErrorWithCode(c echo.Context, status int, code, message string) error {
	return c.JSON(status, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationErrors returns a validation error response
func ValidationErrors(c echo.Context, errors []ValidationError) error {
	return c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: errors,
		},
	})
}

// BadRequest returns a 400 error
func BadRequest(c echo.Context, message string) error {
	return Error(c, http.StatusBadRequest, message)
}

// Unauthorized returns a 401 error
func Unauthorized(c echo.Context, message string) error {
	if message == "" {
		message = "Unauthorized"
	}
	return Error(c, http.StatusUnauthorized, message)
}

// Forbidden returns a 403 error
func Forbidden(c echo.Context, message string) error {
	if message == "" {
		message = "Forbidden"
	}
	return Error(c, http.StatusForbidden, message)
}

// NotFound returns a 404 error
func NotFound(c echo.Context, message string) error {
	if message == "" {
		message = "Resource not found"
	}
	return Error(c, http.StatusNotFound, message)
}

// InternalError returns a 500 error
func InternalError(c echo.Context, message string) error {
	if message == "" {
		message = "Internal server error"
	}
	return Error(c, http.StatusInternalServerError, message)
}

// NoContent returns 204 No Content
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
