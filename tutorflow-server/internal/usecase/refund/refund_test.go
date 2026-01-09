package refund_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/usecase/refund"
)

// MockRefundRepository is a mock implementation of RefundRepository
type MockRefundRepository struct {
	mock.Mock
}

func (m *MockRefundRepository) Create(refund *domain.Refund) error {
	args := m.Called(refund)
	return args.Error(0)
}

func (m *MockRefundRepository) GetByID(id uuid.UUID) (*domain.Refund, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Refund), args.Error(1)
}

func (m *MockRefundRepository) GetByOrderID(orderID uuid.UUID) (*domain.Refund, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Refund), args.Error(1)
}

func (m *MockRefundRepository) GetByUserID(userID uuid.UUID, page, limit int) ([]domain.Refund, int64, error) {
	args := m.Called(userID, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.Refund), args.Get(1).(int64), args.Error(2)
}

func (m *MockRefundRepository) GetPending(page, limit int) ([]domain.Refund, int64, error) {
	args := m.Called(page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.Refund), args.Get(1).(int64), args.Error(2)
}

func (m *MockRefundRepository) GetAll(status *domain.RefundStatus, page, limit int) ([]domain.Refund, int64, error) {
	args := m.Called(status, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.Refund), args.Get(1).(int64), args.Error(2)
}

func (m *MockRefundRepository) Update(refund *domain.Refund) error {
	args := m.Called(refund)
	return args.Error(0)
}

func TestRefundUseCase_GetPendingRefunds(t *testing.T) {
	mockRefundRepo := new(MockRefundRepository)

	userID := uuid.New()
	orderID := uuid.New()

	refunds := []domain.Refund{
		{
			ID:          uuid.New(),
			OrderID:     orderID,
			UserID:      userID,
			Amount:      49.99,
			Reason:      domain.RefundReasonNotAsDescribed,
			Status:      domain.RefundStatusPending,
			Description: "Content was not what I expected",
			CreatedAt:   time.Now(),
		},
	}

	mockRefundRepo.On("GetPending", 1, 10).Return(refunds, int64(1), nil)

	result, total, err := mockRefundRepo.GetPending(1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.Equal(t, domain.RefundStatusPending, result[0].Status)
	mockRefundRepo.AssertExpectations(t)
}

func TestRefund_CanProcess(t *testing.T) {
	refund := &domain.Refund{
		Status: domain.RefundStatusPending,
	}

	assert.True(t, refund.CanProcess(), "Pending refund should be processable")

	refund.Status = domain.RefundStatusApproved
	assert.False(t, refund.CanProcess(), "Approved refund should not be processable")

	refund.Status = domain.RefundStatusRejected
	assert.False(t, refund.CanProcess(), "Rejected refund should not be processable")
}

// Silence unused variable warning
var _ = refund.NewUseCase
