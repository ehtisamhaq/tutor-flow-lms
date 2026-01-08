package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/usecase/announcement"
)

// AnnouncementHandler handles announcement HTTP requests
type AnnouncementHandler struct {
	announcementUC *announcement.UseCase
}

// NewAnnouncementHandler creates a new announcement handler
func NewAnnouncementHandler(announcementUC *announcement.UseCase) *AnnouncementHandler {
	return &AnnouncementHandler{announcementUC: announcementUC}
}

// RegisterRoutes registers announcement routes
func (h *AnnouncementHandler) RegisterRoutes(g *echo.Group, authMW, tutorMW echo.MiddlewareFunc) {
	announcements := g.Group("/announcements")
	announcements.GET("/feed", h.GetMyFeed, authMW)
	announcements.GET("/my", h.GetMyAnnouncements, authMW, tutorMW)
	announcements.GET("/global", h.GetGlobalAnnouncements, authMW)
	announcements.GET("/course/:courseId", h.GetCourseAnnouncements, authMW)
	announcements.GET("/:id", h.GetAnnouncement, authMW)
	announcements.POST("", h.CreateAnnouncement, authMW, tutorMW)
	announcements.PUT("/:id", h.UpdateAnnouncement, authMW, tutorMW)
	announcements.DELETE("/:id", h.DeleteAnnouncement, authMW, tutorMW)
	announcements.POST("/:id/pin", h.PinAnnouncement, authMW, tutorMW)
	announcements.DELETE("/:id/pin", h.UnpinAnnouncement, authMW, tutorMW)
}

// GetMyFeed godoc
// @Summary Get my announcements feed
// @Tags Announcements
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /announcements/feed [get]
func (h *AnnouncementHandler) GetMyFeed(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	page, limit := 1, 20
	announcements, total, err := h.announcementUC.GetMyFeed(c.Request().Context(), claims.UserID, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get announcements")
	}

	return response.Paginated(c, announcements, page, limit, total)
}

// GetMyAnnouncements godoc
// @Summary Get announcements I created
// @Tags Announcements
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /announcements/my [get]
func (h *AnnouncementHandler) GetMyAnnouncements(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	page, limit := 1, 20
	announcements, total, err := h.announcementUC.GetMyAnnouncements(c.Request().Context(), claims.UserID, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get announcements")
	}

	return response.Paginated(c, announcements, page, limit, total)
}

// GetGlobalAnnouncements godoc
// @Summary Get global announcements
// @Tags Announcements
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /announcements/global [get]
func (h *AnnouncementHandler) GetGlobalAnnouncements(c echo.Context) error {
	page, limit := 1, 20
	announcements, total, err := h.announcementUC.GetGlobalAnnouncements(c.Request().Context(), page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get announcements")
	}

	return response.Paginated(c, announcements, page, limit, total)
}

// GetCourseAnnouncements godoc
// @Summary Get course announcements
// @Tags Announcements
// @Security BearerAuth
// @Param courseId path string true "Course ID"
// @Success 200 {object} response.Response
// @Router /announcements/course/{courseId} [get]
func (h *AnnouncementHandler) GetCourseAnnouncements(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	page, limit := 1, 20
	announcements, total, err := h.announcementUC.GetCourseAnnouncements(c.Request().Context(), courseID, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get announcements")
	}

	return response.Paginated(c, announcements, page, limit, total)
}

// GetAnnouncement godoc
// @Summary Get announcement by ID
// @Tags Announcements
// @Security BearerAuth
// @Param id path string true "Announcement ID"
// @Success 200 {object} response.Response{data=domain.Announcement}
// @Router /announcements/{id} [get]
func (h *AnnouncementHandler) GetAnnouncement(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid announcement ID")
	}

	ann, err := h.announcementUC.GetAnnouncement(c.Request().Context(), id)
	if err != nil || ann == nil {
		return response.NotFound(c, "Announcement not found")
	}

	return response.Success(c, ann)
}

// CreateAnnouncement godoc
// @Summary Create an announcement
// @Tags Announcements
// @Security BearerAuth
// @Accept json
// @Param request body announcement.CreateAnnouncementInput true "Announcement data"
// @Success 201 {object} response.Response{data=domain.Announcement}
// @Router /announcements [post]
func (h *AnnouncementHandler) CreateAnnouncement(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	var input announcement.CreateAnnouncementInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return validator.FormatValidationErrors(err)
	}

	ann, err := h.announcementUC.CreateAnnouncement(c.Request().Context(), claims.UserID, isAdmin, input)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, ann)
}

// UpdateAnnouncement godoc
// @Summary Update an announcement
// @Tags Announcements
// @Security BearerAuth
// @Accept json
// @Param id path string true "Announcement ID"
// @Param request body announcement.UpdateAnnouncementInput true "Update data"
// @Success 200 {object} response.Response{data=domain.Announcement}
// @Router /announcements/{id} [put]
func (h *AnnouncementHandler) UpdateAnnouncement(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid announcement ID")
	}

	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	var input announcement.UpdateAnnouncementInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	ann, err := h.announcementUC.UpdateAnnouncement(c.Request().Context(), id, claims.UserID, isAdmin, input)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, ann)
}

// DeleteAnnouncement godoc
// @Summary Delete an announcement
// @Tags Announcements
// @Security BearerAuth
// @Param id path string true "Announcement ID"
// @Success 204
// @Router /announcements/{id} [delete]
func (h *AnnouncementHandler) DeleteAnnouncement(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid announcement ID")
	}

	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	if err := h.announcementUC.DeleteAnnouncement(c.Request().Context(), id, claims.UserID, isAdmin); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.NoContent(c)
}

// PinAnnouncement godoc
// @Summary Pin an announcement
// @Tags Announcements
// @Security BearerAuth
// @Param id path string true "Announcement ID"
// @Success 200 {object} response.Response
// @Router /announcements/{id}/pin [post]
func (h *AnnouncementHandler) PinAnnouncement(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid announcement ID")
	}

	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	if err := h.announcementUC.PinAnnouncement(c.Request().Context(), id, claims.UserID, isAdmin, true); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Announcement pinned", nil)
}

// UnpinAnnouncement godoc
// @Summary Unpin an announcement
// @Tags Announcements
// @Security BearerAuth
// @Param id path string true "Announcement ID"
// @Success 200 {object} response.Response
// @Router /announcements/{id}/pin [delete]
func (h *AnnouncementHandler) UnpinAnnouncement(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid announcement ID")
	}

	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	if err := h.announcementUC.PinAnnouncement(c.Request().Context(), id, claims.UserID, isAdmin, false); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Announcement unpinned", nil)
}
