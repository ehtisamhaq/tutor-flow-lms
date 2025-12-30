package quiz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines quiz business logic
type UseCase struct {
	quizRepo       repository.QuizRepository
	attemptRepo    repository.QuizAttemptRepository
	assignmentRepo repository.AssignmentRepository
	submissionRepo repository.SubmissionRepository
	enrollmentRepo repository.EnrollmentRepository
	progressRepo   repository.LessonProgressRepository
}

// NewUseCase creates a new quiz use case
func NewUseCase(
	quizRepo repository.QuizRepository,
	attemptRepo repository.QuizAttemptRepository,
	assignmentRepo repository.AssignmentRepository,
	submissionRepo repository.SubmissionRepository,
	enrollmentRepo repository.EnrollmentRepository,
	progressRepo repository.LessonProgressRepository,
) *UseCase {
	return &UseCase{
		quizRepo:       quizRepo,
		attemptRepo:    attemptRepo,
		assignmentRepo: assignmentRepo,
		submissionRepo: submissionRepo,
		enrollmentRepo: enrollmentRepo,
		progressRepo:   progressRepo,
	}
}

// --- Quiz Management ---

// GetQuiz returns quiz by ID
func (uc *UseCase) GetQuiz(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	return uc.quizRepo.GetByID(ctx, id)
}

// GetQuizByLesson returns quiz for a lesson
func (uc *UseCase) GetQuizByLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Quiz, error) {
	return uc.quizRepo.GetByLesson(ctx, lessonID)
}

// CreateQuizInput for creating a quiz
type CreateQuizInput struct {
	LessonID           uuid.UUID `json:"lesson_id" validate:"required"`
	Title              string    `json:"title" validate:"required,min=3,max=255"`
	Description        *string   `json:"description"`
	TimeLimit          *int      `json:"time_limit"`
	PassingScore       float64   `json:"passing_score" validate:"gte=0,lte=100"`
	MaxAttempts        int       `json:"max_attempts" validate:"gte=1"`
	ShuffleQuestions   bool      `json:"shuffle_questions"`
	ShowCorrectAnswers bool      `json:"show_correct_answers"`
}

// CreateQuiz creates a new quiz
func (uc *UseCase) CreateQuiz(ctx context.Context, input CreateQuizInput) (*domain.Quiz, error) {
	quiz := &domain.Quiz{
		LessonID:           input.LessonID,
		Title:              input.Title,
		Description:        input.Description,
		TimeLimit:          input.TimeLimit,
		PassingScore:       input.PassingScore,
		MaxAttempts:        input.MaxAttempts,
		ShuffleQuestions:   input.ShuffleQuestions,
		ShowCorrectAnswers: input.ShowCorrectAnswers,
	}

	if quiz.PassingScore == 0 {
		quiz.PassingScore = 60
	}
	if quiz.MaxAttempts == 0 {
		quiz.MaxAttempts = 1
	}

	if err := uc.quizRepo.Create(ctx, quiz); err != nil {
		return nil, err
	}

	return quiz, nil
}

// UpdateQuizInput for updating a quiz
type UpdateQuizInput struct {
	Title              *string  `json:"title" validate:"omitempty,min=3,max=255"`
	Description        *string  `json:"description"`
	TimeLimit          *int     `json:"time_limit"`
	PassingScore       *float64 `json:"passing_score"`
	MaxAttempts        *int     `json:"max_attempts"`
	ShuffleQuestions   *bool    `json:"shuffle_questions"`
	ShowCorrectAnswers *bool    `json:"show_correct_answers"`
	IsPublished        *bool    `json:"is_published"`
}

// UpdateQuiz updates a quiz
func (uc *UseCase) UpdateQuiz(ctx context.Context, id uuid.UUID, input UpdateQuizInput) (*domain.Quiz, error) {
	quiz, err := uc.quizRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		quiz.Title = *input.Title
	}
	if input.Description != nil {
		quiz.Description = input.Description
	}
	if input.TimeLimit != nil {
		quiz.TimeLimit = input.TimeLimit
	}
	if input.PassingScore != nil {
		quiz.PassingScore = *input.PassingScore
	}
	if input.MaxAttempts != nil {
		quiz.MaxAttempts = *input.MaxAttempts
	}
	if input.ShuffleQuestions != nil {
		quiz.ShuffleQuestions = *input.ShuffleQuestions
	}
	if input.ShowCorrectAnswers != nil {
		quiz.ShowCorrectAnswers = *input.ShowCorrectAnswers
	}
	if input.IsPublished != nil {
		quiz.IsPublished = *input.IsPublished
	}

	if err := uc.quizRepo.Update(ctx, quiz); err != nil {
		return nil, err
	}

	return quiz, nil
}

// DeleteQuiz deletes a quiz
func (uc *UseCase) DeleteQuiz(ctx context.Context, id uuid.UUID) error {
	return uc.quizRepo.Delete(ctx, id)
}

// --- Question Management ---

// AddQuestionInput for adding a question
type AddQuestionInput struct {
	QuestionType domain.QuestionType `json:"question_type" validate:"required,oneof=single_choice multiple_choice true_false short_answer essay"`
	QuestionText string              `json:"question_text" validate:"required"`
	Explanation  *string             `json:"explanation"`
	Points       float64             `json:"points" validate:"gte=0"`
	Options      []OptionInput       `json:"options"`
}

type OptionInput struct {
	OptionText string `json:"option_text" validate:"required"`
	IsCorrect  bool   `json:"is_correct"`
}

// AddQuestion adds a question to a quiz
func (uc *UseCase) AddQuestion(ctx context.Context, quizID uuid.UUID, input AddQuestionInput) (*domain.QuizQuestion, error) {
	quiz, err := uc.quizRepo.GetByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	sortOrder := len(quiz.Questions)
	if input.Points == 0 {
		input.Points = 1
	}

	question := &domain.QuizQuestion{
		QuizID:       quizID,
		QuestionType: input.QuestionType,
		QuestionText: input.QuestionText,
		Explanation:  input.Explanation,
		Points:       input.Points,
		SortOrder:    sortOrder,
	}

	if err := uc.quizRepo.AddQuestion(ctx, question); err != nil {
		return nil, err
	}

	// Add options
	for i, opt := range input.Options {
		option := &domain.QuizOption{
			QuestionID: question.ID,
			OptionText: opt.OptionText,
			IsCorrect:  opt.IsCorrect,
			SortOrder:  i,
		}
		if err := uc.quizRepo.AddOption(ctx, option); err != nil {
			return nil, err
		}
		question.Options = append(question.Options, *option)
	}

	return question, nil
}

// UpdateQuestionInput for updating a question
type UpdateQuestionInput struct {
	QuestionText *string              `json:"question_text"`
	Explanation  *string              `json:"explanation"`
	Points       *float64             `json:"points"`
	QuestionType *domain.QuestionType `json:"question_type"`
}

// UpdateQuestion updates a question
func (uc *UseCase) UpdateQuestion(ctx context.Context, id uuid.UUID, input UpdateQuestionInput) (*domain.QuizQuestion, error) {
	question, err := uc.quizRepo.GetQuestion(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.QuestionText != nil {
		question.QuestionText = *input.QuestionText
	}
	if input.Explanation != nil {
		question.Explanation = input.Explanation
	}
	if input.Points != nil {
		question.Points = *input.Points
	}
	if input.QuestionType != nil {
		question.QuestionType = *input.QuestionType
	}

	if err := uc.quizRepo.UpdateQuestion(ctx, question); err != nil {
		return nil, err
	}

	return question, nil
}

// DeleteQuestion deletes a question
func (uc *UseCase) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	return uc.quizRepo.DeleteQuestion(ctx, id)
}

// UpdateOption updates an option
func (uc *UseCase) UpdateOption(ctx context.Context, id uuid.UUID, optionText string, isCorrect bool) error {
	option := &domain.QuizOption{
		ID:         id,
		OptionText: optionText,
		IsCorrect:  isCorrect,
	}
	return uc.quizRepo.UpdateOption(ctx, option)
}

// DeleteOption deletes an option
func (uc *UseCase) DeleteOption(ctx context.Context, id uuid.UUID) error {
	return uc.quizRepo.DeleteOption(ctx, id)
}

// --- Quiz Attempts ---

// StartAttemptInput for starting a quiz
type StartAttemptInput struct {
	QuizID uuid.UUID `json:"quiz_id" validate:"required"`
}

// StartAttempt starts a new quiz attempt
func (uc *UseCase) StartAttempt(ctx context.Context, userID uuid.UUID, quizID uuid.UUID) (*domain.QuizAttempt, error) {
	quiz, err := uc.quizRepo.GetByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	if !quiz.IsPublished {
		return nil, fmt.Errorf("quiz is not published")
	}

	// Check max attempts
	attemptCount, err := uc.attemptRepo.CountByUserAndQuiz(ctx, userID, quizID)
	if err != nil {
		return nil, err
	}

	if attemptCount >= quiz.MaxAttempts {
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

	// Return quiz with questions for the attempt
	attempt.Quiz = quiz
	return attempt, nil
}

// SubmitAnswerInput for submitting all answers
type SubmitAnswerInput struct {
	Answers map[string]interface{} `json:"answers" validate:"required"` // questionID -> answer(s)
}

// SubmitAttempt submits answers and grades the quiz
func (uc *UseCase) SubmitAttempt(ctx context.Context, attemptID uuid.UUID, answers map[string]interface{}) (*domain.QuizAttempt, error) {
	attempt, err := uc.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	if attempt.CompletedAt != nil {
		return nil, fmt.Errorf("attempt already completed")
	}

	quiz, err := uc.quizRepo.GetByID(ctx, attempt.QuizID)
	if err != nil {
		return nil, err
	}

	// Check time limit
	if quiz.TimeLimit != nil {
		deadline := attempt.StartedAt.Add(time.Duration(*quiz.TimeLimit) * time.Minute)
		if time.Now().After(deadline.Add(1 * time.Minute)) { // 1 minute grace period
			return nil, fmt.Errorf("time limit exceeded")
		}
	}

	// Grade the quiz
	score, maxScore := uc.gradeQuiz(quiz, answers)
	percentage := (score / maxScore) * 100
	passed := percentage >= quiz.PassingScore

	// Save answers
	answersJSON, _ := json.Marshal(answers)
	answersStr := string(answersJSON)
	now := time.Now()

	attempt.Score = &score
	attempt.MaxScore = &maxScore
	attempt.Percentage = &percentage
	attempt.Passed = &passed
	attempt.CompletedAt = &now
	attempt.Answers = &answersStr

	if err := uc.attemptRepo.Update(ctx, attempt); err != nil {
		return nil, err
	}

	return attempt, nil
}

// gradeQuiz calculates score based on answers
func (uc *UseCase) gradeQuiz(quiz *domain.Quiz, answers map[string]interface{}) (float64, float64) {
	var score, maxScore float64

	for _, question := range quiz.Questions {
		maxScore += question.Points

		answer, ok := answers[question.ID.String()]
		if !ok {
			continue
		}

		switch question.QuestionType {
		case domain.QuestionTypeSingleChoice, domain.QuestionTypeTrueFalse:
			answerStr, ok := answer.(string)
			if !ok {
				continue
			}
			for _, opt := range question.Options {
				if opt.ID.String() == answerStr && opt.IsCorrect {
					score += question.Points
					break
				}
			}

		case domain.QuestionTypeMultipleChoice:
			answerSlice, ok := answer.([]interface{})
			if !ok {
				continue
			}
			answerIDs := make(map[string]bool)
			for _, a := range answerSlice {
				if s, ok := a.(string); ok {
					answerIDs[s] = true
				}
			}

			// Check if all correct answers are selected and no incorrect ones
			correctCount := 0
			selectedCorrect := 0
			incorrectSelected := false

			for _, opt := range question.Options {
				if opt.IsCorrect {
					correctCount++
					if answerIDs[opt.ID.String()] {
						selectedCorrect++
					}
				} else if answerIDs[opt.ID.String()] {
					incorrectSelected = true
				}
			}

			if !incorrectSelected && selectedCorrect == correctCount && len(answerIDs) == correctCount {
				score += question.Points
			}

		case domain.QuestionTypeShortAnswer:
			// Short answers require manual grading, but we can do exact match
			answerStr, ok := answer.(string)
			if !ok {
				continue
			}
			for _, opt := range question.Options {
				if opt.IsCorrect && opt.OptionText == answerStr {
					score += question.Points
					break
				}
			}

		case domain.QuestionTypeEssay:
			// Essays require manual grading
			continue
		}
	}

	return score, maxScore
}

// GetAttempt returns attempt by ID
func (uc *UseCase) GetAttempt(ctx context.Context, id uuid.UUID) (*domain.QuizAttempt, error) {
	return uc.attemptRepo.GetByID(ctx, id)
}

// GetMyAttempts returns user's attempts for a quiz
func (uc *UseCase) GetMyAttempts(ctx context.Context, userID, quizID uuid.UUID) ([]domain.QuizAttempt, error) {
	return uc.attemptRepo.GetByUserAndQuiz(ctx, userID, quizID)
}

// --- Assignments ---

// GetAssignment returns assignment by ID
func (uc *UseCase) GetAssignment(ctx context.Context, id uuid.UUID) (*domain.Assignment, error) {
	return uc.assignmentRepo.GetByID(ctx, id)
}

// GetAssignmentByLesson returns assignment for a lesson
func (uc *UseCase) GetAssignmentByLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Assignment, error) {
	return uc.assignmentRepo.GetByLesson(ctx, lessonID)
}

// CreateAssignmentInput for creating an assignment
type CreateAssignmentInput struct {
	LessonID            uuid.UUID  `json:"lesson_id" validate:"required"`
	Title               string     `json:"title" validate:"required,min=3,max=255"`
	Description         string     `json:"description" validate:"required"`
	Instructions        *string    `json:"instructions"`
	DueDate             *time.Time `json:"due_date"`
	MaxScore            float64    `json:"max_score"`
	AllowLateSubmission bool       `json:"allow_late_submission"`
	LatePenaltyPercent  float64    `json:"late_penalty_percent"`
	AllowedFileTypes    []string   `json:"allowed_file_types"`
}

// CreateAssignment creates a new assignment
func (uc *UseCase) CreateAssignment(ctx context.Context, input CreateAssignmentInput) (*domain.Assignment, error) {
	assignment := &domain.Assignment{
		LessonID:            input.LessonID,
		Title:               input.Title,
		Description:         input.Description,
		Instructions:        input.Instructions,
		DueDate:             input.DueDate,
		MaxScore:            input.MaxScore,
		AllowLateSubmission: input.AllowLateSubmission,
		LatePenaltyPercent:  input.LatePenaltyPercent,
		AllowedFileTypes:    input.AllowedFileTypes,
	}

	if assignment.MaxScore == 0 {
		assignment.MaxScore = 100
	}

	if err := uc.assignmentRepo.Create(ctx, assignment); err != nil {
		return nil, err
	}

	return assignment, nil
}

// UpdateAssignment updates an assignment
func (uc *UseCase) UpdateAssignment(ctx context.Context, id uuid.UUID, input CreateAssignmentInput) (*domain.Assignment, error) {
	assignment, err := uc.assignmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	assignment.Title = input.Title
	assignment.Description = input.Description
	assignment.Instructions = input.Instructions
	assignment.DueDate = input.DueDate
	assignment.MaxScore = input.MaxScore
	assignment.AllowLateSubmission = input.AllowLateSubmission
	assignment.LatePenaltyPercent = input.LatePenaltyPercent
	assignment.AllowedFileTypes = input.AllowedFileTypes

	if err := uc.assignmentRepo.Update(ctx, assignment); err != nil {
		return nil, err
	}

	return assignment, nil
}

// DeleteAssignment deletes an assignment
func (uc *UseCase) DeleteAssignment(ctx context.Context, id uuid.UUID) error {
	return uc.assignmentRepo.Delete(ctx, id)
}

// --- Submissions ---

// SubmitAssignmentInput for submitting an assignment
type SubmitAssignmentInput struct {
	AssignmentID uuid.UUID `json:"assignment_id" validate:"required"`
	Content      *string   `json:"content"`
	FileURL      *string   `json:"file_url"`
	FileName     *string   `json:"file_name"`
}

// SubmitAssignment submits or updates a student's assignment
func (uc *UseCase) SubmitAssignment(ctx context.Context, userID uuid.UUID, input SubmitAssignmentInput) (*domain.Submission, error) {
	assignment, err := uc.assignmentRepo.GetByID(ctx, input.AssignmentID)
	if err != nil {
		return nil, err
	}

	// Check if overdue
	if assignment.IsOverdue() && !assignment.AllowLateSubmission {
		return nil, fmt.Errorf("assignment is past due date")
	}

	// Check existing submission
	existing, _ := uc.submissionRepo.GetByUserAndAssignment(ctx, userID, input.AssignmentID)
	now := time.Now()

	if existing != nil {
		// Update existing submission
		existing.Content = input.Content
		existing.FileURL = input.FileURL
		existing.FileName = input.FileName
		existing.SubmittedAt = &now
		existing.Status = domain.SubmissionStatusSubmitted

		if err := uc.submissionRepo.Update(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Create new submission
	submission := &domain.Submission{
		AssignmentID: input.AssignmentID,
		UserID:       userID,
		Content:      input.Content,
		FileURL:      input.FileURL,
		FileName:     input.FileName,
		SubmittedAt:  &now,
		Status:       domain.SubmissionStatusSubmitted,
	}

	if err := uc.submissionRepo.Create(ctx, submission); err != nil {
		return nil, err
	}

	return submission, nil
}

// GetSubmission returns submission by ID
func (uc *UseCase) GetSubmission(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	return uc.submissionRepo.GetByID(ctx, id)
}

// GetMySubmission returns user's submission for an assignment
func (uc *UseCase) GetMySubmission(ctx context.Context, userID, assignmentID uuid.UUID) (*domain.Submission, error) {
	return uc.submissionRepo.GetByUserAndAssignment(ctx, userID, assignmentID)
}

// GetSubmissionsByAssignment returns all submissions for an assignment (instructor view)
func (uc *UseCase) GetSubmissionsByAssignment(ctx context.Context, assignmentID uuid.UUID, page, limit int) ([]domain.Submission, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.submissionRepo.GetByAssignment(ctx, assignmentID, page, limit)
}

// GradeSubmissionInput for grading a submission
type GradeSubmissionInput struct {
	Score    float64 `json:"score" validate:"gte=0"`
	Feedback *string `json:"feedback"`
}

// GradeSubmission grades a student's submission
func (uc *UseCase) GradeSubmission(ctx context.Context, submissionID uuid.UUID, graderID uuid.UUID, input GradeSubmissionInput) (*domain.Submission, error) {
	submission, err := uc.submissionRepo.GetByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	submission.Score = &input.Score
	submission.Feedback = input.Feedback
	submission.GradedBy = &graderID
	submission.GradedAt = &now
	submission.Status = domain.SubmissionStatusGraded

	if err := uc.submissionRepo.Update(ctx, submission); err != nil {
		return nil, err
	}

	return submission, nil
}
