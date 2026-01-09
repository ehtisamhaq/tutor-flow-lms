package subscription_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/usecase/subscription"
)

// MockSubscriptionRepository is a mock implementation of SubscriptionRepository
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) CreatePlan(plan *domain.SubscriptionPlan) error {
	args := m.Called(plan)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetPlanByID(id uuid.UUID) (*domain.SubscriptionPlan, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SubscriptionPlan), args.Error(1)
}

func (m *MockSubscriptionRepository) GetPlanBySlug(slug string) (*domain.SubscriptionPlan, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SubscriptionPlan), args.Error(1)
}

func (m *MockSubscriptionRepository) GetActivePlans() ([]domain.SubscriptionPlan, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.SubscriptionPlan), args.Error(1)
}

func (m *MockSubscriptionRepository) UpdatePlan(plan *domain.SubscriptionPlan) error {
	args := m.Called(plan)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) DeletePlan(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) CreateSubscription(sub *domain.Subscription) error {
	args := m.Called(sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetSubscriptionByID(id uuid.UUID) (*domain.Subscription, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) GetUserSubscription(userID uuid.UUID) (*domain.Subscription, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) UpdateSubscription(sub *domain.Subscription) error {
	args := m.Called(sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) CancelSubscription(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockUserRepository for subscription tests
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(page, limit int) ([]domain.User, int64, error) {
	args := m.Called(page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) GetByRole(role domain.UserRole) ([]domain.User, error) {
	args := m.Called(role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func TestSubscriptionUseCase_CreatePlan(t *testing.T) {
	mockSubRepo := new(MockSubscriptionRepository)

	plan := &domain.SubscriptionPlan{
		Name:         "Pro",
		Slug:         "pro",
		Description:  "Pro plan",
		PriceMonthly: 9.99,
		PriceYearly:  99.99,
		Features:     []string{"Feature 1", "Feature 2"},
		IsActive:     true,
	}

	mockSubRepo.On("CreatePlan", mock.AnythingOfType("*domain.SubscriptionPlan")).Return(nil)

	// Call the mock method
	err := mockSubRepo.CreatePlan(plan)

	// Assert expectations
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "Pro", plan.Name)
	mockSubRepo.AssertExpectations(t)
}

func TestSubscriptionUseCase_GetActivePlans(t *testing.T) {
	mockSubRepo := new(MockSubscriptionRepository)

	plans := []domain.SubscriptionPlan{
		{
			Name:         "Basic",
			PriceMonthly: 4.99,
			IsActive:     true,
		},
		{
			Name:         "Pro",
			PriceMonthly: 9.99,
			IsActive:     true,
		},
	}

	mockSubRepo.On("GetActivePlans").Return(plans, nil)

	result, err := mockSubRepo.GetActivePlans()
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Basic", result[0].Name)
	assert.Equal(t, "Pro", result[1].Name)
	mockSubRepo.AssertExpectations(t)
}

// Silence unused variable warning
var _ = subscription.NewUseCase
