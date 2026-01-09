package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type videoRepository struct {
	db *gorm.DB
}

// NewVideoRepository created a new video repository
func NewVideoRepository(db *gorm.DB) repository.VideoRepository {
	return &videoRepository{db: db}
}

// Asset methods
func (r *videoRepository) CreateAsset(ctx context.Context, asset *domain.HLSVideoAsset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

func (r *videoRepository) GetAssetByID(ctx context.Context, id uuid.UUID) (*domain.HLSVideoAsset, error) {
	var asset domain.HLSVideoAsset
	err := r.db.WithContext(ctx).Preload("Qualities").Preload("HLSEncryption").First(&asset, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *videoRepository) GetAssetByLessonID(ctx context.Context, lessonID uuid.UUID) (*domain.HLSVideoAsset, error) {
	var asset domain.HLSVideoAsset
	err := r.db.WithContext(ctx).Preload("Qualities").Preload("HLSEncryption").First(&asset, "lesson_id = ?", lessonID).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *videoRepository) UpdateAsset(ctx context.Context, asset *domain.HLSVideoAsset) error {
	return r.db.WithContext(ctx).Save(asset).Error
}

func (r *videoRepository) DeleteAsset(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.HLSVideoAsset{}, "id = ?", id).Error
}

// Quality methods
func (r *videoRepository) CreateQuality(ctx context.Context, quality *domain.VideoQuality) error {
	return r.db.WithContext(ctx).Create(quality).Error
}

func (r *videoRepository) GetQualitiesByVideoID(ctx context.Context, videoID uuid.UUID) ([]domain.VideoQuality, error) {
	var qualities []domain.VideoQuality
	err := r.db.WithContext(ctx).Where("video_id = ?", videoID).Find(&qualities).Error
	return qualities, err
}

// Encryption methods
func (r *videoRepository) CreateEncryption(ctx context.Context, encryption *domain.VideoEncryption) error {
	return r.db.WithContext(ctx).Create(encryption).Error
}

func (r *videoRepository) GetEncryptionByVideoID(ctx context.Context, videoID uuid.UUID) (*domain.VideoEncryption, error) {
	var encryption domain.VideoEncryption
	err := r.db.WithContext(ctx).First(&encryption, "video_id = ?", videoID).Error
	if err != nil {
		return nil, err
	}
	return &encryption, nil
}

func (r *videoRepository) UpdateEncryption(ctx context.Context, encryption *domain.VideoEncryption) error {
	return r.db.WithContext(ctx).Save(encryption).Error
}

// Signed URL methods
func (r *videoRepository) CreateSignedURL(ctx context.Context, signedURL *domain.SignedURL) error {
	return r.db.WithContext(ctx).Create(signedURL).Error
}

func (r *videoRepository) GetSignedURLByToken(ctx context.Context, token string) (*domain.SignedURL, error) {
	var signedURL domain.SignedURL
	err := r.db.WithContext(ctx).First(&signedURL, "token = ?", token).Error
	if err != nil {
		return nil, err
	}
	return &signedURL, nil
}

func (r *videoRepository) MarkSignedURLUsed(ctx context.Context, token string) error {
	now := time.Now()
	// Using map to force update even if zero value (though used_at is pointer so it's fine)
	return r.db.WithContext(ctx).Model(&domain.SignedURL{}).Where("token = ?", token).Update("used_at", now).Error
}

func (r *videoRepository) CleanupExpiredURLs(ctx context.Context) error {
	return r.db.WithContext(ctx).Delete(&domain.SignedURL{}, "expires_at < ?", time.Now()).Error
}

// Device Session methods
func (r *videoRepository) CreateDeviceSession(ctx context.Context, session *domain.DeviceSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *videoRepository) GetDeviceSession(ctx context.Context, userID uuid.UUID, deviceID string) (*domain.DeviceSession, error) {
	var session domain.DeviceSession
	err := r.db.WithContext(ctx).First(&session, "user_id = ? AND device_id = ?", userID, deviceID).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *videoRepository) GetUserDeviceSessions(ctx context.Context, userID uuid.UUID) ([]domain.DeviceSession, error) {
	var sessions []domain.DeviceSession
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Find(&sessions).Error
	return sessions, err
}

func (r *videoRepository) CountActiveDevices(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.DeviceSession{}).Where("user_id = ? AND is_active = ?", userID, true).Count(&count).Error
	return count, err
}

func (r *videoRepository) UpdateDeviceSession(ctx context.Context, session *domain.DeviceSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *videoRepository) DeactivateDeviceSession(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.DeviceSession{}).Where("id = ?", id).Update("is_active", false).Error
}
