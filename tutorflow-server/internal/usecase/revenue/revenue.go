package revenue

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type revenueUseCase struct {
	earningRepo repository.EarningRepository
	payoutRepo  repository.PayoutRepository
	orderRepo   repository.OrderRepository
	userRepo    repository.UserRepository
}

// NewRevenueUseCase creates a new revenue use case
func NewUseCase(
	earningRepo repository.EarningRepository,
	payoutRepo repository.PayoutRepository,
	orderRepo repository.OrderRepository,
	userRepo repository.UserRepository,
) domain.RevenueUseCase {
	return &revenueUseCase{
		earningRepo: earningRepo,
		payoutRepo:  payoutRepo,
		orderRepo:   orderRepo,
		userRepo:    userRepo,
	}
}

func (uc *revenueUseCase) GetInstructorStats(ctx context.Context, instructorID uuid.UUID) (*domain.InstructorStats, error) {
	return uc.earningRepo.GetStats(ctx, instructorID)
}

func (uc *revenueUseCase) GetEarnings(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]domain.InstructorEarning, int64, error) {
	return uc.earningRepo.GetByInstructor(ctx, instructorID, page, limit)
}

func (uc *revenueUseCase) GetPayouts(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]domain.Payout, int64, error) {
	return uc.payoutRepo.GetByInstructor(ctx, instructorID, page, limit)
}

func (uc *revenueUseCase) RequestPayout(ctx context.Context, instructorID uuid.UUID, amount float64) (*domain.Payout, error) {
	stats, err := uc.earningRepo.GetStats(ctx, instructorID)
	if err != nil {
		return nil, err
	}

	if amount > stats.AvailableEarnings {
		return nil, fmt.Errorf("insufficient available earnings")
	}

	if amount < 50 {
		return nil, fmt.Errorf("minimum payout amount is $50")
	}

	payout := &domain.Payout{
		InstructorID: instructorID,
		Amount:       amount,
		Status:       "pending",
	}

	if err := uc.payoutRepo.Create(ctx, payout); err != nil {
		return nil, err
	}

	// Move earnings from 'available' to 'pending_payout' in a real system
	// For now, we'll just track the payout request.
	return payout, nil
}

func (uc *revenueUseCase) HandlePaymentSuccess(ctx context.Context, order *domain.Order) error {
	// Create earnings for each instructor in the order
	platformFeePercent := 0.30

	for _, item := range order.Items {
		earning := &domain.InstructorEarning{
			InstructorID: item.Course.InstructorID,
			OrderItemID:  item.ID,
			Amount:       item.InstructorShare,
			PlatformFee:  item.Price * platformFeePercent,
			Status:       "pending", // Initially pending
		}

		if err := uc.earningRepo.Create(ctx, earning); err != nil {
			// Log error but continue with other items
			fmt.Printf("Failed to create earning for instructor %s: %v\n", item.Course.InstructorID, err)
		}
	}

	return nil
}
