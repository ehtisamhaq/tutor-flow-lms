package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/usecase/quiz"
)

// QuizHandler handles quiz and assignment HTTP requests
type QuizHandler struct {
	quizUC *quiz.UseCase
}

// NewQuizHandler creates a new quiz handler
func NewQuizHandler(quizUC *quiz.UseCase) *QuizHandler {
	return &QuizHandler{quizUC: quizUC}
}

// RegisterRoutes registers quiz and assignment routes
func (h *QuizHandler) RegisterRoutes(g *echo.Group, authMW, tutorMW echo.MiddlewareFunc) {
	// Quiz routes
	quizzes := g.Group("/quizzes")
	quizzes.GET("/:id", h.GetQuiz, authMW)
	quizzes.GET("/lesson/:lessonId", h.GetQuizByLesson, authMW)
	quizzes.POST("", h.CreateQuiz, authMW, tutorMW)
	quizzes.PUT("/:id", h.UpdateQuiz, authMW, tutorMW)
	quizzes.DELETE("/:id", h.DeleteQuiz, authMW, tutorMW)

	// Question routes
	quizzes.POST("/:id/questions", h.AddQuestion, authMW, tutorMW)
	quizzes.PUT("/questions/:questionId", h.UpdateQuestion, authMW, tutorMW)
	quizzes.DELETE("/questions/:questionId", h.DeleteQuestion, authMW, tutorMW)

	// Attempt routes
	quizzes.POST("/:id/attempts", h.StartAttempt, authMW)
	quizzes.POST("/attempts/:attemptId/submit", h.SubmitAttempt, authMW)
	quizzes.GET("/attempts/:attemptId", h.GetAttempt, authMW)
	quizzes.GET("/:id/my-attempts", h.GetMyAttempts, authMW)

	// Assignment routes
	assignments := g.Group("/assignments")
	assignments.GET("/:id", h.GetAssignment, authMW)
	assignments.GET("/lesson/:lessonId", h.GetAssignmentByLesson, authMW)
	assignments.POST("", h.CreateAssignment, authMW, tutorMW)
	assignments.PUT("/:id", h.UpdateAssignment, authMW, tutorMW)
	assignments.DELETE("/:id", h.DeleteAssignment, authMW, tutorMW)

	// Submission routes
	assignments.POST("/:id/submit", h.SubmitAssignment, authMW)
	assignments.GET("/:id/my-submission", h.GetMySubmission, authMW)
	assignments.GET("/:id/submissions", h.GetSubmissions, authMW, tutorMW)
	assignments.POST("/submissions/:submissionId/grade", h.GradeSubmission, authMW, tutorMW)
}

// --- Quiz Handlers ---

// GetQuiz godoc
// @Summary Get quiz by ID
// @Tags Quizzes
// @Security BearerAuth
// @Param id path string true "Quiz ID"
// @Success 200 {object} response.Response{data=domain.Quiz}
// @Router /quizzes/{id} [get]
func (h *QuizHandler) GetQuiz(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid quiz ID")
	}

	quizObj, err := h.quizUC.GetQuiz(c.Request().Context(), id)
	if err != nil {
		return response.NotFound(c, "Quiz not found")
	}

	// Hide correct answers for students
	claims, _ := middleware.GetClaims(c)
	if claims.Role == domain.RoleStudent {
		for i := range quizObj.Questions {
			for j := range quizObj.Questions[i].Options {
				quizObj.Questions[i].Options[j].IsCorrect = false
			}
		}
	}

	return response.Success(c, quizObj)
}

// GetQuizByLesson godoc
// @Summary Get quiz by lesson
// @Tags Quizzes
// @Security BearerAuth
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} response.Response{data=domain.Quiz}
// @Router /quizzes/lesson/{lessonId} [get]
func (h *QuizHandler) GetQuizByLesson(c echo.Context) error {
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return response.BadRequest(c, "Invalid lesson ID")
	}

	quizObj, err := h.quizUC.GetQuizByLesson(c.Request().Context(), lessonID)
	if err != nil {
		return response.NotFound(c, "Quiz not found")
	}

	return response.Success(c, quizObj)
}

// CreateQuiz godoc
// @Summary Create quiz
// @Tags Quizzes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body quiz.CreateQuizInput true "Quiz data"
// @Success 201 {object} response.Response{data=domain.Quiz}
// @Router /quizzes [post]
func (h *QuizHandler) CreateQuiz(c echo.Context) error {
	var input quiz.CreateQuizInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	quizObj, err := h.quizUC.CreateQuiz(c.Request().Context(), input)
	if err != nil {
		return response.InternalError(c, "Failed to create quiz")
	}

	return response.Created(c, quizObj)
}

// UpdateQuiz godoc
// @Summary Update quiz
// @Tags Quizzes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID"
// @Param request body quiz.UpdateQuizInput true "Quiz data"
// @Success 200 {object} response.Response{data=domain.Quiz}
// @Router /quizzes/{id} [put]
func (h *QuizHandler) UpdateQuiz(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid quiz ID")
	}

	var input quiz.UpdateQuizInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	quizObj, err := h.quizUC.UpdateQuiz(c.Request().Context(), id, input)
	if err != nil {
		return response.InternalError(c, "Failed to update quiz")
	}

	return response.Success(c, quizObj)
}

// DeleteQuiz godoc
// @Summary Delete quiz
// @Tags Quizzes
// @Security BearerAuth
// @Param id path string true "Quiz ID"
// @Success 204
// @Router /quizzes/{id} [delete]
func (h *QuizHandler) DeleteQuiz(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid quiz ID")
	}

	if err := h.quizUC.DeleteQuiz(c.Request().Context(), id); err != nil {
		return response.InternalError(c, "Failed to delete quiz")
	}

	return response.NoContent(c)
}

// --- Question Handlers ---

// AddQuestion godoc
// @Summary Add question to quiz
// @Tags Quizzes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Quiz ID"
// @Param request body quiz.AddQuestionInput true "Question data"
// @Success 201 {object} response.Response{data=domain.QuizQuestion}
// @Router /quizzes/{id}/questions [post]
func (h *QuizHandler) AddQuestion(c echo.Context) error {
	quizID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid quiz ID")
	}

	var input quiz.AddQuestionInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	question, err := h.quizUC.AddQuestion(c.Request().Context(), quizID, input)
	if err != nil {
		return response.InternalError(c, "Failed to add question")
	}

	return response.Created(c, question)
}

// UpdateQuestion updates a question
func (h *QuizHandler) UpdateQuestion(c echo.Context) error {
	questionID, err := uuid.Parse(c.Param("questionId"))
	if err != nil {
		return response.BadRequest(c, "Invalid question ID")
	}

	var input quiz.UpdateQuestionInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	question, err := h.quizUC.UpdateQuestion(c.Request().Context(), questionID, input)
	if err != nil {
		return response.InternalError(c, "Failed to update question")
	}

	return response.Success(c, question)
}

// DeleteQuestion deletes a question
func (h *QuizHandler) DeleteQuestion(c echo.Context) error {
	questionID, err := uuid.Parse(c.Param("questionId"))
	if err != nil {
		return response.BadRequest(c, "Invalid question ID")
	}

	if err := h.quizUC.DeleteQuestion(c.Request().Context(), questionID); err != nil {
		return response.InternalError(c, "Failed to delete question")
	}

	return response.NoContent(c)
}

// --- Attempt Handlers ---

// StartAttempt godoc
// @Summary Start quiz attempt
// @Tags Quizzes
// @Security BearerAuth
// @Param id path string true "Quiz ID"
// @Success 201 {object} response.Response{data=domain.QuizAttempt}
// @Router /quizzes/{id}/attempts [post]
func (h *QuizHandler) StartAttempt(c echo.Context) error {
	quizID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid quiz ID")
	}

	claims, _ := middleware.GetClaims(c)

	attempt, err := h.quizUC.StartAttempt(c.Request().Context(), claims.UserID, quizID)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, attempt)
}

// SubmitAttempt godoc
// @Summary Submit quiz answers
// @Tags Quizzes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param attemptId path string true "Attempt ID"
// @Param request body quiz.SubmitAnswerInput true "Answers"
// @Success 200 {object} response.Response{data=domain.QuizAttempt}
// @Router /quizzes/attempts/{attemptId}/submit [post]
func (h *QuizHandler) SubmitAttempt(c echo.Context) error {
	attemptID, err := uuid.Parse(c.Param("attemptId"))
	if err != nil {
		return response.BadRequest(c, "Invalid attempt ID")
	}

	var input quiz.SubmitAnswerInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	attempt, err := h.quizUC.SubmitAttempt(c.Request().Context(), attemptID, input.Answers)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, attempt)
}

// GetAttempt godoc
// @Summary Get attempt by ID
// @Tags Quizzes
// @Security BearerAuth
// @Param attemptId path string true "Attempt ID"
// @Success 200 {object} response.Response{data=domain.QuizAttempt}
// @Router /quizzes/attempts/{attemptId} [get]
func (h *QuizHandler) GetAttempt(c echo.Context) error {
	attemptID, err := uuid.Parse(c.Param("attemptId"))
	if err != nil {
		return response.BadRequest(c, "Invalid attempt ID")
	}

	attempt, err := h.quizUC.GetAttempt(c.Request().Context(), attemptID)
	if err != nil {
		return response.NotFound(c, "Attempt not found")
	}

	return response.Success(c, attempt)
}

// GetMyAttempts godoc
// @Summary Get my attempts for a quiz
// @Tags Quizzes
// @Security BearerAuth
// @Param id path string true "Quiz ID"
// @Success 200 {object} response.Response
// @Router /quizzes/{id}/my-attempts [get]
func (h *QuizHandler) GetMyAttempts(c echo.Context) error {
	quizID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid quiz ID")
	}

	claims, _ := middleware.GetClaims(c)

	attempts, err := h.quizUC.GetMyAttempts(c.Request().Context(), claims.UserID, quizID)
	if err != nil {
		return response.InternalError(c, "Failed to get attempts")
	}

	return response.Success(c, attempts)
}

// --- Assignment Handlers ---

// GetAssignment godoc
// @Summary Get assignment by ID
// @Tags Assignments
// @Security BearerAuth
// @Param id path string true "Assignment ID"
// @Success 200 {object} response.Response{data=domain.Assignment}
// @Router /assignments/{id} [get]
func (h *QuizHandler) GetAssignment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid assignment ID")
	}

	assignment, err := h.quizUC.GetAssignment(c.Request().Context(), id)
	if err != nil {
		return response.NotFound(c, "Assignment not found")
	}

	return response.Success(c, assignment)
}

// GetAssignmentByLesson godoc
// @Summary Get assignment by lesson
// @Tags Assignments
// @Security BearerAuth
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} response.Response{data=domain.Assignment}
// @Router /assignments/lesson/{lessonId} [get]
func (h *QuizHandler) GetAssignmentByLesson(c echo.Context) error {
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return response.BadRequest(c, "Invalid lesson ID")
	}

	assignment, err := h.quizUC.GetAssignmentByLesson(c.Request().Context(), lessonID)
	if err != nil {
		return response.NotFound(c, "Assignment not found")
	}

	return response.Success(c, assignment)
}

// CreateAssignment godoc
// @Summary Create assignment
// @Tags Assignments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body quiz.CreateAssignmentInput true "Assignment data"
// @Success 201 {object} response.Response{data=domain.Assignment}
// @Router /assignments [post]
func (h *QuizHandler) CreateAssignment(c echo.Context) error {
	var input quiz.CreateAssignmentInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	assignment, err := h.quizUC.CreateAssignment(c.Request().Context(), input)
	if err != nil {
		return response.InternalError(c, "Failed to create assignment")
	}

	return response.Created(c, assignment)
}

// UpdateAssignment updates an assignment
func (h *QuizHandler) UpdateAssignment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid assignment ID")
	}

	var input quiz.CreateAssignmentInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	assignment, err := h.quizUC.UpdateAssignment(c.Request().Context(), id, input)
	if err != nil {
		return response.InternalError(c, "Failed to update assignment")
	}

	return response.Success(c, assignment)
}

// DeleteAssignment deletes an assignment
func (h *QuizHandler) DeleteAssignment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid assignment ID")
	}

	if err := h.quizUC.DeleteAssignment(c.Request().Context(), id); err != nil {
		return response.InternalError(c, "Failed to delete assignment")
	}

	return response.NoContent(c)
}

// --- Submission Handlers ---

// SubmitAssignment godoc
// @Summary Submit assignment
// @Tags Assignments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Assignment ID"
// @Param request body quiz.SubmitAssignmentInput true "Submission data"
// @Success 201 {object} response.Response{data=domain.Submission}
// @Router /assignments/{id}/submit [post]
func (h *QuizHandler) SubmitAssignment(c echo.Context) error {
	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid assignment ID")
	}

	claims, _ := middleware.GetClaims(c)

	var input quiz.SubmitAssignmentInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	input.AssignmentID = assignmentID

	submission, err := h.quizUC.SubmitAssignment(c.Request().Context(), claims.UserID, input)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, submission)
}

// GetMySubmission godoc
// @Summary Get my submission for an assignment
// @Tags Assignments
// @Security BearerAuth
// @Param id path string true "Assignment ID"
// @Success 200 {object} response.Response{data=domain.Submission}
// @Router /assignments/{id}/my-submission [get]
func (h *QuizHandler) GetMySubmission(c echo.Context) error {
	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid assignment ID")
	}

	claims, _ := middleware.GetClaims(c)

	submission, err := h.quizUC.GetMySubmission(c.Request().Context(), claims.UserID, assignmentID)
	if err != nil {
		return response.InternalError(c, "Failed to get submission")
	}

	if submission == nil {
		return response.Success(c, nil)
	}

	return response.Success(c, submission)
}

// GetSubmissions godoc
// @Summary Get all submissions for an assignment (instructor)
// @Tags Assignments
// @Security BearerAuth
// @Param id path string true "Assignment ID"
// @Success 200 {object} response.Response
// @Router /assignments/{id}/submissions [get]
func (h *QuizHandler) GetSubmissions(c echo.Context) error {
	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid assignment ID")
	}

	submissions, total, err := h.quizUC.GetSubmissionsByAssignment(c.Request().Context(), assignmentID, 1, 50)
	if err != nil {
		return response.InternalError(c, "Failed to get submissions")
	}

	return response.Paginated(c, submissions, 1, 50, total)
}

// GradeSubmission godoc
// @Summary Grade a submission
// @Tags Assignments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param submissionId path string true "Submission ID"
// @Param request body quiz.GradeSubmissionInput true "Grade data"
// @Success 200 {object} response.Response{data=domain.Submission}
// @Router /assignments/submissions/{submissionId}/grade [post]
func (h *QuizHandler) GradeSubmission(c echo.Context) error {
	submissionID, err := uuid.Parse(c.Param("submissionId"))
	if err != nil {
		return response.BadRequest(c, "Invalid submission ID")
	}

	claims, _ := middleware.GetClaims(c)

	var input quiz.GradeSubmissionInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	submission, err := h.quizUC.GradeSubmission(c.Request().Context(), submissionID, claims.UserID, input)
	if err != nil {
		return response.InternalError(c, "Failed to grade submission")
	}

	return response.Success(c, submission)
}
