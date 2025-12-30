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

// RecentlyViewedRepository
type recentlyViewedRepository struct {
	db *gorm.DB
}

func NewRecentlyViewedRepository(db *gorm.DB) repository.RecentlyViewedRepository {
	return &recentlyViewedRepository{db: db}
}

func (r *recentlyViewedRepository) Track(ctx context.Context, userID, courseID uuid.UUID) error {
	// Check if already exists
	var existing domain.RecentlyViewed
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		First(&existing).Error

	if err == nil {
		// Update existing
		existing.ViewedAt = time.Now()
		return r.db.WithContext(ctx).Save(&existing).Error
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Create new
	rv := &domain.RecentlyViewed{
		UserID:   userID,
		CourseID: courseID,
		ViewedAt: time.Now(),
	}
	return r.db.WithContext(ctx).Create(rv).Error
}

func (r *recentlyViewedRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]domain.RecentlyViewed, error) {
	var items []domain.RecentlyViewed
	err := r.db.WithContext(ctx).
		Preload("Course").
		Where("user_id = ?", userID).
		Order("viewed_at DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *recentlyViewedRepository) Clear(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&domain.RecentlyViewed{}, "user_id = ?", userID).Error
}

func (r *recentlyViewedRepository) Delete(ctx context.Context, userID, courseID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&domain.RecentlyViewed{}, "user_id = ? AND course_id = ?", userID, courseID).Error
}

// ScheduledReportRepository
type scheduledReportRepository struct {
	db *gorm.DB
}

func NewScheduledReportRepository(db *gorm.DB) repository.ScheduledReportRepository {
	return &scheduledReportRepository{db: db}
}

func (r *scheduledReportRepository) Create(ctx context.Context, report *domain.ScheduledReport) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *scheduledReportRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ScheduledReport, error) {
	var report domain.ScheduledReport
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&report).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &report, nil
}

func (r *scheduledReportRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.ScheduledReport, error) {
	var reports []domain.ScheduledReport
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&reports).Error
	return reports, err
}

func (r *scheduledReportRepository) Update(ctx context.Context, report *domain.ScheduledReport) error {
	return r.db.WithContext(ctx).Save(report).Error
}

func (r *scheduledReportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.ScheduledReport{}, "id = ?", id).Error
}

func (r *scheduledReportRepository) GetDueReports(ctx context.Context) ([]domain.ScheduledReport, error) {
	var reports []domain.ScheduledReport
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND (next_run_at IS NULL OR next_run_at <= ?)", true, now).
		Find(&reports).Error
	return reports, err
}

func (r *scheduledReportRepository) UpdateLastRun(ctx context.Context, id uuid.UUID, runAt time.Time, nextRunAt time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.ScheduledReport{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_run_at": runAt,
			"next_run_at": nextRunAt,
		}).Error
}
