package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/pkg/hash"
	"github.com/tutorflow/tutorflow-server/internal/pkg/jwt"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines auth business logic
type UseCase struct {
	userRepo   repository.UserRepository
	tokenRepo  repository.RefreshTokenRepository
	jwtManager *jwt.Manager
}

// NewUseCase creates a new auth use case
func NewUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	jwtManager *jwt.Manager,
) *UseCase {
	return &UseCase{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
	}
}

// RegisterInput for user registration
type RegisterInput struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,password"`
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
}

// RegisterOutput returned after registration
type RegisterOutput struct {
	User   *domain.User   `json:"user"`
	Tokens *jwt.TokenPair `json:"tokens"`
}

// Register creates a new user
func (uc *UseCase) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	// Check if user already exists
	existing, _ := uc.userRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	passwordHash, err := hash.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Email:        input.Email,
		PasswordHash: passwordHash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Role:         domain.RoleStudent,
		Status:       domain.StatusActive, // Auto-activate for now
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	tokens, err := uc.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	if err := uc.storeRefreshToken(ctx, user.ID, tokens.RefreshToken, tokens.ExpiresAt); err != nil {
		return nil, err
	}

	return &RegisterOutput{
		User:   user,
		Tokens: tokens,
	}, nil
}

// LoginInput for user login
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginOutput returned after login
type LoginOutput struct {
	User   *domain.User   `json:"user"`
	Tokens *jwt.TokenPair `json:"tokens"`
}

// Login authenticates a user
func (uc *UseCase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	// Find user
	user, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Check password
	if !hash.CheckPassword(input.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	// Check user status
	if user.Status == domain.StatusSuspended {
		return nil, domain.ErrUserSuspended
	}
	if user.Status == domain.StatusInactive {
		return nil, domain.ErrUserInactive
	}

	// Generate tokens
	tokens, err := uc.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	if err := uc.storeRefreshToken(ctx, user.ID, tokens.RefreshToken, tokens.ExpiresAt); err != nil {
		return nil, err
	}

	// Update last login
	_ = uc.userRepo.UpdateLastLogin(ctx, user.ID)

	return &LoginOutput{
		User:   user,
		Tokens: tokens,
	}, nil
}

// RefreshInput for token refresh
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshOutput returned after token refresh
type RefreshOutput struct {
	Tokens *jwt.TokenPair `json:"tokens"`
}

// Refresh generates new access token from refresh token
func (uc *UseCase) Refresh(ctx context.Context, input RefreshInput) (*RefreshOutput, error) {
	// Validate refresh token
	claims, err := uc.jwtManager.ValidateToken(input.RefreshToken)
	if err != nil {
		return nil, domain.ErrRefreshTokenInvalid
	}

	// Check if token is in database and not revoked
	tokenHash := hash.HashToken(input.RefreshToken)
	storedToken, err := uc.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, domain.ErrRefreshTokenInvalid
	}

	if !storedToken.IsValid() {
		return nil, domain.ErrRefreshTokenInvalid
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// Revoke old token
	_ = uc.tokenRepo.Revoke(ctx, storedToken.ID)

	// Generate new tokens
	tokens, err := uc.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Store new refresh token
	if err := uc.storeRefreshToken(ctx, user.ID, tokens.RefreshToken, tokens.ExpiresAt); err != nil {
		return nil, err
	}

	return &RefreshOutput{Tokens: tokens}, nil
}

// Logout revokes refresh token
func (uc *UseCase) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hash.HashToken(refreshToken)
	storedToken, err := uc.tokenRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil // Token doesn't exist, consider it logged out
	}

	return uc.tokenRepo.Revoke(ctx, storedToken.ID)
}

// LogoutAll revokes all user's refresh tokens
func (uc *UseCase) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return uc.tokenRepo.RevokeAllForUser(ctx, userID)
}

// GetCurrentUser gets user by ID
func (uc *UseCase) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

// ChangePasswordInput for password change
type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,password"`
}

// ChangePassword changes user password
func (uc *UseCase) ChangePassword(ctx context.Context, userID uuid.UUID, input ChangePasswordInput) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if !hash.CheckPassword(input.CurrentPassword, user.PasswordHash) {
		return domain.ErrInvalidCredentials
	}

	// Hash new password
	newHash, err := hash.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	if err := uc.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		return err
	}

	// Revoke all refresh tokens
	return uc.tokenRepo.RevokeAllForUser(ctx, userID)
}

func (uc *UseCase) storeRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	tokenHash := hash.HashToken(token)
	refreshToken := &domain.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	return uc.tokenRepo.Create(ctx, refreshToken)
}
