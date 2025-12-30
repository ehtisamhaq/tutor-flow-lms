package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/usecase/notification"
)

// NotificationHandler handles notification HTTP requests
type NotificationHandler struct {
	notificationUC *notification.UseCase
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationUC *notification.UseCase) *NotificationHandler {
	return &NotificationHandler{notificationUC: notificationUC}
}

// RegisterRoutes registers notification routes
func (h *NotificationHandler) RegisterRoutes(g *echo.Group, authMW echo.MiddlewareFunc) {
	notifications := g.Group("/notifications", authMW)
	notifications.GET("", h.List)
	notifications.GET("/unread-count", h.GetUnreadCount)
	notifications.POST("/:id/read", h.MarkAsRead)
	notifications.POST("/read-all", h.MarkAllAsRead)
	notifications.DELETE("/:id", h.Delete)
}

// List godoc
// @Summary Get notifications
// @Tags Notifications
// @Security BearerAuth
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {object} response.Response
// @Router /notifications [get]
func (h *NotificationHandler) List(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	// Parse query params
	page := 1
	limit := 20

	notifications, total, err := h.notificationUC.GetNotifications(c.Request().Context(), claims.UserID, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get notifications")
	}

	return response.Paginated(c, notifications, page, limit, total)
}

// GetUnreadCount godoc
// @Summary Get unread notification count
// @Tags Notifications
// @Security BearerAuth
// @Success 200 {object} response.Response{data=map[string]int64}
// @Router /notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	count, err := h.notificationUC.GetUnreadCount(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.InternalError(c, "Failed to get unread count")
	}

	return response.Success(c, map[string]int64{"unread_count": count})
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Tags Notifications
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} response.Response
// @Router /notifications/{id}/read [post]
func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid notification ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.notificationUC.MarkAsRead(c.Request().Context(), id, claims.UserID); err != nil {
		return response.InternalError(c, "Failed to mark as read")
	}

	return response.SuccessWithMessage(c, "Marked as read", nil)
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Tags Notifications
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /notifications/read-all [post]
func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	if err := h.notificationUC.MarkAllAsRead(c.Request().Context(), claims.UserID); err != nil {
		return response.InternalError(c, "Failed to mark all as read")
	}

	return response.SuccessWithMessage(c, "All notifications marked as read", nil)
}

// Delete godoc
// @Summary Delete notification
// @Tags Notifications
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 204
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid notification ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.notificationUC.DeleteNotification(c.Request().Context(), id, claims.UserID); err != nil {
		return response.InternalError(c, "Failed to delete notification")
	}

	return response.NoContent(c)
}
