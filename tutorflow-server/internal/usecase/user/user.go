package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/pkg/hash"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines user management business logic
type UseCase struct {
	userRepo  repository.UserRepository
	tutorRepo repository.TutorProfileRepository
}

// NewUseCase creates a new user use case
func NewUseCase(
	userRepo repository.UserRepository,
	tutorRepo repository.TutorProfileRepository,
) *UseCase {
	return &UseCase{
		userRepo:  userRepo,
		tutorRepo: tutorRepo,
	}
}

// ListInput for listing users
type ListInput struct {
	Role   *domain.UserRole   `query:"role"`
	Status *domain.UserStatus `query:"status"`
	Search string             `query:"search"`
	Page   int                `query:"page"`
	Limit  int                `query:"limit"`
}

// List returns paginated users
func (uc *UseCase) List(ctx context.Context, input ListInput) ([]domain.User, int64, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 100 {
		input.Limit = 20
	}

	return uc.userRepo.List(ctx, repository.UserFilters{
		Role:   input.Role,
		Status: input.Status,
		Search: input.Search,
		Page:   input.Page,
		Limit:  input.Limit,
	})
}

// GetByID returns a user by ID
func (uc *UseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

// UpdateInput for updating a user
type UpdateInput struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=2,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,min=2,max=100"`
	Phone     *string `json:"phone" validate:"omitempty,max=20"`
	Bio       *string `json:"bio" validate:"omitempty,max=1000"`
	AvatarURL *string `json:"avatar_url" validate:"omitempty,url"`
}

// Update updates a user
func (uc *UseCase) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
	}
	if input.Phone != nil {
		user.Phone = input.Phone
	}
	if input.Bio != nil {
		user.Bio = input.Bio
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateStatusInput for updating user status
type UpdateStatusInput struct {
	Status domain.UserStatus `json:"status" validate:"required,oneof=active inactive suspended pending"`
}

// UpdateStatus updates a user's status (admin only)
func (uc *UseCase) UpdateStatus(ctx context.Context, id uuid.UUID, input UpdateStatusInput) error {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.Status = input.Status
	return uc.userRepo.Update(ctx, user)
}

// UpdateRoleInput for updating user role
type UpdateRoleInput struct {
	Role domain.UserRole `json:"role" validate:"required,oneof=admin manager tutor student"`
}

// UpdateRole updates a user's role (admin only)
func (uc *UseCase) UpdateRole(ctx context.Context, id uuid.UUID, input UpdateRoleInput) error {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	oldRole := user.Role
	user.Role = input.Role

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Create tutor profile if promoting to tutor
	if oldRole != domain.RoleTutor && input.Role == domain.RoleTutor {
		profile := &domain.TutorProfile{
			UserID: id,
		}
		_ = uc.tutorRepo.Create(ctx, profile)
	}

	return nil
}

// Delete soft-deletes a user
func (uc *UseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.userRepo.Delete(ctx, id)
}

// CreateUserInput for admin user creation
type CreateUserInput struct {
	Email     string          `json:"email" validate:"required,email"`
	Password  string          `json:"password" validate:"required,min=8"`
	FirstName string          `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string          `json:"last_name" validate:"required,min=2,max=100"`
	Role      domain.UserRole `json:"role" validate:"required,oneof=admin manager tutor student"`
}

// CreateUser creates a new user (admin only)
func (uc *UseCase) CreateUser(ctx context.Context, input CreateUserInput) (*domain.User, error) {
	// Check if user exists
	existing, _ := uc.userRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	passwordHash, err := hash.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        input.Email,
		PasswordHash: passwordHash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Role:         input.Role,
		Status:       domain.StatusActive,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create tutor profile if tutor
	if input.Role == domain.RoleTutor {
		profile := &domain.TutorProfile{
			UserID: user.ID,
		}
		_ = uc.tutorRepo.Create(ctx, profile)
	}

	return user, nil
}

// GetTutorProfile returns tutor profile
func (uc *UseCase) GetTutorProfile(ctx context.Context, userID uuid.UUID) (*domain.TutorProfile, error) {
	return uc.tutorRepo.GetByUserID(ctx, userID)
}

// UpdateTutorProfileInput for updating tutor profile
type UpdateTutorProfileInput struct {
	Qualifications    []string `json:"qualifications"`
	Specializations   []string `json:"specializations"`
	YearsOfExperience *int     `json:"years_of_experience"`
	HourlyRate        *float64 `json:"hourly_rate"`
	PayoutEmail       *string  `json:"payout_email" validate:"omitempty,email"`
}

// UpdateTutorProfile updates tutor profile
func (uc *UseCase) UpdateTutorProfile(ctx context.Context, userID uuid.UUID, input UpdateTutorProfileInput) (*domain.TutorProfile, error) {
	profile, err := uc.tutorRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.Qualifications != nil {
		profile.Qualifications = input.Qualifications
	}
	if input.Specializations != nil {
		profile.Specializations = input.Specializations
	}
	if input.YearsOfExperience != nil {
		profile.YearsOfExperience = input.YearsOfExperience
	}
	if input.HourlyRate != nil {
		profile.HourlyRate = input.HourlyRate
	}
	if input.PayoutEmail != nil {
		profile.PayoutEmail = input.PayoutEmail
	}

	if err := uc.tutorRepo.Update(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}
