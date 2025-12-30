package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// QuizRepository
type quizRepository struct {
	db *gorm.DB
}

func NewQuizRepository(db *gorm.DB) repository.QuizRepository {
	return &quizRepository{db: db}
}

func (r *quizRepository) Create(ctx context.Context, quiz *domain.Quiz) error {
	return r.db.WithContext(ctx).Create(quiz).Error
}

func (r *quizRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	var quiz domain.Quiz
	err := r.db.WithContext(ctx).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Questions.Options", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("id = ?", id).
		First(&quiz).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrQuizNotFound
		}
		return nil, err
	}
	return &quiz, nil
}

func (r *quizRepository) GetByLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Quiz, error) {
	var quiz domain.Quiz
	err := r.db.WithContext(ctx).
		Preload("Questions", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Questions.Options", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("lesson_id = ?", lessonID).
		First(&quiz).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrQuizNotFound
		}
		return nil, err
	}
	return &quiz, nil
}

func (r *quizRepository) Update(ctx context.Context, quiz *domain.Quiz) error {
	return r.db.WithContext(ctx).Save(quiz).Error
}

func (r *quizRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete options
		if err := tx.Exec(`
			DELETE FROM quiz_options WHERE question_id IN 
			(SELECT id FROM quiz_questions WHERE quiz_id = ?)
		`, id).Error; err != nil {
			return err
		}
		// Delete questions
		if err := tx.Where("quiz_id = ?", id).Delete(&domain.QuizQuestion{}).Error; err != nil {
			return err
		}
		// Delete quiz
		return tx.Delete(&domain.Quiz{}, "id = ?", id).Error
	})
}

// AddQuestion adds a question to a quiz
func (r *quizRepository) AddQuestion(ctx context.Context, question *domain.QuizQuestion) error {
	return r.db.WithContext(ctx).Create(question).Error
}

// UpdateQuestion updates a question
func (r *quizRepository) UpdateQuestion(ctx context.Context, question *domain.QuizQuestion) error {
	return r.db.WithContext(ctx).Save(question).Error
}

// DeleteQuestion deletes a question and its options
func (r *quizRepository) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete options first
		if err := tx.Where("question_id = ?", id).Delete(&domain.QuizOption{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.QuizQuestion{}, "id = ?", id).Error
	})
}

// AddOption adds an option to a question
func (r *quizRepository) AddOption(ctx context.Context, option *domain.QuizOption) error {
	return r.db.WithContext(ctx).Create(option).Error
}

// UpdateOption updates an option
func (r *quizRepository) UpdateOption(ctx context.Context, option *domain.QuizOption) error {
	return r.db.WithContext(ctx).Save(option).Error
}

// DeleteOption deletes an option
func (r *quizRepository) DeleteOption(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.QuizOption{}, "id = ?", id).Error
}

// GetQuestion gets a question by ID
func (r *quizRepository) GetQuestion(ctx context.Context, id uuid.UUID) (*domain.QuizQuestion, error) {
	var question domain.QuizQuestion
	err := r.db.WithContext(ctx).
		Preload("Options").
		Where("id = ?", id).
		First(&question).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

// QuizAttemptRepository
type quizAttemptRepository struct {
	db *gorm.DB
}

func NewQuizAttemptRepository(db *gorm.DB) repository.QuizAttemptRepository {
	return &quizAttemptRepository{db: db}
}

func (r *quizAttemptRepository) Create(ctx context.Context, attempt *domain.QuizAttempt) error {
	return r.db.WithContext(ctx).Create(attempt).Error
}

func (r *quizAttemptRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.QuizAttempt, error) {
	var attempt domain.QuizAttempt
	err := r.db.WithContext(ctx).
		Preload("Quiz").
		Where("id = ?", id).
		First(&attempt).Error
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *quizAttemptRepository) Update(ctx context.Context, attempt *domain.QuizAttempt) error {
	return r.db.WithContext(ctx).Save(attempt).Error
}

func (r *quizAttemptRepository) GetByUserAndQuiz(ctx context.Context, userID, quizID uuid.UUID) ([]domain.QuizAttempt, error) {
	var attempts []domain.QuizAttempt
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND quiz_id = ?", userID, quizID).
		Order("started_at DESC").
		Find(&attempts).Error
	return attempts, err
}

func (r *quizAttemptRepository) CountByUserAndQuiz(ctx context.Context, userID, quizID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.QuizAttempt{}).
		Where("user_id = ? AND quiz_id = ?", userID, quizID).
		Count(&count).Error
	return int(count), err
}

func (r *quizAttemptRepository) GetLatestByUserAndQuiz(ctx context.Context, userID, quizID uuid.UUID) (*domain.QuizAttempt, error) {
	var attempt domain.QuizAttempt
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND quiz_id = ?", userID, quizID).
		Order("started_at DESC").
		First(&attempt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &attempt, nil
}

// AssignmentRepository
type assignmentRepository struct {
	db *gorm.DB
}

func NewAssignmentRepository(db *gorm.DB) repository.AssignmentRepository {
	return &assignmentRepository{db: db}
}

func (r *assignmentRepository) Create(ctx context.Context, assignment *domain.Assignment) error {
	return r.db.WithContext(ctx).Create(assignment).Error
}

func (r *assignmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Assignment, error) {
	var assignment domain.Assignment
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&assignment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAssignmentNotFound
		}
		return nil, err
	}
	return &assignment, nil
}

func (r *assignmentRepository) GetByLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Assignment, error) {
	var assignment domain.Assignment
	err := r.db.WithContext(ctx).
		Where("lesson_id = ?", lessonID).
		First(&assignment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAssignmentNotFound
		}
		return nil, err
	}
	return &assignment, nil
}

func (r *assignmentRepository) Update(ctx context.Context, assignment *domain.Assignment) error {
	return r.db.WithContext(ctx).Save(assignment).Error
}

func (r *assignmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Assignment{}, "id = ?", id).Error
}

// SubmissionRepository
type submissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) repository.SubmissionRepository {
	return &submissionRepository{db: db}
}

func (r *submissionRepository) Create(ctx context.Context, submission *domain.Submission) error {
	return r.db.WithContext(ctx).Create(submission).Error
}

func (r *submissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	var submission domain.Submission
	err := r.db.WithContext(ctx).
		Preload("Assignment").
		Preload("User").
		Preload("Grader").
		Where("id = ?", id).
		First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *submissionRepository) Update(ctx context.Context, submission *domain.Submission) error {
	return r.db.WithContext(ctx).Save(submission).Error
}

func (r *submissionRepository) GetByUserAndAssignment(ctx context.Context, userID, assignmentID uuid.UUID) (*domain.Submission, error) {
	var submission domain.Submission
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND assignment_id = ?", userID, assignmentID).
		First(&submission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &submission, nil
}

func (r *submissionRepository) GetByAssignment(ctx context.Context, assignmentID uuid.UUID, page, limit int) ([]domain.Submission, int64, error) {
	var submissions []domain.Submission
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Submission{}).Where("assignment_id = ?", assignmentID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Order("submitted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&submissions).Error

	return submissions, total, err
}

func (r *submissionRepository) GetPendingByAssignment(ctx context.Context, assignmentID uuid.UUID) ([]domain.Submission, error) {
	var submissions []domain.Submission
	err := r.db.WithContext(ctx).
		Where("assignment_id = ? AND status IN ?", assignmentID, []string{"pending", "submitted"}).
		Preload("User").
		Order("submitted_at ASC").
		Find(&submissions).Error
	return submissions, err
}

// Helper function to convert answers to JSON
func answersToJSON(answers map[string]interface{}) *string {
	data, err := json.Marshal(answers)
	if err != nil {
		return nil
	}
	str := string(data)
	return &str
}

// Helper function to parse answers from JSON
func parseAnswers(answersJSON *string) map[string]interface{} {
	if answersJSON == nil {
		return nil
	}
	var answers map[string]interface{}
	if err := json.Unmarshal([]byte(*answersJSON), &answers); err != nil {
		return nil
	}
	return answers
}
