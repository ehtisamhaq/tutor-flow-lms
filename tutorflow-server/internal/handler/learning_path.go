package handler

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/repository"
	"github.com/tutorflow/tutorflow-server/internal/usecase/learningpath"
)

// LearningPathHandler handles learning path HTTP requests
type LearningPathHandler struct {
	pathUC *learningpath.UseCase
}

// NewLearningPathHandler creates a new learning path handler
func NewLearningPathHandler(pathUC *learningpath.UseCase) *LearningPathHandler {
	return &LearningPathHandler{pathUC: pathUC}
}

// RegisterRoutes registers learning path routes
func (h *LearningPathHandler) RegisterRoutes(g *echo.Group, authMW, optionalAuthMW, adminMW echo.MiddlewareFunc) {
	paths := g.Group("/learning-paths")

	// Public routes
	paths.GET("", h.ListPaths, optionalAuthMW)
	paths.GET("/featured", h.GetFeaturedPaths)
	paths.GET("/category/:categoryId", h.GetPathsByCategory)
	paths.GET("/:idOrSlug", h.GetPath, optionalAuthMW)
	paths.GET("/:id/courses", h.GetPathCourses)

	// Authenticated student routes
	paths.POST("/:id/enroll", h.Enroll, authMW)
	paths.GET("/:id/progress", h.GetProgress, authMW)
	paths.GET("/my/enrollments", h.GetMyEnrollments, authMW)

	// Admin routes
	paths.POST("", h.CreatePath, authMW, adminMW)
	paths.PUT("/:id", h.UpdatePath, authMW, adminMW)
	paths.DELETE("/:id", h.DeletePath, authMW, adminMW)
	paths.POST("/:id/courses", h.AddCourse, authMW, adminMW)
	paths.DELETE("/:id/courses/:courseId", h.RemoveCourse, authMW, adminMW)
	paths.PUT("/:id/courses/reorder", h.ReorderCourses, authMW, adminMW)
}

// ListPaths godoc
// @Summary List learning paths
// @Tags Learning Paths
// @Param category_id query string false "Category ID"
// @Param level query string false "Level"
// @Param search query string false "Search"
// @Success 200 {object} response.Response
// @Router /learning-paths [get]
func (h *LearningPathHandler) ListPaths(c echo.Context) error {
	var categoryID *uuid.UUID
	if catStr := c.QueryParam("category_id"); catStr != "" {
		if id, err := uuid.Parse(catStr); err == nil {
			categoryID = &id
		}
	}

	isPublished := true
	filters := repository.LearningPathFilters{
		CategoryID:  categoryID,
		Level:       c.QueryParam("level"),
		Search:      c.QueryParam("search"),
		IsPublished: &isPublished,
		Page:        1,
		Limit:       20,
	}

	if p := c.QueryParam("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			filters.Page = val
		}
	}
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			filters.Limit = val
		}
	}

	paths, total, err := h.pathUC.ListPaths(c.Request().Context(), filters)
	if err != nil {
		return response.InternalError(c, "Failed to list paths")
	}

	return response.Paginated(c, paths, filters.Page, filters.Limit, total)
}

// GetFeaturedPaths godoc
// @Summary Get featured learning paths
// @Tags Learning Paths
// @Success 200 {object} response.Response
// @Router /learning-paths/featured [get]
func (h *LearningPathHandler) GetFeaturedPaths(c echo.Context) error {
	limit := 6
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	paths, err := h.pathUC.GetFeaturedPaths(c.Request().Context(), limit)
	if err != nil {
		return response.InternalError(c, "Failed to get featured paths")
	}

	return response.Success(c, paths)
}

// GetPathsByCategory godoc
// @Summary Get paths by category
// @Tags Learning Paths
// @Param categoryId path string true "Category ID"
// @Success 200 {object} response.Response
// @Router /learning-paths/category/{categoryId} [get]
func (h *LearningPathHandler) GetPathsByCategory(c echo.Context) error {
	categoryID, err := uuid.Parse(c.Param("categoryId"))
	if err != nil {
		return response.BadRequest(c, "Invalid category ID")
	}

	limit := 10
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	paths, err := h.pathUC.GetPathsByCategory(c.Request().Context(), categoryID, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get paths")
	}

	return response.Success(c, paths)
}

// GetPath godoc
// @Summary Get a learning path
// @Tags Learning Paths
// @Param idOrSlug path string true "Path ID or slug"
// @Success 200 {object} response.Response{data=domain.LearningPath}
// @Router /learning-paths/{idOrSlug} [get]
func (h *LearningPathHandler) GetPath(c echo.Context) error {
	idOrSlug := c.Param("idOrSlug")

	var path *domain.LearningPath
	var err error

	// Try as UUID first
	if id, parseErr := uuid.Parse(idOrSlug); parseErr == nil {
		path, err = h.pathUC.GetPath(c.Request().Context(), id)
	} else {
		path, err = h.pathUC.GetPathBySlug(c.Request().Context(), idOrSlug)
	}

	if err != nil || path == nil {
		return response.NotFound(c, "Learning path not found")
	}

	return response.Success(c, path)
}

// GetPathCourses godoc
// @Summary Get courses in a learning path
// @Tags Learning Paths
// @Param id path string true "Path ID"
// @Success 200 {object} response.Response
// @Router /learning-paths/{id}/courses [get]
func (h *LearningPathHandler) GetPathCourses(c echo.Context) error {
	pathID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid path ID")
	}

	courses, err := h.pathUC.GetPathCourses(c.Request().Context(), pathID)
	if err != nil {
		return response.InternalError(c, "Failed to get courses")
	}

	return response.Success(c, courses)
}

// CreatePath godoc
// @Summary Create a learning path
// @Tags Learning Paths
// @Security BearerAuth
// @Accept json
// @Param request body learningpath.CreatePathInput true "Path data"
// @Success 201 {object} response.Response{data=domain.LearningPath}
// @Router /learning-paths [post]
func (h *LearningPathHandler) CreatePath(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	var input learningpath.CreatePathInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	path, err := h.pathUC.CreatePath(c.Request().Context(), claims.UserID, isAdmin, input)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, path)
}

// UpdatePath godoc
// @Summary Update a learning path
// @Tags Learning Paths
// @Security BearerAuth
// @Accept json
// @Param id path string true "Path ID"
// @Param request body learningpath.UpdatePathInput true "Update data"
// @Success 200 {object} response.Response{data=domain.LearningPath}
// @Router /learning-paths/{id} [put]
func (h *LearningPathHandler) UpdatePath(c echo.Context) error {
	pathID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid path ID")
	}

	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	var input learningpath.UpdatePathInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	path, err := h.pathUC.UpdatePath(c.Request().Context(), pathID, isAdmin, input)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, path)
}

// DeletePath godoc
// @Summary Delete a learning path
// @Tags Learning Paths
// @Security BearerAuth
// @Param id path string true "Path ID"
// @Success 204
// @Router /learning-paths/{id} [delete]
func (h *LearningPathHandler) DeletePath(c echo.Context) error {
	pathID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid path ID")
	}

	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	if err := h.pathUC.DeletePath(c.Request().Context(), pathID, isAdmin); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.NoContent(c)
}

// AddCourse godoc
// @Summary Add a course to a learning path
// @Tags Learning Paths
// @Security BearerAuth
// @Accept json
// @Param id path string true "Path ID"
// @Param request body learningpath.AddCourseInput true "Course data"
// @Success 201 {object} response.Response
// @Router /learning-paths/{id}/courses [post]
func (h *LearningPathHandler) AddCourse(c echo.Context) error {
	pathID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid path ID")
	}

	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	var input learningpath.AddCourseInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.pathUC.AddCourse(c.Request().Context(), pathID, isAdmin, input); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, map[string]string{"message": "Course added to path"})
}

// RemoveCourse godoc
// @Summary Remove a course from a learning path
// @Tags Learning Paths
// @Security BearerAuth
// @Param id path string true "Path ID"
// @Param courseId path string true "Course ID"
// @Success 204
// @Router /learning-paths/{id}/courses/{courseId} [delete]
func (h *LearningPathHandler) RemoveCourse(c echo.Context) error {
	pathID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid path ID")
	}
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	if err := h.pathUC.RemoveCourse(c.Request().Context(), pathID, courseID, isAdmin); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.NoContent(c)
}

// ReorderInput for reordering courses
type ReorderInput struct {
	CourseOrder []uuid.UUID `json:"course_order" validate:"required"`
}

// ReorderCourses godoc
// @Summary Reorder courses in a learning path
// @Tags Learning Paths
// @Security BearerAuth
// @Accept json
// @Param id path string true "Path ID"
// @Param request body ReorderInput true "Course order"
// @Success 200 {object} response.Response
// @Router /learning-paths/{id}/courses/reorder [put]
func (h *LearningPathHandler) ReorderCourses(c echo.Context) error {
	pathID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid path ID")
	}

	claims, _ := middleware.GetClaims(c)
	isAdmin := claims.Role == domain.RoleAdmin

	var input ReorderInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.pathUC.ReorderCourses(c.Request().Context(), pathID, isAdmin, input.CourseOrder); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.SuccessWithMessage(c, "Courses reordered", nil)
}

// Enroll godoc
// @Summary Enroll in a learning path
// @Tags Learning Paths
// @Security BearerAuth
// @Param id path string true "Path ID"
// @Success 201 {object} response.Response
// @Router /learning-paths/{id}/enroll [post]
func (h *LearningPathHandler) Enroll(c echo.Context) error {
	pathID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid path ID")
	}

	claims, _ := middleware.GetClaims(c)

	enrollment, err := h.pathUC.EnrollInPath(c.Request().Context(), pathID, claims.UserID)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, enrollment)
}

// GetMyEnrollments godoc
// @Summary Get my learning path enrollments
// @Tags Learning Paths
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /learning-paths/my/enrollments [get]
func (h *LearningPathHandler) GetMyEnrollments(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	page, limit := 1, 20
	if p := c.QueryParam("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			page = val
		}
	}
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	enrollments, total, err := h.pathUC.GetMyEnrollments(c.Request().Context(), claims.UserID, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get enrollments")
	}

	return response.Paginated(c, enrollments, page, limit, total)
}

// GetProgress godoc
// @Summary Get my progress in a learning path
// @Tags Learning Paths
// @Security BearerAuth
// @Param id path string true "Path ID"
// @Success 200 {object} response.Response{data=domain.LearningPathProgress}
// @Router /learning-paths/{id}/progress [get]
func (h *LearningPathHandler) GetProgress(c echo.Context) error {
	pathID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid path ID")
	}

	claims, _ := middleware.GetClaims(c)

	progress, err := h.pathUC.GetProgress(c.Request().Context(), pathID, claims.UserID)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, progress)
}
