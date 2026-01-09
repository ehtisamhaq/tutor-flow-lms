package bundle_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/usecase/bundle"
)

// MockBundleRepository is a mock implementation of BundleRepository
type MockBundleRepository struct {
	mock.Mock
}

func (m *MockBundleRepository) Create(bundle *domain.Bundle) error {
	args := m.Called(bundle)
	return args.Error(0)
}

func (m *MockBundleRepository) GetByID(id uuid.UUID) (*domain.Bundle, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Bundle), args.Error(1)
}

func (m *MockBundleRepository) GetBySlug(slug string) (*domain.Bundle, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Bundle), args.Error(1)
}

func (m *MockBundleRepository) GetActive(page, limit int) ([]domain.Bundle, int64, error) {
	args := m.Called(page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.Bundle), args.Get(1).(int64), args.Error(2)
}

func (m *MockBundleRepository) Update(bundle *domain.Bundle) error {
	args := m.Called(bundle)
	return args.Error(0)
}

func (m *MockBundleRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockBundleRepository) AddCourse(bundleID, courseID uuid.UUID, order int) error {
	args := m.Called(bundleID, courseID, order)
	return args.Error(0)
}

func (m *MockBundleRepository) RemoveCourse(bundleID, courseID uuid.UUID) error {
	args := m.Called(bundleID, courseID)
	return args.Error(0)
}

func (m *MockBundleRepository) GetCourses(bundleID uuid.UUID) ([]domain.Course, error) {
	args := m.Called(bundleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Course), args.Error(1)
}

func (m *MockBundleRepository) RecordPurchase(purchase *domain.BundlePurchase) error {
	args := m.Called(purchase)
	return args.Error(0)
}

func (m *MockBundleRepository) GetUserPurchases(userID uuid.UUID) ([]domain.BundlePurchase, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.BundlePurchase), args.Error(1)
}

func (m *MockBundleRepository) IncrementPurchaseCount(bundleID uuid.UUID) error {
	args := m.Called(bundleID)
	return args.Error(0)
}

func TestBundle_IsAvailable(t *testing.T) {
	maxPurchases := 100
	bundle := &domain.Bundle{
		IsActive:      true,
		MaxPurchases:  &maxPurchases,
		PurchaseCount: 50,
	}

	assert.True(t, bundle.IsAvailable(), "Active bundle with purchases available should be available")

	bundle.IsActive = false
	assert.False(t, bundle.IsAvailable(), "Inactive bundle should not be available")

	bundle.IsActive = true
	*bundle.MaxPurchases = 50
	bundle.PurchaseCount = 50
	assert.False(t, bundle.IsAvailable(), "Bundle at max purchases should not be available")
}

func TestBundle_Savings(t *testing.T) {
	bundle := &domain.Bundle{
		OriginalPrice: 100,
		BundlePrice:   80,
	}

	savings := bundle.Savings()
	assert.Equal(t, 20.0, savings, "Savings should be original - bundle price")
}

func TestBundleUseCase_GetActiveBundles(t *testing.T) {
	mockBundleRepo := new(MockBundleRepository)

	bundles := []domain.Bundle{
		{
			ID:              uuid.New(),
			Title:           "Web Dev Bundle",
			OriginalPrice:   199.99,
			BundlePrice:     149.99,
			DiscountPercent: 25,
			IsActive:        true,
		},
	}

	mockBundleRepo.On("GetActive", 1, 10).Return(bundles, int64(1), nil)

	result, total, err := mockBundleRepo.GetActive(1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.Equal(t, "Web Dev Bundle", result[0].Title)
	mockBundleRepo.AssertExpectations(t)
}

// Silence unused variable warning
var _ = bundle.NewUseCase
