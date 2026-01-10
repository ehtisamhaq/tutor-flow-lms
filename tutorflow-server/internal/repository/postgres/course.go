package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type courseRepository struct {
	db *gorm.DB
}

func NewCourseRepository(db *gorm.DB) repository.CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) Create(ctx context.Context, course *domain.Course) error {
	// Generate slug if not provided
	if course.Slug == "" {
		course.Slug = r.generateUniqueSlug(ctx, course.Title)
	}
	return r.db.WithContext(ctx).Create(course).Error
}

func (r *courseRepository) generateUniqueSlug(ctx context.Context, title string) string {
	baseSlug := slug.Make(title)
	finalSlug := baseSlug
	counter := 1

	for {
		var count int64
		r.db.WithContext(ctx).Model(&domain.Course{}).Where("slug = ?", finalSlug).Count(&count)
		if count == 0 {
			break
		}
		finalSlug = baseSlug + "-" + string(rune('0'+counter))
		counter++
	}

	return finalSlug
}

func (r *courseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Course, error) {
	var course domain.Course
	err := r.db.WithContext(ctx).
		Preload("Instructor").
		Preload("Categories").
		Preload("Modules", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Modules.Lessons", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("id = ?", id).
		First(&course).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCourseNotFound
		}
		return nil, err
	}
	return &course, nil
}

func (r *courseRepository) GetBySlug(ctx context.Context, slug string) (*domain.Course, error) {
	var course domain.Course
	err := r.db.WithContext(ctx).
		Preload("Instructor").
		Preload("Categories").
		Preload("Modules", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Modules.Lessons", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("slug = ?", slug).
		First(&course).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCourseNotFound
		}
		return nil, err
	}
	return &course, nil
}

func (r *courseRepository) Update(ctx context.Context, course *domain.Course) error {
	return r.db.WithContext(ctx).Save(course).Error
}

func (r *courseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Course{}, "id = ?", id).Error
}

func (r *courseRepository) List(ctx context.Context, filters repository.CourseFilters) ([]domain.Course, int64, error) {
	var courses []domain.Course
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Course{})

	// Apply filters
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.Level != nil {
		query = query.Where("level = ?", *filters.Level)
	}
	if filters.InstructorID != nil {
		query = query.Where("instructor_id = ?", *filters.InstructorID)
	}
	if filters.CategoryID != nil {
		query = query.Joins("JOIN course_categories cc ON cc.course_id = courses.id").
			Where("cc.category_id = ?", *filters.CategoryID)
	}
	if filters.IsFeatured != nil {
		query = query.Where("is_featured = ?", *filters.IsFeatured)
	}
	if filters.Search != "" {
		search := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(short_description) LIKE ?", search, search)
	}
	if filters.MinPrice != nil {
		query = query.Where("price >= ?", *filters.MinPrice)
	}
	if filters.MaxPrice != nil {
		query = query.Where("price <= ?", *filters.MaxPrice)
	}
	if filters.MinRating != nil {
		query = query.Where("rating >= ?", *filters.MinRating)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderClause := "created_at DESC"
	if filters.SortBy != "" {
		direction := "DESC"
		if filters.SortOrder == "asc" {
			direction = "ASC"
		}
		switch filters.SortBy {
		case "price":
			orderClause = "price " + direction
		case "rating":
			orderClause = "rating " + direction
		case "students":
			orderClause = "total_students " + direction
		case "created_at":
			orderClause = "created_at " + direction
		}
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.Limit
	err := query.
		Preload("Instructor").
		Preload("Categories").
		Order(orderClause).
		Offset(offset).
		Limit(filters.Limit).
		Find(&courses).Error

	if err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

func (r *courseRepository) GetByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]domain.Course, int64, error) {
	var courses []domain.Course
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Course{}).Where("instructor_id = ?", instructorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Categories").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&courses).Error

	if err != nil {
		return nil, 0, err
	}

	return courses, total, nil
}

func (r *courseRepository) UpdateStats(ctx context.Context, id uuid.UUID) error {
	// Update total lessons count
	var lessonCount int64
	r.db.WithContext(ctx).Model(&domain.Lesson{}).
		Joins("JOIN modules ON modules.id = lessons.module_id").
		Where("modules.course_id = ?", id).
		Count(&lessonCount)

	// Update total students count
	var studentCount int64
	r.db.WithContext(ctx).Model(&domain.Enrollment{}).
		Where("course_id = ? AND status IN ?", id, []domain.EnrollmentStatus{
			domain.EnrollmentStatusActive,
			domain.EnrollmentStatusCompleted,
		}).
		Count(&studentCount)

	return r.db.WithContext(ctx).Model(&domain.Course{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_lessons":  lessonCount,
			"total_students": studentCount,
		}).Error
}

func (r *courseRepository) IncrementStudentCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Course{}).
		Where("id = ?", id).
		UpdateColumn("total_students", gorm.Expr("total_students + 1")).Error
}
