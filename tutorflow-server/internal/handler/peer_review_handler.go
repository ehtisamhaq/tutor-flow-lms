package handler

import (
	"net/http"

	"github.com/tutorflow/tutorflow-server/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// PeerReviewHandler handles peer review HTTP requests
type PeerReviewHandler struct {
	peerReviewUC domain.PeerReviewUseCase
}

// NewPeerReviewHandler creates a new peer review handler
func NewPeerReviewHandler(uc domain.PeerReviewUseCase) *PeerReviewHandler {
	return &PeerReviewHandler{peerReviewUC: uc}
}

// RegisterRoutes registers peer review routes
func (h *PeerReviewHandler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	// Tutor/admin routes (configure peer review)
	config := e.Group("/lessons/:lessonId/peer-review", authMiddleware)
	config.POST("/config", h.ConfigurePeerReview)
	config.GET("/config", h.GetConfig)
	config.POST("/criteria", h.AddCriteria)
	config.PUT("/criteria/:id", h.UpdateCriteria)
	config.DELETE("/criteria/:id", h.DeleteCriteria)

	// Student routes
	review := e.Group("/peer-reviews", authMiddleware)
	review.GET("/my-assignments", h.GetMyAssignments)
	review.GET("/my-submissions/:lessonId", h.GetReviewsForMySubmission)
	review.POST("/:assignmentId/submit", h.SubmitReview)
	review.POST("/:reviewId/dispute", h.DisputeReview)
}

// ConfigureRequest represents peer review configuration
type ConfigureRequest struct {
	ReviewsRequired   int  `json:"reviews_required"`
	ReviewsToComplete int  `json:"reviews_to_complete"`
	DueDays           int  `json:"due_days"`
	IsAnonymous       bool `json:"is_anonymous"`
	ShowScores        bool `json:"show_scores"`
	MinFeedbackLength int  `json:"min_feedback_length"`
}

// ConfigurePeerReview configures peer review for a lesson
func (h *PeerReviewHandler) ConfigurePeerReview(c echo.Context) error {
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid lesson ID"},
		})
	}

	var req ConfigureRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	config := &domain.PeerReviewConfig{
		LessonID:          lessonID,
		ReviewsRequired:   req.ReviewsRequired,
		ReviewsToComplete: req.ReviewsToComplete,
		DueDays:           req.DueDays,
		IsAnonymous:       req.IsAnonymous,
		ShowScores:        req.ShowScores,
		MinFeedbackLength: req.MinFeedbackLength,
	}

	if err := h.peerReviewUC.ConfigurePeerReview(lessonID, config); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    config,
	})
}

// GetConfig returns peer review configuration
func (h *PeerReviewHandler) GetConfig(c echo.Context) error {
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid lesson ID"},
		})
	}

	config, err := h.peerReviewUC.GetPeerReviewConfig(lessonID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Config not found"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    config,
	})
}

// CriteriaRequest represents a criteria creation/update request
type CriteriaRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	MaxScore    float64 `json:"max_score"`
	Weight      float64 `json:"weight"`
	Order       int     `json:"order"`
}

// AddCriteria adds a review criteria
func (h *PeerReviewHandler) AddCriteria(c echo.Context) error {
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid lesson ID"},
		})
	}

	var req CriteriaRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	criteria := &domain.PeerReviewCriteria{
		LessonID:    lessonID,
		Title:       req.Title,
		Description: req.Description,
		MaxScore:    req.MaxScore,
		Weight:      req.Weight,
		Order:       req.Order,
	}

	if err := h.peerReviewUC.AddCriteria(lessonID, criteria); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    criteria,
	})
}

// UpdateCriteria updates a review criteria
func (h *PeerReviewHandler) UpdateCriteria(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid criteria ID"},
		})
	}

	var req CriteriaRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	criteria := &domain.PeerReviewCriteria{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		MaxScore:    req.MaxScore,
		Weight:      req.Weight,
		Order:       req.Order,
	}

	if err := h.peerReviewUC.UpdateCriteria(criteria); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    criteria,
	})
}

// DeleteCriteria removes a review criteria
func (h *PeerReviewHandler) DeleteCriteria(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid criteria ID"},
		})
	}

	if err := h.peerReviewUC.RemoveCriteria(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Criteria deleted",
	})
}

// GetMyAssignments returns user's review assignments
func (h *PeerReviewHandler) GetMyAssignments(c echo.Context) error {
	userID := getUserIDFromContext(c)

	assignments, err := h.peerReviewUC.GetMyReviewAssignments(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Failed to get assignments"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    assignments,
	})
}

// GetReviewsForMySubmission returns reviews received for user's submission
func (h *PeerReviewHandler) GetReviewsForMySubmission(c echo.Context) error {
	userID := getUserIDFromContext(c)
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid lesson ID"},
		})
	}

	reviews, err := h.peerReviewUC.GetReviewsForMySubmission(userID, lessonID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Failed to get reviews"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    reviews,
	})
}

// SubmitReviewRequest represents a review submission
type SubmitReviewRequest struct {
	Feedback     string `json:"feedback"`
	Strengths    string `json:"strengths"`
	Improvements string `json:"improvements"`
	Scores       []struct {
		CriteriaID uuid.UUID `json:"criteria_id"`
		Score      float64   `json:"score"`
		Comment    string    `json:"comment"`
	} `json:"scores"`
}

// SubmitReview submits a peer review
func (h *PeerReviewHandler) SubmitReview(c echo.Context) error {
	assignmentID, err := uuid.Parse(c.Param("assignmentId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid assignment ID"},
		})
	}

	var req SubmitReviewRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	review := &domain.PeerReview{
		AssignmentID: assignmentID,
		Feedback:     req.Feedback,
		Strengths:    req.Strengths,
		Improvements: req.Improvements,
	}

	scores := make([]domain.PeerReviewScore, len(req.Scores))
	for i, s := range req.Scores {
		scores[i] = domain.PeerReviewScore{
			CriteriaID: s.CriteriaID,
			Score:      s.Score,
			Comment:    s.Comment,
		}
	}

	if err := h.peerReviewUC.SubmitReview(assignmentID, review, scores); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    review,
	})
}

// DisputeRequest represents a dispute request
type DisputeRequest struct {
	Reason string `json:"reason"`
}

// DisputeReview disputes a peer review
func (h *PeerReviewHandler) DisputeReview(c echo.Context) error {
	reviewID, err := uuid.Parse(c.Param("reviewId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid review ID"},
		})
	}

	var req DisputeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	if err := h.peerReviewUC.DisputeReview(reviewID, req.Reason); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Review disputed",
	})
}
