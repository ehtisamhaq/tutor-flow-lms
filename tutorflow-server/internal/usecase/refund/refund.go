package refund

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type refundUseCase struct {
	refundRepo     repository.RefundRepository
	orderRepo      repository.OrderRepository
	enrollmentRepo repository.EnrollmentRepository
}

// NewRefundUseCase creates a new refund use case
func NewUseCase(
	refundRepo repository.RefundRepository,
	orderRepo repository.OrderRepository,
	enrollmentRepo repository.EnrollmentRepository,
) domain.RefundUseCase {
	return &refundUseCase{
		refundRepo:     refundRepo,
		orderRepo:      orderRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

// RequestRefund creates a new refund request
func (uc *refundUseCase) RequestRefund(
	ctx context.Context,
	userID, orderID uuid.UUID,
	reason domain.RefundReason,
	description string,
) (*domain.Refund, error) {
	// Check if refund already exists for this order
	existing, _ := uc.refundRepo.GetByOrderID(ctx, orderID)
	if existing != nil {
		return nil, errors.New("refund request already exists for this order")
	}

	// Get the order
	order, err := uc.orderRepo.GetByID(ctx, orderID)
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

	if err := uc.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}

	return refund, nil
}

// ApproveRefund approves a refund request
func (uc *refundUseCase) ApproveRefund(
	ctx context.Context,
	refundID, adminID uuid.UUID,
	notes string,
) (*domain.Refund, error) {
	refund, err := uc.refundRepo.GetByID(ctx, refundID)
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

	if err := uc.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}

	// Update order status
	if refund.Order != nil {
		refund.Order.Status = domain.OrderStatusRefunded
		uc.orderRepo.Update(ctx, refund.Order)
	}

	return refund, nil
}

// RejectRefund rejects a refund request
func (uc *refundUseCase) RejectRefund(
	ctx context.Context,
	refundID, adminID uuid.UUID,
	notes string,
) (*domain.Refund, error) {
	refund, err := uc.refundRepo.GetByID(ctx, refundID)
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

	if err := uc.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}

	return refund, nil
}

// GetUserRefunds returns user's refund requests
func (uc *refundUseCase) GetUserRefunds(
	ctx context.Context,
	userID uuid.UUID,
	page, limit int,
) ([]domain.Refund, int64, error) {
	return uc.refundRepo.GetByUserID(ctx, userID, page, limit)
}

// GetPendingRefunds returns pending refunds for admin
func (uc *refundUseCase) GetPendingRefunds(ctx context.Context, page, limit int) ([]domain.Refund, int64, error) {
	status := domain.RefundStatusPending
	return uc.refundRepo.List(ctx, &status, page, limit)
}

// GetAllRefunds returns all refunds with optional status filter
func (uc *refundUseCase) GetAllRefunds(
	ctx context.Context,
	status *domain.RefundStatus,
	page, limit int,
) ([]domain.Refund, int64, error) {
	return uc.refundRepo.List(ctx, status, page, limit)
}

// GetRefundByID returns a refund by ID
func (uc *refundUseCase) GetRefundByID(ctx context.Context, id uuid.UUID) (*domain.Refund, error) {
	return uc.refundRepo.GetByID(ctx, id)
}
