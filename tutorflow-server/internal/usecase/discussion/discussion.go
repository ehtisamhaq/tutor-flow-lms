package discussion

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines discussion business logic
type UseCase struct {
	discussionRepo repository.DiscussionRepository
	enrollmentRepo repository.EnrollmentRepository
	courseRepo     repository.CourseRepository
}

// NewUseCase creates a new discussion use case
func NewUseCase(
	discussionRepo repository.DiscussionRepository,
	enrollmentRepo repository.EnrollmentRepository,
	courseRepo repository.CourseRepository,
) *UseCase {
	return &UseCase{
		discussionRepo: discussionRepo,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
	}
}

// GetDiscussion returns a discussion with replies
func (uc *UseCase) GetDiscussion(ctx context.Context, id uuid.UUID) (*domain.Discussion, error) {
	return uc.discussionRepo.GetByID(ctx, id)
}

// GetCourseDiscussions returns discussions for a course
func (uc *UseCase) GetCourseDiscussions(ctx context.Context, courseID uuid.UUID, page, limit int) ([]domain.Discussion, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.discussionRepo.GetByCourse(ctx, courseID, page, limit)
}

// GetLessonDiscussions returns Q&A for a specific lesson
func (uc *UseCase) GetLessonDiscussions(ctx context.Context, lessonID uuid.UUID, page, limit int) ([]domain.Discussion, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.discussionRepo.GetByLesson(ctx, lessonID, page, limit)
}

// GetReplies returns replies for a discussion
func (uc *UseCase) GetReplies(ctx context.Context, discussionID uuid.UUID, page, limit int) ([]domain.Discussion, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.discussionRepo.GetReplies(ctx, discussionID, page, limit)
}

// CreateDiscussionInput for creating a discussion
type CreateDiscussionInput struct {
	CourseID uuid.UUID  `json:"course_id" validate:"required"`
	LessonID *uuid.UUID `json:"lesson_id"`
	ParentID *uuid.UUID `json:"parent_id"`
	Content  string     `json:"content" validate:"required,min=5,max=5000"`
}

// CreateDiscussion creates a new discussion or reply
func (uc *UseCase) CreateDiscussion(ctx context.Context, userID uuid.UUID, input CreateDiscussionInput) (*domain.Discussion, error) {
	// Verify user is enrolled in the course
	enrollment, err := uc.enrollmentRepo.GetByUserAndCourse(ctx, userID, input.CourseID)
	if err != nil || enrollment == nil {
		// Check if user is the instructor
		course, _ := uc.courseRepo.GetByID(ctx, input.CourseID)
		if course == nil || course.InstructorID != userID {
			return nil, fmt.Errorf("you must be enrolled in the course to participate in discussions")
		}
	}

	discussion := &domain.Discussion{
		CourseID: input.CourseID,
		LessonID: input.LessonID,
		UserID:   userID,
		ParentID: input.ParentID,
		Content:  input.Content,
	}

	if err := uc.discussionRepo.Create(ctx, discussion); err != nil {
		return nil, err
	}

	// Return with user info
	return uc.discussionRepo.GetByID(ctx, discussion.ID)
}

// UpdateDiscussionInput for updating content
type UpdateDiscussionInput struct {
	Content string `json:"content" validate:"required,min=5,max=5000"`
}

// UpdateDiscussion updates discussion content
func (uc *UseCase) UpdateDiscussion(ctx context.Context, id, userID uuid.UUID, content string) (*domain.Discussion, error) {
	discussion, err := uc.discussionRepo.GetByID(ctx, id)
	if err != nil || discussion == nil {
		return nil, fmt.Errorf("discussion not found")
	}

	if discussion.UserID != userID {
		return nil, fmt.Errorf("you can only edit your own posts")
	}

	discussion.Content = content
	if err := uc.discussionRepo.Update(ctx, discussion); err != nil {
		return nil, err
	}

	return discussion, nil
}

// DeleteDiscussion deletes a discussion
func (uc *UseCase) DeleteDiscussion(ctx context.Context, id, userID uuid.UUID, isAdmin bool) error {
	discussion, err := uc.discussionRepo.GetByID(ctx, id)
	if err != nil || discussion == nil {
		return fmt.Errorf("discussion not found")
	}

	// Check if user is owner, admin, or instructor
	if discussion.UserID != userID {
		if !isAdmin {
			course, _ := uc.courseRepo.GetByID(ctx, discussion.CourseID)
			if course == nil || course.InstructorID != userID {
				return fmt.Errorf("you cannot delete this post")
			}
		}
	}

	return uc.discussionRepo.Delete(ctx, id)
}

// Upvote adds an upvote to a discussion
func (uc *UseCase) Upvote(ctx context.Context, id uuid.UUID) error {
	return uc.discussionRepo.Upvote(ctx, id)
}

// RemoveUpvote removes an upvote from a discussion
func (uc *UseCase) RemoveUpvote(ctx context.Context, id uuid.UUID) error {
	return uc.discussionRepo.RemoveUpvote(ctx, id)
}

// MarkResolved marks a question as resolved (instructor only)
func (uc *UseCase) MarkResolved(ctx context.Context, id, userID uuid.UUID) error {
	discussion, err := uc.discussionRepo.GetByID(ctx, id)
	if err != nil || discussion == nil {
		return fmt.Errorf("discussion not found")
	}

	// Only instructor or original poster can mark as resolved
	course, _ := uc.courseRepo.GetByID(ctx, discussion.CourseID)
	if discussion.UserID != userID && (course == nil || course.InstructorID != userID) {
		return fmt.Errorf("only the instructor or original poster can mark as resolved")
	}

	return uc.discussionRepo.MarkResolved(ctx, id, true)
}

// Unresolve marks a question as unresolved
func (uc *UseCase) Unresolve(ctx context.Context, id, userID uuid.UUID) error {
	discussion, err := uc.discussionRepo.GetByID(ctx, id)
	if err != nil || discussion == nil {
		return fmt.Errorf("discussion not found")
	}

	course, _ := uc.courseRepo.GetByID(ctx, discussion.CourseID)
	if discussion.UserID != userID && (course == nil || course.InstructorID != userID) {
		return fmt.Errorf("only the instructor or original poster can change resolution status")
	}

	return uc.discussionRepo.MarkResolved(ctx, id, false)
}

// Pin pins a discussion (instructor only)
func (uc *UseCase) Pin(ctx context.Context, id, userID uuid.UUID) error {
	discussion, err := uc.discussionRepo.GetByID(ctx, id)
	if err != nil || discussion == nil {
		return fmt.Errorf("discussion not found")
	}

	course, _ := uc.courseRepo.GetByID(ctx, discussion.CourseID)
	if course == nil || course.InstructorID != userID {
		return fmt.Errorf("only the instructor can pin discussions")
	}

	return uc.discussionRepo.Pin(ctx, id, true)
}

// Unpin unpins a discussion (instructor only)
func (uc *UseCase) Unpin(ctx context.Context, id, userID uuid.UUID) error {
	discussion, err := uc.discussionRepo.GetByID(ctx, id)
	if err != nil || discussion == nil {
		return fmt.Errorf("discussion not found")
	}

	course, _ := uc.courseRepo.GetByID(ctx, discussion.CourseID)
	if course == nil || course.InstructorID != userID {
		return fmt.Errorf("only the instructor can unpin discussions")
	}

	return uc.discussionRepo.Pin(ctx, id, false)
}

// GetStats returns discussion stats
type DiscussionStats struct {
	TotalQuestions int64 `json:"total_questions"`
	Resolved       int64 `json:"resolved"`
	Unresolved     int64 `json:"unresolved"`
}

func (uc *UseCase) GetCourseStats(ctx context.Context, courseID uuid.UUID) (*DiscussionStats, error) {
	total, err := uc.discussionRepo.CountByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	return &DiscussionStats{TotalQuestions: total}, nil
}
