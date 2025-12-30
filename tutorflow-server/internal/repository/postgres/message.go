package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// MessageRepository
type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) repository.MessageRepository {
	return &messageRepository{db: db}
}

// Conversations

func (r *messageRepository) CreateConversation(ctx context.Context, conv *domain.Conversation) error {
	return r.db.WithContext(ctx).Create(conv).Error
}

func (r *messageRepository) GetConversationByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	var conv domain.Conversation
	err := r.db.WithContext(ctx).
		Preload("User1").
		Preload("User2").
		Preload("Course").
		Where("id = ?", id).
		First(&conv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

func (r *messageRepository) GetConversationBetween(ctx context.Context, user1, user2 uuid.UUID) (*domain.Conversation, error) {
	var conv domain.Conversation
	err := r.db.WithContext(ctx).
		Preload("User1").
		Preload("User2").
		Where("(participant1 = ? AND participant2 = ?) OR (participant1 = ? AND participant2 = ?)",
			user1, user2, user2, user1).
		First(&conv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

func (r *messageRepository) GetUserConversations(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.ConversationWithUnread, int64, error) {
	var total int64

	// Count total conversations
	r.db.WithContext(ctx).Model(&domain.Conversation{}).
		Where("participant1 = ? OR participant2 = ?", userID, userID).
		Count(&total)

	// Get conversations with unread count
	offset := (page - 1) * limit
	var conversations []domain.Conversation

	err := r.db.WithContext(ctx).
		Preload("User1").
		Preload("User2").
		Preload("Course").
		Where("participant1 = ? OR participant2 = ?", userID, userID).
		Order("last_message_at DESC NULLS LAST").
		Offset(offset).
		Limit(limit).
		Find(&conversations).Error

	if err != nil {
		return nil, 0, err
	}

	// Build result with unread counts
	result := make([]domain.ConversationWithUnread, len(conversations))
	for i, conv := range conversations {
		var unread int64
		r.db.WithContext(ctx).Model(&domain.Message{}).
			Where("conversation_id = ? AND sender_id != ? AND read_at IS NULL", conv.ID, userID).
			Count(&unread)

		// Get last message
		var lastMsg domain.Message
		if err := r.db.WithContext(ctx).
			Preload("Sender").
			Where("conversation_id = ?", conv.ID).
			Order("created_at DESC").
			First(&lastMsg).Error; err == nil {
			conv.LastMessage = &lastMsg
		}

		result[i] = domain.ConversationWithUnread{
			Conversation: conv,
			UnreadCount:  int(unread),
		}
	}

	return result, total, nil
}

func (r *messageRepository) UpdateConversation(ctx context.Context, conv *domain.Conversation) error {
	return r.db.WithContext(ctx).Save(conv).Error
}

// Messages

func (r *messageRepository) CreateMessage(ctx context.Context, msg *domain.Message) error {
	if err := r.db.WithContext(ctx).Create(msg).Error; err != nil {
		return err
	}

	// Update conversation's last_message_at
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Conversation{}).
		Where("id = ?", msg.ConversationID).
		Update("last_message_at", now).Error
}

func (r *messageRepository) GetMessageByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	var msg domain.Message
	err := r.db.WithContext(ctx).
		Preload("Sender").
		Where("id = ?", id).
		First(&msg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepository) GetConversationMessages(ctx context.Context, convID uuid.UUID, page, limit int) ([]domain.Message, int64, error) {
	var messages []domain.Message
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Message{}).Where("conversation_id = ?", convID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Sender").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error

	return messages, total, err
}

func (r *messageRepository) MarkAsRead(ctx context.Context, msgID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Message{}).
		Where("id = ?", msgID).
		Update("read_at", now).Error
}

func (r *messageRepository) MarkConversationAsRead(ctx context.Context, convID, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Message{}).
		Where("conversation_id = ? AND sender_id != ? AND read_at IS NULL", convID, userID).
		Update("read_at", now).Error
}

func (r *messageRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Message{}).
		Joins("JOIN conversations ON conversations.id = messages.conversation_id").
		Where("(conversations.participant1 = ? OR conversations.participant2 = ?) AND messages.sender_id != ? AND messages.read_at IS NULL",
			userID, userID, userID).
		Count(&count).Error
	return count, err
}

func (r *messageRepository) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Message{}, "id = ?", id).Error
}
