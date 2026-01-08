package handler

import (
	"net/http"

	"github.com/tutorflow/tutorflow-server/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SubscriptionHandler handles subscription-related HTTP requests
type SubscriptionHandler struct {
	subscriptionUC domain.SubscriptionUseCase
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(uc domain.SubscriptionUseCase) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionUC: uc}
}

// RegisterRoutes registers subscription routes
func (h *SubscriptionHandler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	plans := e.Group("/subscription-plans")
	plans.GET("", h.GetPlans)
	plans.GET("/:slug", h.GetPlanBySlug)

	// Admin routes
	admin := e.Group("/admin/subscription-plans", authMiddleware)
	admin.POST("", h.CreatePlan)
	admin.PUT("/:id", h.UpdatePlan)

	// User subscription routes
	subs := e.Group("/subscriptions", authMiddleware)
	subs.GET("/my", h.GetMySubscription)
	subs.POST("/subscribe", h.Subscribe)
	subs.POST("/cancel", h.Cancel)
	subs.POST("/resume", h.Resume)
	subs.POST("/change", h.ChangePlan)

	// Webhook
	e.POST("/webhooks/stripe/subscription", h.HandleStripeWebhook)
}

// GetPlans returns all active subscription plans
func (h *SubscriptionHandler) GetPlans(c echo.Context) error {
	plans, err := h.subscriptionUC.GetPlans()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Failed to get plans"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    plans,
	})
}

// GetPlanBySlug returns a plan by slug
func (h *SubscriptionHandler) GetPlanBySlug(c echo.Context) error {
	slug := c.Param("slug")

	plan, err := h.subscriptionUC.GetPlanBySlug(slug)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Plan not found"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    plan,
	})
}

// CreatePlan creates a new subscription plan (admin only)
func (h *SubscriptionHandler) CreatePlan(c echo.Context) error {
	var plan domain.SubscriptionPlan
	if err := c.Bind(&plan); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	if err := h.subscriptionUC.CreatePlan(&plan); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    plan,
	})
}

// UpdatePlan updates an existing plan (admin only)
func (h *SubscriptionHandler) UpdatePlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid plan ID"},
		})
	}

	var plan domain.SubscriptionPlan
	if err := c.Bind(&plan); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	plan.ID = id
	if err := h.subscriptionUC.UpdatePlan(&plan); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    plan,
	})
}

// GetMySubscription returns the current user's subscription
func (h *SubscriptionHandler) GetMySubscription(c echo.Context) error {
	userID := getUserIDFromContext(c)

	sub, err := h.subscriptionUC.GetUserSubscription(userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "No active subscription"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    sub,
	})
}

// SubscribeRequest represents a subscription request
type SubscribeRequest struct {
	PlanSlug string                      `json:"plan_slug"`
	Interval domain.SubscriptionInterval `json:"interval"`
}

// Subscribe creates a new subscription
func (h *SubscriptionHandler) Subscribe(c echo.Context) error {
	userID := getUserIDFromContext(c)

	var req SubscribeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	sub, err := h.subscriptionUC.Subscribe(userID, req.PlanSlug, req.Interval)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    sub,
	})
}

// Cancel cancels the current subscription
func (h *SubscriptionHandler) Cancel(c echo.Context) error {
	userID := getUserIDFromContext(c)

	if err := h.subscriptionUC.CancelSubscription(userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Subscription will be canceled at period end",
	})
}

// Resume resumes a canceled subscription
func (h *SubscriptionHandler) Resume(c echo.Context) error {
	userID := getUserIDFromContext(c)

	if err := h.subscriptionUC.ResumeSubscription(userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Subscription resumed",
	})
}

// ChangePlanRequest represents a plan change request
type ChangePlanRequest struct {
	NewPlanSlug string `json:"new_plan_slug"`
}

// ChangePlan changes the subscription plan
func (h *SubscriptionHandler) ChangePlan(c echo.Context) error {
	userID := getUserIDFromContext(c)

	var req ChangePlanRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	sub, err := h.subscriptionUC.ChangeSubscription(userID, req.NewPlanSlug)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    sub,
	})
}

// HandleStripeWebhook handles Stripe subscription webhooks
func (h *SubscriptionHandler) HandleStripeWebhook(c echo.Context) error {
	// In production, verify the webhook signature
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid payload"},
		})
	}

	eventType, _ := payload["type"].(string)
	data, _ := payload["data"].(map[string]interface{})

	if err := h.subscriptionUC.HandleWebhook(eventType, data); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// Helper to get user ID from context
func getUserIDFromContext(c echo.Context) uuid.UUID {
	user := c.Get("user")
	if user == nil {
		return uuid.Nil
	}
	if u, ok := user.(*domain.User); ok {
		return u.ID
	}
	return uuid.Nil
}
