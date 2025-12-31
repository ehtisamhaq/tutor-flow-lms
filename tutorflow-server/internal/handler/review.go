package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/usecase/review"
)

// ReviewHandler handles review HTTP requests
type ReviewHandler struct {
	reviewUC *review.UseCase
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(reviewUC *review.UseCase) *ReviewHandler {
	return &ReviewHandler{reviewUC: reviewUC}
}

// RegisterRoutes registers review routes
func (h *ReviewHandler) RegisterRoutes(g *echo.Group, authMW, tutorMW echo.MiddlewareFunc) {
	reviews := g.Group("/reviews")
	reviews.GET("/course/:courseId", h.GetCourseReviews)
	reviews.GET("/course/:courseId/summary", h.GetRatingSummary)
	reviews.GET("/:id", h.GetReview)
	reviews.GET("/course/:courseId/my", h.GetMyReview, authMW)
	reviews.POST("", h.CreateReview, authMW)
	reviews.PUT("/:id", h.UpdateReview, authMW)
	reviews.DELETE("/:id", h.DeleteReview, authMW)
	reviews.POST("/:id/vote", h.VoteReview, authMW)
	reviews.POST("/:id/reply", h.ReplyToReview, authMW, tutorMW)
	reviews.PATCH("/:id/feature", h.FeatureReview, authMW, tutorMW)
}

// GetCourseReviews godoc
// @Summary Get reviews for a course
// @Tags Reviews
// @Param courseId path string true "Course ID"
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {object} response.Response
// @Router /reviews/course/{courseId} [get]
func (h *ReviewHandler) GetCourseReviews(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	page := 1
	limit := 10
	// Parse query params if needed

	reviews, total, err := h.reviewUC.GetCourseReviews(c.Request().Context(), courseID, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get reviews")
	}

	return response.Paginated(c, reviews, page, limit, total)
}

// GetRatingSummary godoc
// @Summary Get rating summary for a course
// @Tags Reviews
// @Param courseId path string true "Course ID"
// @Success 200 {object} response.Response{data=review.RatingSummary}
// @Router /reviews/course/{courseId}/summary [get]
func (h *ReviewHandler) GetRatingSummary(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	summary, err := h.reviewUC.GetCourseRatingSummary(c.Request().Context(), courseID)
	if err != nil {
		return response.InternalError(c, "Failed to get rating summary")
	}

	return response.Success(c, summary)
}

// GetReview godoc
// @Summary Get review by ID
// @Tags Reviews
// @Param id path string true "Review ID"
// @Success 200 {object} response.Response{data=domain.CourseReview}
// @Router /reviews/{id} [get]
func (h *ReviewHandler) GetReview(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid review ID")
	}

	reviewObj, err := h.reviewUC.GetReview(c.Request().Context(), id)
	if err != nil {
		return response.NotFound(c, "Review not found")
	}

	return response.Success(c, reviewObj)
}

// GetMyReview godoc
// @Summary Get my review for a course
// @Tags Reviews
// @Security BearerAuth
// @Param courseId path string true "Course ID"
// @Success 200 {object} response.Response{data=domain.CourseReview}
// @Router /reviews/course/{courseId}/my [get]
func (h *ReviewHandler) GetMyReview(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)

	reviewObj, err := h.reviewUC.GetUserReview(c.Request().Context(), claims.UserID, courseID)
	if err != nil || reviewObj == nil {
		return response.Success(c, nil)
	}

	return response.Success(c, reviewObj)
}

// CreateReview godoc
// @Summary Create a review
// @Tags Reviews
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body review.CreateReviewInput true "Review data"
// @Success 201 {object} response.Response{data=domain.CourseReview}
// @Router /reviews [post]
func (h *ReviewHandler) CreateReview(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	var input review.CreateReviewInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	reviewObj, err := h.reviewUC.CreateReview(c.Request().Context(), claims.UserID, input)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, reviewObj)
}

// UpdateReview godoc
// @Summary Update a review
// @Tags Reviews
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Review ID"
// @Param request body review.UpdateReviewInput true "Review data"
// @Success 200 {object} response.Response{data=domain.CourseReview}
// @Router /reviews/{id} [put]
func (h *ReviewHandler) UpdateReview(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid review ID")
	}

	claims, _ := middleware.GetClaims(c)

	var input review.UpdateReviewInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	reviewObj, err := h.reviewUC.UpdateReview(c.Request().Context(), id, claims.UserID, input)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, reviewObj)
}

// DeleteReview godoc
// @Summary Delete a review
// @Tags Reviews
// @Security BearerAuth
// @Param id path string true "Review ID"
// @Success 204
// @Router /reviews/{id} [delete]
func (h *ReviewHandler) DeleteReview(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid review ID")
	}

	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	if err := h.reviewUC.DeleteReview(c.Request().Context(), id, claims.UserID, isAdmin); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.NoContent(c)
}

// VoteReviewInput for voting on a review
type VoteReviewInput struct {
	IsHelpful bool `json:"is_helpful"`
}

// VoteReview godoc
// @Summary Vote on a review
// @Tags Reviews
// @Security BearerAuth
// @Accept json
// @Param id path string true "Review ID"
// @Param request body VoteReviewInput true "Vote"
// @Success 200 {object} response.Response
// @Router /reviews/{id}/vote [post]
func (h *ReviewHandler) VoteReview(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid review ID")
	}

	claims, _ := middleware.GetClaims(c)

	var input VoteReviewInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.reviewUC.VoteReview(c.Request().Context(), id, claims.UserID, input.IsHelpful); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Vote recorded", nil)
}

// ReplyToReview godoc
// @Summary Reply to a review (instructor)
// @Tags Reviews
// @Security BearerAuth
// @Accept json
// @Param id path string true "Review ID"
// @Param request body review.ReplyToReviewInput true "Reply"
// @Success 200 {object} response.Response{data=domain.CourseReview}
// @Router /reviews/{id}/reply [post]
func (h *ReviewHandler) ReplyToReview(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid review ID")
	}

	claims, _ := middleware.GetClaims(c)

	var input review.ReplyToReviewInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	reviewObj, err := h.reviewUC.ReplyToReview(c.Request().Context(), id, claims.UserID, input.Reply)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, reviewObj)
}

// FeatureReviewInput for featuring a review
type FeatureReviewInput struct {
	Featured bool `json:"featured"`
}

// FeatureReview godoc
// @Summary Feature/unfeature a review
// @Tags Reviews
// @Security BearerAuth
// @Accept json
// @Param id path string true "Review ID"
// @Param request body FeatureReviewInput true "Featured status"
// @Success 200 {object} response.Response
// @Router /reviews/{id}/feature [patch]
func (h *ReviewHandler) FeatureReview(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid review ID")
	}

	var input FeatureReviewInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.reviewUC.FeatureReview(c.Request().Context(), id, input.Featured); err != nil {
		return response.InternalError(c, "Failed to update review")
	}

	return response.SuccessWithMessage(c, "Review updated", nil)
}
