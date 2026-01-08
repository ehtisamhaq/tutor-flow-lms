package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v76/webhook"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/service/payment"
	"github.com/tutorflow/tutorflow-server/internal/usecase/order"
)

// OrderHandler handles order HTTP requests
type OrderHandler struct {
	orderUC    *order.UseCase
	paymentSvc *payment.Service
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderUC *order.UseCase, paymentSvc *payment.Service) *OrderHandler {
	return &OrderHandler{
		orderUC:    orderUC,
		paymentSvc: paymentSvc,
	}
}

// RegisterRoutes registers order routes
func (h *OrderHandler) RegisterRoutes(g *echo.Group, authMW, adminMW echo.MiddlewareFunc) {
	// Order routes
	g.GET("/my", h.MyOrders, authMW)
	g.POST("", h.CreateOrder, authMW)
	g.GET("/:id", h.GetOrder, authMW)
	g.POST("/confirm", h.ConfirmPayment, authMW)

	// Webhook (no auth)
	g.POST("/webhook", h.HandleWebhook)

	// Coupon routes
	coupons := g.Group("/coupons")
	coupons.POST("/validate", h.ValidateCoupon, authMW)
	coupons.GET("", h.ListCoupons, authMW, adminMW)
	coupons.POST("", h.CreateCoupon, authMW, adminMW)
	coupons.DELETE("/:id", h.DeleteCoupon, authMW, adminMW)
	coupons.PATCH("/:id/toggle", h.ToggleCoupon, authMW, adminMW)
}

// MyOrders godoc
// @Summary Get my orders
// @Tags Orders
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /orders/my [get]
func (h *OrderHandler) MyOrders(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	orders, total, err := h.orderUC.GetMyOrders(c.Request().Context(), claims.UserID, 1, 20)
	if err != nil {
		return response.InternalError(c, "Failed to get orders")
	}

	return response.Paginated(c, orders, 1, 20, total)
}

// CreateOrder godoc
// @Summary Create order from cart
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body order.CreateOrderInput true "Order data"
// @Success 201 {object} response.Response{data=order.CreateOrderOutput}
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	var input order.CreateOrderInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	output, err := h.orderUC.CreateOrder(c.Request().Context(), claims.UserID, claims.Email, input)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, output)
}

// GetOrder godoc
// @Summary Get order by ID
// @Tags Orders
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response{data=domain.Order}
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid order ID")
	}

	claims, _ := middleware.GetClaims(c)

	orderObj, err := h.orderUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return response.NotFound(c, "Order not found")
	}

	// Check authorization
	if orderObj.UserID != claims.UserID && claims.Role != domain.RoleAdmin {
		return response.Forbidden(c, "")
	}

	return response.Success(c, orderObj)
}

// ConfirmPaymentInput for confirming payment
type ConfirmPaymentInput struct {
	PaymentIntentID string `json:"payment_intent_id" validate:"required"`
}

// ConfirmPayment godoc
// @Summary Confirm payment
// @Tags Orders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ConfirmPaymentInput true "Payment Intent ID"
// @Success 200 {object} response.Response{data=domain.Order}
// @Router /orders/confirm [post]
func (h *OrderHandler) ConfirmPayment(c echo.Context) error {
	var input ConfirmPaymentInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return validator.FormatValidationErrors(err)
	}

	orderObj, err := h.orderUC.ConfirmPayment(c.Request().Context(), input.PaymentIntentID)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, orderObj)
}

// HandleWebhook godoc
// @Summary Handle Stripe webhook
// @Tags Orders
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /orders/webhook [post]
func (h *OrderHandler) HandleWebhook(c echo.Context) error {
	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to read body"})
	}

	// Verify signature
	sigHeader := c.Request().Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, h.paymentSvc.GetWebhookSecret())
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid signature"})
	}

	// Handle event types
	switch event.Type {
	case "payment_intent.succeeded", "payment_intent.payment_failed":
		var pi struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		}

		if err := h.orderUC.HandleWebhook(c.Request().Context(), string(event.Type), pi.ID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "received"})
}

// ValidateCouponInput for coupon validation
type ValidateCouponInput struct {
	Code     string  `json:"code" validate:"required"`
	Subtotal float64 `json:"subtotal" validate:"required,gt=0"`
}

// --- Coupon Handlers ---

// ValidateCoupon godoc
// @Summary Validate coupon code
// @Tags Coupons
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ValidateCouponInput true "Coupon data"
// @Success 200 {object} response.Response
// @Router /orders/coupons/validate [post]
func (h *OrderHandler) ValidateCoupon(c echo.Context) error {
	var input ValidateCouponInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return validator.FormatValidationErrors(err)
	}

	coupon, discount, err := h.orderUC.ValidateCoupon(c.Request().Context(), input.Code, input.Subtotal)
	if err != nil {
		return response.BadRequest(c, "Invalid or expired coupon")
	}

	return response.Success(c, map[string]interface{}{
		"coupon":   coupon,
		"discount": discount,
		"total":    input.Subtotal - discount,
	})
}

// ListCoupons godoc
// @Summary List coupons (admin)
// @Tags Coupons
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /orders/coupons [get]
func (h *OrderHandler) ListCoupons(c echo.Context) error {
	coupons, total, err := h.orderUC.ListCoupons(c.Request().Context(), 1, 50)
	if err != nil {
		return response.InternalError(c, "Failed to list coupons")
	}

	return response.Paginated(c, coupons, 1, 50, total)
}

// CreateCoupon godoc
// @Summary Create coupon (admin)
// @Tags Coupons
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body order.CreateCouponInput true "Coupon data"
// @Success 201 {object} response.Response{data=domain.Coupon}
// @Router /orders/coupons [post]
func (h *OrderHandler) CreateCoupon(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	var input order.CreateCouponInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return validator.FormatValidationErrors(err)
	}

	coupon, err := h.orderUC.CreateCoupon(c.Request().Context(), input, claims.UserID)
	if err != nil {
		return response.InternalError(c, "Failed to create coupon")
	}

	return response.Created(c, coupon)
}

// DeleteCoupon godoc
// @Summary Delete coupon (admin)
// @Tags Coupons
// @Security BearerAuth
// @Param id path string true "Coupon ID"
// @Success 204
// @Router /orders/coupons/{id} [delete]
func (h *OrderHandler) DeleteCoupon(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid coupon ID")
	}

	if err := h.orderUC.DeleteCoupon(c.Request().Context(), id); err != nil {
		return response.InternalError(c, "Failed to delete coupon")
	}

	return response.NoContent(c)
}

// ToggleCouponInput for toggling coupon status
type ToggleCouponInput struct {
	IsActive bool `json:"is_active"`
}

// ToggleCoupon godoc
// @Summary Toggle coupon active status (admin)
// @Tags Coupons
// @Security BearerAuth
// @Accept json
// @Param id path string true "Coupon ID"
// @Param request body ToggleCouponInput true "Active status"
// @Success 200 {object} response.Response
// @Router /orders/coupons/{id}/toggle [patch]
func (h *OrderHandler) ToggleCoupon(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid coupon ID")
	}

	var input ToggleCouponInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.orderUC.ToggleCoupon(c.Request().Context(), id, input.IsActive); err != nil {
		return response.InternalError(c, "Failed to toggle coupon")
	}

	return response.SuccessWithMessage(c, "Coupon updated", nil)
}
