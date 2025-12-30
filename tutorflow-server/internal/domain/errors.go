package domain

import "errors"

// Common domain errors
var (
	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotVerified    = errors.New("email not verified")
	ErrUserSuspended      = errors.New("user account is suspended")
	ErrUserInactive       = errors.New("user account is inactive")

	// Auth errors
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token has expired")
	ErrTokenRevoked        = errors.New("token has been revoked")
	ErrRefreshTokenInvalid = errors.New("refresh token is invalid")

	// Course errors
	ErrCourseNotFound     = errors.New("course not found")
	ErrCourseNotPublished = errors.New("course is not published")
	ErrNotCourseOwner     = errors.New("not the course owner")

	// Enrollment errors
	ErrAlreadyEnrolled   = errors.New("already enrolled in this course")
	ErrNotEnrolled       = errors.New("not enrolled in this course")
	ErrEnrollmentExpired = errors.New("enrollment has expired")

	// Content errors
	ErrLessonNotFound = errors.New("lesson not found")
	ErrModuleNotFound = errors.New("module not found")
	ErrNoAccess       = errors.New("no access to this content")
	ErrContentLocked  = errors.New("content is locked")

	// Assessment errors
	ErrQuizNotFound       = errors.New("quiz not found")
	ErrAssignmentNotFound = errors.New("assignment not found")
	ErrMaxAttemptsReached = errors.New("maximum attempts reached")
	ErrSubmissionNotFound = errors.New("submission not found")

	// Payment errors
	ErrPaymentFailed       = errors.New("payment failed")
	ErrOrderNotFound       = errors.New("order not found")
	ErrCouponInvalid       = errors.New("coupon is invalid or expired")
	ErrCouponNotApplicable = errors.New("coupon is not applicable")

	// Device/DRM errors
	ErrDeviceLimitReached    = errors.New("device limit reached")
	ErrConcurrentStreamLimit = errors.New("concurrent stream limit reached")

	// Permission errors
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
)

// ValidationError for request validation
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}
	return e[0].Error()
}
