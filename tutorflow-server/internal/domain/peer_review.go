package domain

import (
	"time"

	"github.com/google/uuid"
)

// PeerReviewStatus defines review states
type PeerReviewStatus string

const (
	PeerReviewStatusPending   PeerReviewStatus = "pending"
	PeerReviewStatusAssigned  PeerReviewStatus = "assigned"
	PeerReviewStatusCompleted PeerReviewStatus = "completed"
	PeerReviewStatusDisputed  PeerReviewStatus = "disputed"
)

// PeerReviewAssignment links assignments to peer reviewers
type PeerReviewAssignment struct {
	ID           uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SubmissionID uuid.UUID        `gorm:"type:uuid;index;not null" json:"submission_id"`
	ReviewerID   uuid.UUID        `gorm:"type:uuid;index;not null" json:"reviewer_id"`
	Status       PeerReviewStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	AssignedAt   time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"assigned_at"`
	DueAt        time.Time        `gorm:"not null" json:"due_at"`
	CompletedAt  *time.Time       `gorm:"" json:"completed_at,omitempty"`
	CreatedAt    time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Submission *Submission `gorm:"foreignKey:SubmissionID" json:"submission,omitempty"`
	Reviewer   *User       `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	Review     *PeerReview `gorm:"foreignKey:AssignmentID" json:"review,omitempty"`
}

// IsOverdue checks if the review is overdue
func (p *PeerReviewAssignment) IsOverdue() bool {
	return p.Status != PeerReviewStatusCompleted && time.Now().After(p.DueAt)
}

// PeerReview represents a peer review submission
type PeerReview struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssignmentID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"assignment_id"`
	Score        float64   `gorm:"type:decimal(5,2)" json:"score"`
	Feedback     string    `gorm:"type:text;not null" json:"feedback"`
	Strengths    string    `gorm:"type:text" json:"strengths,omitempty"`
	Improvements string    `gorm:"type:text" json:"improvements,omitempty"`
	IsAnonymous  bool      `gorm:"default:true" json:"is_anonymous"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Assignment *PeerReviewAssignment `gorm:"foreignKey:AssignmentID" json:"assignment,omitempty"`
}

// PeerReviewCriteria defines rubric criteria for peer reviews
type PeerReviewCriteria struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LessonID    uuid.UUID `gorm:"type:uuid;index;not null" json:"lesson_id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	MaxScore    float64   `gorm:"type:decimal(5,2);not null;default:10" json:"max_score"`
	Weight      float64   `gorm:"type:decimal(3,2);not null;default:1" json:"weight"`
	Order       int       `gorm:"default:0" json:"order"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Lesson *Lesson `gorm:"foreignKey:LessonID" json:"-"`
}

// PeerReviewScore stores individual criteria scores
type PeerReviewScore struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReviewID   uuid.UUID `gorm:"type:uuid;index;not null" json:"review_id"`
	CriteriaID uuid.UUID `gorm:"type:uuid;not null" json:"criteria_id"`
	Score      float64   `gorm:"type:decimal(5,2);not null" json:"score"`
	Comment    string    `gorm:"type:text" json:"comment,omitempty"`

	Review   *PeerReview         `gorm:"foreignKey:ReviewID" json:"-"`
	Criteria *PeerReviewCriteria `gorm:"foreignKey:CriteriaID" json:"criteria,omitempty"`
}

// PeerReviewConfig defines peer review settings for an assignment
type PeerReviewConfig struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LessonID          uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"lesson_id"`
	ReviewsRequired   int       `gorm:"default:3" json:"reviews_required"`    // Reviews each submission needs
	ReviewsToComplete int       `gorm:"default:3" json:"reviews_to_complete"` // Reviews each student must do
	DueDays           int       `gorm:"default:7" json:"due_days"`            // Days after submission deadline
	IsAnonymous       bool      `gorm:"default:true" json:"is_anonymous"`
	ShowScores        bool      `gorm:"default:false" json:"show_scores"` // Show scores to submitter
	MinFeedbackLength int       `gorm:"default:50" json:"min_feedback_length"`
	CreatedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Lesson   *Lesson              `gorm:"foreignKey:LessonID" json:"-"`
	Criteria []PeerReviewCriteria `gorm:"foreignKey:LessonID;references:LessonID" json:"criteria,omitempty"`
}

// PeerReviewRepository interface
type PeerReviewRepository interface {
	// Config
	CreateConfig(config *PeerReviewConfig) error
	GetConfigByLessonID(lessonID uuid.UUID) (*PeerReviewConfig, error)
	UpdateConfig(config *PeerReviewConfig) error

	// Criteria
	CreateCriteria(criteria *PeerReviewCriteria) error
	GetCriteriaByLessonID(lessonID uuid.UUID) ([]PeerReviewCriteria, error)
	UpdateCriteria(criteria *PeerReviewCriteria) error
	DeleteCriteria(id uuid.UUID) error

	// Assignments
	CreateAssignment(assignment *PeerReviewAssignment) error
	GetAssignmentByID(id uuid.UUID) (*PeerReviewAssignment, error)
	GetAssignmentsForReviewer(reviewerID uuid.UUID) ([]PeerReviewAssignment, error)
	GetAssignmentsForSubmission(submissionID uuid.UUID) ([]PeerReviewAssignment, error)
	UpdateAssignment(assignment *PeerReviewAssignment) error

	// Reviews
	CreateReview(review *PeerReview) error
	GetReviewByAssignmentID(assignmentID uuid.UUID) (*PeerReview, error)
	GetReviewsForSubmission(submissionID uuid.UUID) ([]PeerReview, error)
	CreateScore(score *PeerReviewScore) error
	GetScoresByReviewID(reviewID uuid.UUID) ([]PeerReviewScore, error)

	// Auto-assignment
	GetPendingSubmissionsForAssignment(lessonID uuid.UUID) ([]Submission, error)
	GetEligibleReviewers(lessonID, excludeUserID uuid.UUID) ([]User, error)
}

// PeerReviewUseCase interface
type PeerReviewUseCase interface {
	// Config
	ConfigurePeerReview(lessonID uuid.UUID, config *PeerReviewConfig) error
	GetPeerReviewConfig(lessonID uuid.UUID) (*PeerReviewConfig, error)

	// Criteria
	AddCriteria(lessonID uuid.UUID, criteria *PeerReviewCriteria) error
	UpdateCriteria(criteria *PeerReviewCriteria) error
	RemoveCriteria(id uuid.UUID) error

	// Review Process
	AssignReviewers(submissionID uuid.UUID) error
	GetMyReviewAssignments(userID uuid.UUID) ([]PeerReviewAssignment, error)
	SubmitReview(assignmentID uuid.UUID, review *PeerReview, scores []PeerReviewScore) error
	GetReviewsForMySubmission(userID, lessonID uuid.UUID) ([]PeerReview, error)
	DisputeReview(reviewID uuid.UUID, reason string) error

	// Auto-assignment
	AutoAssignPendingReviews() error
	CalculateFinalScore(submissionID uuid.UUID) (float64, error)
}
