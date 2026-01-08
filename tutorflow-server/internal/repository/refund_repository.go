package repository

import (
	"github.com/tutorflow/tutorflow-server/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type refundRepository struct {
	db *gorm.DB
}

// NewRefundRepository creates a new refund repository
func NewRefundRepository(db *gorm.DB) domain.RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(refund *domain.Refund) error {
	return r.db.Create(refund).Error
}

func (r *refundRepository) GetByID(id uuid.UUID) (*domain.Refund, error) {
	var refund domain.Refund
	err := r.db.Preload("Order").Preload("Order.Items").Preload("User").
		Where("id = ?", id).First(&refund).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepository) GetByOrderID(orderID uuid.UUID) (*domain.Refund, error) {
	var refund domain.Refund
	err := r.db.Preload("User").Where("order_id = ?", orderID).First(&refund).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepository) GetByUserID(userID uuid.UUID, page, limit int) ([]domain.Refund, int64, error) {
	var refunds []domain.Refund
	var total int64

	offset := (page - 1) * limit

	r.db.Model(&domain.Refund{}).Where("user_id = ?", userID).Count(&total)

	err := r.db.Preload("Order").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&refunds).Error

	return refunds, total, err
}

func (r *refundRepository) GetPending(page, limit int) ([]domain.Refund, int64, error) {
	var refunds []domain.Refund
	var total int64

	offset := (page - 1) * limit

	r.db.Model(&domain.Refund{}).Where("status = ?", domain.RefundStatusPending).Count(&total)

	err := r.db.Preload("Order").Preload("User").
		Where("status = ?", domain.RefundStatusPending).
		Order("created_at ASC").
		Offset(offset).Limit(limit).
		Find(&refunds).Error

	return refunds, total, err
}

func (r *refundRepository) GetAll(status *domain.RefundStatus, page, limit int) ([]domain.Refund, int64, error) {
	var refunds []domain.Refund
	var total int64

	offset := (page - 1) * limit
	query := r.db.Model(&domain.Refund{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	query.Count(&total)

	err := r.db.Preload("Order").Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&refunds).Error

	return refunds, total, err
}

func (r *refundRepository) Update(refund *domain.Refund) error {
	return r.db.Save(refund).Error
}
