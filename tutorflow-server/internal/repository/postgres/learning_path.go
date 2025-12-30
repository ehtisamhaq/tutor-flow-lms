package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// LearningPathRepository
type learningPathRepository struct {
	db *gorm.DB
}

func NewLearningPathRepository(db *gorm.DB) repository.LearningPathRepository {
	return &learningPathRepository{db: db}
}

// Learning Paths

func (r *learningPathRepository) Create(ctx context.Context, path *domain.LearningPath) error {
	return r.db.WithContext(ctx).Create(path).Error
}

func (r *learningPathRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.LearningPath, error) {
	var path domain.LearningPath
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Creator").
		Preload("Courses", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Courses.Course").
		Where("id = ?", id).
		First(&path).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &path, nil
}

func (r *learningPathRepository) GetBySlug(ctx context.Context, slug string) (*domain.LearningPath, error) {
	var path domain.LearningPath
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Creator").
		Preload("Courses", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Courses.Course").
		Where("slug = ?", slug).
		First(&path).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &path, nil
}

func (r *learningPathRepository) Update(ctx context.Context, path *domain.LearningPath) error {
	return r.db.WithContext(ctx).Save(path).Error
}

func (r *learningPathRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.LearningPath{}, "id = ?", id).Error
}

func (r *learningPathRepository) List(ctx context.Context, filters repository.LearningPathFilters) ([]domain.LearningPath, int64, error) {
	var paths []domain.LearningPath
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.LearningPath{})

	if filters.CategoryID != nil {
		query = query.Where("category_id = ?", *filters.CategoryID)
	}
	if filters.Level != "" {
		query = query.Where("level = ?", filters.Level)
	}
	if filters.IsPublished != nil {
		query = query.Where("is_published = ?", *filters.IsPublished)
	}
	if filters.IsFeatured != nil {
		query = query.Where("is_featured = ?", *filters.IsFeatured)
	}
	if filters.Search != "" {
		search := "%" + filters.Search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filters.Page - 1) * filters.Limit
	err := query.
		Preload("Category").
		Preload("Creator").
		Order("created_at DESC").
		Offset(offset).
		Limit(filters.Limit).
		Find(&paths).Error

	return paths, total, err
}

func (r *learningPathRepository) GetFeatured(ctx context.Context, limit int) ([]domain.LearningPath, error) {
	var paths []domain.LearningPath
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("is_published = ? AND is_featured = ?", true, true).
		Order("total_students DESC").
		Limit(limit).
		Find(&paths).Error
	return paths, err
}

func (r *learningPathRepository) GetByCategory(ctx context.Context, categoryID uuid.UUID, limit int) ([]domain.LearningPath, error) {
	var paths []domain.LearningPath
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("is_published = ? AND category_id = ?", true, categoryID).
		Order("total_students DESC").
		Limit(limit).
		Find(&paths).Error
	return paths, err
}

// Path Courses

func (r *learningPathRepository) AddCourse(ctx context.Context, pathCourse *domain.LearningPathCourse) error {
	err := r.db.WithContext(ctx).Create(pathCourse).Error
	if err != nil {
		return err
	}
	// Update total courses count
	return r.updatePathStats(ctx, pathCourse.PathID)
}

func (r *learningPathRepository) RemoveCourse(ctx context.Context, pathID, courseID uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Delete(&domain.LearningPathCourse{}, "path_id = ? AND course_id = ?", pathID, courseID).Error
	if err != nil {
		return err
	}
	return r.updatePathStats(ctx, pathID)
}

func (r *learningPathRepository) UpdateCoursePosition(ctx context.Context, pathID, courseID uuid.UUID, position int) error {
	return r.db.WithContext(ctx).Model(&domain.LearningPathCourse{}).
		Where("path_id = ? AND course_id = ?", pathID, courseID).
		Update("position", position).Error
}

func (r *learningPathRepository) GetPathCourses(ctx context.Context, pathID uuid.UUID) ([]domain.LearningPathCourse, error) {
	var courses []domain.LearningPathCourse
	err := r.db.WithContext(ctx).
		Preload("Course").
		Where("path_id = ?", pathID).
		Order("position ASC").
		Find(&courses).Error
	return courses, err
}

// Enrollments

func (r *learningPathRepository) Enroll(ctx context.Context, enrollment *domain.LearningPathEnrollment) error {
	err := r.db.WithContext(ctx).Create(enrollment).Error
	if err != nil {
		return err
	}
	// Update total students count
	return r.db.WithContext(ctx).Model(&domain.LearningPath{}).
		Where("id = ?", enrollment.PathID).
		Update("total_students", gorm.Expr("total_students + 1")).Error
}

func (r *learningPathRepository) GetEnrollment(ctx context.Context, pathID, userID uuid.UUID) (*domain.LearningPathEnrollment, error) {
	var enrollment domain.LearningPathEnrollment
	err := r.db.WithContext(ctx).
		Where("path_id = ? AND user_id = ?", pathID, userID).
		First(&enrollment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &enrollment, nil
}

func (r *learningPathRepository) UpdateEnrollment(ctx context.Context, enrollment *domain.LearningPathEnrollment) error {
	return r.db.WithContext(ctx).Save(enrollment).Error
}

func (r *learningPathRepository) GetUserEnrollments(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.LearningPathEnrollment, int64, error) {
	var enrollments []domain.LearningPathEnrollment
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.LearningPathEnrollment{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Path").
		Preload("Path.Category").
		Order("enrolled_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&enrollments).Error

	return enrollments, total, err
}

// Progress

func (r *learningPathRepository) GetProgress(ctx context.Context, pathID, userID uuid.UUID) (*domain.LearningPathProgress, error) {
	// Get path courses
	var pathCourses []domain.LearningPathCourse
	if err := r.db.WithContext(ctx).
		Preload("Course").
		Where("path_id = ?", pathID).
		Order("position ASC").
		Find(&pathCourses).Error; err != nil {
		return nil, err
	}

	// Get user's completed enrollments for these courses
	courseIDs := make([]uuid.UUID, len(pathCourses))
	for i, pc := range pathCourses {
		courseIDs[i] = pc.CourseID
	}

	var completedEnrollments []domain.Enrollment
	r.db.WithContext(ctx).
		Where("user_id = ? AND course_id IN ? AND status = ?", userID, courseIDs, domain.EnrollmentStatusCompleted).
		Find(&completedEnrollments)

	completedMap := make(map[uuid.UUID]bool)
	for _, e := range completedEnrollments {
		completedMap[e.CourseID] = true
	}

	// Build progress
	progress := &domain.LearningPathProgress{
		PathID:         pathID,
		UserID:         userID,
		TotalCourses:   len(pathCourses),
		CourseProgress: make([]domain.CourseProgressItem, len(pathCourses)),
	}

	for i, pc := range pathCourses {
		isCompleted := completedMap[pc.CourseID]
		if isCompleted {
			progress.CompletedCourses++
		}

		courseTitle := ""
		if pc.Course != nil {
			courseTitle = pc.Course.Title
		}

		progress.CourseProgress[i] = domain.CourseProgressItem{
			CourseID:    pc.CourseID,
			CourseTitle: courseTitle,
			Position:    pc.Position,
			IsRequired:  pc.IsRequired,
			IsCompleted: isCompleted,
			Progress:    0, // Could be enhanced to show partial progress
		}
		if isCompleted {
			progress.CourseProgress[i].Progress = 100
		}
	}

	if progress.TotalCourses > 0 {
		progress.Progress = float64(progress.CompletedCourses) / float64(progress.TotalCourses) * 100
	}

	return progress, nil
}

func (r *learningPathRepository) updatePathStats(ctx context.Context, pathID uuid.UUID) error {
	var count int64
	r.db.WithContext(ctx).Model(&domain.LearningPathCourse{}).
		Where("path_id = ?", pathID).
		Count(&count)

	// Calculate estimated hours
	var totalHours struct{ Sum int }
	r.db.WithContext(ctx).Model(&domain.LearningPathCourse{}).
		Select("COALESCE(SUM(courses.duration), 0) as sum").
		Joins("JOIN courses ON courses.id = learning_path_courses.course_id").
		Where("learning_path_courses.path_id = ?", pathID).
		Scan(&totalHours)

	return r.db.WithContext(ctx).Model(&domain.LearningPath{}).
		Where("id = ?", pathID).
		Updates(map[string]interface{}{
			"total_courses":   count,
			"estimated_hours": totalHours.Sum / 60, // Convert minutes to hours
		}).Error
}
