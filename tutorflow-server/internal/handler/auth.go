package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/pkg/validator"
	"github.com/tutorflow/tutorflow-server/internal/usecase/auth"
)

// AuthHandler handles auth-related HTTP requests
type AuthHandler struct {
	authUC *auth.UseCase
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authUC *auth.UseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

// RegisterRoutes registers auth routes
func (h *AuthHandler) RegisterRoutes(g *echo.Group, authMiddleware echo.MiddlewareFunc) {
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/refresh", h.Refresh)
	g.POST("/logout", h.Logout, authMiddleware)
	g.GET("/me", h.Me, authMiddleware)
	g.PUT("/password", h.ChangePassword, authMiddleware)
}

// Register godoc
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.RegisterInput true "Registration data"
// @Success 201 {object} response.Response{data=auth.RegisterOutput}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var input auth.RegisterInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	output, err := h.authUC.Register(c.Request().Context(), input)
	if err != nil {
		switch err {
		case domain.ErrUserAlreadyExists:
			return response.ErrorWithCode(c, http.StatusConflict, "USER_EXISTS", "User with this email already exists")
		default:
			return response.InternalError(c, "Failed to register user")
		}
	}

	return response.Created(c, output)
}

// Login godoc
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.LoginInput true "Login credentials"
// @Success 200 {object} response.Response{data=auth.LoginOutput}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var input auth.LoginInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	output, err := h.authUC.Login(c.Request().Context(), input)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			return response.Unauthorized(c, "Invalid email or password")
		case domain.ErrUserSuspended:
			return response.ErrorWithCode(c, http.StatusForbidden, "USER_SUSPENDED", "Your account has been suspended")
		case domain.ErrUserInactive:
			return response.ErrorWithCode(c, http.StatusForbidden, "USER_INACTIVE", "Your account is inactive")
		default:
			return response.InternalError(c, "Failed to login")
		}
	}

	return response.Success(c, output)
}

// Refresh godoc
// @Summary Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.RefreshInput true "Refresh token"
// @Success 200 {object} response.Response{data=auth.RefreshOutput}
// @Failure 401 {object} response.Response
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c echo.Context) error {
	var input auth.RefreshInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	tokens, err := h.authUC.Refresh(c.Request().Context(), input)
	if err != nil {
		return response.Unauthorized(c, "Invalid or expired refresh token")
	}

	return response.Success(c, tokens)
}

// Logout godoc
// @Summary Logout user
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body auth.RefreshInput true "Refresh token to revoke"
// @Success 204
// @Failure 401 {object} response.Response
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	var input auth.RefreshInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	_ = h.authUC.Logout(c.Request().Context(), input.RefreshToken)
	return response.NoContent(c)
}

// Me godoc
// @Summary Get current user
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 401 {object} response.Response
// @Router /auth/me [get]
func (h *AuthHandler) Me(c echo.Context) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "")
	}

	user, err := h.authUC.GetCurrentUser(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.NotFound(c, "User not found")
	}

	return response.Success(c, user)
}

// ChangePassword godoc
// @Summary Change password
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body auth.ChangePasswordInput true "Password change data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/password [put]
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.Unauthorized(c, "")
	}

	var input auth.ChangePasswordInput
	if err := c.Bind(&input); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := validator.Validate(input); err != nil {
		return response.ValidationErrors(c, validator.FormatValidationErrors(err))
	}

	err := h.authUC.ChangePassword(c.Request().Context(), claims.UserID, input)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			return response.BadRequest(c, "Current password is incorrect")
		}
		return response.InternalError(c, "Failed to change password")
	}

	return response.SuccessWithMessage(c, "Password changed successfully", nil)
}
