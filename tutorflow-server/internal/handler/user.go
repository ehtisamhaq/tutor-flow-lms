package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/usecase/user"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userUC *user.UseCase
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUC *user.UseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

// RegisterRoutes registers user routes
func (h *UserHandler) RegisterRoutes(g *echo.Group, authMW, adminMW, managerMW echo.MiddlewareFunc) {
	// Admin routes
	g.GET("", h.List, authMW, managerMW)
	g.POST("", h.Create, authMW, adminMW)
	g.GET("/:id", h.GetByID, authMW)
	g.PUT("/:id", h.Update, authMW)
	g.DELETE("/:id", h.Delete, authMW, adminMW)
	g.PATCH("/:id/status", h.UpdateStatus, authMW, managerMW)
	g.PATCH("/:id/role", h.UpdateRole, authMW, adminMW)

	// Tutor routes
	tutors := g.Group("/tutors")
	tutors.GET("", h.ListTutors)
	tutors.GET("/:id", h.GetTutor)
	tutors.PUT("/:id/profile", h.UpdateTutorProfile, authMW)
}

// List godoc
// @Summary List users
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param role query string false "Filter by role"
// @Param status query string false "Filter by status"
// @Param search query string false "Search term"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response
// @Router /users [get]
func (h *UserHandler) List(c echo.Context) error {
	var input user.ListInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid query parameters")
	}

	users, total, err := h.userUC.List(c.Request().Context(), input)
	if err != nil {
		return response.InternalError(c, "Failed to list users")
	}

	return response.Paginated(c, users, input.Page, input.Limit, total)
}

// GetByID godoc
// @Summary Get user by ID
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=domain.User}
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid user ID")
	}

	// Check if user can access this profile
	claims, _ := middleware.GetClaims(c)
	if claims.UserID != id && claims.Role != domain.RoleAdmin && claims.Role != domain.RoleManager {
		return response.Forbidden(c, "")
	}

	u, err := h.userUC.GetByID(c.Request().Context(), id)
	if err != nil {
		return response.NotFound(c, "User not found")
	}

	return response.Success(c, u)
}

// Create godoc
// @Summary Create user (admin)
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body user.CreateUserInput true "User data"
// @Success 201 {object} response.Response{data=domain.User}
// @Router /users [post]
func (h *UserHandler) Create(c echo.Context) error {
	var input user.CreateUserInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	u, err := h.userUC.CreateUser(c.Request().Context(), input)
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			return response.ErrorWithCode(c, http.StatusConflict, "USER_EXISTS", "User with this email already exists")
		}
		return response.InternalError(c, "Failed to create user")
	}

	return response.Created(c, u)
}

// Update godoc
// @Summary Update user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body user.UpdateInput true "Update data"
// @Success 200 {object} response.Response{data=domain.User}
// @Router /users/{id} [put]
func (h *UserHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid user ID")
	}

	// Check authorization
	claims, _ := middleware.GetClaims(c)
	if claims.UserID != id && claims.Role != domain.RoleAdmin {
		return response.Forbidden(c, "")
	}

	var input user.UpdateInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	u, err := h.userUC.Update(c.Request().Context(), id, input)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return response.NotFound(c, "User not found")
		}
		return response.InternalError(c, "Failed to update user")
	}

	return response.Success(c, u)
}

// Delete godoc
// @Summary Delete user
// @Tags Users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 204
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid user ID")
	}

	if err := h.userUC.Delete(c.Request().Context(), id); err != nil {
		return response.InternalError(c, "Failed to delete user")
	}

	return response.NoContent(c)
}

// UpdateStatus godoc
// @Summary Update user status
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body user.UpdateStatusInput true "Status"
// @Success 200 {object} response.Response
// @Router /users/{id}/status [patch]
func (h *UserHandler) UpdateStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid user ID")
	}

	var input user.UpdateStatusInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	if err := h.userUC.UpdateStatus(c.Request().Context(), id, input); err != nil {
		return response.InternalError(c, "Failed to update status")
	}

	return response.SuccessWithMessage(c, "Status updated successfully", nil)
}

// UpdateRole godoc
// @Summary Update user role
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body user.UpdateRoleInput true "Role"
// @Success 200 {object} response.Response
// @Router /users/{id}/role [patch]
func (h *UserHandler) UpdateRole(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid user ID")
	}

	var input user.UpdateRoleInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	if err := h.userUC.UpdateRole(c.Request().Context(), id, input); err != nil {
		return response.InternalError(c, "Failed to update role")
	}

	return response.SuccessWithMessage(c, "Role updated successfully", nil)
}

// ListTutors godoc
// @Summary List tutors
// @Tags Users
// @Produce json
// @Success 200 {object} response.Response
// @Router /users/tutors [get]
func (h *UserHandler) ListTutors(c echo.Context) error {
	tutorRole := domain.RoleTutor
	activeStatus := domain.StatusActive

	users, total, err := h.userUC.List(c.Request().Context(), user.ListInput{
		Role:   &tutorRole,
		Status: &activeStatus,
		Page:   1,
		Limit:  50,
	})
	if err != nil {
		return response.InternalError(c, "Failed to list tutors")
	}

	return response.Paginated(c, users, 1, 50, total)
}

// GetTutor godoc
// @Summary Get tutor profile
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response
// @Router /users/tutors/{id} [get]
func (h *UserHandler) GetTutor(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid user ID")
	}

	profile, err := h.userUC.GetTutorProfile(c.Request().Context(), id)
	if err != nil {
		return response.NotFound(c, "Tutor not found")
	}

	return response.Success(c, profile)
}

// UpdateTutorProfile godoc
// @Summary Update tutor profile
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body user.UpdateTutorProfileInput true "Profile data"
// @Success 200 {object} response.Response
// @Router /users/tutors/{id}/profile [put]
func (h *UserHandler) UpdateTutorProfile(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid user ID")
	}

	// Check authorization - only self or admin
	claims, _ := middleware.GetClaims(c)
	if claims.UserID != id && claims.Role != domain.RoleAdmin {
		return response.Forbidden(c, "")
	}

	var input user.UpdateTutorProfileInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	profile, err := h.userUC.UpdateTutorProfile(c.Request().Context(), id, input)
	if err != nil {
		return response.InternalError(c, "Failed to update profile")
	}

	return response.Success(c, profile)
}
