package domain

import (
	"time"

	"github.com/google/uuid"
)

// PushSubscription represents a user's web push subscription
type PushSubscription struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Endpoint  string    `gorm:"type:text;not null" json:"endpoint"`
	P256dh    string    `gorm:"type:text;not null" json:"p256dh"` // Public key
	Auth      string    `gorm:"type:text;not null" json:"auth"`   // Auth secret
	UserAgent string    `gorm:"type:varchar(500)" json:"user_agent,omitempty"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}

// PushNotification represents a push notification to be sent
type PushNotification struct {
	Title   string                 `json:"title"`
	Body    string                 `json:"body"`
	Icon    string                 `json:"icon,omitempty"`
	Badge   string                 `json:"badge,omitempty"`
	Tag     string                 `json:"tag,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	URL     string                 `json:"url,omitempty"`
	Actions []NotificationAction   `json:"actions,omitempty"`
}

// NotificationAction for interactive notifications
type NotificationAction struct {
	Action string `json:"action"`
	Title  string `json:"title"`
	Icon   string `json:"icon,omitempty"`
}
