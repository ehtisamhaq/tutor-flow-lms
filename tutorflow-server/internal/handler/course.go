package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"strconv"
	"strings"

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
		return err
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
		return err
	}

	// Check if unpublished course can be viewed
	if crs.Status != domain.CourseStatusPublished {
		claims, ok := middleware.GetClaims(c)
		if !ok || (crs.InstructorID != claims.UserID && claims.Role != domain.RoleAdmin) {
			return domain.ErrCourseNotFound
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
		return err
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

	f, _ := os.OpenFile("/tmp/course_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		defer f.Close()
		fmt.Fprintf(f, "\n--- Create Course Request Received %v ---\n", time.Now())
		fmt.Fprintf(f, "Content-Type: %s\n", c.Request().Header.Get(echo.HeaderContentType))
		fmt.Fprintf(f, "Headers: %+v\n", c.Request().Header)

		if err := c.Request().ParseMultipartForm(32 << 20); err != nil {
			fmt.Fprintf(f, "ParseMultipartForm error: %v\n", err)
		}

		fmt.Fprintf(f, "Form Keys: ")
		for k := range c.Request().Form {
			fmt.Fprintf(f, "%s, ", k)
		}
		fmt.Fprintln(f)

		if c.Request().MultipartForm != nil {
			fmt.Fprintf(f, "Multipart Form Keys: ")
			for k := range c.Request().MultipartForm.Value {
				fmt.Fprintf(f, "%s, ", k)
			}
			fmt.Fprintln(f)
		}
	}

	if c.Request().MultipartForm != nil {
		fmt.Printf("Multipart Form Keys: ")
		for k := range c.Request().MultipartForm.Value {
			fmt.Printf("%s, ", k)
		}
		fmt.Println()
	}

	var input course.CreateInput

	// Use c.FormValue which is reliable in Echo for both multipart and urlencoded
	input.Title = c.FormValue("title")
	input.Level = c.FormValue("level")
	input.Language = c.FormValue("language")

	if f != nil {
		fmt.Fprintf(f, "Extracted values: title=%q, level=%q, language=%q\n", input.Title, input.Level, input.Language)
	}
	fmt.Printf("Extracted: title=%q, level=%q\n", input.Title, input.Level)

	desc := c.FormValue("description")
	if desc != "" {
		input.Description = &desc
	}

	shortDesc := c.FormValue("short_description")
	if shortDesc != "" {
		input.ShortDescription = &shortDesc
	}

	if priceStr := c.FormValue("price"); priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			input.Price = price
		}
	}

	if discountPriceStr := c.FormValue("discount_price"); discountPriceStr != "" {
		if discountPrice, err := strconv.ParseFloat(discountPriceStr, 64); err == nil {
			input.DiscountPrice = &discountPrice
		}
	}

	catID := c.FormValue("category_id")
	if catID != "" {
		input.CategoryID = &catID
	}

	// Parse slices from form
	if c.Request().MultipartForm != nil {
		input.Requirements = c.Request().MultipartForm.Value["requirements"]
		input.WhatYouLearn = c.Request().MultipartForm.Value["what_you_learn"]
		input.CategoryIDs = c.Request().MultipartForm.Value["category_ids"]
	} else {
		input.Requirements = c.Request().Form["requirements"]
		input.WhatYouLearn = c.Request().Form["what_you_learn"]
		input.CategoryIDs = c.Request().Form["category_ids"]
	}

	// Handle thumbnail upload
	file, err := c.FormFile("thumbnail")
	if err == nil {
		thumbnailURL, err := h.uploadFile(file)
		if err != nil {
			return response.InternalError(c, "Failed to upload thumbnail")
		}
		input.ThumbnailURL = &thumbnailURL
	}

	// Parse modules JSON if provided in multipart form
	if modulesJSON := c.FormValue("modules"); modulesJSON != "" {
		if f != nil {
			fmt.Fprintf(f, "Modules JSON: %s\n", modulesJSON)
		}
		if err := json.Unmarshal([]byte(modulesJSON), &input.Modules); err != nil {
			if f != nil {
				fmt.Fprintf(f, "Unmarshal modules error: %v\n", err)
			}
			return response.BadRequest(c, "Invalid modules JSON")
		}
	}

	if err := validator.Validate(input); err != nil {
		return validator.FormatValidationErrors(err)
	}

	crs, err := h.courseUC.Create(c.Request().Context(), claims.UserID, input)
	if err != nil {
		return err
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

	f, _ := os.OpenFile("/tmp/course_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		defer f.Close()
		fmt.Fprintf(f, "\n--- Update Course Request Received %v ID=%s ---\n", time.Now(), id)
		fmt.Fprintf(f, "Content-Type: %s\n", c.Request().Header.Get(echo.HeaderContentType))

		if err := c.Request().ParseMultipartForm(32 << 20); err != nil {
			fmt.Fprintf(f, "ParseMultipartForm error: %v\n", err)
		}

		fmt.Fprintf(f, "Form Keys: ")
		for k := range c.Request().Form {
			fmt.Fprintf(f, "%s, ", k)
		}
		fmt.Fprintln(f)
	}

	var input course.UpdateInput
	contentType := c.Request().Header.Get(echo.HeaderContentType)

	if strings.HasPrefix(contentType, echo.MIMEApplicationJSON) {
		if err := c.Bind(&input); err != nil {
			return response.BadRequest(c, "Invalid JSON body")
		}
	} else {
		// Manual parsing fallback
		if err := c.Request().ParseMultipartForm(32 << 20); err != nil {
			_ = c.Request().ParseForm()
		}

		title := c.Request().FormValue("title")
		if title != "" {
			input.Title = &title
		}

		level := c.Request().FormValue("level")
		if level != "" {
			input.Level = &level
		}

		language := c.Request().FormValue("language")
		if language != "" {
			input.Language = &language
		}

		description := c.Request().FormValue("description")
		if description != "" {
			input.Description = &description
		}

		shortDescription := c.Request().FormValue("short_description")
		if shortDescription != "" {
			input.ShortDescription = &shortDescription
		}

		if priceStr := c.Request().FormValue("price"); priceStr != "" {
			if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
				input.Price = &price
			}
		}

		if discountPriceStr := c.Request().FormValue("discount_price"); discountPriceStr != "" {
			if discountPrice, err := strconv.ParseFloat(discountPriceStr, 64); err == nil {
				input.DiscountPrice = &discountPrice
			}
		}

		if c.Request().MultipartForm != nil {
			input.Requirements = c.Request().MultipartForm.Value["requirements"]
			input.WhatYouLearn = c.Request().MultipartForm.Value["what_you_learn"]
		} else {
			input.Requirements = c.Request().PostForm["requirements"]
			input.WhatYouLearn = c.Request().PostForm["what_you_learn"]
		}
	}

	// Handle thumbnail upload
	file, err := c.FormFile("thumbnail")
	if err == nil {
		thumbnailURL, err := h.uploadFile(file)
		if err != nil {
			return response.InternalError(c, "Failed to upload thumbnail")
		}
		input.ThumbnailURL = &thumbnailURL
	}

	// Parse modules JSON
	if modulesJSON := c.Request().FormValue("modules"); modulesJSON != "" {
		if f != nil {
			fmt.Fprintf(f, "Modules JSON: %s\n", modulesJSON)
		}
		if err := json.Unmarshal([]byte(modulesJSON), &input.Modules); err != nil {
			if f != nil {
				fmt.Fprintf(f, "Unmarshal modules error: %v\n", err)
			}
			return response.BadRequest(c, "Invalid modules JSON")
		}
	}

	if err := validator.Validate(input); err != nil {
		return validator.FormatValidationErrors(err)
	}

	crs, err := h.courseUC.Update(c.Request().Context(), id, input)
	if err != nil {
		return err
	}

	return response.Success(c, crs)
}

func (h *CourseHandler) uploadFile(header *multipart.FileHeader) (string, error) {
	// Simple local file upload implementation
	// Ensure directory exists
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Generate filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(header.Filename))
	dstPath := filepath.Join(uploadDir, filename)

	// Open source file
	src, err := header.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy content
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	// Return URL (relative for simplicity in dev)
	return fmt.Sprintf("/%s/%s", uploadDir, filename), nil
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
		return err
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
		return err
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
		return err
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
		return err
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
		return validator.FormatValidationErrors(err)
	}

	module, err := h.courseUC.CreateModule(c.Request().Context(), courseID, input)
	if err != nil {
		return err
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
		return err
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
		return err
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
		return err
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
		return err
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
		return validator.FormatValidationErrors(err)
	}

	lesson, err := h.courseUC.CreateLesson(c.Request().Context(), moduleID, input)
	if err != nil {
		return err
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
		return err
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
		return err
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
		return err
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
		return err
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
		return validator.FormatValidationErrors(err)
	}

	category, err := h.courseUC.CreateCategory(c.Request().Context(), input)
	if err != nil {
		return err
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
		return err
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
		return err
	}

	return response.NoContent(c)
}

// Helper to check course ownership
func (h *CourseHandler) checkOwnership(c echo.Context, courseID, userID uuid.UUID, role domain.UserRole) error {
	if role == domain.RoleAdmin {
		return nil
	}

	if err := h.courseUC.ValidateOwnership(c.Request().Context(), courseID, userID); err != nil {
		return err
	}

	return nil
}
