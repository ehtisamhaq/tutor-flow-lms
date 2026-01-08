package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tutorflow/tutorflow-server/internal/domain"
)

type videoUseCase struct {
	videoRepo      domain.VideoRepository
	lessonRepo     domain.LessonRepository
	enrollmentRepo domain.EnrollmentRepository
	config         domain.HLSConfig
	signingSecret  string
}

// NewVideoUseCase creates a new video use case
func NewVideoUseCase(
	videoRepo domain.VideoRepository,
	lessonRepo domain.LessonRepository,
	enrollmentRepo domain.EnrollmentRepository,
	signingSecret string,
) domain.VideoUseCase {
	return &videoUseCase{
		videoRepo:      videoRepo,
		lessonRepo:     lessonRepo,
		enrollmentRepo: enrollmentRepo,
		config:         domain.DefaultHLSConfig(),
		signingSecret:  signingSecret,
	}
}

// UploadVideo uploads a video for a lesson
func (uc *videoUseCase) UploadVideo(lessonID uuid.UUID, fileURL string) (*domain.HLSVideoAsset, error) {
	// Check if video already exists for lesson
	existing, _ := uc.videoRepo.GetAssetByLessonID(lessonID)
	if existing != nil {
		return nil, errors.New("video already exists for this lesson")
	}

	// Create video asset record
	asset := &domain.HLSVideoAsset{
		LessonID:    lessonID,
		OriginalURL: fileURL,
		Status:      domain.VideoStatusPending,
	}

	if err := uc.videoRepo.CreateAsset(asset); err != nil {
		return nil, err
	}

	// In production, trigger async video processing job here
	// For now, mark as completed
	asset.Status = domain.VideoStatusCompleted

	return asset, nil
}

// ProcessVideo processes a video into HLS format (async in production)
func (uc *videoUseCase) ProcessVideo(videoID uuid.UUID) error {
	asset, err := uc.videoRepo.GetAssetByID(videoID)
	if err != nil {
		return errors.New("video not found")
	}

	// Update status to processing
	asset.Status = domain.VideoStatusProcessing
	if err := uc.videoRepo.UpdateAsset(asset); err != nil {
		return err
	}

	// In production, this would:
	// 1. Download the original video
	// 2. Transcode to multiple qualities
	// 3. Generate HLS playlists
	// 4. Apply encryption if enabled
	// 5. Upload segments to CDN

	// For now, simulate completion
	asset.Status = domain.VideoStatusCompleted
	asset.Duration = 600 // 10 minutes placeholder
	asset.Resolution = "1920x1080"
	asset.UpdatedAt = time.Now()

	return uc.videoRepo.UpdateAsset(asset)
}

// GetProcessingStatus returns video processing status
func (uc *videoUseCase) GetProcessingStatus(lessonID uuid.UUID) (*domain.HLSVideoAsset, error) {
	return uc.videoRepo.GetAssetByLessonID(lessonID)
}

// GetPlaybackURL returns a signed playback URL for a video
func (uc *videoUseCase) GetPlaybackURL(lessonID, userID uuid.UUID, deviceID string) (string, error) {
	// Verify user has access to the lesson
	lesson, err := uc.lessonRepo.GetByID(lessonID)
	if err != nil {
		return "", errors.New("lesson not found")
	}

	// Check enrollment (unless it's a preview)
	if !lesson.IsPreview {
		enrolled, _ := uc.enrollmentRepo.IsEnrolled(userID, lesson.Module.CourseID)
		if !enrolled {
			return "", errors.New("user is not enrolled in this course")
		}
	}

	// Get video asset
	asset, err := uc.videoRepo.GetAssetByLessonID(lessonID)
	if err != nil {
		return "", errors.New("video not found")
	}

	if asset.Status != domain.VideoStatusCompleted {
		return "", errors.New("video is not ready for playback")
	}

	// Validate device limit
	if err := uc.ValidateDeviceLimit(userID); err != nil {
		return "", err
	}

	// Register device session
	uc.RegisterDevice(userID, deviceID, "Unknown", "unknown")

	// Generate signed URL
	token := uc.generateToken(asset.ID, userID, deviceID)
	expiresAt := time.Now().Add(time.Duration(uc.config.SignedURLExpiry) * time.Second)

	signedURL := &domain.SignedURL{
		VideoID:   asset.ID,
		UserID:    userID,
		SessionID: uuid.New().String(),
		DeviceID:  deviceID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := uc.videoRepo.CreateSignedURL(signedURL); err != nil {
		return "", err
	}

	// Return the playback URL with token
	// In production, this would be a CDN URL
	playbackURL := fmt.Sprintf("/api/v1/videos/stream/%s?token=%s", asset.ID, token)
	return playbackURL, nil
}

// GetEncryptionKey returns the encryption key for a video
func (uc *videoUseCase) GetEncryptionKey(token string) ([]byte, error) {
	signedURL, err := uc.videoRepo.GetSignedURLByToken(token)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	if !signedURL.IsValid() {
		return nil, errors.New("token expired")
	}

	// Get encryption info
	encryption, err := uc.videoRepo.GetEncryptionByVideoID(signedURL.VideoID)
	if err != nil {
		return nil, errors.New("encryption not configured")
	}

	// Decode and return the key
	key, err := hex.DecodeString(encryption.EncryptionKey)
	if err != nil {
		return nil, errors.New("invalid encryption key")
	}

	// Mark URL as used (for single-use tokens)
	uc.videoRepo.MarkSignedURLUsed(token)

	return key, nil
}

// ValidatePlayback validates if a playback session is valid
func (uc *videoUseCase) ValidatePlayback(token string) error {
	signedURL, err := uc.videoRepo.GetSignedURLByToken(token)
	if err != nil {
		return errors.New("invalid token")
	}

	if !signedURL.IsValid() {
		return errors.New("token expired")
	}

	return nil
}

// EnableEncryption enables encryption for a video
func (uc *videoUseCase) EnableEncryption(videoID uuid.UUID, encType domain.HLSEncryptionType) error {
	_, err := uc.videoRepo.GetAssetByID(videoID)
	if err != nil {
		return errors.New("video not found")
	}

	// Generate encryption key
	key, iv, err := domain.GenerateEncryptionKey()
	if err != nil {
		return err
	}

	encryption := &domain.VideoEncryption{
		VideoID:        videoID,
		EncryptionType: encType,
		KeyID:          uuid.New().String(),
		EncryptionKey:  key,
		IV:             iv,
		KeyURL:         fmt.Sprintf("/api/v1/drm/key/%s", videoID),
	}

	// Check if encryption already exists
	existing, _ := uc.videoRepo.GetEncryptionByVideoID(videoID)
	if existing != nil {
		encryption.ID = existing.ID
		encryption.UpdatedAt = time.Now()
		return uc.videoRepo.UpdateEncryption(encryption)
	}

	return uc.videoRepo.CreateEncryption(encryption)
}

// RotateEncryptionKey rotates the encryption key for a video
func (uc *videoUseCase) RotateEncryptionKey(videoID uuid.UUID) error {
	encryption, err := uc.videoRepo.GetEncryptionByVideoID(videoID)
	if err != nil {
		return errors.New("encryption not configured")
	}

	// Generate new key
	key, iv, err := domain.GenerateEncryptionKey()
	if err != nil {
		return err
	}

	encryption.EncryptionKey = key
	encryption.IV = iv
	encryption.KeyID = uuid.New().String()
	encryption.UpdatedAt = time.Now()

	return uc.videoRepo.UpdateEncryption(encryption)
}

// RegisterDevice registers a device for a user
func (uc *videoUseCase) RegisterDevice(userID uuid.UUID, deviceID, deviceName, deviceType string) error {
	// Check if device already exists
	existing, _ := uc.videoRepo.GetDeviceSession(userID, deviceID)
	if existing != nil {
		// Update last seen
		existing.LastSeenAt = time.Now()
		existing.IsActive = true
		return uc.videoRepo.UpdateDeviceSession(existing)
	}

	session := &domain.DeviceSession{
		UserID:     userID,
		DeviceID:   deviceID,
		DeviceName: deviceName,
		DeviceType: deviceType,
		IsActive:   true,
		LastSeenAt: time.Now(),
	}

	return uc.videoRepo.CreateDeviceSession(session)
}

// GetUserDevices returns user's registered devices
func (uc *videoUseCase) GetUserDevices(userID uuid.UUID) ([]domain.DeviceSession, error) {
	return uc.videoRepo.GetUserDeviceSessions(userID)
}

// RemoveDevice removes a device from user's account
func (uc *videoUseCase) RemoveDevice(userID uuid.UUID, deviceID string) error {
	session, err := uc.videoRepo.GetDeviceSession(userID, deviceID)
	if err != nil {
		return errors.New("device not found")
	}

	return uc.videoRepo.DeactivateDeviceSession(session.ID)
}

// ValidateDeviceLimit checks if user has reached device limit
func (uc *videoUseCase) ValidateDeviceLimit(userID uuid.UUID) error {
	count, err := uc.videoRepo.CountActiveDevices(userID)
	if err != nil {
		return err
	}

	if int(count) >= uc.config.MaxDevices {
		return errors.New("device limit reached")
	}

	return nil
}

// Helper to generate signed token
func (uc *videoUseCase) generateToken(videoID, userID uuid.UUID, deviceID string) string {
	data := fmt.Sprintf("%s:%s:%s:%d", videoID, userID, deviceID, time.Now().Unix())
	h := hmac.New(sha256.New, []byte(uc.signingSecret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
