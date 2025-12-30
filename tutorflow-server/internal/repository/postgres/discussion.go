package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// DiscussionRepository
type discussionRepository struct {
	db *gorm.DB
}

func NewDiscussionRepository(db *gorm.DB) repository.DiscussionRepository {
	return &discussionRepository{db: db}
}

func (r *discussionRepository) Create(ctx context.Context, discussion *domain.Discussion) error {
	return r.db.WithContext(ctx).Create(discussion).Error
}

func (r *discussionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Discussion, error) {
	var discussion domain.Discussion
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC").Preload("User")
		}).
		Where("id = ?", id).
		First(&discussion).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &discussion, nil
}

func (r *discussionRepository) Update(ctx context.Context, discussion *domain.Discussion) error {
	return r.db.WithContext(ctx).Save(discussion).Error
}

func (r *discussionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete replies first
	if err := r.db.WithContext(ctx).Where("parent_id = ?", id).Delete(&domain.Discussion{}).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&domain.Discussion{}, "id = ?", id).Error
}

func (r *discussionRepository) GetByCourse(ctx context.Context, courseID uuid.UUID, page, limit int) ([]domain.Discussion, int64, error) {
	var discussions []domain.Discussion
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Discussion{}).
		Where("course_id = ? AND parent_id IS NULL", courseID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC").Limit(3).Preload("User")
		}).
		Order("is_pinned DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&discussions).Error

	return discussions, total, err
}

func (r *discussionRepository) GetByLesson(ctx context.Context, lessonID uuid.UUID, page, limit int) ([]domain.Discussion, int64, error) {
	var discussions []domain.Discussion
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Discussion{}).
		Where("lesson_id = ? AND parent_id IS NULL", lessonID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC").Limit(3).Preload("User")
		}).
		Order("is_pinned DESC, upvotes DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&discussions).Error

	return discussions, total, err
}

func (r *discussionRepository) GetReplies(ctx context.Context, parentID uuid.UUID, page, limit int) ([]domain.Discussion, int64, error) {
	var replies []domain.Discussion
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Discussion{}).Where("parent_id = ?", parentID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&replies).Error

	return replies, total, err
}

func (r *discussionRepository) Upvote(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Discussion{}).
		Where("id = ?", id).
		UpdateColumn("upvotes", gorm.Expr("upvotes + 1")).Error
}

func (r *discussionRepository) RemoveUpvote(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Discussion{}).
		Where("id = ? AND upvotes > 0", id).
		UpdateColumn("upvotes", gorm.Expr("upvotes - 1")).Error
}

func (r *discussionRepository) MarkResolved(ctx context.Context, id uuid.UUID, resolved bool) error {
	return r.db.WithContext(ctx).Model(&domain.Discussion{}).
		Where("id = ?", id).
		Update("is_resolved", resolved).Error
}

func (r *discussionRepository) Pin(ctx context.Context, id uuid.UUID, pinned bool) error {
	return r.db.WithContext(ctx).Model(&domain.Discussion{}).
		Where("id = ?", id).
		Update("is_pinned", pinned).Error
}

func (r *discussionRepository) CountByCourse(ctx context.Context, courseID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Discussion{}).
		Where("course_id = ?", courseID).
		Count(&count).Error
	return count, err
}

func (r *discussionRepository) CountByLesson(ctx context.Context, lessonID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Discussion{}).
		Where("lesson_id = ?", lessonID).
		Count(&count).Error
	return count, err
}
