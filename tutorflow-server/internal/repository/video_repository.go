package repository

import (
	"time"

	"github.com/tutorflow/tutorflow-server/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type videoRepository struct {
	db *gorm.DB
}

// NewVideoRepository creates a new video repository
func NewVideoRepository(db *gorm.DB) domain.VideoRepository {
	return &videoRepository{db: db}
}

// Video Assets

func (r *videoRepository) CreateAsset(asset *domain.HLSVideoAsset) error {
	return r.db.Create(asset).Error
}

func (r *videoRepository) GetAssetByID(id uuid.UUID) (*domain.HLSVideoAsset, error) {
	var asset domain.HLSVideoAsset
	err := r.db.Preload("Qualities").Preload("Encryption").
		Where("id = ?", id).First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *videoRepository) GetAssetByLessonID(lessonID uuid.UUID) (*domain.HLSVideoAsset, error) {
	var asset domain.HLSVideoAsset
	err := r.db.Preload("Qualities").Preload("Encryption").
		Where("lesson_id = ?", lessonID).First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *videoRepository) UpdateAsset(asset *domain.HLSVideoAsset) error {
	return r.db.Save(asset).Error
}

func (r *videoRepository) DeleteAsset(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete related records
		if err := tx.Where("video_id = ?", id).Delete(&domain.VideoQuality{}).Error; err != nil {
			return err
		}
		if err := tx.Where("video_id = ?", id).Delete(&domain.VideoEncryption{}).Error; err != nil {
			return err
		}
		if err := tx.Where("video_id = ?", id).Delete(&domain.SignedURL{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.HLSVideoAsset{}, id).Error
	})
}

// Quality variants

func (r *videoRepository) CreateQuality(quality *domain.VideoQuality) error {
	return r.db.Create(quality).Error
}

func (r *videoRepository) GetQualitiesByVideoID(videoID uuid.UUID) ([]domain.VideoQuality, error) {
	var qualities []domain.VideoQuality
	err := r.db.Where("video_id = ?", videoID).Order("bitrate DESC").Find(&qualities).Error
	return qualities, err
}

// Encryption

func (r *videoRepository) CreateEncryption(encryption *domain.VideoEncryption) error {
	return r.db.Create(encryption).Error
}

func (r *videoRepository) GetEncryptionByVideoID(videoID uuid.UUID) (*domain.VideoEncryption, error) {
	var enc domain.VideoEncryption
	err := r.db.Where("video_id = ?", videoID).First(&enc).Error
	if err != nil {
		return nil, err
	}
	return &enc, nil
}

func (r *videoRepository) UpdateEncryption(encryption *domain.VideoEncryption) error {
	return r.db.Save(encryption).Error
}

// Signed URLs

func (r *videoRepository) CreateSignedURL(signedURL *domain.SignedURL) error {
	return r.db.Create(signedURL).Error
}

func (r *videoRepository) GetSignedURLByToken(token string) (*domain.SignedURL, error) {
	var url domain.SignedURL
	err := r.db.Preload("Video").Preload("User").
		Where("token = ?", token).First(&url).Error
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (r *videoRepository) MarkSignedURLUsed(token string) error {
	now := time.Now()
	return r.db.Model(&domain.SignedURL{}).
		Where("token = ?", token).
		Update("used_at", &now).Error
}

func (r *videoRepository) CleanupExpiredURLs() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&domain.SignedURL{}).Error
}

// Device Sessions

func (r *videoRepository) CreateDeviceSession(session *domain.DeviceSession) error {
	return r.db.Create(session).Error
}

func (r *videoRepository) GetDeviceSession(userID uuid.UUID, deviceID string) (*domain.DeviceSession, error) {
	var session domain.DeviceSession
	err := r.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *videoRepository) GetUserDeviceSessions(userID uuid.UUID) ([]domain.DeviceSession, error) {
	var sessions []domain.DeviceSession
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_seen_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *videoRepository) CountActiveDevices(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&domain.DeviceSession{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Count(&count).Error
	return count, err
}

func (r *videoRepository) UpdateDeviceSession(session *domain.DeviceSession) error {
	return r.db.Save(session).Error
}

func (r *videoRepository) DeactivateDeviceSession(id uuid.UUID) error {
	return r.db.Model(&domain.DeviceSession{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}
