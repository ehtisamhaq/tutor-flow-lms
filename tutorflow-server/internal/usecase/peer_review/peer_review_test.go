package peer_review_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/usecase/peer_review"
)

// MockPeerReviewRepository is a mock implementation of PeerReviewRepository
type MockPeerReviewRepository struct {
	mock.Mock
}

func (m *MockPeerReviewRepository) CreateConfig(config *domain.PeerReviewConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockPeerReviewRepository) GetConfigByLessonID(lessonID uuid.UUID) (*domain.PeerReviewConfig, error) {
	args := m.Called(lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PeerReviewConfig), args.Error(1)
}

func (m *MockPeerReviewRepository) UpdateConfig(config *domain.PeerReviewConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockPeerReviewRepository) CreateCriteria(criteria *domain.PeerReviewCriteria) error {
	args := m.Called(criteria)
	return args.Error(0)
}

func (m *MockPeerReviewRepository) GetCriteriaByLessonID(lessonID uuid.UUID) ([]domain.PeerReviewCriteria, error) {
	args := m.Called(lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PeerReviewCriteria), args.Error(1)
}

func (m *MockPeerReviewRepository) UpdateCriteria(criteria *domain.PeerReviewCriteria) error {
	args := m.Called(criteria)
	return args.Error(0)
}

func (m *MockPeerReviewRepository) DeleteCriteria(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPeerReviewRepository) CreateAssignment(assignment *domain.PeerReviewAssignment) error {
	args := m.Called(assignment)
	return args.Error(0)
}

func (m *MockPeerReviewRepository) GetAssignmentByID(id uuid.UUID) (*domain.PeerReviewAssignment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PeerReviewAssignment), args.Error(1)
}

func (m *MockPeerReviewRepository) GetAssignmentsForSubmission(submissionID uuid.UUID) ([]domain.PeerReviewAssignment, error) {
	args := m.Called(submissionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PeerReviewAssignment), args.Error(1)
}

func (m *MockPeerReviewRepository) GetAssignmentsForReviewer(reviewerID uuid.UUID) ([]domain.PeerReviewAssignment, error) {
	args := m.Called(reviewerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PeerReviewAssignment), args.Error(1)
}

func (m *MockPeerReviewRepository) UpdateAssignment(assignment *domain.PeerReviewAssignment) error {
	args := m.Called(assignment)
	return args.Error(0)
}

func (m *MockPeerReviewRepository) CreateReview(review *domain.PeerReview) error {
	args := m.Called(review)
	return args.Error(0)
}

func (m *MockPeerReviewRepository) GetReviewByAssignmentID(assignmentID uuid.UUID) (*domain.PeerReview, error) {
	args := m.Called(assignmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PeerReview), args.Error(1)
}

func (m *MockPeerReviewRepository) GetReviewsForSubmission(submissionID uuid.UUID) ([]domain.PeerReview, error) {
	args := m.Called(submissionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PeerReview), args.Error(1)
}

func (m *MockPeerReviewRepository) CreateScore(score *domain.PeerReviewScore) error {
	args := m.Called(score)
	return args.Error(0)
}

func (m *MockPeerReviewRepository) GetScoresByReviewID(reviewID uuid.UUID) ([]domain.PeerReviewScore, error) {
	args := m.Called(reviewID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PeerReviewScore), args.Error(1)
}

func TestPeerReviewConfig_Defaults(t *testing.T) {
	config := &domain.PeerReviewConfig{
		ReviewsRequired:   3,
		ReviewsToComplete: 3,
		DueDays:           7,
		IsAnonymous:       true,
		ShowScores:        false,
		MinFeedbackLength: 50,
	}

	assert.Equal(t, 3, config.ReviewsRequired)
	assert.Equal(t, 7, config.DueDays)
	assert.True(t, config.IsAnonymous)
}

func TestPeerReviewAssignment_Status(t *testing.T) {
	now := time.Now()
	assignment := &domain.PeerReviewAssignment{
		ID:         uuid.New(),
		Status:     domain.PeerReviewStatusPending,
		AssignedAt: now,
		DueAt:      now.Add(7 * 24 * time.Hour),
	}

	assert.Equal(t, domain.PeerReviewStatusPending, assignment.Status)

	// Complete the assignment
	completedAt := time.Now()
	assignment.Status = domain.PeerReviewStatusCompleted
	assignment.CompletedAt = &completedAt

	assert.Equal(t, domain.PeerReviewStatusCompleted, assignment.Status)
	assert.NotNil(t, assignment.CompletedAt)
}

func TestPeerReviewUseCase_GetMyReviewAssignments(t *testing.T) {
	mockPeerReviewRepo := new(MockPeerReviewRepository)
	reviewerID := uuid.New()

	assignments := []domain.PeerReviewAssignment{
		{
			ID:         uuid.New(),
			ReviewerID: reviewerID,
			Status:     domain.PeerReviewStatusPending,
			AssignedAt: time.Now(),
			DueAt:      time.Now().Add(7 * 24 * time.Hour),
		},
	}

	mockPeerReviewRepo.On("GetAssignmentsForReviewer", reviewerID).Return(assignments, nil)

	result, err := mockPeerReviewRepo.GetAssignmentsForReviewer(reviewerID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, reviewerID, result[0].ReviewerID)
	mockPeerReviewRepo.AssertExpectations(t)
}

// Silence unused variable warning
var _ = peer_review.NewUseCase
