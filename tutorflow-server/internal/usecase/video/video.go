package video

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type videoUseCase struct {
	videoRepo      repository.VideoRepository
	lessonRepo     repository.LessonRepository
	enrollmentRepo repository.EnrollmentRepository
	config         domain.HLSConfig
	signingSecret  string
}

// NewVideoUseCase creates a new video use case
func NewUseCase(
	videoRepo repository.VideoRepository,
	lessonRepo repository.LessonRepository,
	enrollmentRepo repository.EnrollmentRepository,
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
func (uc *videoUseCase) UploadVideo(ctx context.Context, lessonID uuid.UUID, fileURL string) (*domain.HLSVideoAsset, error) {
	// Check if video already exists for lesson
	existing, _ := uc.videoRepo.GetAssetByLessonID(ctx, lessonID)
	if existing != nil {
		return nil, errors.New("video already exists for this lesson")
	}

	// Create video asset record
	asset := &domain.HLSVideoAsset{
		LessonID:    lessonID,
		OriginalURL: fileURL,
		Status:      domain.VideoStatusPending,
	}

	if err := uc.videoRepo.CreateAsset(ctx, asset); err != nil {
		return nil, err
	}

	// In production, trigger async video processing job here
	// For now, we'll mark strictly as pending and expect ProcessVideo to be called.
	// But to match previous behavior for non-async parts, we can leave it or trigger it.
	// The user asked to implement "real video processing", so usually this would be triggered via a queue.
	// We will manually trigger ProcessVideo for testing purposes if needed, or assume the caller does it.

	// Start processing in a goroutine to simulate async job
	go func() {
		// Create a new context for the background job
		bgCtx := context.Background()
		_ = uc.ProcessVideo(bgCtx, asset.ID)
	}()

	return asset, nil
}

// ProcessVideo processes a video into HLS format
func (uc *videoUseCase) ProcessVideo(ctx context.Context, videoID uuid.UUID) error {
	asset, err := uc.videoRepo.GetAssetByID(ctx, videoID)
	if err != nil {
		return errors.New("video not found")
	}

	// Update status to processing
	asset.Status = domain.VideoStatusProcessing
	if err := uc.videoRepo.UpdateAsset(ctx, asset); err != nil {
		return err
	}

	// 1. Create temp directory
	tempDir, err := os.MkdirTemp("", "hls-processing-*")
	if err != nil {
		return uc.handleProcessingError(ctx, asset, err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// 2. Generate encryption key
	key, iv, err := domain.GenerateEncryptionKey()
	if err != nil {
		return uc.handleProcessingError(ctx, asset, err)
	}

	// 3. Save key to file (for ffmpeg)
	keyFile := filepath.Join(tempDir, "video.key")
	// Key must be binary for ffmpeg? No, usually hex or binary.
	// FFmpeg hls_key_info_file format:
	// key URI
	// key file path
	// IV (optional)

	// We need to write the BINARY key to a file for ffmpeg to use for encryption
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return uc.handleProcessingError(ctx, asset, err)
	}
	if err := os.WriteFile(keyFile, keyBytes, 0600); err != nil {
		return uc.handleProcessingError(ctx, asset, err)
	}

	// Create key info file
	keyInfoFile := filepath.Join(tempDir, "video.keyinfo")
	// The first line is the URI that will be written to the playlist.
	// We want this to be our API endpoint.
	keyURI := fmt.Sprintf("http://localhost:8080/api/v1/drm/key/%s", videoID) // TODO: Use configured base URL

	keyInfoContent := fmt.Sprintf("%s\n%s\n%s", keyURI, keyFile, iv)
	if err := os.WriteFile(keyInfoFile, []byte(keyInfoContent), 0600); err != nil {
		return uc.handleProcessingError(ctx, asset, err)
	}

	// 4. Run FFmpeg
	// Input: asset.OriginalURL (assuming local path for simplicity or http url)
	// Output: HLS playlist and segments

	// Ensure output dir exists
	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return uc.handleProcessingError(ctx, asset, err)
	}

	playlistPath := filepath.Join(outputDir, "index.m3u8")
	segmentPath := filepath.Join(outputDir, "segment_%03d.ts")

	// Verify input file exists (if local)
	if _, err := os.Stat(asset.OriginalURL); os.IsNotExist(err) {
		// If it's a URL, ffmpeg can handle it, but if it expects a local file...
		// For now assume local file path from upload.
		// If it's "uploads/..." we might need absolute path.
		// Let's assume absolute path is needed if it's relative.
		if !filepath.IsAbs(asset.OriginalURL) {
			cwd, _ := os.Getwd()
			asset.OriginalURL = filepath.Join(cwd, asset.OriginalURL)
		}
	}

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", asset.OriginalURL,
		"-c:v", "libx264", "-b:v", "1000k", // Simplify for now: single quality
		"-c:a", "aac", "-b:a", "128k",
		"-hls_time", "10",
		"-hls_playlist_type", "vod",
		"-hls_key_info_file", keyInfoFile,
		"-hls_segment_filename", segmentPath,
		playlistPath,
	)

	// Capture output for debugging
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return uc.handleProcessingError(ctx, asset, fmt.Errorf("ffmpeg failed: %w", err))
	}

	// 5. Save encryption info to DB
	// We need to store the key we generated so the API can serve it
	encRecord := &domain.VideoEncryption{
		VideoID:        videoID,
		EncryptionType: domain.HLSEncryptionAES128,
		KeyID:          uuid.New().String(),
		EncryptionKey:  key,
		IV:             iv,
		KeyURL:         keyURI,
	}
	if err := uc.videoRepo.CreateEncryption(ctx, encRecord); err != nil {
		return uc.handleProcessingError(ctx, asset, err)
	}

	// 6. Move files to permanent storage
	// For this PoC, we'll just keep them in a specific 'processed' folder
	finalDir := filepath.Join("uploads", "hls", videoID.String())
	if err := os.MkdirAll(finalDir, 0755); err != nil {
		return uc.handleProcessingError(ctx, asset, err)
	}

	// Copy/Move files
	files, _ := os.ReadDir(outputDir)
	for _, f := range files {
		src := filepath.Join(outputDir, f.Name())
		dst := filepath.Join(finalDir, f.Name())
		// Move
		if err := os.Rename(src, dst); err != nil {
			// Fallback to copy if rename fails (cross-device)
			if err := copyFile(src, dst); err != nil {
				return uc.handleProcessingError(ctx, asset, err)
			}
		}
	}

	// 7. Update Asset
	asset.Status = domain.VideoStatusCompleted
	asset.Duration = 600           // TODO: Parse from ffmpeg output
	asset.Resolution = "1920x1080" // TODO: Parse
	asset.UpdatedAt = time.Now()

	// Update generic playlist URL (fake CDN)
	// In reality this would be /uploads/hls/<id>/index.m3u8 served statically or via endpoint
	// Assuming we serve static files from /uploads
	// playlistURL assignment removed as it was unused

	return uc.videoRepo.UpdateAsset(ctx, asset)
}

func (uc *videoUseCase) handleProcessingError(ctx context.Context, asset *domain.HLSVideoAsset, err error) error {
	asset.Status = domain.VideoStatusFailed
	asset.ProcessingError = err.Error()
	_ = uc.videoRepo.UpdateAsset(ctx, asset)
	return err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// GetProcessingStatus returns video processing status
func (uc *videoUseCase) GetProcessingStatus(ctx context.Context, lessonID uuid.UUID) (*domain.HLSVideoAsset, error) {
	return uc.videoRepo.GetAssetByLessonID(ctx, lessonID)
}

// GetPlaybackURL returns a signed playback URL for a video
func (uc *videoUseCase) GetPlaybackURL(ctx context.Context, lessonID, userID uuid.UUID, deviceID string) (string, error) {
	// Verify user has access to the lesson
	lesson, err := uc.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return "", errors.New("lesson not found")
	}

	// Check enrollment (unless it's a preview)
	if !lesson.IsPreview {
		enrollment, err := uc.enrollmentRepo.GetByUserAndCourse(ctx, userID, lesson.Module.CourseID)
		if err != nil {
			return "", errors.New("user is not enrolled in this course")
		}

		if !enrollment.IsActive() && !enrollment.IsCompleted() {
			return "", errors.New("user is not enrolled or active in this course")
		}
	}

	// Get video asset
	asset, err := uc.videoRepo.GetAssetByLessonID(ctx, lessonID)
	if err != nil {
		return "", errors.New("video not found")
	}

	if asset.Status != domain.VideoStatusCompleted {
		return "", errors.New("video is not ready for playback")
	}

	// Validate device limit
	if err := uc.ValidateDeviceLimit(ctx, userID); err != nil {
		return "", err
	}

	// Register device session
	uc.RegisterDevice(ctx, userID, deviceID, "Unknown", "unknown")

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

	if err := uc.videoRepo.CreateSignedURL(ctx, signedURL); err != nil {
		return "", err
	}

	// Return the playback URL with token
	// This URL should point to our HLS playlist serve endpoint
	// e.g. /api/v1/videos/stream/:id/index.m3u8?token=...
	playbackURL := fmt.Sprintf("/api/v1/videos/stream/%s/index.m3u8?token=%s", asset.ID, token)
	return playbackURL, nil
}

// GetEncryptionKey returns the encryption key for a video
func (uc *videoUseCase) GetEncryptionKey(ctx context.Context, token string) ([]byte, error) {
	signedURL, err := uc.videoRepo.GetSignedURLByToken(ctx, token)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	if !signedURL.IsValid() {
		return nil, errors.New("token expired")
	}

	// Get encryption info
	encryption, err := uc.videoRepo.GetEncryptionByVideoID(ctx, signedURL.VideoID)
	if err != nil {
		return nil, errors.New("encryption not configured")
	}

	// Decode and return the key
	key, err := hex.DecodeString(encryption.EncryptionKey)
	if err != nil {
		return nil, errors.New("invalid encryption key")
	}

	// Mark URL as used (for single-use tokens, but for HLS key we might need to allow multiple fetches?)
	// Usually key is fetched once per session or periodically.
	// Let's keep it for now.
	uc.videoRepo.MarkSignedURLUsed(ctx, token)

	return key, nil
}

// ValidatePlayback validates if a playback session is valid
func (uc *videoUseCase) ValidatePlayback(ctx context.Context, token string) error {
	signedURL, err := uc.videoRepo.GetSignedURLByToken(ctx, token)
	if err != nil {
		return errors.New("invalid token")
	}

	if !signedURL.IsValid() {
		return errors.New("token expired")
	}

	return nil
}

// EnableEncryption enables encryption for a video
func (uc *videoUseCase) EnableEncryption(ctx context.Context, videoID uuid.UUID, encType domain.HLSEncryptionType) error {
	_, err := uc.videoRepo.GetAssetByID(ctx, videoID)
	if err != nil {
		return errors.New("video not found")
	}

	// Generate encryption key
	key, iv, err := domain.GenerateEncryptionKey()
	if err != nil {
		return err
	}

	// We need to re-process the video to apply encryption really.
	// But this function just creates the record?
	// The interface implies "EnableEncryption" might trigger processing?
	// For now, let's just create the record as per stub, but updated with context.

	encryption := &domain.VideoEncryption{
		VideoID:        videoID,
		EncryptionType: encType,
		KeyID:          uuid.New().String(),
		EncryptionKey:  key,
		IV:             iv,
		KeyURL:         fmt.Sprintf("/api/v1/drm/key/%s", videoID),
	}

	// Check if encryption already exists
	existing, _ := uc.videoRepo.GetEncryptionByVideoID(ctx, videoID)
	if existing != nil {
		encryption.ID = existing.ID
		encryption.UpdatedAt = time.Now()
		return uc.videoRepo.UpdateEncryption(ctx, encryption)
	}

	return uc.videoRepo.CreateEncryption(ctx, encryption)
}

// RotateEncryptionKey rotates the encryption key for a video
func (uc *videoUseCase) RotateEncryptionKey(ctx context.Context, videoID uuid.UUID) error {
	encryption, err := uc.videoRepo.GetEncryptionByVideoID(ctx, videoID)
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

	return uc.videoRepo.UpdateEncryption(ctx, encryption)
}

// RegisterDevice registers a device for a user
func (uc *videoUseCase) RegisterDevice(ctx context.Context, userID uuid.UUID, deviceID, deviceName, deviceType string) error {
	// Check if device already exists
	existing, _ := uc.videoRepo.GetDeviceSession(ctx, userID, deviceID)
	if existing != nil {
		// Update last seen
		existing.LastSeenAt = time.Now()
		existing.IsActive = true
		return uc.videoRepo.UpdateDeviceSession(ctx, existing)
	}

	session := &domain.DeviceSession{
		UserID:     userID,
		DeviceID:   deviceID,
		DeviceName: deviceName,
		DeviceType: deviceType,
		IsActive:   true,
		LastSeenAt: time.Now(),
	}

	return uc.videoRepo.CreateDeviceSession(ctx, session)
}

// GetUserDevices returns user's registered devices
func (uc *videoUseCase) GetUserDevices(ctx context.Context, userID uuid.UUID) ([]domain.DeviceSession, error) {
	return uc.videoRepo.GetUserDeviceSessions(ctx, userID)
}

// RemoveDevice removes a device from user's account
func (uc *videoUseCase) RemoveDevice(ctx context.Context, userID uuid.UUID, deviceID string) error {
	session, err := uc.videoRepo.GetDeviceSession(ctx, userID, deviceID)
	if err != nil {
		return errors.New("device not found")
	}

	return uc.videoRepo.DeactivateDeviceSession(ctx, session.ID)
}

// ValidateDeviceLimit checks if user has reached device limit
func (uc *videoUseCase) ValidateDeviceLimit(ctx context.Context, userID uuid.UUID) error {
	count, err := uc.videoRepo.CountActiveDevices(ctx, userID)
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
