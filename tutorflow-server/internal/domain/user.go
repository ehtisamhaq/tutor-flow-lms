package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole enum
type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
	RoleTutor   UserRole = "tutor"
	RoleStudent UserRole = "student"
)

// UserStatus enum
type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
	StatusPending   UserStatus = "pending"
)

// User represents a user in the system
type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email           string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash    string         `gorm:"type:varchar(255);not null" json:"-"`
	FirstName       string         `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName        string         `gorm:"type:varchar(100);not null" json:"last_name"`
	Role            UserRole       `gorm:"type:user_role;not null;default:'student'" json:"role"`
	Status          UserStatus     `gorm:"type:user_status;not null;default:'pending'" json:"status"`
	AvatarURL       *string        `gorm:"type:varchar(500)" json:"avatar_url,omitempty"`
	Phone           *string        `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Bio             *string        `gorm:"type:text" json:"bio,omitempty"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt       time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	TutorProfile  *TutorProfile  `gorm:"foreignKey:UserID" json:"tutor_profile,omitempty"`
	Courses       []Course       `gorm:"foreignKey:InstructorID" json:"courses,omitempty"`
	Enrollments   []Enrollment   `gorm:"foreignKey:UserID" json:"enrollments,omitempty"`
	RefreshTokens []RefreshToken `gorm:"foreignKey:UserID" json:"-"`
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsManager() bool {
	return u.Role == RoleManager
}

func (u *User) IsTutor() bool {
	return u.Role == RoleTutor
}

func (u *User) IsStudent() bool {
	return u.Role == RoleStudent
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// TutorProfile contains tutor-specific data
type TutorProfile struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	Qualifications    []string  `gorm:"type:text[]" json:"qualifications,omitempty"`
	Specializations   []string  `gorm:"type:text[]" json:"specializations,omitempty"`
	YearsOfExperience *int      `json:"years_of_experience,omitempty"`
	HourlyRate        *float64  `gorm:"type:decimal(10,2)" json:"hourly_rate,omitempty"`
	Rating            float64   `gorm:"type:decimal(3,2);default:0" json:"rating"`
	TotalReviews      int       `gorm:"default:0" json:"total_reviews"`
	TotalStudents     int       `gorm:"default:0" json:"total_students"`
	TotalCourses      int       `gorm:"default:0" json:"total_courses"`
	TotalEarnings     float64   `gorm:"type:decimal(12,2);default:0" json:"total_earnings"`
	PayoutEmail       *string   `gorm:"type:varchar(255)" json:"payout_email,omitempty"`
	CreatedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// RefreshToken for JWT authentication
type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	TokenHash string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

func (r *RefreshToken) IsRevoked() bool {
	return r.RevokedAt != nil
}

func (r *RefreshToken) IsValid() bool {
	return !r.IsExpired() && !r.IsRevoked()
}

// UserDevice for DRM device tracking
type UserDevice struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	DeviceFingerprint string     `gorm:"type:varchar(255);not null" json:"device_fingerprint"`
	DeviceName        *string    `gorm:"type:varchar(100)" json:"device_name,omitempty"`
	DeviceType        *string    `gorm:"type:varchar(50)" json:"device_type,omitempty"`
	Browser           *string    `gorm:"type:varchar(100)" json:"browser,omitempty"`
	OS                *string    `gorm:"type:varchar(100)" json:"os,omitempty"`
	LastActiveAt      *time.Time `json:"last_active_at,omitempty"`
	IsTrusted         bool       `gorm:"default:false" json:"is_trusted"`
	CreatedAt         time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (UserDevice) TableName() string {
	return "user_devices"
}

func init() {
	// Composite unique constraint
}

// UserRepository interface
type UserRepository interface {
	Create(user *User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll(page, limit int) ([]User, int64, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
	UpdatePassword(id uuid.UUID, passwordHash string) error
	VerifyEmail(id uuid.UUID) error
}
