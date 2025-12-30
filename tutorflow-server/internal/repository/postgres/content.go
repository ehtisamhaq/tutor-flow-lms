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

// TutorProfileRepository
type tutorProfileRepository struct {
	db *gorm.DB
}

func NewTutorProfileRepository(db *gorm.DB) repository.TutorProfileRepository {
	return &tutorProfileRepository{db: db}
}

func (r *tutorProfileRepository) Create(ctx context.Context, profile *domain.TutorProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *tutorProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.TutorProfile, error) {
	var profile domain.TutorProfile
	err := r.db.WithContext(ctx).Preload("User").Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *tutorProfileRepository) Update(ctx context.Context, profile *domain.TutorProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *tutorProfileRepository) List(ctx context.Context, page, limit int) ([]domain.TutorProfile, int64, error) {
	var profiles []domain.TutorProfile
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.TutorProfile{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("User").Offset(offset).Limit(limit).Find(&profiles).Error
	return profiles, total, err
}

// CategoryRepository
type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) repository.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Category{}, "id = ?", id).Error
}

func (r *categoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetWithSubcategories(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category
	err := r.db.WithContext(ctx).
		Where("parent_id IS NULL").
		Preload("Subcategories").
		Order("sort_order ASC, name ASC").
		Find(&categories).Error
	return categories, err
}

// ModuleRepository
type moduleRepository struct {
	db *gorm.DB
}

func NewModuleRepository(db *gorm.DB) repository.ModuleRepository {
	return &moduleRepository{db: db}
}

func (r *moduleRepository) Create(ctx context.Context, module *domain.Module) error {
	return r.db.WithContext(ctx).Create(module).Error
}

func (r *moduleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Module, error) {
	var module domain.Module
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&module).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrModuleNotFound
		}
		return nil, err
	}
	return &module, nil
}

func (r *moduleRepository) Update(ctx context.Context, module *domain.Module) error {
	return r.db.WithContext(ctx).Save(module).Error
}

func (r *moduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Module{}, "id = ?", id).Error
}

func (r *moduleRepository) GetByCourse(ctx context.Context, courseID uuid.UUID) ([]domain.Module, error) {
	var modules []domain.Module
	err := r.db.WithContext(ctx).
		Where("course_id = ?", courseID).
		Order("sort_order ASC").
		Find(&modules).Error
	return modules, err
}

func (r *moduleRepository) Reorder(ctx context.Context, courseID uuid.UUID, moduleIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range moduleIDs {
			if err := tx.Model(&domain.Module{}).
				Where("id = ? AND course_id = ?", id, courseID).
				Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// LessonRepository
type lessonRepository struct {
	db *gorm.DB
}

func NewLessonRepository(db *gorm.DB) repository.LessonRepository {
	return &lessonRepository{db: db}
}

func (r *lessonRepository) Create(ctx context.Context, lesson *domain.Lesson) error {
	return r.db.WithContext(ctx).Create(lesson).Error
}

func (r *lessonRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lesson, error) {
	var lesson domain.Lesson
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&lesson).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLessonNotFound
		}
		return nil, err
	}
	return &lesson, nil
}

func (r *lessonRepository) Update(ctx context.Context, lesson *domain.Lesson) error {
	return r.db.WithContext(ctx).Save(lesson).Error
}

func (r *lessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Lesson{}, "id = ?", id).Error
}

func (r *lessonRepository) GetByModule(ctx context.Context, moduleID uuid.UUID) ([]domain.Lesson, error) {
	var lessons []domain.Lesson
	err := r.db.WithContext(ctx).
		Where("module_id = ?", moduleID).
		Order("sort_order ASC").
		Find(&lessons).Error
	return lessons, err
}

func (r *lessonRepository) Reorder(ctx context.Context, moduleID uuid.UUID, lessonIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range lessonIDs {
			if err := tx.Model(&domain.Lesson{}).
				Where("id = ? AND module_id = ?", id, moduleID).
				Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// EnrollmentRepository
type enrollmentRepository struct {
	db *gorm.DB
}

func NewEnrollmentRepository(db *gorm.DB) repository.EnrollmentRepository {
	return &enrollmentRepository{db: db}
}

func (r *enrollmentRepository) Create(ctx context.Context, enrollment *domain.Enrollment) error {
	return r.db.WithContext(ctx).Create(enrollment).Error
}

func (r *enrollmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Enrollment, error) {
	var enrollment domain.Enrollment
	err := r.db.WithContext(ctx).
		Preload("Course").
		Preload("User").
		Where("id = ?", id).
		First(&enrollment).Error
	if err != nil {
		return nil, err
	}
	return &enrollment, nil
}

func (r *enrollmentRepository) GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (*domain.Enrollment, error) {
	var enrollment domain.Enrollment
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		First(&enrollment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotEnrolled
		}
		return nil, err
	}
	return &enrollment, nil
}

func (r *enrollmentRepository) Update(ctx context.Context, enrollment *domain.Enrollment) error {
	return r.db.WithContext(ctx).Save(enrollment).Error
}

func (r *enrollmentRepository) List(ctx context.Context, filters repository.EnrollmentFilters) ([]domain.Enrollment, int64, error) {
	var enrollments []domain.Enrollment
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Enrollment{})

	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.CourseID != nil {
		query = query.Where("course_id = ?", *filters.CourseID)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filters.Page - 1) * filters.Limit
	err := query.
		Preload("Course").
		Preload("User").
		Order("enrolled_at DESC").
		Offset(offset).
		Limit(filters.Limit).
		Find(&enrollments).Error

	return enrollments, total, err
}

func (r *enrollmentRepository) GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Enrollment, int64, error) {
	var enrollments []domain.Enrollment
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Enrollment{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Course").
		Preload("Course.Instructor").
		Order("last_accessed_at DESC NULLS LAST, enrolled_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&enrollments).Error

	return enrollments, total, err
}

func (r *enrollmentRepository) GetByCourse(ctx context.Context, courseID uuid.UUID, page, limit int) ([]domain.Enrollment, int64, error) {
	var enrollments []domain.Enrollment
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Enrollment{}).Where("course_id = ?", courseID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Order("enrolled_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&enrollments).Error

	return enrollments, total, err
}

func (r *enrollmentRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress float64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Enrollment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"progress_percent": progress,
			"last_accessed_at": now,
		}).Error
}

// LessonProgressRepository
type lessonProgressRepository struct {
	db *gorm.DB
}

func NewLessonProgressRepository(db *gorm.DB) repository.LessonProgressRepository {
	return &lessonProgressRepository{db: db}
}

func (r *lessonProgressRepository) Upsert(ctx context.Context, progress *domain.LessonProgress) error {
	return r.db.WithContext(ctx).
		Where(domain.LessonProgress{EnrollmentID: progress.EnrollmentID, LessonID: progress.LessonID}).
		Assign(progress).
		FirstOrCreate(progress).Error
}

func (r *lessonProgressRepository) GetByEnrollmentAndLesson(ctx context.Context, enrollmentID, lessonID uuid.UUID) (*domain.LessonProgress, error) {
	var progress domain.LessonProgress
	err := r.db.WithContext(ctx).
		Where("enrollment_id = ? AND lesson_id = ?", enrollmentID, lessonID).
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *lessonProgressRepository) GetByEnrollment(ctx context.Context, enrollmentID uuid.UUID) ([]domain.LessonProgress, error) {
	var progress []domain.LessonProgress
	err := r.db.WithContext(ctx).
		Where("enrollment_id = ?", enrollmentID).
		Find(&progress).Error
	return progress, err
}

func (r *lessonProgressRepository) MarkComplete(ctx context.Context, enrollmentID, lessonID uuid.UUID) error {
	now := time.Now()
	progress := &domain.LessonProgress{
		EnrollmentID: enrollmentID,
		LessonID:     lessonID,
		IsCompleted:  true,
		CompletedAt:  &now,
	}
	return r.Upsert(ctx, progress)
}

func (r *lessonProgressRepository) UpdateVideoPosition(ctx context.Context, enrollmentID, lessonID uuid.UUID, position int) error {
	progress := &domain.LessonProgress{
		EnrollmentID:  enrollmentID,
		LessonID:      lessonID,
		VideoPosition: position,
	}
	return r.db.WithContext(ctx).
		Where(domain.LessonProgress{EnrollmentID: enrollmentID, LessonID: lessonID}).
		Assign(map[string]interface{}{"video_position": position}).
		FirstOrCreate(progress).Error
}
