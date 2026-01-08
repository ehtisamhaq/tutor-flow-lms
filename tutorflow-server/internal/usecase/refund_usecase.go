package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tutorflow/tutorflow-server/internal/domain"
)

type refundUseCase struct {
	refundRepo     domain.RefundRepository
	orderRepo      domain.OrderRepository
	enrollmentRepo domain.EnrollmentRepository
}

// NewRefundUseCase creates a new refund use case
func NewRefundUseCase(
	refundRepo domain.RefundRepository,
	orderRepo domain.OrderRepository,
	enrollmentRepo domain.EnrollmentRepository,
) domain.RefundUseCase {
	return &refundUseCase{
		refundRepo:     refundRepo,
		orderRepo:      orderRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

// RequestRefund creates a new refund request
func (uc *refundUseCase) RequestRefund(
	userID, orderID uuid.UUID,
	reason domain.RefundReason,
	description string,
) (*domain.Refund, error) {
	// Check if refund already exists for this order
	existing, _ := uc.refundRepo.GetByOrderID(orderID)
	if existing != nil {
		return nil, errors.New("refund request already exists for this order")
	}

	// Get the order
	order, err := uc.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	// Verify order belongs to user
	if order.UserID != userID {
		return nil, errors.New("order does not belong to user")
	}

	// Check if order is eligible for refund
	policy := domain.DefaultRefundPolicy()

	// Check time limit
	daysSincePurchase := int(time.Since(order.CreatedAt).Hours() / 24)
	if daysSincePurchase > policy.MaxDaysAfterPurchase {
		return nil, errors.New("order is outside refund window")
	}

	// Check if order is already refunded
	if order.Status == domain.OrderStatusRefunded {
		return nil, errors.New("order has already been refunded")
	}

	// Create refund request
	refund := &domain.Refund{
		OrderID:     orderID,
		UserID:      userID,
		Amount:      order.Total,
		Reason:      reason,
		Description: description,
		Status:      domain.RefundStatusPending,
	}

	// Auto-approve small refunds
	if order.Total <= policy.AutoApproveUnder && !policy.RequiresApproval {
		refund.Status = domain.RefundStatusApproved
		now := time.Now()
		refund.ProcessedAt = &now
	}

	if err := uc.refundRepo.Create(refund); err != nil {
		return nil, err
	}

	return refund, nil
}

// ApproveRefund approves a refund request
func (uc *refundUseCase) ApproveRefund(
	refundID, adminID uuid.UUID,
	notes string,
) (*domain.Refund, error) {
	refund, err := uc.refundRepo.GetByID(refundID)
	if err != nil {
		return nil, errors.New("refund not found")
	}

	if !refund.CanProcess() {
		return nil, errors.New("refund cannot be processed")
	}

	now := time.Now()
	refund.Status = domain.RefundStatusApproved
	refund.ProcessedBy = &adminID
	refund.ProcessedAt = &now
	refund.AdminNotes = notes
	refund.UpdatedAt = now

	if err := uc.refundRepo.Update(refund); err != nil {
		return nil, err
	}

	// Update order status
	if refund.Order != nil {
		refund.Order.Status = domain.OrderStatusRefunded
		uc.orderRepo.Update(refund.Order)
	}

	// Revoke enrollments for courses in the order
	// In production, you'd also process the Stripe refund here

	return refund, nil
}

// RejectRefund rejects a refund request
func (uc *refundUseCase) RejectRefund(
	refundID, adminID uuid.UUID,
	notes string,
) (*domain.Refund, error) {
	refund, err := uc.refundRepo.GetByID(refundID)
	if err != nil {
		return nil, errors.New("refund not found")
	}

	if !refund.CanProcess() {
		return nil, errors.New("refund cannot be processed")
	}

	now := time.Now()
	refund.Status = domain.RefundStatusRejected
	refund.ProcessedBy = &adminID
	refund.ProcessedAt = &now
	refund.AdminNotes = notes
	refund.UpdatedAt = now

	if err := uc.refundRepo.Update(refund); err != nil {
		return nil, err
	}

	return refund, nil
}

// GetUserRefunds returns user's refund requests
func (uc *refundUseCase) GetUserRefunds(
	userID uuid.UUID,
	page, limit int,
) ([]domain.Refund, int64, error) {
	return uc.refundRepo.GetByUserID(userID, page, limit)
}

// GetPendingRefunds returns pending refunds for admin
func (uc *refundUseCase) GetPendingRefunds(page, limit int) ([]domain.Refund, int64, error) {
	return uc.refundRepo.GetPending(page, limit)
}

// GetAllRefunds returns all refunds with optional status filter
func (uc *refundUseCase) GetAllRefunds(
	status *domain.RefundStatus,
	page, limit int,
) ([]domain.Refund, int64, error) {
	return uc.refundRepo.GetAll(status, page, limit)
}

// GetRefundByID returns a refund by ID
func (uc *refundUseCase) GetRefundByID(id uuid.UUID) (*domain.Refund, error) {
	return uc.refundRepo.GetByID(id)
}
