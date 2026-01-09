package handler

import (
	"net/http"
	"strconv"

	"github.com/tutorflow/tutorflow-server/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BundleHandler handles bundle-related HTTP requests
type BundleHandler struct {
	bundleUC domain.BundleUseCase
}

// NewBundleHandler creates a new bundle handler
func NewBundleHandler(uc domain.BundleUseCase) *BundleHandler {
	return &BundleHandler{bundleUC: uc}
}

// RegisterRoutes registers bundle routes
func (h *BundleHandler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	// Public routes
	bundles := e.Group("/bundles")
	bundles.GET("", h.GetActiveBundles)
	bundles.GET("/:slug", h.GetBundle)

	// Authenticated routes
	auth := e.Group("/bundles", authMiddleware)
	auth.POST("/:id/purchase", h.PurchaseBundle)
	auth.GET("/my", h.GetMyBundles)

	// Admin routes
	admin := e.Group("/admin/bundles", authMiddleware)
	admin.POST("", h.CreateBundle)
	admin.PUT("/:id", h.UpdateBundle)
	admin.DELETE("/:id", h.DeleteBundle)
	admin.POST("/:id/courses", h.AddCourse)
	admin.DELETE("/:id/courses/:courseId", h.RemoveCourse)
}

// GetActiveBundles returns all active bundles
func (h *BundleHandler) GetActiveBundles(c echo.Context) error {
	page := 1
	limit := 20

	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	bundles, total, err := h.bundleUC.GetActiveBundles(c.Request().Context(), page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Failed to get bundles"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"items": bundles,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GetBundle returns a bundle by slug
func (h *BundleHandler) GetBundle(c echo.Context) error {
	slug := c.Param("slug")

	bundle, err := h.bundleUC.GetBundleBySlug(c.Request().Context(), slug)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Bundle not found"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    bundle,
	})
}

// CreateBundleRequest represents a bundle creation request
type CreateBundleRequest struct {
	Title           string      `json:"title"`
	Slug            string      `json:"slug"`
	Description     string      `json:"description"`
	ThumbnailURL    string      `json:"thumbnail_url"`
	DiscountPercent float64     `json:"discount_percent"`
	CourseIDs       []uuid.UUID `json:"course_ids"`
	StartDate       *string     `json:"start_date"`
	EndDate         *string     `json:"end_date"`
	MaxPurchases    *int        `json:"max_purchases"`
}

// CreateBundle creates a new bundle (admin)
func (h *BundleHandler) CreateBundle(c echo.Context) error {
	userID := getUserIDFromContext(c)

	var req CreateBundleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	bundle := &domain.Bundle{
		Title:           req.Title,
		Slug:            req.Slug,
		Description:     req.Description,
		ThumbnailURL:    req.ThumbnailURL,
		DiscountPercent: req.DiscountPercent,
		MaxPurchases:    req.MaxPurchases,
		CreatedBy:       userID,
		IsActive:        true,
	}

	created, err := h.bundleUC.CreateBundle(c.Request().Context(), bundle, req.CourseIDs)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    created,
	})
}

// UpdateBundle updates an existing bundle (admin)
func (h *BundleHandler) UpdateBundle(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid bundle ID"},
		})
	}

	var bundle domain.Bundle
	if err := c.Bind(&bundle); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	bundle.ID = id
	if err := h.bundleUC.UpdateBundle(c.Request().Context(), &bundle); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    bundle,
	})
}

// DeleteBundle deletes a bundle (admin)
func (h *BundleHandler) DeleteBundle(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid bundle ID"},
		})
	}

	if err := h.bundleUC.DeleteBundle(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Bundle deleted",
	})
}

// AddCourseRequest represents a course addition request
type AddCourseRequest struct {
	CourseID uuid.UUID `json:"course_id"`
}

// AddCourse adds a course to a bundle (admin)
func (h *BundleHandler) AddCourse(c echo.Context) error {
	bundleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid bundle ID"},
		})
	}

	var req AddCourseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	if err := h.bundleUC.AddCourseToBundle(c.Request().Context(), bundleID, req.CourseID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Course added to bundle",
	})
}

// RemoveCourse removes a course from a bundle (admin)
func (h *BundleHandler) RemoveCourse(c echo.Context) error {
	bundleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid bundle ID"},
		})
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid course ID"},
		})
	}

	if err := h.bundleUC.RemoveCourseFromBundle(c.Request().Context(), bundleID, courseID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Course removed from bundle",
	})
}

// PurchaseBundle purchases a bundle
func (h *BundleHandler) PurchaseBundle(c echo.Context) error {
	bundleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid bundle ID"},
		})
	}

	userID := getUserIDFromContext(c)

	purchase, err := h.bundleUC.PurchaseBundle(c.Request().Context(), userID, bundleID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    purchase,
	})
}

// GetMyBundles returns the current user's bundle purchases
func (h *BundleHandler) GetMyBundles(c echo.Context) error {
	userID := getUserIDFromContext(c)

	purchases, err := h.bundleUC.GetUserBundles(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Failed to get bundles"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    purchases,
	})
}
