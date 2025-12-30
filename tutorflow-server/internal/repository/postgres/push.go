package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// PushSubscriptionRepository
type pushSubscriptionRepository struct {
	db *gorm.DB
}

func NewPushSubscriptionRepository(db *gorm.DB) repository.PushSubscriptionRepository {
	return &pushSubscriptionRepository{db: db}
}

func (r *pushSubscriptionRepository) Create(ctx context.Context, sub *domain.PushSubscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *pushSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.PushSubscription, error) {
	var sub domain.PushSubscription
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *pushSubscriptionRepository) GetByEndpoint(ctx context.Context, endpoint string) (*domain.PushSubscription, error) {
	var sub domain.PushSubscription
	err := r.db.WithContext(ctx).Where("endpoint = ?", endpoint).First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *pushSubscriptionRepository) GetByUser(ctx context.Context, userID uuid.UUID) ([]domain.PushSubscription, error) {
	var subs []domain.PushSubscription
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&subs).Error
	return subs, err
}

func (r *pushSubscriptionRepository) Update(ctx context.Context, sub *domain.PushSubscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *pushSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PushSubscription{}, "id = ?", id).Error
}

func (r *pushSubscriptionRepository) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	return r.db.WithContext(ctx).Delete(&domain.PushSubscription{}, "endpoint = ?", endpoint).Error
}

func (r *pushSubscriptionRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PushSubscription{}, "user_id = ?", userID).Error
}

func (r *pushSubscriptionRepository) GetAllUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := r.db.WithContext(ctx).Model(&domain.PushSubscription{}).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}
