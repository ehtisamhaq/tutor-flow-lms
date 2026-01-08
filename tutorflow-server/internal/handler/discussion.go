package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/usecase/discussion"
)

// DiscussionHandler handles discussion HTTP requests
type DiscussionHandler struct {
	discussionUC *discussion.UseCase
}

// NewDiscussionHandler creates a new discussion handler
func NewDiscussionHandler(discussionUC *discussion.UseCase) *DiscussionHandler {
	return &DiscussionHandler{discussionUC: discussionUC}
}

// RegisterRoutes registers discussion routes
func (h *DiscussionHandler) RegisterRoutes(g *echo.Group, authMW echo.MiddlewareFunc) {
	discussions := g.Group("/discussions", authMW)
	discussions.GET("/:id", h.GetDiscussion)
	discussions.GET("/course/:courseId", h.GetCourseDiscussions)
	discussions.GET("/lesson/:lessonId", h.GetLessonDiscussions)
	discussions.GET("/:id/replies", h.GetReplies)
	discussions.POST("", h.CreateDiscussion)
	discussions.PUT("/:id", h.UpdateDiscussion)
	discussions.DELETE("/:id", h.DeleteDiscussion)
	discussions.POST("/:id/upvote", h.Upvote)
	discussions.DELETE("/:id/upvote", h.RemoveUpvote)
	discussions.POST("/:id/resolve", h.MarkResolved)
	discussions.DELETE("/:id/resolve", h.Unresolve)
	discussions.POST("/:id/pin", h.Pin)
	discussions.DELETE("/:id/pin", h.Unpin)
}

// GetDiscussion godoc
// @Summary Get discussion with replies
// @Tags Discussions
// @Security BearerAuth
// @Param id path string true "Discussion ID"
// @Success 200 {object} response.Response{data=domain.Discussion}
// @Router /discussions/{id} [get]
func (h *DiscussionHandler) GetDiscussion(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid discussion ID")
	}

	disc, err := h.discussionUC.GetDiscussion(c.Request().Context(), id)
	if err != nil || disc == nil {
		return response.NotFound(c, "Discussion not found")
	}

	return response.Success(c, disc)
}

// GetCourseDiscussions godoc
// @Summary Get discussions for a course
// @Tags Discussions
// @Security BearerAuth
// @Param courseId path string true "Course ID"
// @Success 200 {object} response.Response
// @Router /discussions/course/{courseId} [get]
func (h *DiscussionHandler) GetCourseDiscussions(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	page, limit := 1, 20
	discussions, total, err := h.discussionUC.GetCourseDiscussions(c.Request().Context(), courseID, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get discussions")
	}

	return response.Paginated(c, discussions, page, limit, total)
}

// GetLessonDiscussions godoc
// @Summary Get Q&A for a lesson
// @Tags Discussions
// @Security BearerAuth
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} response.Response
// @Router /discussions/lesson/{lessonId} [get]
func (h *DiscussionHandler) GetLessonDiscussions(c echo.Context) error {
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return response.BadRequest(c, "Invalid lesson ID")
	}

	page, limit := 1, 20
	discussions, total, err := h.discussionUC.GetLessonDiscussions(c.Request().Context(), lessonID, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get discussions")
	}

	return response.Paginated(c, discussions, page, limit, total)
}

// GetReplies godoc
// @Summary Get replies for a discussion
// @Tags Discussions
// @Security BearerAuth
// @Param id path string true "Discussion ID"
// @Success 200 {object} response.Response
// @Router /discussions/{id}/replies [get]
func (h *DiscussionHandler) GetReplies(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid discussion ID")
	}

	page, limit := 1, 50
	replies, total, err := h.discussionUC.GetReplies(c.Request().Context(), id, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get replies")
	}

	return response.Paginated(c, replies, page, limit, total)
}

// CreateDiscussion godoc
// @Summary Create a discussion or reply
// @Tags Discussions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body discussion.CreateDiscussionInput true "Discussion data"
// @Success 201 {object} response.Response{data=domain.Discussion}
// @Router /discussions [post]
func (h *DiscussionHandler) CreateDiscussion(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	var input discussion.CreateDiscussionInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return validator.FormatValidationErrors(err)
	}

	disc, err := h.discussionUC.CreateDiscussion(c.Request().Context(), claims.UserID, input)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, disc)
}

// UpdateDiscussion godoc
// @Summary Update a discussion
// @Tags Discussions
// @Security BearerAuth
// @Accept json
// @Param id path string true "Discussion ID"
// @Param request body discussion.UpdateDiscussionInput true "Content"
// @Success 200 {object} response.Response{data=domain.Discussion}
// @Router /discussions/{id} [put]
func (h *DiscussionHandler) UpdateDiscussion(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid discussion ID")
	}

	claims, _ := middleware.GetClaims(c)

	var input discussion.UpdateDiscussionInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	disc, err := h.discussionUC.UpdateDiscussion(c.Request().Context(), id, claims.UserID, input.Content)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, disc)
}

// DeleteDiscussion godoc
// @Summary Delete a discussion
// @Tags Discussions
// @Security BearerAuth
// @Param id path string true "Discussion ID"
// @Success 204
// @Router /discussions/{id} [delete]
func (h *DiscussionHandler) DeleteDiscussion(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid discussion ID")
	}

	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	if err := h.discussionUC.DeleteDiscussion(c.Request().Context(), id, claims.UserID, isAdmin); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.NoContent(c)
}

// Upvote godoc
// @Summary Upvote a discussion
// @Tags Discussions
// @Security BearerAuth
// @Param id path string true "Discussion ID"
// @Success 200 {object} response.Response
// @Router /discussions/{id}/upvote [post]
func (h *DiscussionHandler) Upvote(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid discussion ID")
	}

	if err := h.discussionUC.Upvote(c.Request().Context(), id); err != nil {
		return response.InternalError(c, "Failed to upvote")
	}

	return response.SuccessWithMessage(c, "Upvoted", nil)
}

// RemoveUpvote godoc
// @Summary Remove upvote from a discussion
// @Tags Discussions
// @Security BearerAuth
// @Param id path string true "Discussion ID"
// @Success 200 {object} response.Response
// @Router /discussions/{id}/upvote [delete]
func (h *DiscussionHandler) RemoveUpvote(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid discussion ID")
	}

	if err := h.discussionUC.RemoveUpvote(c.Request().Context(), id); err != nil {
		return response.InternalError(c, "Failed to remove upvote")
	}

	return response.SuccessWithMessage(c, "Upvote removed", nil)
}

// MarkResolved godoc
// @Summary Mark discussion as resolved
// @Tags Discussions
// @Security BearerAuth
// @Param id path string true "Discussion ID"
// @Success 200 {object} response.Response
// @Router /discussions/{id}/resolve [post]
func (h *DiscussionHandler) MarkResolved(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid discussion ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.discussionUC.MarkResolved(c.Request().Context(), id, claims.UserID); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Marked as resolved", nil)
}

// Unresolve godoc
// @Summary Unmark discussion as resolved
// @Tags Discussions
// @Security BearerAuth
// @Param id path string true "Discussion ID"
// @Success 200 {object} response.Response
// @Router /discussions/{id}/resolve [delete]
func (h *DiscussionHandler) Unresolve(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid discussion ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.discussionUC.Unresolve(c.Request().Context(), id, claims.UserID); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Unmarked as resolved", nil)
}

// Pin godoc
// @Summary Pin a discussion (instructor only)
// @Tags Discussions
// @Security BearerAuth
// @Param id path string true "Discussion ID"
// @Success 200 {object} response.Response
// @Router /discussions/{id}/pin [post]
func (h *DiscussionHandler) Pin(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid discussion ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.discussionUC.Pin(c.Request().Context(), id, claims.UserID); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Discussion pinned", nil)
}

// Unpin godoc
// @Summary Unpin a discussion (instructor only)
// @Tags Discussions
// @Security BearerAuth
// @Param id path string true "Discussion ID"
// @Success 200 {object} response.Response
// @Router /discussions/{id}/pin [delete]
func (h *DiscussionHandler) Unpin(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid discussion ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.discussionUC.Unpin(c.Request().Context(), id, claims.UserID); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Discussion unpinned", nil)
}
