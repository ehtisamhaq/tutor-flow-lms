package enrollment

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines enrollment business logic
type UseCase struct {
	enrollmentRepo   repository.EnrollmentRepository
	progressRepo     repository.LessonProgressRepository
	courseRepo       repository.CourseRepository
	notificationRepo repository.NotificationRepository
}

// NewUseCase creates a new enrollment use case
func NewUseCase(
	enrollmentRepo repository.EnrollmentRepository,
	progressRepo repository.LessonProgressRepository,
	courseRepo repository.CourseRepository,
	notificationRepo repository.NotificationRepository,
) *UseCase {
	return &UseCase{
		enrollmentRepo:   enrollmentRepo,
		progressRepo:     progressRepo,
		courseRepo:       courseRepo,
		notificationRepo: notificationRepo,
	}
}

// ListInput for listing enrollments
type ListInput struct {
	UserID   *uuid.UUID               `query:"user_id"`
	CourseID *uuid.UUID               `query:"course_id"`
	Status   *domain.EnrollmentStatus `query:"status"`
	Page     int                      `query:"page"`
	Limit    int                      `query:"limit"`
}

// List returns paginated enrollments
func (uc *UseCase) List(ctx context.Context, input ListInput) ([]domain.Enrollment, int64, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 100 {
		input.Limit = 20
	}

	return uc.enrollmentRepo.List(ctx, repository.EnrollmentFilters{
		UserID:   input.UserID,
		CourseID: input.CourseID,
		Status:   input.Status,
		Page:     input.Page,
		Limit:    input.Limit,
	})
}

// GetMyEnrollments returns current user's enrollments
func (uc *UseCase) GetMyEnrollments(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Enrollment, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 12
	}
	return uc.enrollmentRepo.GetByUser(ctx, userID, page, limit)
}

// GetByID returns enrollment by ID
func (uc *UseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Enrollment, error) {
	return uc.enrollmentRepo.GetByID(ctx, id)
}

// GetByUserAndCourse returns enrollment for user and course
func (uc *UseCase) GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (*domain.Enrollment, error) {
	return uc.enrollmentRepo.GetByUserAndCourse(ctx, userID, courseID)
}

// EnrollInput for enrolling in a course
type EnrollInput struct {
	CourseID uuid.UUID `json:"course_id" validate:"required"`
}

// Enroll enrolls a user in a course
func (uc *UseCase) Enroll(ctx context.Context, userID uuid.UUID, input EnrollInput) (*domain.Enrollment, error) {
	// Check if already enrolled
	existing, _ := uc.enrollmentRepo.GetByUserAndCourse(ctx, userID, input.CourseID)
	if existing != nil {
		return nil, domain.ErrAlreadyEnrolled
	}

	// Get course
	course, err := uc.courseRepo.GetByID(ctx, input.CourseID)
	if err != nil {
		return nil, domain.ErrCourseNotFound
	}

	if course.Status != domain.CourseStatusPublished {
		return nil, domain.ErrCourseNotPublished
	}

	// Determine enrollment status
	status := domain.EnrollmentStatusActive
	if course.Price > 0 {
		// Paid course requires payment - handled separately
		status = domain.EnrollmentStatusPending
	}

	enrollment := &domain.Enrollment{
		UserID:   userID,
		CourseID: input.CourseID,
		Status:   status,
	}

	if status == domain.EnrollmentStatusActive {
		now := time.Now()
		enrollment.StartedAt = &now
	}

	if err := uc.enrollmentRepo.Create(ctx, enrollment); err != nil {
		return nil, err
	}

	// Update course student count for free courses
	if status == domain.EnrollmentStatusActive {
		_ = uc.courseRepo.IncrementStudentCount(ctx, input.CourseID)
	}

	return enrollment, nil
}

// ActivateEnrollment activates a pending enrollment (after payment)
func (uc *UseCase) ActivateEnrollment(ctx context.Context, id uuid.UUID, orderID *uuid.UUID) error {
	enrollment, err := uc.enrollmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	enrollment.Status = domain.EnrollmentStatusActive
	enrollment.StartedAt = &now
	enrollment.OrderID = orderID

	if err := uc.enrollmentRepo.Update(ctx, enrollment); err != nil {
		return err
	}

	// Update course stats
	_ = uc.courseRepo.IncrementStudentCount(ctx, enrollment.CourseID)

	// Send notification
	_ = uc.notificationRepo.Create(ctx, &domain.Notification{
		UserID:  enrollment.UserID,
		Type:    domain.NotificationEnrollmentApproved,
		Title:   "Enrollment Confirmed",
		Message: stringPtr("Your enrollment has been confirmed. Start learning now!"),
	})

	return nil
}

// Cancel cancels an enrollment
func (uc *UseCase) Cancel(ctx context.Context, id uuid.UUID) error {
	enrollment, err := uc.enrollmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	enrollment.Status = domain.EnrollmentStatusCancelled
	return uc.enrollmentRepo.Update(ctx, enrollment)
}

// Complete marks enrollment as completed
func (uc *UseCase) Complete(ctx context.Context, id uuid.UUID) error {
	enrollment, err := uc.enrollmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	enrollment.Status = domain.EnrollmentStatusCompleted
	enrollment.CompletedAt = &now
	enrollment.ProgressPercent = 100

	return uc.enrollmentRepo.Update(ctx, enrollment)
}

// MarkLessonCompleteInput for marking lesson as complete
type MarkLessonCompleteInput struct {
	LessonID uuid.UUID `json:"lesson_id" validate:"required"`
}

// MarkLessonComplete marks a lesson as complete and updates progress
func (uc *UseCase) MarkLessonComplete(ctx context.Context, userID, courseID uuid.UUID, lessonID uuid.UUID) error {
	enrollment, err := uc.enrollmentRepo.GetByUserAndCourse(ctx, userID, courseID)
	if err != nil {
		return domain.ErrNotEnrolled
	}

	if !enrollment.CanAccess() {
		return domain.ErrEnrollmentExpired
	}

	// Mark lesson as complete
	if err := uc.progressRepo.MarkComplete(ctx, enrollment.ID, lessonID); err != nil {
		return err
	}

	// Calculate new progress
	progress, err := uc.calculateProgress(ctx, enrollment.ID, courseID)
	if err != nil {
		return err
	}

	// Update enrollment progress
	if err := uc.enrollmentRepo.UpdateProgress(ctx, enrollment.ID, progress); err != nil {
		return err
	}

	// Check if course is completed
	if progress >= 100 {
		return uc.Complete(ctx, enrollment.ID)
	}

	return nil
}

// UpdateVideoPositionInput for updating video position
type UpdateVideoPositionInput struct {
	LessonID uuid.UUID `json:"lesson_id" validate:"required"`
	Position int       `json:"position" validate:"gte=0"`
}

// UpdateVideoPosition updates video playback position
func (uc *UseCase) UpdateVideoPosition(ctx context.Context, userID, courseID uuid.UUID, input UpdateVideoPositionInput) error {
	enrollment, err := uc.enrollmentRepo.GetByUserAndCourse(ctx, userID, courseID)
	if err != nil {
		return domain.ErrNotEnrolled
	}

	if !enrollment.CanAccess() {
		return domain.ErrEnrollmentExpired
	}

	return uc.progressRepo.UpdateVideoPosition(ctx, enrollment.ID, input.LessonID, input.Position)
}

// GetProgress returns detailed progress for an enrollment
func (uc *UseCase) GetProgress(ctx context.Context, enrollmentID uuid.UUID) ([]domain.LessonProgress, error) {
	return uc.progressRepo.GetByEnrollment(ctx, enrollmentID)
}

// calculateProgress calculates completion percentage
func (uc *UseCase) calculateProgress(ctx context.Context, enrollmentID, courseID uuid.UUID) (float64, error) {
	// Get all lessons in course
	course, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return 0, err
	}

	totalLessons := course.TotalLessons
	if totalLessons == 0 {
		return 100, nil
	}

	// Get completed lessons
	progress, err := uc.progressRepo.GetByEnrollment(ctx, enrollmentID)
	if err != nil {
		return 0, err
	}

	completedCount := 0
	for _, p := range progress {
		if p.IsCompleted {
			completedCount++
		}
	}

	return float64(completedCount) / float64(totalLessons) * 100, nil
}

// CanAccessLesson checks if user can access a specific lesson
func (uc *UseCase) CanAccessLesson(ctx context.Context, userID, lessonID uuid.UUID) (bool, error) {
	// TODO: Get lesson and check access type
	// For now, check if enrolled in the lesson's course
	return true, nil
}

func stringPtr(s string) *string {
	return &s
}
