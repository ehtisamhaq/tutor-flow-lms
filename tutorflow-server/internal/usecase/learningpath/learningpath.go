package learningpath

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines learning path business logic
type UseCase struct {
	pathRepo       repository.LearningPathRepository
	enrollmentRepo repository.EnrollmentRepository
	certRepo       repository.CertificateRepository
}

// NewUseCase creates a new learning path use case
func NewUseCase(
	pathRepo repository.LearningPathRepository,
	enrollmentRepo repository.EnrollmentRepository,
	certRepo repository.CertificateRepository,
) *UseCase {
	return &UseCase{
		pathRepo:       pathRepo,
		enrollmentRepo: enrollmentRepo,
		certRepo:       certRepo,
	}
}

// GetPath returns a learning path by ID
func (uc *UseCase) GetPath(ctx context.Context, id uuid.UUID) (*domain.LearningPath, error) {
	return uc.pathRepo.GetByID(ctx, id)
}

// GetPathBySlug returns a learning path by slug
func (uc *UseCase) GetPathBySlug(ctx context.Context, pathSlug string) (*domain.LearningPath, error) {
	return uc.pathRepo.GetBySlug(ctx, pathSlug)
}

// ListPaths returns learning paths with filters
func (uc *UseCase) ListPaths(ctx context.Context, filters repository.LearningPathFilters) ([]domain.LearningPath, int64, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 50 {
		filters.Limit = 20
	}
	return uc.pathRepo.List(ctx, filters)
}

// GetFeaturedPaths returns featured learning paths
func (uc *UseCase) GetFeaturedPaths(ctx context.Context, limit int) ([]domain.LearningPath, error) {
	if limit < 1 || limit > 20 {
		limit = 6
	}
	return uc.pathRepo.GetFeatured(ctx, limit)
}

// GetPathsByCategory returns paths in a category
func (uc *UseCase) GetPathsByCategory(ctx context.Context, categoryID uuid.UUID, limit int) ([]domain.LearningPath, error) {
	if limit < 1 || limit > 20 {
		limit = 10
	}
	return uc.pathRepo.GetByCategory(ctx, categoryID, limit)
}

// CreatePathInput for creating a path
type CreatePathInput struct {
	Title            string     `json:"title" validate:"required,min=5,max=255"`
	Description      string     `json:"description" validate:"omitempty,max=5000"`
	ShortDescription string     `json:"short_description" validate:"omitempty,max=500"`
	ThumbnailURL     *string    `json:"thumbnail_url,omitempty"`
	CategoryID       *uuid.UUID `json:"category_id,omitempty"`
	Level            string     `json:"level" validate:"omitempty,oneof=beginner intermediate advanced"`
	IsPublished      bool       `json:"is_published,omitempty"`
	IsFeatured       bool       `json:"is_featured,omitempty"`
}

// CreatePath creates a new learning path
func (uc *UseCase) CreatePath(ctx context.Context, creatorID uuid.UUID, isAdmin bool, input CreatePathInput) (*domain.LearningPath, error) {
	if !isAdmin {
		return nil, fmt.Errorf("only admins can create learning paths")
	}

	pathSlug := slug.Make(input.Title)

	path := &domain.LearningPath{
		Title:            input.Title,
		Slug:             pathSlug,
		Description:      input.Description,
		ShortDescription: input.ShortDescription,
		ThumbnailURL:     input.ThumbnailURL,
		CategoryID:       input.CategoryID,
		Level:            input.Level,
		IsPublished:      input.IsPublished,
		IsFeatured:       input.IsFeatured,
		CreatedBy:        creatorID,
	}

	if path.Level == "" {
		path.Level = "beginner"
	}

	if err := uc.pathRepo.Create(ctx, path); err != nil {
		// Handle duplicate slug
		if strings.Contains(err.Error(), "duplicate") {
			path.Slug = pathSlug + "-" + uuid.New().String()[:8]
			if err := uc.pathRepo.Create(ctx, path); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return uc.pathRepo.GetByID(ctx, path.ID)
}

// UpdatePathInput for updating a path
type UpdatePathInput struct {
	Title            *string    `json:"title,omitempty" validate:"omitempty,min=5,max=255"`
	Description      *string    `json:"description,omitempty"`
	ShortDescription *string    `json:"short_description,omitempty"`
	ThumbnailURL     *string    `json:"thumbnail_url,omitempty"`
	CategoryID       *uuid.UUID `json:"category_id,omitempty"`
	Level            *string    `json:"level,omitempty"`
	IsPublished      *bool      `json:"is_published,omitempty"`
	IsFeatured       *bool      `json:"is_featured,omitempty"`
}

// UpdatePath updates a learning path
func (uc *UseCase) UpdatePath(ctx context.Context, pathID uuid.UUID, isAdmin bool, input UpdatePathInput) (*domain.LearningPath, error) {
	if !isAdmin {
		return nil, fmt.Errorf("only admins can update learning paths")
	}

	path, err := uc.pathRepo.GetByID(ctx, pathID)
	if err != nil || path == nil {
		return nil, fmt.Errorf("path not found")
	}

	if input.Title != nil {
		path.Title = *input.Title
		path.Slug = slug.Make(*input.Title)
	}
	if input.Description != nil {
		path.Description = *input.Description
	}
	if input.ShortDescription != nil {
		path.ShortDescription = *input.ShortDescription
	}
	if input.ThumbnailURL != nil {
		path.ThumbnailURL = input.ThumbnailURL
	}
	if input.CategoryID != nil {
		path.CategoryID = input.CategoryID
	}
	if input.Level != nil {
		path.Level = *input.Level
	}
	if input.IsPublished != nil {
		path.IsPublished = *input.IsPublished
	}
	if input.IsFeatured != nil {
		path.IsFeatured = *input.IsFeatured
	}

	if err := uc.pathRepo.Update(ctx, path); err != nil {
		return nil, err
	}

	return path, nil
}

// DeletePath deletes a learning path
func (uc *UseCase) DeletePath(ctx context.Context, pathID uuid.UUID, isAdmin bool) error {
	if !isAdmin {
		return fmt.Errorf("only admins can delete learning paths")
	}
	return uc.pathRepo.Delete(ctx, pathID)
}

// AddCourseInput for adding a course to a path
type AddCourseInput struct {
	CourseID    uuid.UUID `json:"course_id" validate:"required"`
	Position    int       `json:"position"`
	IsRequired  bool      `json:"is_required"`
	Description *string   `json:"description,omitempty"`
}

// AddCourse adds a course to a learning path
func (uc *UseCase) AddCourse(ctx context.Context, pathID uuid.UUID, isAdmin bool, input AddCourseInput) error {
	if !isAdmin {
		return fmt.Errorf("only admins can modify learning paths")
	}

	pathCourse := &domain.LearningPathCourse{
		PathID:      pathID,
		CourseID:    input.CourseID,
		Position:    input.Position,
		IsRequired:  input.IsRequired,
		Description: input.Description,
	}

	return uc.pathRepo.AddCourse(ctx, pathCourse)
}

// RemoveCourse removes a course from a learning path
func (uc *UseCase) RemoveCourse(ctx context.Context, pathID, courseID uuid.UUID, isAdmin bool) error {
	if !isAdmin {
		return fmt.Errorf("only admins can modify learning paths")
	}
	return uc.pathRepo.RemoveCourse(ctx, pathID, courseID)
}

// ReorderCourses reorders courses in a path
func (uc *UseCase) ReorderCourses(ctx context.Context, pathID uuid.UUID, isAdmin bool, courseOrder []uuid.UUID) error {
	if !isAdmin {
		return fmt.Errorf("only admins can modify learning paths")
	}

	for i, courseID := range courseOrder {
		if err := uc.pathRepo.UpdateCoursePosition(ctx, pathID, courseID, i); err != nil {
			return err
		}
	}
	return nil
}

// GetPathCourses returns courses in a path
func (uc *UseCase) GetPathCourses(ctx context.Context, pathID uuid.UUID) ([]domain.LearningPathCourse, error) {
	return uc.pathRepo.GetPathCourses(ctx, pathID)
}

// EnrollInPath enrolls a user in a learning path
func (uc *UseCase) EnrollInPath(ctx context.Context, pathID, userID uuid.UUID) (*domain.LearningPathEnrollment, error) {
	// Check if already enrolled
	existing, _ := uc.pathRepo.GetEnrollment(ctx, pathID, userID)
	if existing != nil {
		return existing, nil
	}

	// Check path exists and is published
	path, err := uc.pathRepo.GetByID(ctx, pathID)
	if err != nil || path == nil {
		return nil, fmt.Errorf("path not found")
	}
	if !path.IsPublished {
		return nil, fmt.Errorf("path is not available")
	}

	enrollment := &domain.LearningPathEnrollment{
		PathID: pathID,
		UserID: userID,
		Status: domain.PathEnrollmentActive,
	}

	if err := uc.pathRepo.Enroll(ctx, enrollment); err != nil {
		return nil, err
	}

	return enrollment, nil
}

// GetMyEnrollments returns user's learning path enrollments
func (uc *UseCase) GetMyEnrollments(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.LearningPathEnrollment, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.pathRepo.GetUserEnrollments(ctx, userID, page, limit)
}

// GetProgress returns user's progress in a learning path
func (uc *UseCase) GetProgress(ctx context.Context, pathID, userID uuid.UUID) (*domain.LearningPathProgress, error) {
	// Verify enrollment
	enrollment, _ := uc.pathRepo.GetEnrollment(ctx, pathID, userID)
	if enrollment == nil {
		return nil, fmt.Errorf("not enrolled in this path")
	}

	progress, err := uc.pathRepo.GetProgress(ctx, pathID, userID)
	if err != nil {
		return nil, err
	}

	// Check for completion and issue certificate
	if progress.Progress >= 100 && enrollment.Status != domain.PathEnrollmentCompleted {
		now := time.Now()
		enrollment.Status = domain.PathEnrollmentCompleted
		enrollment.CompletedAt = &now
		enrollment.Progress = 100
		_ = uc.pathRepo.UpdateEnrollment(ctx, enrollment)
	} else {
		enrollment.Progress = progress.Progress
		_ = uc.pathRepo.UpdateEnrollment(ctx, enrollment)
	}

	return progress, nil
}
