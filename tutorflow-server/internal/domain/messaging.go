package domain

import (
	"time"

	"github.com/google/uuid"
)

// Conversation represents a messaging thread between users
type Conversation struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CourseID      *uuid.UUID `gorm:"type:uuid;index" json:"course_id,omitempty"` // Optional: context for conversation
	Participant1  uuid.UUID  `gorm:"type:uuid;not null;index" json:"participant1_id"`
	Participant2  uuid.UUID  `gorm:"type:uuid;not null;index" json:"participant2_id"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Course      *Course   `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User1       *User     `gorm:"foreignKey:Participant1" json:"user1,omitempty"`
	User2       *User     `gorm:"foreignKey:Participant2" json:"user2,omitempty"`
	Messages    []Message `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
	LastMessage *Message  `gorm:"-" json:"last_message,omitempty"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID  `gorm:"type:uuid;index;not null" json:"conversation_id"`
	SenderID       uuid.UUID  `gorm:"type:uuid;not null" json:"sender_id"`
	Content        string     `gorm:"type:text;not null" json:"content"`
	AttachmentURL  *string    `gorm:"type:varchar(500)" json:"attachment_url,omitempty"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// Relationships
	Conversation *Conversation `gorm:"foreignKey:ConversationID" json:"-"`
	Sender       *User         `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}

// IsRead returns true if the message has been read
func (m *Message) IsRead() bool {
	return m.ReadAt != nil
}

// GetOtherParticipant returns the other user in a conversation
func (c *Conversation) GetOtherParticipant(userID uuid.UUID) uuid.UUID {
	if c.Participant1 == userID {
		return c.Participant2
	}
	return c.Participant1
}

// ConversationWithUnread for list view
type ConversationWithUnread struct {
	Conversation
	UnreadCount int `json:"unread_count"`
}
