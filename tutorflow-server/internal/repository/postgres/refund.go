package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type refundRepository struct {
	db *gorm.DB
}

// NewRefundRepository creates a new refund repository
func NewRefundRepository(db *gorm.DB) repository.RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(ctx context.Context, refund *domain.Refund) error {
	return r.db.WithContext(ctx).Create(refund).Error
}

func (r *refundRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Refund, error) {
	var refund domain.Refund
	err := r.db.WithContext(ctx).Preload("Order").Preload("Order.Items").Preload("User").
		Where("id = ?", id).First(&refund).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Refund, error) {
	var refund domain.Refund
	err := r.db.WithContext(ctx).Preload("User").Where("order_id = ?", orderID).First(&refund).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Refund, int64, error) {
	var refunds []domain.Refund
	var total int64

	offset := (page - 1) * limit

	r.db.WithContext(ctx).Model(&domain.Refund{}).Where("user_id = ?", userID).Count(&total)

	err := r.db.WithContext(ctx).Preload("Order").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&refunds).Error

	return refunds, total, err
}

func (r *refundRepository) List(ctx context.Context, status *domain.RefundStatus, page, limit int) ([]domain.Refund, int64, error) {
	var refunds []domain.Refund
	var total int64

	offset := (page - 1) * limit
	query := r.db.WithContext(ctx).Model(&domain.Refund{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	query.Count(&total)

	err := r.db.WithContext(ctx).Preload("Order").Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&refunds).Error

	return refunds, total, err
}

func (r *refundRepository) Update(ctx context.Context, refund *domain.Refund) error {
	return r.db.WithContext(ctx).Save(refund).Error
}
