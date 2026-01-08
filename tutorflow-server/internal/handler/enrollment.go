package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/usecase/enrollment"
)

// EnrollmentHandler handles enrollment-related HTTP requests
type EnrollmentHandler struct {
	enrollmentUC *enrollment.UseCase
}

// NewEnrollmentHandler creates a new enrollment handler
func NewEnrollmentHandler(enrollmentUC *enrollment.UseCase) *EnrollmentHandler {
	return &EnrollmentHandler{enrollmentUC: enrollmentUC}
}

// RegisterRoutes registers enrollment routes
func (h *EnrollmentHandler) RegisterRoutes(g *echo.Group, authMW, managerMW echo.MiddlewareFunc) {
	g.GET("", h.List, authMW, managerMW)
	g.GET("/my", h.MyEnrollments, authMW)
	g.POST("", h.Enroll, authMW)
	g.GET("/:id", h.GetByID, authMW)
	g.PATCH("/:id/cancel", h.Cancel, authMW)
	g.GET("/:id/progress", h.GetProgress, authMW)
	g.POST("/:id/lessons/:lessonId/complete", h.MarkLessonComplete, authMW)
	g.PUT("/:id/lessons/:lessonId/position", h.UpdateVideoPosition, authMW)
}

// List godoc
// @Summary List enrollments
// @Tags Enrollments
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /enrollments [get]
func (h *EnrollmentHandler) List(c echo.Context) error {
	var input enrollment.ListInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid query parameters")
	}

	enrollments, total, err := h.enrollmentUC.List(c.Request().Context(), input)
	if err != nil {
		return response.InternalError(c, "Failed to list enrollments")
	}

	return response.Paginated(c, enrollments, input.Page, input.Limit, total)
}

// MyEnrollments godoc
// @Summary Get my enrollments
// @Tags Enrollments
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /enrollments/my [get]
func (h *EnrollmentHandler) MyEnrollments(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	page := 1
	limit := 20

	enrollments, total, err := h.enrollmentUC.GetMyEnrollments(c.Request().Context(), claims.UserID, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get enrollments")
	}

	return response.Paginated(c, enrollments, page, limit, total)
}

// Enroll godoc
// @Summary Enroll in a course
// @Tags Enrollments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body enrollment.EnrollInput true "Enrollment data"
// @Success 201 {object} response.Response{data=domain.Enrollment}
// @Router /enrollments [post]
func (h *EnrollmentHandler) Enroll(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	var input enrollment.EnrollInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return validator.FormatValidationErrors(err)
	}

	enroll, err := h.enrollmentUC.Enroll(c.Request().Context(), claims.UserID, input)
	if err != nil {
		switch err {
		case domain.ErrAlreadyEnrolled:
			return response.BadRequest(c, "Already enrolled in this course")
		case domain.ErrCourseNotFound:
			return response.NotFound(c, "Course not found")
		case domain.ErrCourseNotPublished:
			return response.BadRequest(c, "Course is not available")
		default:
			return response.InternalError(c, "Failed to enroll")
		}
	}

	return response.Created(c, enroll)
}

// GetByID godoc
// @Summary Get enrollment by ID
// @Tags Enrollments
// @Security BearerAuth
// @Produce json
// @Param id path string true "Enrollment ID"
// @Success 200 {object} response.Response{data=domain.Enrollment}
// @Router /enrollments/{id} [get]
func (h *EnrollmentHandler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid enrollment ID")
	}

	claims, _ := middleware.GetClaims(c)

	enroll, err := h.enrollmentUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return response.NotFound(c, "Enrollment not found")
	}

	// Check authorization
	if enroll.UserID != claims.UserID && claims.Role != domain.RoleAdmin && claims.Role != domain.RoleManager {
		return response.Forbidden(c, "")
	}

	return response.Success(c, enroll)
}

// Cancel godoc
// @Summary Cancel enrollment
// @Tags Enrollments
// @Security BearerAuth
// @Param id path string true "Enrollment ID"
// @Success 200 {object} response.Response
// @Router /enrollments/{id}/cancel [patch]
func (h *EnrollmentHandler) Cancel(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid enrollment ID")
	}

	claims, _ := middleware.GetClaims(c)

	enroll, err := h.enrollmentUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return response.NotFound(c, "Enrollment not found")
	}

	// Check authorization
	if enroll.UserID != claims.UserID && claims.Role != domain.RoleAdmin {
		return response.Forbidden(c, "")
	}

	if err := h.enrollmentUC.Cancel(c.Request().Context(), id); err != nil {
		return response.InternalError(c, "Failed to cancel enrollment")
	}

	return response.SuccessWithMessage(c, "Enrollment cancelled", nil)
}

// GetProgress godoc
// @Summary Get enrollment progress
// @Tags Enrollments
// @Security BearerAuth
// @Produce json
// @Param id path string true "Enrollment ID"
// @Success 200 {object} response.Response
// @Router /enrollments/{id}/progress [get]
func (h *EnrollmentHandler) GetProgress(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid enrollment ID")
	}

	progress, err := h.enrollmentUC.GetProgress(c.Request().Context(), id)
	if err != nil {
		return response.InternalError(c, "Failed to get progress")
	}

	return response.Success(c, progress)
}

// MarkLessonComplete godoc
// @Summary Mark lesson as complete
// @Tags Enrollments
// @Security BearerAuth
// @Param id path string true "Enrollment ID"
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} response.Response
// @Router /enrollments/{id}/lessons/{lessonId}/complete [post]
func (h *EnrollmentHandler) MarkLessonComplete(c echo.Context) error {
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid enrollment ID")
	}

	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return response.BadRequest(c, "Invalid lesson ID")
	}

	claims, _ := middleware.GetClaims(c)

	// Get enrollment to verify ownership and get course ID
	enroll, err := h.enrollmentUC.GetByID(c.Request().Context(), enrollmentID)
	if err != nil {
		return response.NotFound(c, "Enrollment not found")
	}

	if enroll.UserID != claims.UserID {
		return response.Forbidden(c, "")
	}

	if err := h.enrollmentUC.MarkLessonComplete(c.Request().Context(), claims.UserID, enroll.CourseID, lessonID); err != nil {
		switch err {
		case domain.ErrNotEnrolled:
			return response.Forbidden(c, "Not enrolled in this course")
		case domain.ErrEnrollmentExpired:
			return response.Forbidden(c, "Enrollment has expired")
		default:
			return response.InternalError(c, "Failed to mark lesson complete")
		}
	}

	return response.SuccessWithMessage(c, "Lesson marked as complete", nil)
}

// UpdateVideoPosition godoc
// @Summary Update video playback position
// @Tags Enrollments
// @Security BearerAuth
// @Accept json
// @Param id path string true "Enrollment ID"
// @Param lessonId path string true "Lesson ID"
// @Param request body enrollment.UpdateVideoPositionInput true "Position"
// @Success 200 {object} response.Response
// @Router /enrollments/{id}/lessons/{lessonId}/position [put]
func (h *EnrollmentHandler) UpdateVideoPosition(c echo.Context) error {
	enrollmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid enrollment ID")
	}

	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return response.BadRequest(c, "Invalid lesson ID")
	}

	var input enrollment.UpdateVideoPositionInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	input.LessonID = lessonID

	claims, _ := middleware.GetClaims(c)

	// Get enrollment
	enroll, err := h.enrollmentUC.GetByID(c.Request().Context(), enrollmentID)
	if err != nil {
		return response.NotFound(c, "Enrollment not found")
	}

	if enroll.UserID != claims.UserID {
		return response.Forbidden(c, "")
	}

	if err := h.enrollmentUC.UpdateVideoPosition(c.Request().Context(), claims.UserID, enroll.CourseID, input); err != nil {
		return response.InternalError(c, "Failed to update position")
	}

	return response.SuccessWithMessage(c, "Position updated", nil)
}
