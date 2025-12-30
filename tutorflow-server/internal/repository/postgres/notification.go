package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// NotificationRepository
type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *notificationRepository) GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Notification, int64, error) {
	var notifications []domain.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error
	return notifications, total, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("id = ?", id).
		Update("read_at", now).Error
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", now).Error
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Notification{}, "id = ?", id).Error
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

// ReviewRepository
type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) repository.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review *domain.CourseReview) error {
	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		return err
	}
	// Update course rating
	return r.updateCourseRating(ctx, review.CourseID)
}

func (r *reviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CourseReview, error) {
	var review domain.CourseReview
	err := r.db.WithContext(ctx).Preload("User").Where("id = ?", id).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) Update(ctx context.Context, review *domain.CourseReview) error {
	if err := r.db.WithContext(ctx).Save(review).Error; err != nil {
		return err
	}
	return r.updateCourseRating(ctx, review.CourseID)
}

func (r *reviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var review domain.CourseReview
	if err := r.db.WithContext(ctx).First(&review, "id = ?", id).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Delete(&review).Error; err != nil {
		return err
	}
	return r.updateCourseRating(ctx, review.CourseID)
}

func (r *reviewRepository) GetByCourse(ctx context.Context, courseID uuid.UUID, page, limit int) ([]domain.CourseReview, int64, error) {
	var reviews []domain.CourseReview
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.CourseReview{}).
		Where("course_id = ? AND status = ?", courseID, "published")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reviews).Error

	return reviews, total, err
}

func (r *reviewRepository) GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (*domain.CourseReview, error) {
	var review domain.CourseReview
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) Vote(ctx context.Context, reviewID, userID uuid.UUID, isHelpful bool) error {
	vote := &domain.ReviewVote{
		ReviewID:  reviewID,
		UserID:    userID,
		IsHelpful: isHelpful,
	}

	// Upsert vote
	err := r.db.WithContext(ctx).
		Where(domain.ReviewVote{ReviewID: reviewID, UserID: userID}).
		Assign(vote).
		FirstOrCreate(vote).Error
	if err != nil {
		return err
	}

	// Update counts
	return r.updateVoteCounts(ctx, reviewID)
}

func (r *reviewRepository) updateCourseRating(ctx context.Context, courseID uuid.UUID) error {
	var result struct {
		AvgRating float64
		Count     int64
	}

	r.db.WithContext(ctx).Model(&domain.CourseReview{}).
		Select("AVG(rating) as avg_rating, COUNT(*) as count").
		Where("course_id = ? AND status = ?", courseID, "published").
		Scan(&result)

	return r.db.WithContext(ctx).Model(&domain.Course{}).
		Where("id = ?", courseID).
		Updates(map[string]interface{}{
			"rating":        result.AvgRating,
			"total_reviews": result.Count,
		}).Error
}

func (r *reviewRepository) updateVoteCounts(ctx context.Context, reviewID uuid.UUID) error {
	var helpful, unhelpful int64

	r.db.WithContext(ctx).Model(&domain.ReviewVote{}).
		Where("review_id = ? AND is_helpful = ?", reviewID, true).
		Count(&helpful)

	r.db.WithContext(ctx).Model(&domain.ReviewVote{}).
		Where("review_id = ? AND is_helpful = ?", reviewID, false).
		Count(&unhelpful)

	return r.db.WithContext(ctx).Model(&domain.CourseReview{}).
		Where("id = ?", reviewID).
		Updates(map[string]interface{}{
			"helpful_count":   helpful,
			"unhelpful_count": unhelpful,
		}).Error
}
