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

// AnnouncementRepository
type announcementRepository struct {
	db *gorm.DB
}

func NewAnnouncementRepository(db *gorm.DB) repository.AnnouncementRepository {
	return &announcementRepository{db: db}
}

func (r *announcementRepository) Create(ctx context.Context, announcement *domain.Announcement) error {
	return r.db.WithContext(ctx).Create(announcement).Error
}

func (r *announcementRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	var announcement domain.Announcement
	err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Course").
		Where("id = ?", id).
		First(&announcement).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &announcement, nil
}

func (r *announcementRepository) Update(ctx context.Context, announcement *domain.Announcement) error {
	return r.db.WithContext(ctx).Save(announcement).Error
}

func (r *announcementRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Announcement{}, "id = ?", id).Error
}

func (r *announcementRepository) GetByCourse(ctx context.Context, courseID uuid.UUID, page, limit int) ([]domain.Announcement, int64, error) {
	var announcements []domain.Announcement
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Announcement{}).
		Where("course_id = ? AND published_at <= ?", courseID, time.Now())

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Author").
		Order("is_pinned DESC, published_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&announcements).Error

	return announcements, total, err
}

func (r *announcementRepository) GetGlobal(ctx context.Context, page, limit int) ([]domain.Announcement, int64, error) {
	var announcements []domain.Announcement
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Announcement{}).
		Where("course_id IS NULL AND published_at <= ?", time.Now())

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Author").
		Order("is_pinned DESC, published_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&announcements).Error

	return announcements, total, err
}

func (r *announcementRepository) GetForUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Announcement, int64, error) {
	var announcements []domain.Announcement
	var total int64

	// Get announcements from enrolled courses + global announcements
	query := r.db.WithContext(ctx).Model(&domain.Announcement{}).
		Joins("LEFT JOIN enrollments ON enrollments.course_id = announcements.course_id").
		Where("(enrollments.user_id = ? AND enrollments.status = ?) OR announcements.course_id IS NULL", userID, domain.EnrollmentStatusActive).
		Where("announcements.published_at <= ?", time.Now())

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).Model(&domain.Announcement{}).
		Preload("Author").
		Preload("Course").
		Joins("LEFT JOIN enrollments ON enrollments.course_id = announcements.course_id").
		Where("(enrollments.user_id = ? AND enrollments.status = ?) OR announcements.course_id IS NULL", userID, domain.EnrollmentStatusActive).
		Where("announcements.published_at <= ?", time.Now()).
		Group("announcements.id").
		Order("announcements.is_pinned DESC, announcements.published_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&announcements).Error

	return announcements, total, err
}

func (r *announcementRepository) GetByAuthor(ctx context.Context, authorID uuid.UUID, page, limit int) ([]domain.Announcement, int64, error) {
	var announcements []domain.Announcement
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Announcement{}).Where("author_id = ?", authorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Course").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&announcements).Error

	return announcements, total, err
}

func (r *announcementRepository) Pin(ctx context.Context, id uuid.UUID, pinned bool) error {
	return r.db.WithContext(ctx).Model(&domain.Announcement{}).
		Where("id = ?", id).
		Update("is_pinned", pinned).Error
}
