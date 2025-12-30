package handler

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/usecase/message"
)

// MessageHandler handles messaging HTTP requests
type MessageHandler struct {
	messageUC *message.UseCase
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageUC *message.UseCase) *MessageHandler {
	return &MessageHandler{messageUC: messageUC}
}

// RegisterRoutes registers messaging routes
func (h *MessageHandler) RegisterRoutes(g *echo.Group, authMW echo.MiddlewareFunc) {
	msgs := g.Group("/messages", authMW)
	msgs.GET("/conversations", h.GetConversations)
	msgs.GET("/conversations/:id", h.GetConversation)
	msgs.GET("/conversations/:id/messages", h.GetMessages)
	msgs.POST("/conversations/:userId", h.StartConversation)
	msgs.POST("", h.SendMessage)
	msgs.POST("/:id/read", h.MarkAsRead)
	msgs.POST("/conversations/:id/read", h.MarkConversationAsRead)
	msgs.GET("/unread-count", h.GetUnreadCount)
	msgs.DELETE("/:id", h.DeleteMessage)
}

// GetConversations godoc
// @Summary Get my conversations
// @Tags Messages
// @Security BearerAuth
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {object} response.Response
// @Router /messages/conversations [get]
func (h *MessageHandler) GetConversations(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	page, limit := 1, 20
	if p := c.QueryParam("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			page = val
		}
	}
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	conversations, total, err := h.messageUC.GetConversations(c.Request().Context(), claims.UserID, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get conversations")
	}

	return response.Paginated(c, conversations, page, limit, total)
}

// GetConversation godoc
// @Summary Get a conversation
// @Tags Messages
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Success 200 {object} response.Response{data=domain.Conversation}
// @Router /messages/conversations/{id} [get]
func (h *MessageHandler) GetConversation(c echo.Context) error {
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid conversation ID")
	}

	claims, _ := middleware.GetClaims(c)

	conv, err := h.messageUC.GetConversation(c.Request().Context(), claims.UserID, convID)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, conv)
}

// GetMessages godoc
// @Summary Get messages in a conversation
// @Tags Messages
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {object} response.Response
// @Router /messages/conversations/{id}/messages [get]
func (h *MessageHandler) GetMessages(c echo.Context) error {
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid conversation ID")
	}

	claims, _ := middleware.GetClaims(c)

	page, limit := 1, 50
	if p := c.QueryParam("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			page = val
		}
	}
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	messages, total, err := h.messageUC.GetMessages(c.Request().Context(), claims.UserID, convID, page, limit)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Paginated(c, messages, page, limit, total)
}

// StartConversation godoc
// @Summary Start a conversation with a user
// @Tags Messages
// @Security BearerAuth
// @Param userId path string true "User ID to message"
// @Success 200 {object} response.Response{data=domain.Conversation}
// @Router /messages/conversations/{userId} [post]
func (h *MessageHandler) StartConversation(c echo.Context) error {
	otherUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return response.BadRequest(c, "Invalid user ID")
	}

	claims, _ := middleware.GetClaims(c)

	conv, err := h.messageUC.GetOrCreateConversation(c.Request().Context(), claims.UserID, otherUserID)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, conv)
}

// SendMessage godoc
// @Summary Send a message
// @Tags Messages
// @Security BearerAuth
// @Accept json
// @Param request body message.SendMessageInput true "Message data"
// @Success 201 {object} response.Response{data=domain.Message}
// @Router /messages [post]
func (h *MessageHandler) SendMessage(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	var input message.SendMessageInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	msg, err := h.messageUC.SendMessage(c.Request().Context(), claims.UserID, input)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, msg)
}

// MarkAsRead godoc
// @Summary Mark a message as read
// @Tags Messages
// @Security BearerAuth
// @Param id path string true "Message ID"
// @Success 200 {object} response.Response
// @Router /messages/{id}/read [post]
func (h *MessageHandler) MarkAsRead(c echo.Context) error {
	msgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid message ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.messageUC.MarkAsRead(c.Request().Context(), claims.UserID, msgID); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Message marked as read", nil)
}

// MarkConversationAsRead godoc
// @Summary Mark all messages in a conversation as read
// @Tags Messages
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Success 200 {object} response.Response
// @Router /messages/conversations/{id}/read [post]
func (h *MessageHandler) MarkConversationAsRead(c echo.Context) error {
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid conversation ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.messageUC.MarkConversationAsRead(c.Request().Context(), claims.UserID, convID); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Conversation marked as read", nil)
}

// GetUnreadCount godoc
// @Summary Get total unread message count
// @Tags Messages
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /messages/unread-count [get]
func (h *MessageHandler) GetUnreadCount(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	count, err := h.messageUC.GetUnreadCount(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.InternalError(c, "Failed to get unread count")
	}

	return response.Success(c, map[string]int64{"unread_count": count})
}

// DeleteMessage godoc
// @Summary Delete a message
// @Tags Messages
// @Security BearerAuth
// @Param id path string true "Message ID"
// @Success 204
// @Router /messages/{id} [delete]
func (h *MessageHandler) DeleteMessage(c echo.Context) error {
	msgID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid message ID")
	}

	claims, _ := middleware.GetClaims(c)

	if err := h.messageUC.DeleteMessage(c.Request().Context(), claims.UserID, msgID); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.NoContent(c)
}
