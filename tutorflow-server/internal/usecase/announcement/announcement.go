package announcement

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines announcement business logic
type UseCase struct {
	announcementRepo repository.AnnouncementRepository
	courseRepo       repository.CourseRepository
	enrollmentRepo   repository.EnrollmentRepository
	notificationRepo repository.NotificationRepository
}

// NewUseCase creates a new announcement use case
func NewUseCase(
	announcementRepo repository.AnnouncementRepository,
	courseRepo repository.CourseRepository,
	enrollmentRepo repository.EnrollmentRepository,
	notificationRepo repository.NotificationRepository,
) *UseCase {
	return &UseCase{
		announcementRepo: announcementRepo,
		courseRepo:       courseRepo,
		enrollmentRepo:   enrollmentRepo,
		notificationRepo: notificationRepo,
	}
}

// GetAnnouncement returns an announcement by ID
func (uc *UseCase) GetAnnouncement(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	return uc.announcementRepo.GetByID(ctx, id)
}

// GetCourseAnnouncements returns announcements for a course
func (uc *UseCase) GetCourseAnnouncements(ctx context.Context, courseID uuid.UUID, page, limit int) ([]domain.Announcement, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.announcementRepo.GetByCourse(ctx, courseID, page, limit)
}

// GetGlobalAnnouncements returns global (platform-wide) announcements
func (uc *UseCase) GetGlobalAnnouncements(ctx context.Context, page, limit int) ([]domain.Announcement, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.announcementRepo.GetGlobal(ctx, page, limit)
}

// GetMyFeed returns announcements relevant to a user (enrolled courses + global)
func (uc *UseCase) GetMyFeed(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Announcement, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.announcementRepo.GetForUser(ctx, userID, page, limit)
}

// GetMyAnnouncements returns announcements created by an instructor
func (uc *UseCase) GetMyAnnouncements(ctx context.Context, authorID uuid.UUID, page, limit int) ([]domain.Announcement, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.announcementRepo.GetByAuthor(ctx, authorID, page, limit)
}

// CreateAnnouncementInput for creating an announcement
type CreateAnnouncementInput struct {
	CourseID    *uuid.UUID `json:"course_id,omitempty"`
	Title       string     `json:"title" validate:"required,min=5,max=255"`
	Content     string     `json:"content" validate:"required,min=10,max=10000"`
	IsPinned    bool       `json:"is_pinned,omitempty"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

// CreateAnnouncement creates an announcement
func (uc *UseCase) CreateAnnouncement(ctx context.Context, authorID uuid.UUID, isAdmin bool, input CreateAnnouncementInput) (*domain.Announcement, error) {
	// Verify permissions
	if input.CourseID != nil {
		// Course-specific announcement - must be instructor
		course, err := uc.courseRepo.GetByID(ctx, *input.CourseID)
		if err != nil || course == nil {
			return nil, fmt.Errorf("course not found")
		}
		if course.InstructorID != authorID && !isAdmin {
			return nil, fmt.Errorf("only the course instructor can create announcements")
		}
	} else {
		// Global announcement - admin only
		if !isAdmin {
			return nil, fmt.Errorf("only admins can create global announcements")
		}
	}

	publishedAt := time.Now()
	if input.ScheduledAt != nil && input.ScheduledAt.After(time.Now()) {
		publishedAt = *input.ScheduledAt
	}

	announcement := &domain.Announcement{
		CourseID:    input.CourseID,
		AuthorID:    authorID,
		Title:       input.Title,
		Content:     input.Content,
		IsPinned:    input.IsPinned,
		PublishedAt: publishedAt,
	}

	if err := uc.announcementRepo.Create(ctx, announcement); err != nil {
		return nil, err
	}

	// Send notifications to enrolled students (async for course announcements)
	if input.CourseID != nil && publishedAt.Before(time.Now().Add(time.Minute)) {
		go uc.notifyEnrolledStudents(context.Background(), *input.CourseID, announcement.Title)
	}

	return uc.announcementRepo.GetByID(ctx, announcement.ID)
}

// notifyEnrolledStudents sends notifications to enrolled students
func (uc *UseCase) notifyEnrolledStudents(ctx context.Context, courseID uuid.UUID, title string) {
	course, _ := uc.courseRepo.GetByID(ctx, courseID)
	if course == nil {
		return
	}

	enrollments, _, _ := uc.enrollmentRepo.GetByCourse(ctx, courseID, 1, 1000)
	for _, e := range enrollments {
		if e.Status == domain.EnrollmentStatusActive {
			notification := &domain.Notification{
				UserID: e.UserID,
				Type:   domain.NotificationAnnouncement,
				Title:  "New Announcement",
			}
			msg := fmt.Sprintf("New announcement in %s: %s", course.Title, title)
			notification.Message = &msg
			_ = uc.notificationRepo.Create(ctx, notification)
		}
	}
}

// UpdateAnnouncementInput for updating an announcement
type UpdateAnnouncementInput struct {
	Title    string `json:"title" validate:"omitempty,min=5,max=255"`
	Content  string `json:"content" validate:"omitempty,min=10,max=10000"`
	IsPinned *bool  `json:"is_pinned,omitempty"`
}

// UpdateAnnouncement updates an announcement
func (uc *UseCase) UpdateAnnouncement(ctx context.Context, id, userID uuid.UUID, isAdmin bool, input UpdateAnnouncementInput) (*domain.Announcement, error) {
	announcement, err := uc.announcementRepo.GetByID(ctx, id)
	if err != nil || announcement == nil {
		return nil, fmt.Errorf("announcement not found")
	}

	if announcement.AuthorID != userID && !isAdmin {
		return nil, fmt.Errorf("you cannot edit this announcement")
	}

	if input.Title != "" {
		announcement.Title = input.Title
	}
	if input.Content != "" {
		announcement.Content = input.Content
	}
	if input.IsPinned != nil {
		announcement.IsPinned = *input.IsPinned
	}

	if err := uc.announcementRepo.Update(ctx, announcement); err != nil {
		return nil, err
	}

	return announcement, nil
}

// DeleteAnnouncement deletes an announcement
func (uc *UseCase) DeleteAnnouncement(ctx context.Context, id, userID uuid.UUID, isAdmin bool) error {
	announcement, err := uc.announcementRepo.GetByID(ctx, id)
	if err != nil || announcement == nil {
		return fmt.Errorf("announcement not found")
	}

	if announcement.AuthorID != userID && !isAdmin {
		return fmt.Errorf("you cannot delete this announcement")
	}

	return uc.announcementRepo.Delete(ctx, id)
}

// PinAnnouncement pins/unpins an announcement
func (uc *UseCase) PinAnnouncement(ctx context.Context, id, userID uuid.UUID, isAdmin bool, pinned bool) error {
	announcement, err := uc.announcementRepo.GetByID(ctx, id)
	if err != nil || announcement == nil {
		return fmt.Errorf("announcement not found")
	}

	if announcement.AuthorID != userID && !isAdmin {
		return fmt.Errorf("you cannot modify this announcement")
	}

	return uc.announcementRepo.Pin(ctx, id, pinned)
}
