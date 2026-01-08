package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
)

// ErrorHandler wraps the default Echo error handler
func ErrorHandler(logger *zap.SugaredLogger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		// Default values
		code := http.StatusInternalServerError
		message := "Internal server error"
		var errorCode string
		var details []response.ValidationError

		// Handle Echo HTTP errors
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			if m, ok := he.Message.(string); ok {
				message = m
			} else {
				message = http.StatusText(code)
			}
		}

		// Handle Domain primary errors
		switch err {
		case domain.ErrUserNotFound, domain.ErrCourseNotFound, domain.ErrLessonNotFound, domain.ErrModuleNotFound, domain.ErrQuizNotFound, domain.ErrAssignmentNotFound, domain.ErrSubmissionNotFound, domain.ErrOrderNotFound:
			code = http.StatusNotFound
			message = err.Error()
		case domain.ErrUserAlreadyExists:
			code = http.StatusConflict
			message = err.Error()
			errorCode = "USER_EXISTS"
		case domain.ErrInvalidCredentials, domain.ErrUnauthorized:
			code = http.StatusUnauthorized
			message = "Invalid credentials"
		case domain.ErrForbidden, domain.ErrNoAccess, domain.ErrNotCourseOwner:
			code = http.StatusForbidden
			message = "Access denied"
		case domain.ErrUserSuspended:
			code = http.StatusForbidden
			message = "Account suspended"
			errorCode = "USER_SUSPENDED"
		case domain.ErrUserInactive:
			code = http.StatusForbidden
			message = "Account inactive"
			errorCode = "USER_INACTIVE"
		case domain.ErrUserNotVerified:
			code = http.StatusForbidden
			message = "Email not verified"
			errorCode = "EMAIL_NOT_VERIFIED"
		case domain.ErrCourseNotPublished:
			code = http.StatusBadRequest
			message = "Course is not available"
		case domain.ErrAlreadyEnrolled:
			code = http.StatusBadRequest
			message = "Already enrolled"
		case domain.ErrEnrollmentExpired:
			code = http.StatusForbidden
			message = "Enrollment has expired"
		case domain.ErrMaxAttemptsReached:
			code = http.StatusBadRequest
			message = "Maximum attempts reached"
		case domain.ErrCouponInvalid:
			code = http.StatusBadRequest
			message = "Invalid or expired coupon"
		case domain.ErrCouponNotApplicable:
			code = http.StatusBadRequest
			message = "Coupon not applicable to this order"
		}

		// Handle Validation Errors
		if ve, ok := err.(domain.ValidationErrors); ok {
			code = http.StatusBadRequest
			message = "Validation failed"
			errorCode = "VALIDATION_ERROR"
			for _, e := range ve {
				details = append(details, response.ValidationError{
					Field:   e.Field,
					Message: e.Message,
				})
			}
		}

		// Log errors
		if code >= 500 {
			logger.Errorf("Internal Server Error: %v, Path: %s", err, c.Path())
		} else {
			logger.Debugf("Request Error: %v, Code: %d, Path: %s", err, code, c.Path())
		}

		// Send standardized response
		resp := response.Response{
			Success: false,
			Error: &response.ErrorInfo{
				Code:    errorCode,
				Message: message,
				Details: details,
			},
		}

		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead {
				err = c.NoContent(code)
			} else {
				err = c.JSON(code, resp)
			}
			if err != nil {
				logger.Errorf("Failed to send error response: %v", err)
			}
		}
	}
}
