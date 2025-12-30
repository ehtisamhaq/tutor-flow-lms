package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/service/push"
)

// PushHandler handles push notification HTTP requests
type PushHandler struct {
	pushSvc *push.Service
}

// NewPushHandler creates a new push handler
func NewPushHandler(pushSvc *push.Service) *PushHandler {
	return &PushHandler{pushSvc: pushSvc}
}

// RegisterRoutes registers push notification routes
func (h *PushHandler) RegisterRoutes(g *echo.Group, authMW echo.MiddlewareFunc) {
	p := g.Group("/push", authMW)
	p.GET("/vapid-key", h.GetVAPIDPublicKey)
	p.POST("/subscribe", h.Subscribe)
	p.POST("/unsubscribe", h.Unsubscribe)
	p.DELETE("/unsubscribe-all", h.UnsubscribeAll)
	p.GET("/subscriptions", h.GetSubscriptions)
}

// GetVAPIDPublicKey godoc
// @Summary Get VAPID public key for push subscription
// @Tags Push
// @Success 200 {object} response.Response
// @Router /push/vapid-key [get]
func (h *PushHandler) GetVAPIDPublicKey(c echo.Context) error {
	key := h.pushSvc.GetVAPIDPublicKey()
	return response.Success(c, map[string]string{"vapid_public_key": key})
}

// Subscribe godoc
// @Summary Subscribe to push notifications
// @Tags Push
// @Security BearerAuth
// @Accept json
// @Param request body push.SubscriptionInput true "Subscription data"
// @Success 201 {object} response.Response
// @Router /push/subscribe [post]
func (h *PushHandler) Subscribe(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	var input push.SubscriptionInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	sub, err := h.pushSvc.Subscribe(c.Request().Context(), claims.UserID, input)
	if err != nil {
		return response.InternalError(c, "Failed to subscribe")
	}

	return response.Created(c, sub)
}

// UnsubscribeInput for unsubscribing
type UnsubscribeInput struct {
	Endpoint string `json:"endpoint" validate:"required"`
}

// Unsubscribe godoc
// @Summary Unsubscribe from push notifications
// @Tags Push
// @Security BearerAuth
// @Accept json
// @Param request body UnsubscribeInput true "Unsubscribe data"
// @Success 200 {object} response.Response
// @Router /push/unsubscribe [post]
func (h *PushHandler) Unsubscribe(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	var input UnsubscribeInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.pushSvc.Unsubscribe(c.Request().Context(), claims.UserID, input.Endpoint); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Unsubscribed successfully", nil)
}

// UnsubscribeAll godoc
// @Summary Unsubscribe all devices from push notifications
// @Tags Push
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /push/unsubscribe-all [delete]
func (h *PushHandler) UnsubscribeAll(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	if err := h.pushSvc.UnsubscribeAll(c.Request().Context(), claims.UserID); err != nil {
		return response.InternalError(c, "Failed to unsubscribe")
	}

	return response.SuccessWithMessage(c, "All devices unsubscribed", nil)
}

// GetSubscriptions godoc
// @Summary Get my push subscriptions
// @Tags Push
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /push/subscriptions [get]
func (h *PushHandler) GetSubscriptions(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	subs, err := h.pushSvc.GetUserSubscriptions(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.InternalError(c, "Failed to get subscriptions")
	}

	return response.Success(c, subs)
}
