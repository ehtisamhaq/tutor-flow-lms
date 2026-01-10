package assessment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type assessmentUseCase struct {
	quizRepo       repository.QuizRepository
	attemptRepo    repository.QuizAttemptRepository
	assignmentRepo repository.AssignmentRepository
	submissionRepo repository.SubmissionRepository
}

func NewUseCase(
	quizRepo repository.QuizRepository,
	attemptRepo repository.QuizAttemptRepository,
	assignmentRepo repository.AssignmentRepository,
	submissionRepo repository.SubmissionRepository,
) domain.AssessmentUseCase {
	return &assessmentUseCase{
		quizRepo:       quizRepo,
		attemptRepo:    attemptRepo,
		assignmentRepo: assignmentRepo,
		submissionRepo: submissionRepo,
	}
}

// --- Quiz ---

func (uc *assessmentUseCase) GetQuiz(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	return uc.quizRepo.GetByID(ctx, id)
}

func (uc *assessmentUseCase) GetQuizByLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Quiz, error) {
	return uc.quizRepo.GetByLesson(ctx, lessonID)
}

func (uc *assessmentUseCase) CreateQuiz(ctx context.Context, quiz *domain.Quiz) error {
	return uc.quizRepo.Create(ctx, quiz)
}

func (uc *assessmentUseCase) UpdateQuiz(ctx context.Context, quiz *domain.Quiz) error {
	return uc.quizRepo.Update(ctx, quiz)
}

func (uc *assessmentUseCase) DeleteQuiz(ctx context.Context, id uuid.UUID) error {
	return uc.quizRepo.Delete(ctx, id)
}

// --- Quiz Attempts ---

func (uc *assessmentUseCase) StartAttempt(ctx context.Context, userID, quizID uuid.UUID) (*domain.QuizAttempt, error) {
	quiz, err := uc.quizRepo.GetByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// Check max attempts
	attempts, err := uc.attemptRepo.GetByUserAndQuiz(ctx, userID, quizID)
	if err == nil && len(attempts) >= quiz.MaxAttempts {
		return nil, fmt.Errorf("maximum attempts reached")
	}

	attempt := &domain.QuizAttempt{
		QuizID:    quizID,
		UserID:    userID,
		StartedAt: time.Now(),
	}

	if err := uc.attemptRepo.Create(ctx, attempt); err != nil {
		return nil, err
	}

	return attempt, nil
}

func (uc *assessmentUseCase) SubmitAttempt(ctx context.Context, attemptID uuid.UUID, answers string) (*domain.QuizAttempt, error) {
	attempt, err := uc.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	attempt.CompletedAt = &now
	attempt.Answers = &answers

	// Grade the attempt (simplified logic)
	// In a real app, we would parse the JSON answers and compare with QuizQuestion answers
	// For now, we'll just mark it as completed.

	if err := uc.attemptRepo.Update(ctx, attempt); err != nil {
		return nil, err
	}

	return attempt, nil
}

func (uc *assessmentUseCase) GetAttempt(ctx context.Context, id uuid.UUID) (*domain.QuizAttempt, error) {
	return uc.attemptRepo.GetByID(ctx, id)
}

func (uc *assessmentUseCase) GetMyAttempts(ctx context.Context, userID, quizID uuid.UUID) ([]domain.QuizAttempt, error) {
	return uc.attemptRepo.GetByUserAndQuiz(ctx, userID, quizID)
}

// --- Assignments ---

func (uc *assessmentUseCase) GetAssignment(ctx context.Context, id uuid.UUID) (*domain.Assignment, error) {
	return uc.assignmentRepo.GetByID(ctx, id)
}

func (uc *assessmentUseCase) GetAssignmentByLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Assignment, error) {
	return uc.assignmentRepo.GetByLesson(ctx, lessonID)
}

func (uc *assessmentUseCase) CreateAssignment(ctx context.Context, assignment *domain.Assignment) error {
	return uc.assignmentRepo.Create(ctx, assignment)
}

func (uc *assessmentUseCase) UpdateAssignment(ctx context.Context, assignment *domain.Assignment) error {
	return uc.assignmentRepo.Update(ctx, assignment)
}

func (uc *assessmentUseCase) DeleteAssignment(ctx context.Context, id uuid.UUID) error {
	return uc.assignmentRepo.Delete(ctx, id)
}

// --- Submissions ---

func (uc *assessmentUseCase) SubmitAssignment(ctx context.Context, userID uuid.UUID, assignmentID uuid.UUID, content, fileURL, fileName *string) (*domain.Submission, error) {
	existing, _ := uc.submissionRepo.GetByUserAndAssignment(ctx, userID, assignmentID)
	now := time.Now()

	if existing != nil {
		existing.Content = content
		existing.FileURL = fileURL
		existing.FileName = fileName
		existing.SubmittedAt = &now
		existing.Status = domain.SubmissionStatusSubmitted
		if err := uc.submissionRepo.Update(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	submission := &domain.Submission{
		AssignmentID: assignmentID,
		UserID:       userID,
		Content:      content,
		FileURL:      fileURL,
		FileName:     fileName,
		SubmittedAt:  &now,
		Status:       domain.SubmissionStatusSubmitted,
	}

	if err := uc.submissionRepo.Create(ctx, submission); err != nil {
		return nil, err
	}

	return submission, nil
}

func (uc *assessmentUseCase) GetSubmission(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	return uc.submissionRepo.GetByID(ctx, id)
}

func (uc *assessmentUseCase) GetMySubmission(ctx context.Context, userID, assignmentID uuid.UUID) (*domain.Submission, error) {
	return uc.submissionRepo.GetByUserAndAssignment(ctx, userID, assignmentID)
}

func (uc *assessmentUseCase) GetSubmissionsByAssignment(ctx context.Context, assignmentID uuid.UUID, page, limit int) ([]domain.Submission, int64, error) {
	return uc.submissionRepo.GetByAssignment(ctx, assignmentID, page, limit)
}

func (uc *assessmentUseCase) GradeSubmission(ctx context.Context, submissionID uuid.UUID, graderID uuid.UUID, score float64, feedback *string) (*domain.Submission, error) {
	submission, err := uc.submissionRepo.GetByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	submission.Score = &score
	submission.Feedback = feedback
	submission.GradedBy = &graderID
	submission.GradedAt = &now
	submission.Status = domain.SubmissionStatusGraded

	if err := uc.submissionRepo.Update(ctx, submission); err != nil {
		return nil, err
	}

	return submission, nil
}
