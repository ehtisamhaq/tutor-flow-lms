package review

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines review business logic
type UseCase struct {
	reviewRepo       repository.ReviewRepository
	enrollmentRepo   repository.EnrollmentRepository
	courseRepo       repository.CourseRepository
	notificationRepo repository.NotificationRepository
}

// NewUseCase creates a new review use case
func NewUseCase(
	reviewRepo repository.ReviewRepository,
	enrollmentRepo repository.EnrollmentRepository,
	courseRepo repository.CourseRepository,
	notificationRepo repository.NotificationRepository,
) *UseCase {
	return &UseCase{
		reviewRepo:       reviewRepo,
		enrollmentRepo:   enrollmentRepo,
		courseRepo:       courseRepo,
		notificationRepo: notificationRepo,
	}
}

// GetReview returns review by ID
func (uc *UseCase) GetReview(ctx context.Context, id uuid.UUID) (*domain.CourseReview, error) {
	return uc.reviewRepo.GetByID(ctx, id)
}

// GetCourseReviews returns reviews for a course
func (uc *UseCase) GetCourseReviews(ctx context.Context, courseID uuid.UUID, page, limit int) ([]domain.CourseReview, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return uc.reviewRepo.GetByCourse(ctx, courseID, page, limit)
}

// GetUserReview returns a user's review for a course
func (uc *UseCase) GetUserReview(ctx context.Context, userID, courseID uuid.UUID) (*domain.CourseReview, error) {
	return uc.reviewRepo.GetByUserAndCourse(ctx, userID, courseID)
}

// CreateReviewInput for creating a review
type CreateReviewInput struct {
	CourseID uuid.UUID `json:"course_id" validate:"required"`
	Rating   float64   `json:"rating" validate:"required,gte=1,lte=5"`
	Title    *string   `json:"title" validate:"omitempty,max=200"`
	Content  *string   `json:"content" validate:"omitempty,max=5000"`
}

// CreateReview creates a new course review
func (uc *UseCase) CreateReview(ctx context.Context, userID uuid.UUID, input CreateReviewInput) (*domain.CourseReview, error) {
	// Check if user already reviewed this course
	existing, _ := uc.reviewRepo.GetByUserAndCourse(ctx, userID, input.CourseID)
	if existing != nil {
		return nil, fmt.Errorf("you have already reviewed this course")
	}

	// Check if user is enrolled
	enrollment, err := uc.enrollmentRepo.GetByUserAndCourse(ctx, userID, input.CourseID)
	isVerified := false
	if err == nil && enrollment != nil && enrollment.CanAccess() {
		isVerified = true
	}

	review := &domain.CourseReview{
		CourseID:           input.CourseID,
		UserID:             userID,
		Rating:             input.Rating,
		Title:              input.Title,
		Content:            input.Content,
		IsVerifiedPurchase: isVerified,
		Status:             "published",
	}

	if err := uc.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	// Get updated review with user info
	review, _ = uc.reviewRepo.GetByID(ctx, review.ID)

	// Notify course instructor
	course, _ := uc.courseRepo.GetByID(ctx, input.CourseID)
	if course != nil {
		notification := &domain.Notification{
			UserID:  course.InstructorID,
			Type:    domain.NotificationReviewReceived,
			Title:   "New Course Review",
			Message: stringPtr(fmt.Sprintf("Your course \"%s\" received a %.1f star review", course.Title, input.Rating)),
		}
		_ = uc.notificationRepo.Create(ctx, notification)
	}

	return review, nil
}

// UpdateReviewInput for updating a review
type UpdateReviewInput struct {
	Rating  *float64 `json:"rating" validate:"omitempty,gte=1,lte=5"`
	Title   *string  `json:"title" validate:"omitempty,max=200"`
	Content *string  `json:"content" validate:"omitempty,max=5000"`
}

// UpdateReview updates a review
func (uc *UseCase) UpdateReview(ctx context.Context, reviewID, userID uuid.UUID, input UpdateReviewInput) (*domain.CourseReview, error) {
	review, err := uc.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if review.UserID != userID {
		return nil, fmt.Errorf("you can only edit your own reviews")
	}

	if input.Rating != nil {
		review.Rating = *input.Rating
	}
	if input.Title != nil {
		review.Title = input.Title
	}
	if input.Content != nil {
		review.Content = input.Content
	}

	if err := uc.reviewRepo.Update(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

// DeleteReview deletes a review
func (uc *UseCase) DeleteReview(ctx context.Context, reviewID, userID uuid.UUID, isAdmin bool) error {
	review, err := uc.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}

	// Check authorization
	if review.UserID != userID && !isAdmin {
		return fmt.Errorf("you can only delete your own reviews")
	}

	return uc.reviewRepo.Delete(ctx, reviewID)
}

// VoteReview votes on a review
func (uc *UseCase) VoteReview(ctx context.Context, reviewID, userID uuid.UUID, isHelpful bool) error {
	// Verify review exists
	review, err := uc.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}

	// Can't vote on your own review
	if review.UserID == userID {
		return fmt.Errorf("you cannot vote on your own review")
	}

	return uc.reviewRepo.Vote(ctx, reviewID, userID, isHelpful)
}

// ReplyToReviewInput for instructor reply
type ReplyToReviewInput struct {
	Reply string `json:"reply" validate:"required,max=2000"`
}

// ReplyToReview allows instructor to reply to a review
func (uc *UseCase) ReplyToReview(ctx context.Context, reviewID, instructorID uuid.UUID, reply string) (*domain.CourseReview, error) {
	review, err := uc.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	// Verify instructor owns the course
	course, err := uc.courseRepo.GetByID(ctx, review.CourseID)
	if err != nil {
		return nil, err
	}

	if course.InstructorID != instructorID {
		return nil, fmt.Errorf("only the course instructor can reply to reviews")
	}

	now := time.Now()
	review.InstructorReply = &reply
	review.InstructorReplyAt = &now

	if err := uc.reviewRepo.Update(ctx, review); err != nil {
		return nil, err
	}

	// Notify reviewer
	notification := &domain.Notification{
		UserID:  review.UserID,
		Type:    domain.NotificationMessage,
		Title:   "Instructor Replied to Your Review",
		Message: stringPtr(fmt.Sprintf("The instructor replied to your review on \"%s\"", course.Title)),
	}
	_ = uc.notificationRepo.Create(ctx, notification)

	return review, nil
}

// GetCourseRatingSummary returns rating breakdown for a course
type RatingSummary struct {
	TotalReviews  int64         `json:"total_reviews"`
	AverageRating float64       `json:"average_rating"`
	Distribution  map[int]int64 `json:"distribution"` // 1-5 star counts
}

func (uc *UseCase) GetCourseRatingSummary(ctx context.Context, courseID uuid.UUID) (*RatingSummary, error) {
	course, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	// Get all reviews to calculate distribution
	reviews, total, err := uc.reviewRepo.GetByCourse(ctx, courseID, 1, 1000)
	if err != nil {
		return nil, err
	}

	distribution := map[int]int64{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	for _, r := range reviews {
		star := int(r.Rating)
		if star >= 1 && star <= 5 {
			distribution[star]++
		}
	}

	return &RatingSummary{
		TotalReviews:  total,
		AverageRating: course.Rating,
		Distribution:  distribution,
	}, nil
}

// FeatureReview marks a review as featured (admin/instructor)
func (uc *UseCase) FeatureReview(ctx context.Context, reviewID uuid.UUID, featured bool) error {
	review, err := uc.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}

	review.IsFeatured = featured
	return uc.reviewRepo.Update(ctx, review)
}

func stringPtr(s string) *string {
	return &s
}
