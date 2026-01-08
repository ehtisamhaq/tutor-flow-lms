package repository

import (
	"github.com/tutorflow/tutorflow-server/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type peerReviewRepository struct {
	db *gorm.DB
}

// NewPeerReviewRepository creates a new peer review repository
func NewPeerReviewRepository(db *gorm.DB) domain.PeerReviewRepository {
	return &peerReviewRepository{db: db}
}

// Config

func (r *peerReviewRepository) CreateConfig(config *domain.PeerReviewConfig) error {
	return r.db.Create(config).Error
}

func (r *peerReviewRepository) GetConfigByLessonID(lessonID uuid.UUID) (*domain.PeerReviewConfig, error) {
	var config domain.PeerReviewConfig
	err := r.db.Preload("Criteria").Where("lesson_id = ?", lessonID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *peerReviewRepository) UpdateConfig(config *domain.PeerReviewConfig) error {
	return r.db.Save(config).Error
}

// Criteria

func (r *peerReviewRepository) CreateCriteria(criteria *domain.PeerReviewCriteria) error {
	return r.db.Create(criteria).Error
}

func (r *peerReviewRepository) GetCriteriaByLessonID(lessonID uuid.UUID) ([]domain.PeerReviewCriteria, error) {
	var criteria []domain.PeerReviewCriteria
	err := r.db.Where("lesson_id = ?", lessonID).Order("\"order\" ASC").Find(&criteria).Error
	return criteria, err
}

func (r *peerReviewRepository) UpdateCriteria(criteria *domain.PeerReviewCriteria) error {
	return r.db.Save(criteria).Error
}

func (r *peerReviewRepository) DeleteCriteria(id uuid.UUID) error {
	return r.db.Delete(&domain.PeerReviewCriteria{}, id).Error
}

// Assignments

func (r *peerReviewRepository) CreateAssignment(assignment *domain.PeerReviewAssignment) error {
	return r.db.Create(assignment).Error
}

func (r *peerReviewRepository) GetAssignmentByID(id uuid.UUID) (*domain.PeerReviewAssignment, error) {
	var assignment domain.PeerReviewAssignment
	err := r.db.Preload("Submission").Preload("Reviewer").Preload("Review").
		Where("id = ?", id).First(&assignment).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *peerReviewRepository) GetAssignmentsForReviewer(reviewerID uuid.UUID) ([]domain.PeerReviewAssignment, error) {
	var assignments []domain.PeerReviewAssignment
	err := r.db.Preload("Submission").Preload("Review").
		Where("reviewer_id = ?", reviewerID).
		Order("due_at ASC").
		Find(&assignments).Error
	return assignments, err
}

func (r *peerReviewRepository) GetAssignmentsForSubmission(submissionID uuid.UUID) ([]domain.PeerReviewAssignment, error) {
	var assignments []domain.PeerReviewAssignment
	err := r.db.Preload("Reviewer").Preload("Review").
		Where("submission_id = ?", submissionID).
		Find(&assignments).Error
	return assignments, err
}

func (r *peerReviewRepository) UpdateAssignment(assignment *domain.PeerReviewAssignment) error {
	return r.db.Save(assignment).Error
}

// Reviews

func (r *peerReviewRepository) CreateReview(review *domain.PeerReview) error {
	return r.db.Create(review).Error
}

func (r *peerReviewRepository) GetReviewByAssignmentID(assignmentID uuid.UUID) (*domain.PeerReview, error) {
	var review domain.PeerReview
	err := r.db.Where("assignment_id = ?", assignmentID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *peerReviewRepository) GetReviewsForSubmission(submissionID uuid.UUID) ([]domain.PeerReview, error) {
	var reviews []domain.PeerReview
	err := r.db.Joins("JOIN peer_review_assignments ON peer_review_assignments.id = peer_reviews.assignment_id").
		Where("peer_review_assignments.submission_id = ?", submissionID).
		Preload("Assignment").
		Find(&reviews).Error
	return reviews, err
}

func (r *peerReviewRepository) CreateScore(score *domain.PeerReviewScore) error {
	return r.db.Create(score).Error
}

func (r *peerReviewRepository) GetScoresByReviewID(reviewID uuid.UUID) ([]domain.PeerReviewScore, error) {
	var scores []domain.PeerReviewScore
	err := r.db.Preload("Criteria").Where("review_id = ?", reviewID).Find(&scores).Error
	return scores, err
}

// Auto-assignment helpers

func (r *peerReviewRepository) GetPendingSubmissionsForAssignment(lessonID uuid.UUID) ([]domain.Submission, error) {
	var submissions []domain.Submission
	// Get submissions that need more reviewers
	err := r.db.Raw(`
		SELECT s.* FROM assignment_submissions s
		LEFT JOIN peer_review_assignments pra ON pra.submission_id = s.id
		WHERE s.lesson_id = ?
		GROUP BY s.id
		HAVING COUNT(pra.id) < (
			SELECT reviews_required FROM peer_review_configs WHERE lesson_id = ?
		)
	`, lessonID, lessonID).Scan(&submissions).Error
	return submissions, err
}

func (r *peerReviewRepository) GetEligibleReviewers(lessonID, excludeUserID uuid.UUID) ([]domain.User, error) {
	var users []domain.User
	// Get users who have submitted and haven't completed their review quota
	err := r.db.Raw(`
		SELECT u.* FROM users u
		JOIN assignment_submissions s ON s.user_id = u.id
		WHERE s.lesson_id = ? AND u.id != ?
		AND u.id NOT IN (
			SELECT reviewer_id FROM peer_review_assignments pra
			JOIN assignment_submissions s2 ON s2.id = pra.submission_id
			WHERE s2.lesson_id = ?
			GROUP BY reviewer_id
			HAVING COUNT(*) >= (
				SELECT reviews_to_complete FROM peer_review_configs WHERE lesson_id = ?
			)
		)
	`, lessonID, excludeUserID, lessonID, lessonID).Scan(&users).Error
	return users, err
}
