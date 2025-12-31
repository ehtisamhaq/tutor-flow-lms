package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/usecase/course"
)

// CourseHandler handles course-related HTTP requests
type CourseHandler struct {
	courseUC *course.UseCase
}

// NewCourseHandler creates a new course handler
func NewCourseHandler(courseUC *course.UseCase) *CourseHandler {
	return &CourseHandler{courseUC: courseUC}
}

// RegisterRoutes registers course routes
func (h *CourseHandler) RegisterRoutes(g *echo.Group, authMW, optionalAuthMW, tutorMW, adminMW echo.MiddlewareFunc) {
	// Public routes
	g.GET("", h.List, optionalAuthMW)
	g.GET("/:idOrSlug", h.Get, optionalAuthMW)
	g.GET("/:id/curriculum", h.GetCurriculum, optionalAuthMW)

	// Tutor routes
	g.POST("", h.Create, authMW, tutorMW)
	g.PUT("/:id", h.Update, authMW, tutorMW)
	g.DELETE("/:id", h.Delete, authMW, tutorMW)
	g.PATCH("/:id/publish", h.Publish, authMW, tutorMW)
	g.PATCH("/:id/archive", h.Archive, authMW, tutorMW)
	g.GET("/my", h.MyCourses, authMW, tutorMW)

	// Module routes
	modules := g.Group("/:courseId/modules", authMW, tutorMW)
	modules.GET("", h.ListModules)
	modules.POST("", h.CreateModule)
	modules.PUT("/:moduleId", h.UpdateModule)
	modules.DELETE("/:moduleId", h.DeleteModule)
	modules.PATCH("/reorder", h.ReorderModules)

	// Lesson routes
	lessons := g.Group("/modules/:moduleId/lessons", authMW, tutorMW)
	lessons.GET("", h.ListLessons)
	lessons.POST("", h.CreateLesson)
	lessons.GET("/:lessonId", h.GetLesson)
	lessons.PUT("/:lessonId", h.UpdateLesson)
	lessons.DELETE("/:lessonId", h.DeleteLesson)

	// Category routes
	categories := g.Group("/categories")
	categories.GET("", h.ListCategories)
	categories.POST("", h.CreateCategory, authMW, adminMW)
	categories.PUT("/:id", h.UpdateCategory, authMW, adminMW)
	categories.DELETE("/:id", h.DeleteCategory, authMW, adminMW)
}

// List godoc
// @Summary List courses
// @Tags Courses
// @Produce json
// @Param search query string false "Search term"
// @Param level query string false "Level filter"
// @Param category_id query string false "Category ID"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param sort_by query string false "Sort by: created_at, price, rating, students"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response
// @Router /courses [get]
func (h *CourseHandler) List(c echo.Context) error {
	var input course.ListInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid query parameters")
	}

	courses, total, err := h.courseUC.List(c.Request().Context(), input, true)
	if err != nil {
		return response.InternalError(c, "Failed to list courses")
	}

	return response.Paginated(c, courses, input.Page, input.Limit, total)
}

// Get godoc
// @Summary Get course
// @Tags Courses
// @Produce json
// @Param idOrSlug path string true "Course ID or slug"
// @Success 200 {object} response.Response{data=domain.Course}
// @Router /courses/{idOrSlug} [get]
func (h *CourseHandler) Get(c echo.Context) error {
	idOrSlug := c.Param("idOrSlug")

	var crs *domain.Course
	var err error

	// Try parsing as UUID first
	if id, parseErr := uuid.Parse(idOrSlug); parseErr == nil {
		crs, err = h.courseUC.GetByID(c.Request().Context(), id)
	} else {
		crs, err = h.courseUC.GetBySlug(c.Request().Context(), idOrSlug)
	}

	if err != nil {
		return response.NotFound(c, "Course not found")
	}

	// Check if unpublished course can be viewed
	if crs.Status != domain.CourseStatusPublished {
		claims, ok := middleware.GetClaims(c)
		if !ok || (crs.InstructorID != claims.UserID && claims.Role != domain.RoleAdmin) {
			return response.NotFound(c, "Course not found")
		}
	}

	return response.Success(c, crs)
}

// GetCurriculum godoc
// @Summary Get course curriculum
// @Tags Courses
// @Produce json
// @Param id path string true "Course ID"
// @Success 200 {object} response.Response
// @Router /courses/{id}/curriculum [get]
func (h *CourseHandler) GetCurriculum(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	modules, err := h.courseUC.GetCurriculum(c.Request().Context(), id)
	if err != nil {
		return response.InternalError(c, "Failed to get curriculum")
	}

	return response.Success(c, modules)
}

// Create godoc
// @Summary Create course
// @Tags Courses
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body course.CreateInput true "Course data"
// @Success 201 {object} response.Response{data=domain.Course}
// @Router /courses [post]
func (h *CourseHandler) Create(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	var input course.CreateInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	crs, err := h.courseUC.Create(c.Request().Context(), claims.UserID, input)
	if err != nil {
		return response.InternalError(c, "Failed to create course")
	}

	return response.Created(c, crs)
}

// Update godoc
// @Summary Update course
// @Tags Courses
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Course ID"
// @Param request body course.UpdateInput true "Course data"
// @Success 200 {object} response.Response{data=domain.Course}
// @Router /courses/{id} [put]
func (h *CourseHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)
	if err := h.checkOwnership(c, id, claims.UserID, claims.Role); err != nil {
		return err
	}

	var input course.UpdateInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	crs, err := h.courseUC.Update(c.Request().Context(), id, input)
	if err != nil {
		return response.InternalError(c, "Failed to update course")
	}

	return response.Success(c, crs)
}

// Delete godoc
// @Summary Delete course
// @Tags Courses
// @Security BearerAuth
// @Param id path string true "Course ID"
// @Success 204
// @Router /courses/{id} [delete]
func (h *CourseHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)
	if err := h.checkOwnership(c, id, claims.UserID, claims.Role); err != nil {
		return err
	}

	if err := h.courseUC.Delete(c.Request().Context(), id); err != nil {
		return response.InternalError(c, "Failed to delete course")
	}

	return response.NoContent(c)
}

// Publish godoc
// @Summary Publish course
// @Tags Courses
// @Security BearerAuth
// @Param id path string true "Course ID"
// @Success 200 {object} response.Response
// @Router /courses/{id}/publish [patch]
func (h *CourseHandler) Publish(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)
	if err := h.checkOwnership(c, id, claims.UserID, claims.Role); err != nil {
		return err
	}

	if err := h.courseUC.Publish(c.Request().Context(), id); err != nil {
		if err == domain.ErrContentLocked {
			return response.BadRequest(c, "Course must have at least one module to publish")
		}
		return response.InternalError(c, "Failed to publish course")
	}

	return response.SuccessWithMessage(c, "Course published successfully", nil)
}

// Archive godoc
// @Summary Archive course
// @Tags Courses
// @Security BearerAuth
// @Param id path string true "Course ID"
// @Success 200 {object} response.Response
// @Router /courses/{id}/archive [patch]
func (h *CourseHandler) Archive(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)
	if err := h.checkOwnership(c, id, claims.UserID, claims.Role); err != nil {
		return err
	}

	if err := h.courseUC.Archive(c.Request().Context(), id); err != nil {
		return response.InternalError(c, "Failed to archive course")
	}

	return response.SuccessWithMessage(c, "Course archived successfully", nil)
}

// MyCourses godoc
// @Summary Get my courses (as instructor)
// @Tags Courses
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /courses/my [get]
func (h *CourseHandler) MyCourses(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	courses, total, err := h.courseUC.GetByInstructor(c.Request().Context(), claims.UserID, 1, 50)
	if err != nil {
		return response.InternalError(c, "Failed to get courses")
	}

	return response.Paginated(c, courses, 1, 50, total)
}

// --- Module Handlers ---

// ListModules godoc
// @Summary List modules for a course
// @Tags Modules
// @Produce json
// @Param courseId path string true "Course ID"
// @Success 200 {object} response.Response
// @Router /courses/{courseId}/modules [get]
func (h *CourseHandler) ListModules(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	modules, err := h.courseUC.GetCurriculum(c.Request().Context(), courseID)
	if err != nil {
		return response.InternalError(c, "Failed to list modules")
	}

	return response.Success(c, modules)
}

// CreateModule godoc
// @Summary Create a module
// @Tags Modules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param courseId path string true "Course ID"
// @Param request body course.CreateModuleInput true "Module data"
// @Success 201 {object} response.Response{data=domain.Module}
// @Router /courses/{courseId}/modules [post]
func (h *CourseHandler) CreateModule(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)
	if err := h.checkOwnership(c, courseID, claims.UserID, claims.Role); err != nil {
		return err
	}

	var input course.CreateModuleInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	module, err := h.courseUC.CreateModule(c.Request().Context(), courseID, input)
	if err != nil {
		return response.InternalError(c, "Failed to create module")
	}

	return response.Created(c, module)
}

// UpdateModule godoc
// @Summary Update a module
// @Tags Modules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param courseId path string true "Course ID"
// @Param moduleId path string true "Module ID"
// @Param request body course.UpdateModuleInput true "Module data"
// @Success 200 {object} response.Response{data=domain.Module}
// @Router /courses/{courseId}/modules/{moduleId} [put]
func (h *CourseHandler) UpdateModule(c echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return response.BadRequest(c, "Invalid module ID")
	}

	var input course.UpdateModuleInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	module, err := h.courseUC.UpdateModule(c.Request().Context(), moduleID, input)
	if err != nil {
		return response.InternalError(c, "Failed to update module")
	}

	return response.Success(c, module)
}

// DeleteModule godoc
// @Summary Delete a module
// @Tags Modules
// @Security BearerAuth
// @Param courseId path string true "Course ID"
// @Param moduleId path string true "Module ID"
// @Success 204
// @Router /courses/{courseId}/modules/{moduleId} [delete]
func (h *CourseHandler) DeleteModule(c echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return response.BadRequest(c, "Invalid module ID")
	}

	if err := h.courseUC.DeleteModule(c.Request().Context(), moduleID); err != nil {
		return response.InternalError(c, "Failed to delete module")
	}

	return response.NoContent(c)
}

// ReorderModules godoc
// @Summary Reorder modules
// @Tags Modules
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param courseId path string true "Course ID"
// @Param request body object{module_ids=[]string} true "Module IDs in order"
// @Success 200 {object} response.Response
// @Router /courses/{courseId}/modules/reorder [patch]
func (h *CourseHandler) ReorderModules(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	var input struct {
		ModuleIDs []uuid.UUID `json:"module_ids"`
	}
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.courseUC.ReorderModules(c.Request().Context(), courseID, input.ModuleIDs); err != nil {
		return response.InternalError(c, "Failed to reorder modules")
	}

	return response.SuccessWithMessage(c, "Modules reordered successfully", nil)
}

// --- Lesson Handlers ---

// ListLessons godoc
// @Summary List lessons for a module
// @Tags Lessons
// @Produce json
// @Param moduleId path string true "Module ID"
// @Success 200 {object} response.Response
// @Router /courses/modules/{moduleId}/lessons [get]
func (h *CourseHandler) ListLessons(c echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return response.BadRequest(c, "Invalid module ID")
	}

	modules, err := h.courseUC.GetCurriculum(c.Request().Context(), moduleID)
	if err != nil {
		return response.InternalError(c, "Failed to list lessons")
	}

	return response.Success(c, modules)
}

// CreateLesson godoc
// @Summary Create a lesson
// @Tags Lessons
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param moduleId path string true "Module ID"
// @Param request body course.CreateLessonInput true "Lesson data"
// @Success 201 {object} response.Response{data=domain.Lesson}
// @Router /courses/modules/{moduleId}/lessons [post]
func (h *CourseHandler) CreateLesson(c echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return response.BadRequest(c, "Invalid module ID")
	}

	var input course.CreateLessonInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	lesson, err := h.courseUC.CreateLesson(c.Request().Context(), moduleID, input)
	if err != nil {
		return response.InternalError(c, "Failed to create lesson")
	}

	return response.Created(c, lesson)
}

// GetLesson godoc
// @Summary Get a lesson
// @Tags Lessons
// @Security BearerAuth
// @Produce json
// @Param moduleId path string true "Module ID"
// @Param lessonId path string true "Lesson ID"
// @Success 200 {object} response.Response{data=domain.Lesson}
// @Router /courses/modules/{moduleId}/lessons/{lessonId} [get]
func (h *CourseHandler) GetLesson(c echo.Context) error {
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return response.BadRequest(c, "Invalid lesson ID")
	}

	lesson, err := h.courseUC.GetLesson(c.Request().Context(), lessonID)
	if err != nil {
		return response.NotFound(c, "Lesson not found")
	}

	return response.Success(c, lesson)
}

// UpdateLesson godoc
// @Summary Update a lesson
// @Tags Lessons
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param moduleId path string true "Module ID"
// @Param lessonId path string true "Lesson ID"
// @Param request body course.UpdateLessonInput true "Lesson data"
// @Success 200 {object} response.Response{data=domain.Lesson}
// @Router /courses/modules/{moduleId}/lessons/{lessonId} [put]
func (h *CourseHandler) UpdateLesson(c echo.Context) error {
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return response.BadRequest(c, "Invalid lesson ID")
	}

	var input course.UpdateLessonInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	lesson, err := h.courseUC.UpdateLesson(c.Request().Context(), lessonID, input)
	if err != nil {
		return response.InternalError(c, "Failed to update lesson")
	}

	return response.Success(c, lesson)
}

// DeleteLesson godoc
// @Summary Delete a lesson
// @Tags Lessons
// @Security BearerAuth
// @Param moduleId path string true "Module ID"
// @Param lessonId path string true "Lesson ID"
// @Success 204
// @Router /courses/modules/{moduleId}/lessons/{lessonId} [delete]
func (h *CourseHandler) DeleteLesson(c echo.Context) error {
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return response.BadRequest(c, "Invalid lesson ID")
	}

	if err := h.courseUC.DeleteLesson(c.Request().Context(), lessonID); err != nil {
		return response.InternalError(c, "Failed to delete lesson")
	}

	return response.NoContent(c)
}

// --- Category Handlers ---

// ListCategories godoc
// @Summary List all categories
// @Tags Categories
// @Produce json
// @Success 200 {object} response.Response{data=[]domain.Category}
// @Router /courses/categories [get]
func (h *CourseHandler) ListCategories(c echo.Context) error {
	categories, err := h.courseUC.ListCategories(c.Request().Context())
	if err != nil {
		return response.InternalError(c, "Failed to list categories")
	}

	return response.Success(c, categories)
}

// CreateCategory godoc
// @Summary Create a category (admin)
// @Tags Categories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body course.CreateCategoryInput true "Category data"
// @Success 201 {object} response.Response{data=domain.Category}
// @Router /courses/categories [post]
func (h *CourseHandler) CreateCategory(c echo.Context) error {
	var input course.CreateCategoryInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	category, err := h.courseUC.CreateCategory(c.Request().Context(), input)
	if err != nil {
		return response.InternalError(c, "Failed to create category")
	}

	return response.Created(c, category)
}

// UpdateCategory godoc
// @Summary Update a category (admin)
// @Tags Categories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param request body course.CreateCategoryInput true "Category data"
// @Success 200 {object} response.Response{data=domain.Category}
// @Router /courses/categories/{id} [put]
func (h *CourseHandler) UpdateCategory(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid category ID")
	}

	var input course.CreateCategoryInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	category, err := h.courseUC.UpdateCategory(c.Request().Context(), id, input)
	if err != nil {
		return response.InternalError(c, "Failed to update category")
	}

	return response.Success(c, category)
}

// DeleteCategory godoc
// @Summary Delete a category (admin)
// @Tags Categories
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 204
// @Router /courses/categories/{id} [delete]
func (h *CourseHandler) DeleteCategory(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid category ID")
	}

	if err := h.courseUC.DeleteCategory(c.Request().Context(), id); err != nil {
		return response.InternalError(c, "Failed to delete category")
	}

	return response.NoContent(c)
}

// Helper to check course ownership
func (h *CourseHandler) checkOwnership(c echo.Context, courseID, userID uuid.UUID, role domain.UserRole) error {
	if role == domain.RoleAdmin {
		return nil
	}

	if err := h.courseUC.ValidateOwnership(c.Request().Context(), courseID, userID); err != nil {
		if err == domain.ErrNotCourseOwner {
			return response.Forbidden(c, "You don't own this course")
		}
		return response.NotFound(c, "Course not found")
	}

	return nil
}
