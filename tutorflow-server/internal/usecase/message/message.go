package message

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines messaging business logic
type UseCase struct {
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
}

// NewUseCase creates a new message use case
func NewUseCase(
	messageRepo repository.MessageRepository,
	userRepo repository.UserRepository,
) *UseCase {
	return &UseCase{
		messageRepo: messageRepo,
		userRepo:    userRepo,
	}
}

// GetConversations returns user's conversations
func (uc *UseCase) GetConversations(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.ConversationWithUnread, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.messageRepo.GetUserConversations(ctx, userID, page, limit)
}

// GetConversation returns a conversation by ID
func (uc *UseCase) GetConversation(ctx context.Context, userID, convID uuid.UUID) (*domain.Conversation, error) {
	conv, err := uc.messageRepo.GetConversationByID(ctx, convID)
	if err != nil || conv == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	// Verify user is participant
	if conv.Participant1 != userID && conv.Participant2 != userID {
		return nil, fmt.Errorf("access denied")
	}

	return conv, nil
}

// GetOrCreateConversation gets or creates a conversation with another user
func (uc *UseCase) GetOrCreateConversation(ctx context.Context, userID, otherUserID uuid.UUID) (*domain.Conversation, error) {
	if userID == otherUserID {
		return nil, fmt.Errorf("cannot message yourself")
	}

	// Check other user exists
	otherUser, err := uc.userRepo.GetByID(ctx, otherUserID)
	if err != nil || otherUser == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Try to find existing conversation
	conv, _ := uc.messageRepo.GetConversationBetween(ctx, userID, otherUserID)
	if conv != nil {
		return conv, nil
	}

	// Create new conversation
	conv = &domain.Conversation{
		Participant1: userID,
		Participant2: otherUserID,
	}

	if err := uc.messageRepo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}

	return uc.messageRepo.GetConversationByID(ctx, conv.ID)
}

// GetMessages returns messages in a conversation
func (uc *UseCase) GetMessages(ctx context.Context, userID, convID uuid.UUID, page, limit int) ([]domain.Message, int64, error) {
	// Verify access
	conv, err := uc.messageRepo.GetConversationByID(ctx, convID)
	if err != nil || conv == nil {
		return nil, 0, fmt.Errorf("conversation not found")
	}

	if conv.Participant1 != userID && conv.Participant2 != userID {
		return nil, 0, fmt.Errorf("access denied")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	return uc.messageRepo.GetConversationMessages(ctx, convID, page, limit)
}

// SendMessageInput for sending a message
type SendMessageInput struct {
	ConversationID *uuid.UUID `json:"conversation_id,omitempty"`
	RecipientID    *uuid.UUID `json:"recipient_id,omitempty"`
	Content        string     `json:"content" validate:"required,min=1,max=5000"`
	AttachmentURL  *string    `json:"attachment_url,omitempty"`
}

// SendMessage sends a message
func (uc *UseCase) SendMessage(ctx context.Context, senderID uuid.UUID, input SendMessageInput) (*domain.Message, error) {
	var convID uuid.UUID

	if input.ConversationID != nil {
		// Verify access to conversation
		conv, err := uc.messageRepo.GetConversationByID(ctx, *input.ConversationID)
		if err != nil || conv == nil {
			return nil, fmt.Errorf("conversation not found")
		}
		if conv.Participant1 != senderID && conv.Participant2 != senderID {
			return nil, fmt.Errorf("access denied")
		}
		convID = conv.ID
	} else if input.RecipientID != nil {
		// Get or create conversation
		conv, err := uc.GetOrCreateConversation(ctx, senderID, *input.RecipientID)
		if err != nil {
			return nil, err
		}
		convID = conv.ID
	} else {
		return nil, fmt.Errorf("conversation_id or recipient_id required")
	}

	msg := &domain.Message{
		ConversationID: convID,
		SenderID:       senderID,
		Content:        input.Content,
		AttachmentURL:  input.AttachmentURL,
	}

	if err := uc.messageRepo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	return uc.messageRepo.GetMessageByID(ctx, msg.ID)
}

// MarkAsRead marks a message as read
func (uc *UseCase) MarkAsRead(ctx context.Context, userID, msgID uuid.UUID) error {
	msg, err := uc.messageRepo.GetMessageByID(ctx, msgID)
	if err != nil || msg == nil {
		return fmt.Errorf("message not found")
	}

	// Verify user is recipient (not sender)
	conv, _ := uc.messageRepo.GetConversationByID(ctx, msg.ConversationID)
	if conv == nil || (conv.Participant1 != userID && conv.Participant2 != userID) {
		return fmt.Errorf("access denied")
	}

	if msg.SenderID == userID {
		return nil // Can't mark own message as read
	}

	return uc.messageRepo.MarkAsRead(ctx, msgID)
}

// MarkConversationAsRead marks all messages in a conversation as read
func (uc *UseCase) MarkConversationAsRead(ctx context.Context, userID, convID uuid.UUID) error {
	conv, err := uc.messageRepo.GetConversationByID(ctx, convID)
	if err != nil || conv == nil {
		return fmt.Errorf("conversation not found")
	}

	if conv.Participant1 != userID && conv.Participant2 != userID {
		return fmt.Errorf("access denied")
	}

	return uc.messageRepo.MarkConversationAsRead(ctx, convID, userID)
}

// GetUnreadCount returns total unread messages for a user
func (uc *UseCase) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return uc.messageRepo.GetUnreadCount(ctx, userID)
}

// DeleteMessage deletes a message
func (uc *UseCase) DeleteMessage(ctx context.Context, userID, msgID uuid.UUID) error {
	msg, err := uc.messageRepo.GetMessageByID(ctx, msgID)
	if err != nil || msg == nil {
		return fmt.Errorf("message not found")
	}

	if msg.SenderID != userID {
		return fmt.Errorf("you can only delete your own messages")
	}

	return uc.messageRepo.DeleteMessage(ctx, msgID)
}
