package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type bundleRepository struct {
	db *gorm.DB
}

// NewBundleRepository creates a new bundle repository
func NewBundleRepository(db *gorm.DB) repository.BundleRepository {
	return &bundleRepository{db: db}
}

func (r *bundleRepository) Create(ctx context.Context, bundle *domain.Bundle) error {
	return r.db.WithContext(ctx).Create(bundle).Error
}

func (r *bundleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bundle, error) {
	var bundle domain.Bundle
	err := r.db.WithContext(ctx).Preload("Courses").Preload("Courses.Course").Preload("Category").
		Where("id = ?", id).First(&bundle).Error
	if err != nil {
		return nil, err
	}
	return &bundle, nil
}

func (r *bundleRepository) GetBySlug(ctx context.Context, slug string) (*domain.Bundle, error) {
	var bundle domain.Bundle
	err := r.db.WithContext(ctx).Preload("Courses").Preload("Courses.Course").Preload("Category").
		Where("slug = ?", slug).First(&bundle).Error
	if err != nil {
		return nil, err
	}
	return &bundle, nil
}

func (r *bundleRepository) GetActive(ctx context.Context, page, limit int) ([]domain.Bundle, int64, error) {
	var bundles []domain.Bundle
	var total int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&domain.Bundle{}).
		Where("is_active = ?", true).
		Where("(start_date IS NULL OR start_date <= NOW())").
		Where("(end_date IS NULL OR end_date >= NOW())")

	query.Count(&total)

	err := r.db.WithContext(ctx).Preload("Courses").Preload("Courses.Course").
		Where("is_active = ?", true).
		Where("(start_date IS NULL OR start_date <= NOW())").
		Where("(end_date IS NULL OR end_date >= NOW())").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&bundles).Error

	return bundles, total, err
}

func (r *bundleRepository) GetByCategory(ctx context.Context, categoryID uuid.UUID) ([]domain.Bundle, error) {
	var bundles []domain.Bundle
	err := r.db.WithContext(ctx).Preload("Courses").Preload("Courses.Course").
		Where("category_id = ? AND is_active = ?", categoryID, true).
		Find(&bundles).Error
	return bundles, err
}

func (r *bundleRepository) Update(ctx context.Context, bundle *domain.Bundle) error {
	return r.db.WithContext(ctx).Save(bundle).Error
}

func (r *bundleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete bundle courses first
		if err := tx.Where("bundle_id = ?", id).Delete(&domain.BundleCourse{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.Bundle{}, id).Error
	})
}

func (r *bundleRepository) AddCourse(ctx context.Context, bundleID, courseID uuid.UUID, order int) error {
	bc := domain.BundleCourse{
		BundleID: bundleID,
		CourseID: courseID,
		Order:    order,
	}
	return r.db.WithContext(ctx).Create(&bc).Error
}

func (r *bundleRepository) RemoveCourse(ctx context.Context, bundleID, courseID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("bundle_id = ? AND course_id = ?", bundleID, courseID).
		Delete(&domain.BundleCourse{}).Error
}

func (r *bundleRepository) GetCourses(ctx context.Context, bundleID uuid.UUID) ([]domain.Course, error) {
	var bundleCourses []domain.BundleCourse
	err := r.db.WithContext(ctx).Preload("Course").Preload("Course.Instructor").
		Where("bundle_id = ?", bundleID).
		Order("\"order\" ASC").
		Find(&bundleCourses).Error
	if err != nil {
		return nil, err
	}

	courses := make([]domain.Course, len(bundleCourses))
	for i, bc := range bundleCourses {
		if bc.Course != nil {
			courses[i] = *bc.Course
		}
	}
	return courses, nil
}

func (r *bundleRepository) RecordPurchase(ctx context.Context, purchase *domain.BundlePurchase) error {
	return r.db.WithContext(ctx).Create(purchase).Error
}

func (r *bundleRepository) GetUserPurchases(ctx context.Context, userID uuid.UUID) ([]domain.BundlePurchase, error) {
	var purchases []domain.BundlePurchase
	err := r.db.WithContext(ctx).Preload("Bundle").Preload("Bundle.Courses").Preload("Bundle.Courses.Course").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&purchases).Error
	return purchases, err
}

func (r *bundleRepository) IncrementPurchaseCount(ctx context.Context, bundleID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Bundle{}).
		Where("id = ?", bundleID).
		Update("purchase_count", gorm.Expr("purchase_count + 1")).Error
}
