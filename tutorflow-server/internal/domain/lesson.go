package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// LessonType enum
type LessonType string

const (
	LessonTypeVideo      LessonType = "video"
	LessonTypeText       LessonType = "text"
	LessonTypeQuiz       LessonType = "quiz"
	LessonTypeAssignment LessonType = "assignment"
	LessonTypeResource   LessonType = "resource"
)

// ContentAccess enum
type ContentAccess string

const (
	ContentAccessFree     ContentAccess = "free"
	ContentAccessEnrolled ContentAccess = "enrolled"
	ContentAccessPremium  ContentAccess = "premium"
)

// Lesson represents a lesson within a module
type Lesson struct {
	ID            uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ModuleID      uuid.UUID     `gorm:"type:uuid;index;not null" json:"module_id"`
	Title         string        `gorm:"type:varchar(255);not null" json:"title"`
	Description   *string       `gorm:"type:text" json:"description,omitempty"`
	Content       *string       `gorm:"type:text" json:"content,omitempty"`
	LessonType    LessonType    `gorm:"type:lesson_type;not null;default:'text'" json:"lesson_type"`
	AccessType    ContentAccess `gorm:"type:content_access;not null;default:'enrolled'" json:"access_type"`
	VideoURL      *string       `gorm:"type:varchar(500)" json:"video_url,omitempty"`
	VideoDuration *int          `json:"video_duration,omitempty"` // seconds
	Attachments   *string       `gorm:"type:jsonb;default:'[]'" json:"attachments,omitempty"`
	SortOrder     int           `gorm:"not null;default:0" json:"sort_order"`
	IsPublished   bool          `gorm:"default:false" json:"is_published"`
	IsPreview     bool          `gorm:"default:false" json:"is_preview"`
	CreatedAt     time.Time     `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Module      *Module      `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	VideoAssets []VideoAsset `gorm:"foreignKey:LessonID" json:"video_assets,omitempty"`
	Quiz        *Quiz        `gorm:"foreignKey:LessonID" json:"quiz,omitempty"`
	Assignment  *Assignment  `gorm:"foreignKey:LessonID" json:"assignment,omitempty"`
}

func (l *Lesson) IsFreeAccess() bool {
	return l.AccessType == ContentAccessFree || l.IsPreview
}

// VideoAsset represents encrypted video files for DRM
type VideoAsset struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LessonID         uuid.UUID `gorm:"type:uuid;index;not null" json:"lesson_id"`
	OriginalFilename *string   `gorm:"type:varchar(255)" json:"original_filename,omitempty"`
	StoragePath      string    `gorm:"type:varchar(500);not null" json:"storage_path"`
	EncryptionKeyID  *string   `gorm:"type:varchar(100)" json:"encryption_key_id,omitempty"`
	DRMPolicyID      *string   `gorm:"type:varchar(100)" json:"drm_policy_id,omitempty"`
	Resolution       *string   `gorm:"type:varchar(20)" json:"resolution,omitempty"`
	Bitrate          *int      `json:"bitrate,omitempty"`
	Duration         *int      `json:"duration,omitempty"` // seconds
	FileSize         *int64    `json:"file_size,omitempty"`
	Status           string    `gorm:"type:varchar(20);default:'processing'" json:"status"`
	HLSPlaylistURL   *string   `gorm:"type:varchar(500)" json:"hls_playlist_url,omitempty"`
	CreatedAt        time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Lesson *Lesson `gorm:"foreignKey:LessonID" json:"-"`
}

// PlaybackSession for concurrent stream limits
type PlaybackSession struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	LessonID      uuid.UUID  `gorm:"type:uuid;not null" json:"lesson_id"`
	DeviceID      *uuid.UUID `gorm:"type:uuid" json:"device_id,omitempty"`
	SessionToken  string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"session_token"`
	StartedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"started_at"`
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	EndedAt       *time.Time `json:"ended_at,omitempty"`

	User   *User       `gorm:"foreignKey:UserID" json:"-"`
	Device *UserDevice `gorm:"foreignKey:DeviceID" json:"-"`
}

func (p *PlaybackSession) IsActive() bool {
	return p.EndedAt == nil
}

// CourseNote for student notes
type CourseNote struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	LessonID       uuid.UUID `gorm:"type:uuid;index;not null" json:"lesson_id"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	VideoTimestamp *int      `json:"video_timestamp,omitempty"` // seconds
	CreatedAt      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	User   *User   `gorm:"foreignKey:UserID" json:"-"`
	Lesson *Lesson `gorm:"foreignKey:LessonID" json:"-"`
}

// VideoBookmark for video bookmarks
type VideoBookmark struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	LessonID  uuid.UUID `gorm:"type:uuid;index;not null" json:"lesson_id"`
	Timestamp int       `gorm:"not null" json:"timestamp"` // seconds
	Label     *string   `gorm:"type:varchar(100)" json:"label,omitempty"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	User   *User   `gorm:"foreignKey:UserID" json:"-"`
	Lesson *Lesson `gorm:"foreignKey:LessonID" json:"-"`
}

// Attachment represents file attachments
type Attachment struct {
	FileName    string `json:"file_name"`
	FileURL     string `json:"file_url"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
}

// Helper function to parse attachments JSON
func (l *Lesson) GetAttachments() []Attachment {
	// TODO: Implement JSON parsing
	return nil
}

// ResourceInfo for downloadable resources
type ResourceInfo struct {
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Files       pq.StringArray `json:"files"`
}
