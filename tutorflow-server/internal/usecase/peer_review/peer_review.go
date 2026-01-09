package peer_review

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type peerReviewUseCase struct {
	peerReviewRepo repository.PeerReviewRepository
	lessonRepo     repository.LessonRepository
}

// NewPeerReviewUseCase creates a new peer review use case
func NewUseCase(
	peerReviewRepo repository.PeerReviewRepository,
	lessonRepo repository.LessonRepository,
) domain.PeerReviewUseCase {
	return &peerReviewUseCase{
		peerReviewRepo: peerReviewRepo,
		lessonRepo:     lessonRepo,
	}
}

// ConfigurePeerReview configures peer review for a lesson
func (uc *peerReviewUseCase) ConfigurePeerReview(
	ctx context.Context,
	lessonID uuid.UUID,
	config *domain.PeerReviewConfig,
) error {
	// Check if config already exists
	existing, _ := uc.peerReviewRepo.GetConfigByLessonID(ctx, lessonID)
	if existing != nil {
		// Update existing config
		existing.ReviewsRequired = config.ReviewsRequired
		existing.ReviewsToComplete = config.ReviewsToComplete
		existing.DueDays = config.DueDays
		existing.IsAnonymous = config.IsAnonymous
		existing.ShowScores = config.ShowScores
		existing.MinFeedbackLength = config.MinFeedbackLength
		existing.UpdatedAt = time.Now()
		return uc.peerReviewRepo.UpdateConfig(ctx, existing)
	}

	// Create new config
	config.LessonID = lessonID
	return uc.peerReviewRepo.CreateConfig(ctx, config)
}

// GetPeerReviewConfig returns peer review configuration for a lesson
func (uc *peerReviewUseCase) GetPeerReviewConfig(ctx context.Context, lessonID uuid.UUID) (*domain.PeerReviewConfig, error) {
	return uc.peerReviewRepo.GetConfigByLessonID(ctx, lessonID)
}

// AddCriteria adds a review criteria to a lesson
func (uc *peerReviewUseCase) AddCriteria(
	ctx context.Context,
	lessonID uuid.UUID,
	criteria *domain.PeerReviewCriteria,
) error {
	criteria.LessonID = lessonID
	if criteria.MaxScore <= 0 {
		criteria.MaxScore = 10
	}
	if criteria.Weight <= 0 {
		criteria.Weight = 1
	}
	return uc.peerReviewRepo.CreateCriteria(ctx, criteria)
}

// UpdateCriteria updates a review criteria
func (uc *peerReviewUseCase) UpdateCriteria(ctx context.Context, criteria *domain.PeerReviewCriteria) error {
	return uc.peerReviewRepo.UpdateCriteria(ctx, criteria)
}

// RemoveCriteria removes a review criteria
func (uc *peerReviewUseCase) RemoveCriteria(ctx context.Context, id uuid.UUID) error {
	return uc.peerReviewRepo.DeleteCriteria(ctx, id)
}

// AssignReviewers assigns reviewers to a submission
func (uc *peerReviewUseCase) AssignReviewers(ctx context.Context, submissionID uuid.UUID) error {
	// Get submission to find lesson
	assignments, err := uc.peerReviewRepo.GetAssignmentsBySubmissionID(ctx, submissionID)
	if err == nil && len(assignments) > 0 {
		// Already has assignments
		return nil
	}

	// This would need to look up the submission to get lessonID
	// For now, return an error indicating implementation needed
	return errors.New("submission lookup not yet implemented")
}

// GetMyReviewAssignments returns user's review assignments
func (uc *peerReviewUseCase) GetMyReviewAssignments(ctx context.Context, userID uuid.UUID) ([]domain.PeerReviewAssignment, error) {
	return uc.peerReviewRepo.GetAssignmentsByReviewerID(ctx, userID)
}

// SubmitReview submits a peer review
func (uc *peerReviewUseCase) SubmitReview(
	ctx context.Context,
	assignmentID uuid.UUID,
	review *domain.PeerReview,
	scores []domain.PeerReviewScore,
) error {
	// Get the assignment
	assignment, err := uc.peerReviewRepo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return errors.New("assignment not found")
	}

	if assignment.Status == domain.PeerReviewStatusCompleted {
		return errors.New("review already submitted")
	}

	// Get config for validation
	// The LessonID would come from the submission's assignment which is from the lesson
	var config *domain.PeerReviewConfig
	if assignment.Submission != nil && assignment.Submission.Assignment != nil && assignment.Submission.Assignment.LessonID != uuid.Nil {
		config, _ = uc.peerReviewRepo.GetConfigByLessonID(ctx, assignment.Submission.Assignment.LessonID)
	}
	if config != nil && len(review.Feedback) < config.MinFeedbackLength {
		return errors.New("feedback is too short")
	}

	// Calculate total score
	var totalScore float64
	var totalWeight float64
	for _, score := range scores {
		totalScore += score.Score * 1 // Weight would come from criteria
		totalWeight += 1
	}
	if totalWeight > 0 {
		review.Score = totalScore / totalWeight
	}

	// Save review
	review.AssignmentID = assignmentID
	review.IsAnonymous = config != nil && config.IsAnonymous

	if err := uc.peerReviewRepo.CreateReview(ctx, review); err != nil {
		return err
	}

	// Save individual scores
	for _, score := range scores {
		score.ReviewID = review.ID
		uc.peerReviewRepo.CreateScore(ctx, &score)
	}

	// Update assignment status
	now := time.Now()
	assignment.Status = domain.PeerReviewStatusCompleted
	assignment.CompletedAt = &now
	assignment.UpdatedAt = now

	return uc.peerReviewRepo.UpdateAssignment(ctx, assignment)
}

// GetReviewsForMySubmission returns reviews received for user's submission
func (uc *peerReviewUseCase) GetReviewsForMySubmission(
	ctx context.Context,
	userID, lessonID uuid.UUID,
) ([]domain.PeerReview, error) {
	// Would need to find submission by userID and lessonID first
	// Then get reviews for that submission
	return nil, errors.New("implementation requires submission lookup")
}

// DisputeReview disputes a peer review
func (uc *peerReviewUseCase) DisputeReview(ctx context.Context, reviewID uuid.UUID, reason string) error {
	review, err := uc.peerReviewRepo.GetReviewByAssignmentID(ctx, reviewID)
	if err != nil {
		return errors.New("review not found")
	}

	// Get the assignment and mark as disputed
	assignment, err := uc.peerReviewRepo.GetAssignmentByID(ctx, review.AssignmentID)
	if err != nil {
		return errors.New("assignment not found")
	}

	assignment.Status = domain.PeerReviewStatusDisputed
	assignment.UpdatedAt = time.Now()

	return uc.peerReviewRepo.UpdateAssignment(ctx, assignment)
}

// AutoAssignPendingReviews automatically assigns reviewers to pending submissions
func (uc *peerReviewUseCase) AutoAssignPendingReviews(ctx context.Context) error {
	// This would iterate through all lessons with peer review enabled
	// and assign reviewers to submissions that need them
	// Implementation would require listing all lessons with peer review
	return nil
}

// CalculateFinalScore calculates the final score for a submission based on peer reviews
func (uc *peerReviewUseCase) CalculateFinalScore(ctx context.Context, submissionID uuid.UUID) (float64, error) {
	reviews, err := uc.peerReviewRepo.GetReviewsForSubmission(ctx, submissionID)
	if err != nil {
		return 0, err
	}

	if len(reviews) == 0 {
		return 0, errors.New("no reviews found")
	}

	var totalScore float64
	for _, review := range reviews {
		totalScore += review.Score
	}

	return totalScore / float64(len(reviews)), nil
}

// Helper to shuffle slice for random assignment
func shuffleUsers(users []domain.User) {
	rand.Shuffle(len(users), func(i, j int) {
		users[i], users[j] = users[j], users[i]
	})
}
