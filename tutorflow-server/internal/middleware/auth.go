package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/pkg/jwt"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
)

// Context keys
const (
	UserIDKey = "user_id"
	UserKey   = "user"
	ClaimsKey = "claims"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(jwtManager *jwt.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return response.Unauthorized(c, "Missing authorization header")
			}

			// Extract token from "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return response.Unauthorized(c, "Invalid authorization header format")
			}

			tokenString := parts[1]

			// Validate token
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				if err == jwt.ErrExpiredToken {
					return response.Unauthorized(c, "Token has expired")
				}
				return response.Unauthorized(c, "Invalid token")
			}

			// Store claims in context
			c.Set(UserIDKey, claims.UserID)
			c.Set(ClaimsKey, claims)

			return next(c)
		}
	}
}

// OptionalAuthMiddleware extracts user info if token present, but doesn't require it
func OptionalAuthMiddleware(jwtManager *jwt.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return next(c)
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return next(c)
			}

			claims, err := jwtManager.ValidateToken(parts[1])
			if err == nil {
				c.Set(UserIDKey, claims.UserID)
				c.Set(ClaimsKey, claims)
			}

			return next(c)
		}
	}
}

// GetUserID extracts user ID from context
func GetUserID(c echo.Context) (string, bool) {
	id := c.Get(UserIDKey)
	if id == nil {
		return "", false
	}
	return id.(string), true
}

// GetClaims extracts claims from context
func GetClaims(c echo.Context) (*jwt.Claims, bool) {
	claims := c.Get(ClaimsKey)
	if claims == nil {
		return nil, false
	}
	return claims.(*jwt.Claims), true
}

// RequireRole creates middleware that requires specific roles
func RequireRole(roles ...domain.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := GetClaims(c)
			if !ok {
				return response.Unauthorized(c, "")
			}

			for _, role := range roles {
				if claims.Role == role {
					return next(c)
				}
			}

			return response.Forbidden(c, "Insufficient permissions")
		}
	}
}

// RequireAdmin requires admin role
func RequireAdmin() echo.MiddlewareFunc {
	return RequireRole(domain.RoleAdmin)
}

// RequireAdminOrManager requires admin or manager role
func RequireAdminOrManager() echo.MiddlewareFunc {
	return RequireRole(domain.RoleAdmin, domain.RoleManager)
}

// RequireTutor requires tutor role (or higher)
func RequireTutor() echo.MiddlewareFunc {
	return RequireRole(domain.RoleAdmin, domain.RoleManager, domain.RoleTutor)
}
