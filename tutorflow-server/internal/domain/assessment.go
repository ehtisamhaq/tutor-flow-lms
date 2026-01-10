package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// QuestionType enum
type QuestionType string

const (
	QuestionTypeSingleChoice   QuestionType = "single_choice"
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeTrueFalse      QuestionType = "true_false"
	QuestionTypeShortAnswer    QuestionType = "short_answer"
	QuestionTypeEssay          QuestionType = "essay"
)

// SubmissionStatus enum
type SubmissionStatus string

const (
	SubmissionStatusPending   SubmissionStatus = "pending"
	SubmissionStatusSubmitted SubmissionStatus = "submitted"
	SubmissionStatusGraded    SubmissionStatus = "graded"
	SubmissionStatusReturned  SubmissionStatus = "returned"
)

// Quiz represents a quiz
type Quiz struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LessonID           uuid.UUID `gorm:"type:uuid;index;not null" json:"lesson_id"`
	Title              string    `gorm:"type:varchar(255);not null" json:"title"`
	Description        *string   `gorm:"type:text" json:"description,omitempty"`
	TimeLimit          *int      `json:"time_limit,omitempty"` // minutes
	PassingScore       float64   `gorm:"type:decimal(5,2);default:60" json:"passing_score"`
	MaxAttempts        int       `gorm:"default:1" json:"max_attempts"`
	ShuffleQuestions   bool      `gorm:"default:false" json:"shuffle_questions"`
	ShowCorrectAnswers bool      `gorm:"default:true" json:"show_correct_answers"`
	IsPublished        bool      `gorm:"default:false" json:"is_published"`
	CreatedAt          time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Lesson    *Lesson        `gorm:"foreignKey:LessonID" json:"-"`
	Questions []QuizQuestion `gorm:"foreignKey:QuizID" json:"questions,omitempty"`
}

func (q *Quiz) TotalPoints() float64 {
	var total float64
	for _, question := range q.Questions {
		total += question.Points
	}
	return total
}

// QuizQuestion represents a question in a quiz
type QuizQuestion struct {
	ID           uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	QuizID       uuid.UUID    `gorm:"type:uuid;index;not null" json:"quiz_id"`
	QuestionType QuestionType `gorm:"type:question_type;not null" json:"question_type"`
	QuestionText string       `gorm:"type:text;not null" json:"question_text"`
	Explanation  *string      `gorm:"type:text" json:"explanation,omitempty"`
	Points       float64      `gorm:"type:decimal(5,2);default:1" json:"points"`
	SortOrder    int          `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Quiz    *Quiz        `gorm:"foreignKey:QuizID" json:"-"`
	Options []QuizOption `gorm:"foreignKey:QuestionID" json:"options,omitempty"`
}

// QuizOption represents an option for a question
type QuizOption struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	QuestionID uuid.UUID `gorm:"type:uuid;index;not null" json:"question_id"`
	OptionText string    `gorm:"type:text;not null" json:"option_text"`
	IsCorrect  bool      `gorm:"default:false" json:"is_correct"`
	SortOrder  int       `gorm:"default:0" json:"sort_order"`

	Question *QuizQuestion `gorm:"foreignKey:QuestionID" json:"-"`
}

// QuizAttempt represents a student's attempt at a quiz
type QuizAttempt struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	QuizID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"quiz_id"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Score       *float64   `gorm:"type:decimal(5,2)" json:"score,omitempty"`
	MaxScore    *float64   `gorm:"type:decimal(5,2)" json:"max_score,omitempty"`
	Percentage  *float64   `gorm:"type:decimal(5,2)" json:"percentage,omitempty"`
	Passed      *bool      `json:"passed,omitempty"`
	StartedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Answers     *string    `gorm:"type:jsonb" json:"answers,omitempty"`

	Quiz *Quiz `gorm:"foreignKey:QuizID" json:"quiz,omitempty"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Assignment represents an assignment
type Assignment struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LessonID            uuid.UUID      `gorm:"type:uuid;index;not null" json:"lesson_id"`
	Title               string         `gorm:"type:varchar(255);not null" json:"title"`
	Description         string         `gorm:"type:text;not null" json:"description"`
	Instructions        *string        `gorm:"type:text" json:"instructions,omitempty"`
	DueDate             *time.Time     `json:"due_date,omitempty"`
	MaxScore            float64        `gorm:"type:decimal(5,2);default:100" json:"max_score"`
	AllowLateSubmission bool           `gorm:"default:true" json:"allow_late_submission"`
	LatePenaltyPercent  float64        `gorm:"type:decimal(5,2);default:0" json:"late_penalty_percent"`
	MaxFileSize         int            `gorm:"default:10485760" json:"max_file_size"` // 10MB
	AllowedFileTypes    pq.StringArray `gorm:"type:text[]" json:"allowed_file_types"`
	CreatedAt           time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Lesson      *Lesson      `gorm:"foreignKey:LessonID" json:"-"`
	Submissions []Submission `gorm:"foreignKey:AssignmentID" json:"submissions,omitempty"`
}

func (a *Assignment) IsOverdue() bool {
	if a.DueDate == nil {
		return false
	}
	return time.Now().After(*a.DueDate)
}

// Submission represents a student's assignment submission
type Submission struct {
	ID           uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssignmentID uuid.UUID        `gorm:"type:uuid;index;not null" json:"assignment_id"`
	UserID       uuid.UUID        `gorm:"type:uuid;index;not null" json:"user_id"`
	Status       SubmissionStatus `gorm:"type:submission_status;not null;default:'pending'" json:"status"`
	Content      *string          `gorm:"type:text" json:"content,omitempty"`
	FileURL      *string          `gorm:"type:varchar(500)" json:"file_url,omitempty"`
	FileName     *string          `gorm:"type:varchar(255)" json:"file_name,omitempty"`
	SubmittedAt  *time.Time       `json:"submitted_at,omitempty"`
	Score        *float64         `gorm:"type:decimal(5,2)" json:"score,omitempty"`
	Feedback     *string          `gorm:"type:text" json:"feedback,omitempty"`
	GradedBy     *uuid.UUID       `gorm:"type:uuid" json:"graded_by,omitempty"`
	GradedAt     *time.Time       `json:"graded_at,omitempty"`
	CreatedAt    time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Assignment *Assignment `gorm:"foreignKey:AssignmentID" json:"assignment,omitempty"`
	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Grader     *User       `gorm:"foreignKey:GradedBy" json:"grader,omitempty"`
}

// AssessmentUseCase interface
type AssessmentUseCase interface {
	// Quiz
	GetQuiz(ctx context.Context, id uuid.UUID) (*Quiz, error)
	GetQuizByLesson(ctx context.Context, lessonID uuid.UUID) (*Quiz, error)
	CreateQuiz(ctx context.Context, quiz *Quiz) error
	UpdateQuiz(ctx context.Context, quiz *Quiz) error
	DeleteQuiz(ctx context.Context, id uuid.UUID) error

	// Quiz Attempts
	StartAttempt(ctx context.Context, userID, quizID uuid.UUID) (*QuizAttempt, error)
	SubmitAttempt(ctx context.Context, attemptID uuid.UUID, answers string) (*QuizAttempt, error)
	GetAttempt(ctx context.Context, id uuid.UUID) (*QuizAttempt, error)
	GetMyAttempts(ctx context.Context, userID, quizID uuid.UUID) ([]QuizAttempt, error)

	// Assignments
	GetAssignment(ctx context.Context, id uuid.UUID) (*Assignment, error)
	GetAssignmentByLesson(ctx context.Context, lessonID uuid.UUID) (*Assignment, error)
	CreateAssignment(ctx context.Context, assignment *Assignment) error
	UpdateAssignment(ctx context.Context, assignment *Assignment) error
	DeleteAssignment(ctx context.Context, id uuid.UUID) error

	// Submissions
	SubmitAssignment(ctx context.Context, userID uuid.UUID, assignmentID uuid.UUID, content, fileURL, fileName *string) (*Submission, error)
	GetSubmission(ctx context.Context, id uuid.UUID) (*Submission, error)
	GetMySubmission(ctx context.Context, userID, assignmentID uuid.UUID) (*Submission, error)
	GetSubmissionsByAssignment(ctx context.Context, assignmentID uuid.UUID, page, limit int) ([]Submission, int64, error)
	GradeSubmission(ctx context.Context, submissionID uuid.UUID, graderID uuid.UUID, score float64, feedback *string) (*Submission, error)
}
