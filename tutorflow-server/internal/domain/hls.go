package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// HLSEncryptionType defines encryption types
type HLSEncryptionType string

const (
	HLSEncryptionNone      HLSEncryptionType = "none"
	HLSEncryptionAES128    HLSEncryptionType = "aes-128"
	HLSEncryptionSampleAES HLSEncryptionType = "sample-aes"
)

// VideoProcessingStatus defines video processing states
type VideoProcessingStatus string

const (
	VideoStatusPending    VideoProcessingStatus = "pending"
	VideoStatusProcessing VideoProcessingStatus = "processing"
	VideoStatusCompleted  VideoProcessingStatus = "completed"
	VideoStatusFailed     VideoProcessingStatus = "failed"
)

// VideoAsset represents a video file with HLS encoding
type HLSVideoAsset struct {
	ID              uuid.UUID             `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LessonID        uuid.UUID             `gorm:"type:uuid;index;not null" json:"lesson_id"`
	OriginalURL     string                `gorm:"size:500;not null" json:"-"`
	Duration        int                   `gorm:"" json:"duration"` // seconds
	FileSize        int64                 `gorm:"" json:"file_size"`
	Resolution      string                `gorm:"size:20" json:"resolution"` // e.g., "1920x1080"
	Status          VideoProcessingStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	ProcessingError string                `gorm:"type:text" json:"processing_error,omitempty"`
	CreatedAt       time.Time             `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time             `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Lesson        *Lesson          `gorm:"foreignKey:LessonID" json:"-"`
	Qualities     []VideoQuality   `gorm:"foreignKey:VideoID" json:"qualities,omitempty"`
	HLSEncryption *VideoEncryption `gorm:"foreignKey:VideoID" json:"-"`
}

// VideoQuality represents an HLS quality variant
type VideoQuality struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID     uuid.UUID `gorm:"type:uuid;index;not null" json:"video_id"`
	Quality     string    `gorm:"size:20;not null" json:"quality"` // e.g., "1080p", "720p", "480p", "360p"
	Bitrate     int       `gorm:"not null" json:"bitrate"`         // kbps
	Resolution  string    `gorm:"size:20" json:"resolution"`       // e.g., "1920x1080"
	PlaylistURL string    `gorm:"size:500;not null" json:"-"`
	SegmentDir  string    `gorm:"size:500" json:"-"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Video *HLSVideoAsset `gorm:"foreignKey:VideoID" json:"-"`
}

// VideoEncryption stores HLS encryption keys
type VideoEncryption struct {
	ID             uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID        uuid.UUID         `gorm:"type:uuid;uniqueIndex;not null" json:"video_id"`
	EncryptionType HLSEncryptionType `gorm:"size:20;not null" json:"encryption_type"`
	KeyID          string            `gorm:"size:64" json:"-"`   // Key identifier
	EncryptionKey  string            `gorm:"size:64" json:"-"`   // AES key (encrypted at rest)
	IV             string            `gorm:"size:32" json:"-"`   // Initialization vector
	KeyURL         string            `gorm:"size:500" json:"-"`  // URL to fetch key
	RotationPeriod int               `gorm:"default:0" json:"-"` // Rotate key every N seconds (0 = no rotation)
	CreatedAt      time.Time         `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time         `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Video *HLSVideoAsset `gorm:"foreignKey:VideoID" json:"-"`
}

// GenerateEncryptionKey generates a new AES-128 key
func GenerateEncryptionKey() (key, iv string, err error) {
	keyBytes := make([]byte, 16) // 128-bit key
	if _, err := rand.Read(keyBytes); err != nil {
		return "", "", err
	}

	ivBytes := make([]byte, 16) // 128-bit IV
	if _, err := rand.Read(ivBytes); err != nil {
		return "", "", err
	}

	return hex.EncodeToString(keyBytes), hex.EncodeToString(ivBytes), nil
}

// SignedURL represents a time-limited video access URL
type SignedURL struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VideoID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"video_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	SessionID string     `gorm:"size:100;not null" json:"-"`
	DeviceID  string     `gorm:"size:100" json:"-"`
	URL       string     `gorm:"size:1000;not null" json:"-"`
	Token     string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"" json:"-"`
	CreatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Video *HLSVideoAsset `gorm:"foreignKey:VideoID" json:"-"`
	User  *User          `gorm:"foreignKey:UserID" json:"-"`
}

// IsValid checks if the signed URL is still valid
func (s *SignedURL) IsValid() bool {
	return time.Now().Before(s.ExpiresAt)
}

// DeviceSession tracks active viewing sessions
type DeviceSession struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	DeviceID   string    `gorm:"size:100;index;not null" json:"device_id"`
	DeviceName string    `gorm:"size:200" json:"device_name"`
	DeviceType string    `gorm:"size:50" json:"device_type"` // desktop, mobile, tablet
	Browser    string    `gorm:"size:100" json:"browser"`
	OS         string    `gorm:"size:100" json:"os"`
	IP         string    `gorm:"size:45" json:"-"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	LastSeenAt time.Time `gorm:"not null" json:"last_seen_at"`
	CreatedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}

// VideoUseCase interface
type VideoUseCase interface {
	// Upload & Processing
	UploadVideo(ctx context.Context, lessonID uuid.UUID, fileURL string) (*HLSVideoAsset, error)
	ProcessVideo(ctx context.Context, videoID uuid.UUID) error
	GetProcessingStatus(ctx context.Context, videoID uuid.UUID) (*HLSVideoAsset, error)

	// Playback
	GetPlaybackURL(ctx context.Context, lessonID, userID uuid.UUID, deviceID string) (string, error)
	GetEncryptionKey(ctx context.Context, token string) ([]byte, error)
	ValidatePlayback(ctx context.Context, token string) error

	// DRM
	EnableEncryption(ctx context.Context, videoID uuid.UUID, encType HLSEncryptionType) error
	RotateEncryptionKey(ctx context.Context, videoID uuid.UUID) error

	// Device Management
	RegisterDevice(ctx context.Context, userID uuid.UUID, deviceID, deviceName, deviceType string) error
	GetUserDevices(ctx context.Context, userID uuid.UUID) ([]DeviceSession, error)
	RemoveDevice(ctx context.Context, userID uuid.UUID, deviceID string) error
	ValidateDeviceLimit(ctx context.Context, userID uuid.UUID) error
}

// HLSConfig defines HLS encoding settings
type HLSConfig struct {
	Qualities            []QualityPreset   `json:"qualities"`
	SegmentDuration      int               `json:"segment_duration"` // seconds
	Encryption           HLSEncryptionType `json:"encryption"`
	MaxConcurrentStreams int               `json:"max_concurrent_streams"`
	MaxDevices           int               `json:"max_devices"`
	SignedURLExpiry      int               `json:"signed_url_expiry"` // seconds
}

// QualityPreset defines encoding quality settings
type QualityPreset struct {
	Name         string `json:"name"` // e.g., "1080p"
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Bitrate      int    `json:"bitrate"`       // kbps
	AudioBitrate int    `json:"audio_bitrate"` // kbps
}

// DefaultHLSConfig returns default HLS configuration
func DefaultHLSConfig() HLSConfig {
	return HLSConfig{
		Qualities: []QualityPreset{
			{Name: "1080p", Width: 1920, Height: 1080, Bitrate: 5000, AudioBitrate: 192},
			{Name: "720p", Width: 1280, Height: 720, Bitrate: 2500, AudioBitrate: 128},
			{Name: "480p", Width: 854, Height: 480, Bitrate: 1000, AudioBitrate: 96},
			{Name: "360p", Width: 640, Height: 360, Bitrate: 600, AudioBitrate: 64},
		},
		SegmentDuration:      10,
		Encryption:           HLSEncryptionAES128,
		MaxConcurrentStreams: 1,
		MaxDevices:           3,
		SignedURLExpiry:      14400, // 4 hours
	}
}
