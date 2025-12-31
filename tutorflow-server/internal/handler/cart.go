package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/usecase/cart"
)

// CartHandler handles cart and wishlist HTTP requests
type CartHandler struct {
	cartUC *cart.UseCase
}

// NewCartHandler creates a new cart handler
func NewCartHandler(cartUC *cart.UseCase) *CartHandler {
	return &CartHandler{cartUC: cartUC}
}

// RegisterRoutes registers cart and wishlist routes
func (h *CartHandler) RegisterRoutes(g *echo.Group, authMW, optionalAuthMW echo.MiddlewareFunc) {
	// Cart routes (support guest checkout)
	cartGroup := g.Group("/cart")
	cartGroup.GET("", h.GetCart, optionalAuthMW)
	cartGroup.GET("/summary", h.GetCartSummary, optionalAuthMW)
	cartGroup.POST("/items", h.AddToCart, optionalAuthMW)
	cartGroup.DELETE("/items/:courseId", h.RemoveFromCart, optionalAuthMW)
	cartGroup.DELETE("", h.ClearCart, optionalAuthMW)
	cartGroup.POST("/merge", h.MergeCart, authMW)

	// Wishlist routes (auth required)
	wishlistGroup := g.Group("/wishlist", authMW)
	wishlistGroup.GET("", h.GetWishlist)
	wishlistGroup.POST("/:courseId", h.AddToWishlist)
	wishlistGroup.DELETE("/:courseId", h.RemoveFromWishlist)
	wishlistGroup.GET("/:courseId/check", h.IsInWishlist)
	wishlistGroup.POST("/:courseId/move-to-cart", h.MoveToCart)
}

// GetCart godoc
// @Summary Get cart
// @Tags Cart
// @Produce json
// @Success 200 {object} response.Response
// @Router /cart [get]
func (h *CartHandler) GetCart(c echo.Context) error {
	userID, sessionID := h.getCartIdentifiers(c)

	cart, err := h.cartUC.GetCart(c.Request().Context(), userID, sessionID)
	if err != nil {
		return response.InternalError(c, "Failed to get cart")
	}

	return response.Success(c, cart)
}

// GetCartSummary godoc
// @Summary Get cart summary with totals
// @Tags Cart
// @Produce json
// @Success 200 {object} response.Response{data=cart.CartSummary}
// @Router /cart/summary [get]
func (h *CartHandler) GetCartSummary(c echo.Context) error {
	userID, sessionID := h.getCartIdentifiers(c)

	summary, err := h.cartUC.GetCartSummary(c.Request().Context(), userID, sessionID)
	if err != nil {
		return response.InternalError(c, "Failed to get cart summary")
	}

	return response.Success(c, summary)
}

// AddToCart godoc
// @Summary Add course to cart
// @Tags Cart
// @Accept json
// @Produce json
// @Param request body cart.AddToCartInput true "Course ID"
// @Success 200 {object} response.Response
// @Router /cart/items [post]
func (h *CartHandler) AddToCart(c echo.Context) error {
	userID, sessionID := h.getCartIdentifiers(c)

	var input cart.AddToCartInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	updatedCart, err := h.cartUC.AddToCart(c.Request().Context(), userID, sessionID, input)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, updatedCart)
}

// RemoveFromCart godoc
// @Summary Remove course from cart
// @Tags Cart
// @Param courseId path string true "Course ID"
// @Success 200 {object} response.Response
// @Router /cart/items/{courseId} [delete]
func (h *CartHandler) RemoveFromCart(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	userID, sessionID := h.getCartIdentifiers(c)

	updatedCart, err := h.cartUC.RemoveFromCart(c.Request().Context(), userID, sessionID, courseID)
	if err != nil {
		return response.InternalError(c, "Failed to remove from cart")
	}

	return response.Success(c, updatedCart)
}

// ClearCart godoc
// @Summary Clear cart
// @Tags Cart
// @Success 204
// @Router /cart [delete]
func (h *CartHandler) ClearCart(c echo.Context) error {
	userID, sessionID := h.getCartIdentifiers(c)

	if err := h.cartUC.ClearCart(c.Request().Context(), userID, sessionID); err != nil {
		return response.InternalError(c, "Failed to clear cart")
	}

	return response.NoContent(c)
}

// MergeCartInput for merging guest cart
type MergeCartInput struct {
	SessionID string `json:"session_id" validate:"required"`
}

// MergeCart godoc
// @Summary Merge guest cart with user cart
// @Tags Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body MergeCartInput true "Session ID"
// @Success 200 {object} response.Response
// @Router /cart/merge [post]
func (h *CartHandler) MergeCart(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	var input MergeCartInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.cartUC.MergeGuestCart(c.Request().Context(), input.SessionID, claims.UserID); err != nil {
		return response.InternalError(c, "Failed to merge cart")
	}

	return response.SuccessWithMessage(c, "Cart merged successfully", nil)
}

// Helper to get cart identifiers
func (h *CartHandler) getCartIdentifiers(c echo.Context) (*uuid.UUID, *string) {
	claims, ok := middleware.GetClaims(c)
	if ok {
		return &claims.UserID, nil
	}

	// Use session ID from header or cookie
	sessionID := c.Request().Header.Get("X-Session-ID")
	if sessionID == "" {
		cookie, err := c.Cookie("session_id")
		if err == nil {
			sessionID = cookie.Value
		}
	}

	if sessionID != "" {
		return nil, &sessionID
	}

	// Generate new session ID
	newSessionID := uuid.New().String()
	return nil, &newSessionID
}

// --- Wishlist Handlers ---

// GetWishlist godoc
// @Summary Get wishlist
// @Tags Wishlist
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /wishlist [get]
func (h *CartHandler) GetWishlist(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	items, total, err := h.cartUC.GetWishlist(c.Request().Context(), claims.UserID, 1, 50)
	if err != nil {
		return response.InternalError(c, "Failed to get wishlist")
	}

	return response.Paginated(c, items, 1, 50, total)
}

// AddToWishlist godoc
// @Summary Add course to wishlist
// @Tags Wishlist
// @Security BearerAuth
// @Param courseId path string true "Course ID"
// @Success 200 {object} response.Response
// @Router /wishlist/{courseId} [post]
func (h *CartHandler) AddToWishlist(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.cartUC.AddToWishlist(c.Request().Context(), claims.UserID, courseID); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Added to wishlist", nil)
}

// RemoveFromWishlist godoc
// @Summary Remove course from wishlist
// @Tags Wishlist
// @Security BearerAuth
// @Param courseId path string true "Course ID"
// @Success 204
// @Router /wishlist/{courseId} [delete]
func (h *CartHandler) RemoveFromWishlist(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.cartUC.RemoveFromWishlist(c.Request().Context(), claims.UserID, courseID); err != nil {
		return response.InternalError(c, "Failed to remove from wishlist")
	}

	return response.NoContent(c)
}

// IsInWishlist godoc
// @Summary Check if course is in wishlist
// @Tags Wishlist
// @Security BearerAuth
// @Param courseId path string true "Course ID"
// @Success 200 {object} response.Response{data=map[string]bool}
// @Router /wishlist/{courseId}/check [get]
func (h *CartHandler) IsInWishlist(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)

	exists, err := h.cartUC.IsInWishlist(c.Request().Context(), claims.UserID, courseID)
	if err != nil {
		return response.InternalError(c, "Failed to check wishlist")
	}

	return response.Success(c, map[string]bool{"in_wishlist": exists})
}

// MoveToCart godoc
// @Summary Move course from wishlist to cart
// @Tags Wishlist
// @Security BearerAuth
// @Param courseId path string true "Course ID"
// @Success 200 {object} response.Response
// @Router /wishlist/{courseId}/move-to-cart [post]
func (h *CartHandler) MoveToCart(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.cartUC.MoveToCart(c.Request().Context(), claims.UserID, courseID); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Moved to cart", nil)
}
