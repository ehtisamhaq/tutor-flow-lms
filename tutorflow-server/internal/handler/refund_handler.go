package handler

import (
	"net/http"
	"strconv"

	"github.com/tutorflow/tutorflow-server/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RefundHandler handles refund-related HTTP requests
type RefundHandler struct {
	refundUC domain.RefundUseCase
}

// NewRefundHandler creates a new refund handler
func NewRefundHandler(uc domain.RefundUseCase) *RefundHandler {
	return &RefundHandler{refundUC: uc}
}

// RegisterRoutes registers refund routes
func (h *RefundHandler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	// User routes
	refunds := e.Group("/refunds", authMiddleware)
	refunds.POST("", h.RequestRefund)
	refunds.GET("/my", h.GetMyRefunds)
	refunds.GET("/:id", h.GetRefund)

	// Admin routes
	admin := e.Group("/admin/refunds", authMiddleware)
	admin.GET("", h.GetAllRefunds)
	admin.GET("/pending", h.GetPendingRefunds)
	admin.POST("/:id/approve", h.ApproveRefund)
	admin.POST("/:id/reject", h.RejectRefund)
}

// RefundRequest represents a refund request
type RefundRequest struct {
	OrderID     uuid.UUID           `json:"order_id"`
	Reason      domain.RefundReason `json:"reason"`
	Description string              `json:"description"`
}

// RequestRefund creates a new refund request
func (h *RefundHandler) RequestRefund(c echo.Context) error {
	userID := getUserIDFromContext(c)

	var req RefundRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	refund, err := h.refundUC.RequestRefund(c.Request().Context(), userID, req.OrderID, req.Reason, req.Description)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    refund,
	})
}

// GetMyRefunds returns the current user's refunds
func (h *RefundHandler) GetMyRefunds(c echo.Context) error {
	userID := getUserIDFromContext(c)
	page, limit := getPagination(c)

	refunds, total, err := h.refundUC.GetUserRefunds(c.Request().Context(), userID, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Failed to get refunds"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"items": refunds,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GetRefund returns a specific refund
func (h *RefundHandler) GetRefund(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid refund ID"},
		})
	}

	refund, err := h.refundUC.GetRefundByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Refund not found"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    refund,
	})
}

// GetAllRefunds returns all refunds (admin)
func (h *RefundHandler) GetAllRefunds(c echo.Context) error {
	page, limit := getPagination(c)

	var status *domain.RefundStatus
	if s := c.QueryParam("status"); s != "" {
		st := domain.RefundStatus(s)
		status = &st
	}

	refunds, total, err := h.refundUC.GetAllRefunds(c.Request().Context(), status, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Failed to get refunds"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"items": refunds,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GetPendingRefunds returns pending refunds (admin)
func (h *RefundHandler) GetPendingRefunds(c echo.Context) error {
	page, limit := getPagination(c)

	refunds, total, err := h.refundUC.GetPendingRefunds(c.Request().Context(), page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Failed to get refunds"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"items": refunds,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// RefundActionRequest represents an approve/reject request
type RefundActionRequest struct {
	Notes string `json:"notes"`
}

// ApproveRefund approves a refund (admin)
func (h *RefundHandler) ApproveRefund(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid refund ID"},
		})
	}

	adminID := getUserIDFromContext(c)

	var req RefundActionRequest
	c.Bind(&req)

	refund, err := h.refundUC.ApproveRefund(c.Request().Context(), id, adminID, req.Notes)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    refund,
	})
}

// RejectRefund rejects a refund (admin)
func (h *RefundHandler) RejectRefund(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid refund ID"},
		})
	}

	adminID := getUserIDFromContext(c)

	var req RefundActionRequest
	c.Bind(&req)

	refund, err := h.refundUC.RejectRefund(c.Request().Context(), id, adminID, req.Notes)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    refund,
	})
}

// Helper function for pagination
func getPagination(c echo.Context) (int, int) {
	page := 1
	limit := 20

	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	return page, limit
}
