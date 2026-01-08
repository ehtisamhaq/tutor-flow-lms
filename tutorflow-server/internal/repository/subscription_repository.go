package repository

import (
	"github.com/tutorflow/tutorflow-server/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type subscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *gorm.DB) domain.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

// Plans

func (r *subscriptionRepository) CreatePlan(plan *domain.SubscriptionPlan) error {
	return r.db.Create(plan).Error
}

func (r *subscriptionRepository) GetPlanByID(id uuid.UUID) (*domain.SubscriptionPlan, error) {
	var plan domain.SubscriptionPlan
	err := r.db.Where("id = ?", id).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionRepository) GetPlanBySlug(slug string) (*domain.SubscriptionPlan, error) {
	var plan domain.SubscriptionPlan
	err := r.db.Where("slug = ?", slug).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionRepository) GetActivePlans() ([]domain.SubscriptionPlan, error) {
	var plans []domain.SubscriptionPlan
	err := r.db.Where("is_active = ?", true).Order("priority ASC").Find(&plans).Error
	return plans, err
}

func (r *subscriptionRepository) UpdatePlan(plan *domain.SubscriptionPlan) error {
	return r.db.Save(plan).Error
}

// Subscriptions

func (r *subscriptionRepository) Create(subscription *domain.Subscription) error {
	return r.db.Create(subscription).Error
}

func (r *subscriptionRepository) GetByID(id uuid.UUID) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := r.db.Preload("Plan").Preload("User").Where("id = ?", id).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) GetByUserID(userID uuid.UUID) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := r.db.Preload("Plan").Where("user_id = ?", userID).Order("created_at DESC").First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) GetActiveByUserID(userID uuid.UUID) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := r.db.Preload("Plan").
		Where("user_id = ? AND status IN (?)", userID, []string{"active", "trialing"}).
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) Update(subscription *domain.Subscription) error {
	return r.db.Save(subscription).Error
}

func (r *subscriptionRepository) Cancel(id uuid.UUID) error {
	return r.db.Model(&domain.Subscription{}).
		Where("id = ?", id).
		Update("status", domain.SubscriptionStatusCanceled).Error
}

func (r *subscriptionRepository) GetExpiringSubscriptions(days int) ([]domain.Subscription, error) {
	var subs []domain.Subscription
	err := r.db.Preload("User").Preload("Plan").
		Where("status = ? AND current_period_end <= NOW() + INTERVAL '? days'",
			domain.SubscriptionStatusActive, days).
		Find(&subs).Error
	return subs, err
}
