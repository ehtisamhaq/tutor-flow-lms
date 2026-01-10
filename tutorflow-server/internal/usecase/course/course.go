package course

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines course management business logic
type UseCase struct {
	courseRepo     repository.CourseRepository
	categoryRepo   repository.CategoryRepository
	moduleRepo     repository.ModuleRepository
	lessonRepo     repository.LessonRepository
	enrollmentRepo repository.EnrollmentRepository
}

// NewUseCase creates a new course use case
func NewUseCase(
	courseRepo repository.CourseRepository,
	categoryRepo repository.CategoryRepository,
	moduleRepo repository.ModuleRepository,
	lessonRepo repository.LessonRepository,
	enrollmentRepo repository.EnrollmentRepository,
) *UseCase {
	return &UseCase{
		courseRepo:     courseRepo,
		categoryRepo:   categoryRepo,
		moduleRepo:     moduleRepo,
		lessonRepo:     lessonRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

// ListInput for listing courses
type ListInput struct {
	Status     *domain.CourseStatus `query:"status"`
	Level      *domain.CourseLevel  `query:"level"`
	CategoryID *uuid.UUID           `query:"category_id"`
	IsFeatured *bool                `query:"is_featured"`
	Search     string               `query:"search"`
	MinPrice   *float64             `query:"min_price"`
	MaxPrice   *float64             `query:"max_price"`
	MinRating  *float64             `query:"min_rating"`
	SortBy     string               `query:"sort_by"`
	SortOrder  string               `query:"sort_order"`
	Page       int                  `query:"page"`
	Limit      int                  `query:"limit"`
}

// List returns paginated courses
func (uc *UseCase) List(ctx context.Context, input ListInput, isPublicOnly bool) ([]domain.Course, int64, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 50 {
		input.Limit = 12
	}

	filters := repository.CourseFilters{
		Level:      input.Level,
		CategoryID: input.CategoryID,
		IsFeatured: input.IsFeatured,
		Search:     input.Search,
		MinPrice:   input.MinPrice,
		MaxPrice:   input.MaxPrice,
		MinRating:  input.MinRating,
		SortBy:     input.SortBy,
		SortOrder:  input.SortOrder,
		Page:       input.Page,
		Limit:      input.Limit,
	}

	if isPublicOnly {
		published := domain.CourseStatusPublished
		filters.Status = &published
	} else if input.Status != nil {
		filters.Status = input.Status
	}

	return uc.courseRepo.List(ctx, filters)
}

// GetByID returns a course by ID
func (uc *UseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Course, error) {
	return uc.courseRepo.GetByID(ctx, id)
}

// GetBySlug returns a course by slug
func (uc *UseCase) GetBySlug(ctx context.Context, slugStr string) (*domain.Course, error) {
	return uc.courseRepo.GetBySlug(ctx, slugStr)
}

// GetByInstructor returns courses by instructor
func (uc *UseCase) GetByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]domain.Course, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 12
	}
	return uc.courseRepo.GetByInstructor(ctx, instructorID, page, limit)
}

// CreateInput for creating a course
type CreateInput struct {
	Title            string        `json:"title" form:"title" validate:"required,min=5,max=255"`
	Description      *string       `json:"description" form:"description"`
	ShortDescription *string       `json:"short_description" form:"short_description" validate:"omitempty,max=500"`
	ThumbnailURL     *string       `json:"thumbnail_url" form:"thumbnail_url"`
	Level            string        `json:"level" form:"level" validate:"required,oneof=beginner intermediate advanced"`
	Price            float64       `json:"price" form:"price" validate:"gte=0"`
	DiscountPrice    *float64      `json:"discount_price" form:"discount_price" validate:"omitempty,gte=0"`
	CategoryIDs      []string      `json:"category_ids" form:"category_ids"`
	CategoryID       *string       `json:"-" form:"category_id"`
	Requirements     []string      `json:"requirements" form:"requirements"`
	WhatYouLearn     []string      `json:"what_you_learn" form:"what_you_learn"`
	Language         string        `json:"language" form:"language"`
	Modules          []ModuleInput `json:"modules" form:"-"`
}

type ModuleInput struct {
	ID          string        `json:"id"`
	Title       string        `json:"title" validate:"required,min=3"`
	Description *string       `json:"description"`
	Order       int           `json:"order"`
	Lessons     []LessonInput `json:"lessons"`
}

type LessonInput struct {
	ID              string  `json:"id"`
	Title           string  `json:"title" validate:"required,min=3"`
	Description     *string `json:"description"`
	Content         *string `json:"content"`
	Type            string  `json:"type" validate:"required,oneof=video text quiz assignment"`
	Order           int     `json:"order"`
	DurationMinutes int     `json:"duration_minutes"`
	VideoURL        *string `json:"video_url"`
}

// Create creates a new course
func (uc *UseCase) Create(ctx context.Context, instructorID uuid.UUID, input CreateInput) (*domain.Course, error) {
	courseSlug := slug.Make(input.Title)

	// Handle both plural and singular categories for flexibility
	categoryIDs := input.CategoryIDs
	if input.CategoryID != nil && *input.CategoryID != "" {
		categoryIDs = append(categoryIDs, *input.CategoryID)
	}

	course := &domain.Course{
		Title:            input.Title,
		Slug:             courseSlug,
		Description:      input.Description,
		ShortDescription: input.ShortDescription,
		ThumbnailURL:     input.ThumbnailURL,
		InstructorID:     instructorID,
		Status:           domain.CourseStatusDraft,
		Level:            domain.CourseLevel(input.Level),
		Price:            input.Price,
		DiscountPrice:    input.DiscountPrice,
		Requirements:     input.Requirements,
		WhatYouLearn:     input.WhatYouLearn,
		Language:         input.Language,
	}

	if course.Language == "" {
		course.Language = "English"
	}

	if err := uc.courseRepo.Create(ctx, course); err != nil {
		return nil, err
	}

	// Save curriculum if provided
	if len(input.Modules) > 0 {
		if err := uc.saveCurriculum(ctx, course.ID, input.Modules); err != nil {
			return course, err // Return course anyway but with error
		}
	}

	return course, nil
}

// UpdateInput for updating a course
type UpdateInput struct {
	Title            *string       `json:"title" form:"title" validate:"omitempty,min=5,max=255"`
	Description      *string       `json:"description" form:"description"`
	ShortDescription *string       `json:"short_description" form:"short_description" validate:"omitempty,max=500"`
	ThumbnailURL     *string       `json:"thumbnail_url" form:"thumbnail_url" validate:"omitempty,url"`
	PreviewVideoURL  *string       `json:"preview_video_url" form:"preview_video_url" validate:"omitempty,url"`
	Level            *string       `json:"level" form:"level" validate:"omitempty,oneof=beginner intermediate advanced"`
	Price            *float64      `json:"price" form:"price" validate:"omitempty,gte=0"`
	DiscountPrice    *float64      `json:"discount_price" form:"discount_price" validate:"omitempty,gte=0"`
	Requirements     []string      `json:"requirements" form:"requirements"`
	WhatYouLearn     []string      `json:"what_you_learn" form:"what_you_learn"`
	Language         *string       `json:"language" form:"language"`
	IsFeatured       *bool         `json:"is_featured" form:"is_featured"`
	Modules          []ModuleInput `json:"modules" form:"-"`
}

// Update updates a course
func (uc *UseCase) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Course, error) {
	course, err := uc.courseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		course.Title = *input.Title
		course.Slug = slug.Make(*input.Title)
	}
	if input.Description != nil {
		course.Description = input.Description
	}
	if input.ShortDescription != nil {
		course.ShortDescription = input.ShortDescription
	}
	if input.ThumbnailURL != nil {
		course.ThumbnailURL = input.ThumbnailURL
	}
	if input.PreviewVideoURL != nil {
		course.PreviewVideoURL = input.PreviewVideoURL
	}
	if input.Level != nil {
		course.Level = domain.CourseLevel(*input.Level)
	}
	if input.Price != nil {
		course.Price = *input.Price
	}
	if input.DiscountPrice != nil {
		course.DiscountPrice = input.DiscountPrice
	}
	if input.Requirements != nil {
		course.Requirements = input.Requirements
	}
	if input.WhatYouLearn != nil {
		course.WhatYouLearn = input.WhatYouLearn
	}
	if input.Language != nil {
		course.Language = *input.Language
	}
	// Prevent GORM from re-saving old modules that we want to replace
	course.Modules = nil

	if err := uc.courseRepo.Update(ctx, course); err != nil {
		return nil, err
	}

	// Save curriculum if provided (this handles deletions, creations, and stats updates)
	if len(input.Modules) > 0 {
		if err := uc.saveCurriculum(ctx, course.ID, input.Modules); err != nil {
			return nil, err
		}
	}

	// Fetch fresh course from DB to return accurate data
	return uc.courseRepo.GetByID(ctx, course.ID)
}

func (uc *UseCase) saveCurriculum(ctx context.Context, courseID uuid.UUID, modules []ModuleInput) error {
	f, _ := os.OpenFile("/tmp/course_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		defer f.Close()
		fmt.Fprintf(f, "[%v] saveCurriculum: started for courseID=%s, received %d modules\n", time.Now().Format("15:04:05"), courseID, len(modules))
	}

	oldModules, err := uc.moduleRepo.GetByCourse(ctx, courseID)
	if err == nil {
		if f != nil {
			fmt.Fprintf(f, "[%v] saveCurriculum: found %d old modules to delete\n", time.Now().Format("15:04:05"), len(oldModules))
		}
		for _, m := range oldModules {
			if err := uc.lessonRepo.DeleteByModule(ctx, m.ID); err != nil {
				if f != nil {
					fmt.Fprintf(f, "[%v] saveCurriculum: FAILED deleting lessons for module %s: %v\n", time.Now().Format("15:04:05"), m.ID, err)
				}
				return err
			}
			if err := uc.moduleRepo.Delete(ctx, m.ID); err != nil {
				if f != nil {
					fmt.Fprintf(f, "[%v] saveCurriculum: FAILED deleting module %s: %v\n", time.Now().Format("15:04:05"), m.ID, err)
				}
				return err
			}
		}
	} else {
		if f != nil {
			fmt.Fprintf(f, "[%v] saveCurriculum: error fetching old modules: %v\n", time.Now().Format("15:04:05"), err)
		}
	}

	for i, mInput := range modules {
		moduleID := uuid.New()
		module := &domain.Module{
			ID:          moduleID,
			CourseID:    courseID,
			Title:       mInput.Title,
			Description: mInput.Description,
			SortOrder:   i,
			IsPublished: true,
		}

		if err := uc.moduleRepo.Create(ctx, module); err != nil {
			if f != nil {
				fmt.Fprintf(f, "[%v] saveCurriculum: FAILED creating module %d: %v\n", time.Now().Format("15:04:05"), i, err)
			}
			return err
		}

		for j, lInput := range mInput.Lessons {
			durationSec := lInput.DurationMinutes * 60
			lesson := &domain.Lesson{
				ID:            uuid.New(),
				ModuleID:      moduleID,
				Title:         lInput.Title,
				Description:   lInput.Description,
				Content:       lInput.Content,
				LessonType:    domain.LessonType(lInput.Type),
				SortOrder:     j,
				VideoURL:      lInput.VideoURL,
				VideoDuration: &durationSec,
				IsPublished:   true,
			}
			if err := uc.lessonRepo.Create(ctx, lesson); err != nil {
				if f != nil {
					fmt.Fprintf(f, "[%v] saveCurriculum: FAILED creating lesson %d in module %d: %v\n", time.Now().Format("15:04:05"), j, i, err)
				}
				return err
			}
		}
	}

	if f != nil {
		fmt.Fprintf(f, "[%v] saveCurriculum: COMPLETED successfully\n", time.Now().Format("15:04:05"))
	}
	return uc.courseRepo.UpdateStats(ctx, courseID)
}

// ValidateOwnership checks if user owns the course
func (uc *UseCase) ValidateOwnership(ctx context.Context, courseID, userID uuid.UUID) error {
	course, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return err
	}

	if course.InstructorID != userID {
		return domain.ErrNotCourseOwner
	}

	return nil
}

// Publish publishes a course
func (uc *UseCase) Publish(ctx context.Context, id uuid.UUID) error {
	course, err := uc.courseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate course has content
	modules, err := uc.moduleRepo.GetByCourse(ctx, id)
	if err != nil {
		return err
	}
	if len(modules) == 0 {
		return domain.ErrContentLocked
	}

	course.Status = domain.CourseStatusPublished
	return uc.courseRepo.Update(ctx, course)
}

// Archive archives a course
func (uc *UseCase) Archive(ctx context.Context, id uuid.UUID) error {
	course, err := uc.courseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	course.Status = domain.CourseStatusArchived
	return uc.courseRepo.Update(ctx, course)
}

// Delete soft-deletes a course
func (uc *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.courseRepo.Delete(ctx, id)
}

// GetCurriculum returns full course curriculum with modules and lessons
func (uc *UseCase) GetCurriculum(ctx context.Context, courseID uuid.UUID) ([]domain.Module, error) {
	modules, err := uc.moduleRepo.GetByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	for i := range modules {
		lessons, err := uc.lessonRepo.GetByModule(ctx, modules[i].ID)
		if err != nil {
			return nil, err
		}
		modules[i].Lessons = lessons
	}

	return modules, nil
}

// CanAccessCourse checks if user can access course content
func (uc *UseCase) CanAccessCourse(ctx context.Context, courseID, userID uuid.UUID) (bool, error) {
	course, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return false, err
	}

	// Owner always has access
	if course.InstructorID == userID {
		return true, nil
	}

	// Check enrollment
	enrollment, err := uc.enrollmentRepo.GetByUserAndCourse(ctx, userID, courseID)
	if err != nil {
		return false, nil
	}

	return enrollment.CanAccess(), nil
}

// --- Module Management ---

// CreateModuleInput for creating a module
type CreateModuleInput struct {
	Title       string  `json:"title" validate:"required,min=3,max=255"`
	Description *string `json:"description"`
}

// CreateModule creates a new module
func (uc *UseCase) CreateModule(ctx context.Context, courseID uuid.UUID, input CreateModuleInput) (*domain.Module, error) {
	// Get current max sort order
	modules, _ := uc.moduleRepo.GetByCourse(ctx, courseID)
	sortOrder := len(modules)

	module := &domain.Module{
		CourseID:    courseID,
		Title:       input.Title,
		Description: input.Description,
		SortOrder:   sortOrder,
	}

	if err := uc.moduleRepo.Create(ctx, module); err != nil {
		return nil, err
	}

	return module, nil
}

// UpdateModuleInput for updating a module
type UpdateModuleInput struct {
	Title       *string `json:"title" validate:"omitempty,min=3,max=255"`
	Description *string `json:"description"`
	IsPublished *bool   `json:"is_published"`
}

// UpdateModule updates a module
func (uc *UseCase) UpdateModule(ctx context.Context, id uuid.UUID, input UpdateModuleInput) (*domain.Module, error) {
	module, err := uc.moduleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		module.Title = *input.Title
	}
	if input.Description != nil {
		module.Description = input.Description
	}
	if input.IsPublished != nil {
		module.IsPublished = *input.IsPublished
	}

	if err := uc.moduleRepo.Update(ctx, module); err != nil {
		return nil, err
	}

	return module, nil
}

// DeleteModule deletes a module
func (uc *UseCase) DeleteModule(ctx context.Context, id uuid.UUID) error {
	// First delete all lessons in this module
	if err := uc.lessonRepo.DeleteByModule(ctx, id); err != nil {
		return fmt.Errorf("failed to delete module lessons: %w", err)
	}

	return uc.moduleRepo.Delete(ctx, id)
}

// ReorderModules reorders modules
func (uc *UseCase) ReorderModules(ctx context.Context, courseID uuid.UUID, moduleIDs []uuid.UUID) error {
	return uc.moduleRepo.Reorder(ctx, courseID, moduleIDs)
}

// --- Lesson Management ---

// CreateLessonInput for creating a lesson
type CreateLessonInput struct {
	Title       string  `json:"title" validate:"required,min=3,max=255"`
	Description *string `json:"description"`
	Content     *string `json:"content"`
	LessonType  string  `json:"lesson_type" validate:"required,oneof=video text quiz assignment resource"`
	AccessType  string  `json:"access_type" validate:"required,oneof=free enrolled premium"`
	VideoURL    *string `json:"video_url" validate:"omitempty,url"`
	IsPreview   bool    `json:"is_preview"`
}

// CreateLesson creates a new lesson
func (uc *UseCase) CreateLesson(ctx context.Context, moduleID uuid.UUID, input CreateLessonInput) (*domain.Lesson, error) {
	// Get current max sort order
	lessons, _ := uc.lessonRepo.GetByModule(ctx, moduleID)
	sortOrder := len(lessons)

	lesson := &domain.Lesson{
		ModuleID:    moduleID,
		Title:       input.Title,
		Description: input.Description,
		Content:     input.Content,
		LessonType:  domain.LessonType(input.LessonType),
		AccessType:  domain.ContentAccess(input.AccessType),
		VideoURL:    input.VideoURL,
		IsPreview:   input.IsPreview,
		SortOrder:   sortOrder,
	}

	if err := uc.lessonRepo.Create(ctx, lesson); err != nil {
		return nil, err
	}

	// Update course stats
	module, _ := uc.moduleRepo.GetByID(ctx, moduleID)
	if module != nil {
		_ = uc.courseRepo.UpdateStats(ctx, module.CourseID)
	}

	return lesson, nil
}

// UpdateLessonInput for updating a lesson
type UpdateLessonInput struct {
	Title       *string `json:"title" validate:"omitempty,min=3,max=255"`
	Description *string `json:"description"`
	Content     *string `json:"content"`
	VideoURL    *string `json:"video_url" validate:"omitempty,url"`
	AccessType  *string `json:"access_type" validate:"omitempty,oneof=free enrolled premium"`
	IsPublished *bool   `json:"is_published"`
	IsPreview   *bool   `json:"is_preview"`
}

// UpdateLesson updates a lesson
func (uc *UseCase) UpdateLesson(ctx context.Context, id uuid.UUID, input UpdateLessonInput) (*domain.Lesson, error) {
	lesson, err := uc.lessonRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		lesson.Title = *input.Title
	}
	if input.Description != nil {
		lesson.Description = input.Description
	}
	if input.Content != nil {
		lesson.Content = input.Content
	}
	if input.VideoURL != nil {
		lesson.VideoURL = input.VideoURL
	}
	if input.AccessType != nil {
		lesson.AccessType = domain.ContentAccess(*input.AccessType)
	}
	if input.IsPublished != nil {
		lesson.IsPublished = *input.IsPublished
	}
	if input.IsPreview != nil {
		lesson.IsPreview = *input.IsPreview
	}

	if err := uc.lessonRepo.Update(ctx, lesson); err != nil {
		return nil, err
	}

	return lesson, nil
}

// DeleteLesson deletes a lesson
func (uc *UseCase) DeleteLesson(ctx context.Context, id uuid.UUID) error {
	lesson, err := uc.lessonRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.lessonRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Update course stats
	module, _ := uc.moduleRepo.GetByID(ctx, lesson.ModuleID)
	if module != nil {
		_ = uc.courseRepo.UpdateStats(ctx, module.CourseID)
	}

	return nil
}

// GetLesson returns a lesson by ID
func (uc *UseCase) GetLesson(ctx context.Context, id uuid.UUID) (*domain.Lesson, error) {
	return uc.lessonRepo.GetByID(ctx, id)
}

// --- Category Management ---

// ListCategories returns all categories
func (uc *UseCase) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return uc.categoryRepo.GetWithSubcategories(ctx)
}

// CreateCategoryInput for creating a category
type CreateCategoryInput struct {
	Name        string     `json:"name" validate:"required,min=2,max=100"`
	Description *string    `json:"description"`
	Icon        *string    `json:"icon"`
	ParentID    *uuid.UUID `json:"parent_id"`
}

// CreateCategory creates a new category (admin only)
func (uc *UseCase) CreateCategory(ctx context.Context, input CreateCategoryInput) (*domain.Category, error) {
	category := &domain.Category{
		Name:        input.Name,
		Slug:        slug.Make(input.Name),
		Description: input.Description,
		Icon:        input.Icon,
		ParentID:    input.ParentID,
	}

	if err := uc.categoryRepo.Create(ctx, category); err != nil {
		// Handle duplicate slug
		if strings.Contains(err.Error(), "duplicate") {
			category.Slug = slug.Make(input.Name + "-" + uuid.New().String()[:4])
			if err := uc.categoryRepo.Create(ctx, category); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return category, nil
}

// UpdateCategory updates a category
func (uc *UseCase) UpdateCategory(ctx context.Context, id uuid.UUID, input CreateCategoryInput) (*domain.Category, error) {
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	category.Name = input.Name
	category.Slug = slug.Make(input.Name)
	category.Description = input.Description
	category.Icon = input.Icon
	category.ParentID = input.ParentID

	if err := uc.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteCategory deletes a category
func (uc *UseCase) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return uc.categoryRepo.Delete(ctx, id)
}
