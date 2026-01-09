package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type subscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *gorm.DB) repository.SubscriptionRepository { // Return matching repository interface
	return &subscriptionRepository{db: db}
}

// Plans methods

func (r *subscriptionRepository) CreatePlan(ctx context.Context, plan *domain.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *subscriptionRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*domain.SubscriptionPlan, error) {
	var plan domain.SubscriptionPlan
	err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionRepository) GetPlanBySlug(ctx context.Context, slug string) (*domain.SubscriptionPlan, error) {
	var plan domain.SubscriptionPlan
	err := r.db.WithContext(ctx).First(&plan, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionRepository) GetActivePlans(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	var plans []domain.SubscriptionPlan
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("priority DESC").Find(&plans).Error
	return plans, err
}

func (r *subscriptionRepository) UpdatePlan(ctx context.Context, plan *domain.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

// Subscriptions methods

func (r *subscriptionRepository) Create(ctx context.Context, subscription *domain.Subscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := r.db.WithContext(ctx).Preload("Plan").First(&sub, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := r.db.WithContext(ctx).Preload("Plan").First(&sub, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	var sub domain.Subscription
	// Find the most recent active subscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("user_id = ? AND status = ? AND current_period_end > ?", userID, domain.SubscriptionStatusActive, time.Now()).
		Order("created_at DESC").
		First(&sub).Error

	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, subscription *domain.Subscription) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

func (r *subscriptionRepository) Cancel(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Subscription{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      domain.SubscriptionStatusCanceled,
			"canceled_at": now,
			"updated_at":  now,
		}).Error
}

func (r *subscriptionRepository) GetExpiringSubscriptions(ctx context.Context, days int) ([]domain.Subscription, error) {
	var subs []domain.Subscription
	target := time.Now().AddDate(0, 0, days)
	// Find active subscriptions expiring on or before target date
	err := r.db.WithContext(ctx).
		Where("status = ? AND current_period_end <= ?", domain.SubscriptionStatusActive, target).
		Find(&subs).Error
	return subs, err
}
