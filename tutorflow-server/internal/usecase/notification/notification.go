package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines notification business logic
type UseCase struct {
	notificationRepo repository.NotificationRepository
	enrollmentRepo   repository.EnrollmentRepository
}

// NewUseCase creates a new notification use case
func NewUseCase(
	notificationRepo repository.NotificationRepository,
	enrollmentRepo repository.EnrollmentRepository,
) *UseCase {
	return &UseCase{
		notificationRepo: notificationRepo,
		enrollmentRepo:   enrollmentRepo,
	}
}

// GetNotifications returns user's notifications
func (uc *UseCase) GetNotifications(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.notificationRepo.GetByUser(ctx, userID, page, limit)
}

// GetUnreadCount returns count of unread notifications
func (uc *UseCase) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return uc.notificationRepo.GetUnreadCount(ctx, userID)
}

// MarkAsRead marks a notification as read
func (uc *UseCase) MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	// TODO: Verify notification belongs to user
	return uc.notificationRepo.MarkAsRead(ctx, notificationID)
}

// MarkAllAsRead marks all notifications as read
func (uc *UseCase) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return uc.notificationRepo.MarkAllAsRead(ctx, userID)
}

// DeleteNotification deletes a notification
func (uc *UseCase) DeleteNotification(ctx context.Context, notificationID, userID uuid.UUID) error {
	// TODO: Verify notification belongs to user
	return uc.notificationRepo.Delete(ctx, notificationID)
}

// --- Notification Sending Helpers ---

// SendNotificationInput for creating notifications
type SendNotificationInput struct {
	UserID  uuid.UUID               `json:"user_id"`
	Type    domain.NotificationType `json:"type"`
	Title   string                  `json:"title"`
	Message string                  `json:"message"`
	Data    map[string]interface{}  `json:"data,omitempty"`
}

// Send creates and sends a notification
func (uc *UseCase) Send(ctx context.Context, input SendNotificationInput) error {
	var dataStr *string
	if input.Data != nil {
		data, _ := json.Marshal(input.Data)
		s := string(data)
		dataStr = &s
	}

	var messagePtr *string
	if input.Message != "" {
		messagePtr = &input.Message
	}

	notification := &domain.Notification{
		UserID:  input.UserID,
		Type:    input.Type,
		Title:   input.Title,
		Message: messagePtr,
		Data:    dataStr,
	}

	return uc.notificationRepo.Create(ctx, notification)
}

// SendToMany sends notification to multiple users
func (uc *UseCase) SendToMany(ctx context.Context, userIDs []uuid.UUID, notifType domain.NotificationType, title, message string, data map[string]interface{}) error {
	for _, userID := range userIDs {
		_ = uc.Send(ctx, SendNotificationInput{
			UserID:  userID,
			Type:    notifType,
			Title:   title,
			Message: message,
			Data:    data,
		})
	}
	return nil
}

// --- Predefined Notification Types ---

// NotifyEnrollmentApproved sends enrollment approval notification
func (uc *UseCase) NotifyEnrollmentApproved(ctx context.Context, userID uuid.UUID, courseName string, courseID uuid.UUID) error {
	return uc.Send(ctx, SendNotificationInput{
		UserID:  userID,
		Type:    domain.NotificationEnrollmentApproved,
		Title:   "Enrollment Approved",
		Message: fmt.Sprintf("Your enrollment in \"%s\" has been approved. Start learning now!", courseName),
		Data:    map[string]interface{}{"course_id": courseID.String()},
	})
}

// NotifyNewLesson sends new lesson notification to enrolled students
func (uc *UseCase) NotifyNewLesson(ctx context.Context, courseID uuid.UUID, courseName, lessonTitle string) error {
	// Get all enrolled students
	enrollments, _, err := uc.enrollmentRepo.GetByCourse(ctx, courseID, 1, 1000)
	if err != nil {
		return err
	}

	for _, e := range enrollments {
		if e.Status == domain.EnrollmentStatusActive {
			_ = uc.Send(ctx, SendNotificationInput{
				UserID:  e.UserID,
				Type:    domain.NotificationNewLesson,
				Title:   "New Lesson Available",
				Message: fmt.Sprintf("A new lesson \"%s\" is now available in \"%s\"", lessonTitle, courseName),
				Data:    map[string]interface{}{"course_id": courseID.String()},
			})
		}
	}
	return nil
}

// NotifyAssignmentDue sends assignment due reminder
func (uc *UseCase) NotifyAssignmentDue(ctx context.Context, userID uuid.UUID, assignmentTitle, courseName string, assignmentID uuid.UUID) error {
	return uc.Send(ctx, SendNotificationInput{
		UserID:  userID,
		Type:    domain.NotificationAssignmentDue,
		Title:   "Assignment Due Soon",
		Message: fmt.Sprintf("Your assignment \"%s\" in \"%s\" is due soon. Don't forget to submit!", assignmentTitle, courseName),
		Data:    map[string]interface{}{"assignment_id": assignmentID.String()},
	})
}

// NotifyGradePosted sends grade posted notification
func (uc *UseCase) NotifyGradePosted(ctx context.Context, userID uuid.UUID, itemTitle, courseName string, score float64) error {
	return uc.Send(ctx, SendNotificationInput{
		UserID:  userID,
		Type:    domain.NotificationGradePosted,
		Title:   "Grade Posted",
		Message: fmt.Sprintf("Your grade for \"%s\" in \"%s\" has been posted: %.1f", itemTitle, courseName, score),
	})
}

// NotifyCourseUpdate sends course update notification
func (uc *UseCase) NotifyCourseUpdate(ctx context.Context, courseID uuid.UUID, courseName, updateMessage string) error {
	enrollments, _, err := uc.enrollmentRepo.GetByCourse(ctx, courseID, 1, 1000)
	if err != nil {
		return err
	}

	for _, e := range enrollments {
		if e.Status == domain.EnrollmentStatusActive {
			_ = uc.Send(ctx, SendNotificationInput{
				UserID:  e.UserID,
				Type:    domain.NotificationCourseUpdate,
				Title:   "Course Update",
				Message: fmt.Sprintf("%s: %s", courseName, updateMessage),
				Data:    map[string]interface{}{"course_id": courseID.String()},
			})
		}
	}
	return nil
}

// NotifyPaymentReceived sends payment confirmation
func (uc *UseCase) NotifyPaymentReceived(ctx context.Context, userID uuid.UUID, amount float64, orderNumber string) error {
	return uc.Send(ctx, SendNotificationInput{
		UserID:  userID,
		Type:    domain.NotificationPaymentReceived,
		Title:   "Payment Confirmed",
		Message: fmt.Sprintf("Your payment of $%.2f for order %s has been received. Thank you!", amount, orderNumber),
		Data:    map[string]interface{}{"order_number": orderNumber},
	})
}

// NotifyReviewReceived sends review notification to instructor
func (uc *UseCase) NotifyReviewReceived(ctx context.Context, instructorID uuid.UUID, courseName string, rating float64) error {
	return uc.Send(ctx, SendNotificationInput{
		UserID:  instructorID,
		Type:    domain.NotificationReviewReceived,
		Title:   "New Course Review",
		Message: fmt.Sprintf("Your course \"%s\" received a %.1f star review", courseName, rating),
	})
}
