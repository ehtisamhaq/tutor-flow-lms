package domain

import (
	"time"

	"github.com/google/uuid"
)

// EnrollmentStatus enum
type EnrollmentStatus string

const (
	EnrollmentStatusPending   EnrollmentStatus = "pending"
	EnrollmentStatusActive    EnrollmentStatus = "active"
	EnrollmentStatusCompleted EnrollmentStatus = "completed"
	EnrollmentStatusCancelled EnrollmentStatus = "cancelled"
	EnrollmentStatusExpired   EnrollmentStatus = "expired"
)

// Enrollment represents a student's enrollment in a course
type Enrollment struct {
	ID             uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID        `gorm:"type:uuid;index;not null" json:"user_id"`
	CourseID       uuid.UUID        `gorm:"type:uuid;index;not null" json:"course_id"`
	Status         EnrollmentStatus `gorm:"type:enrollment_status;not null;default:'pending'" json:"status"`
	EnrolledAt     time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"enrolled_at"`
	StartedAt      *time.Time       `json:"started_at,omitempty"`
	CompletedAt    *time.Time       `json:"completed_at,omitempty"`
	ExpiresAt      *time.Time       `json:"expires_at,omitempty"`
	Progress       float64          `gorm:"column:progress_percent;type:decimal(5,2);default:0" json:"progress"`
	LastAccessedAt *time.Time       `json:"last_accessed_at,omitempty"`
	OrderID        *uuid.UUID       `gorm:"type:uuid" json:"order_id,omitempty"` // Link to purchase

	// Relationships
	User             *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Course           *Course          `gorm:"foreignKey:CourseID" json:"course"`
	Certificate      *Certificate     `gorm:"foreignKey:EnrollmentID" json:"certificate,omitempty"`
	LessonProgresses []LessonProgress `gorm:"foreignKey:EnrollmentID" json:"lesson_progresses"`
}

func (e *Enrollment) IsActive() bool {
	return e.Status == EnrollmentStatusActive
}

func (e *Enrollment) IsCompleted() bool {
	return e.Status == EnrollmentStatusCompleted
}

func (e *Enrollment) IsExpired() bool {
	if e.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*e.ExpiresAt)
}

func (e *Enrollment) CanAccess() bool {
	return e.IsActive() && !e.IsExpired()
}

// LessonProgress tracks progress for each lesson
type LessonProgress struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EnrollmentID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"enrollment_id"`
	LessonID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"lesson_id"`
	IsCompleted   bool       `gorm:"default:false" json:"is_completed"`
	TimeSpent     int        `gorm:"default:0" json:"time_spent"`     // seconds
	VideoPosition int        `gorm:"default:0" json:"video_position"` // seconds
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Enrollment *Enrollment `gorm:"foreignKey:EnrollmentID" json:"-"`
	Lesson     *Lesson     `gorm:"foreignKey:LessonID" json:"lesson"`
}

// EnrollmentRepository interface
type EnrollmentRepository interface {
	Create(enrollment *Enrollment) error
	GetByID(id uuid.UUID) (*Enrollment, error)
	GetByUserAndCourse(userID, courseID uuid.UUID) (*Enrollment, error)
	GetByUserID(userID uuid.UUID, page, limit int) ([]Enrollment, int64, error)
	GetByCourseID(courseID uuid.UUID, page, limit int) ([]Enrollment, int64, error)
	Update(enrollment *Enrollment) error
	Delete(id uuid.UUID) error
	IsEnrolled(userID, courseID uuid.UUID) (bool, error)
	UpdateProgress(enrollmentID uuid.UUID, percent float64) error
}

// StudentDashboardStats represents statistics for the student dashboard
type StudentDashboardStats struct {
	EnrolledCourses int     `json:"enrolled_courses"`
	HoursCompleted  int     `json:"hours_completed"`
	Certificates    int     `json:"certificates"`
	AverageProgress float64 `json:"average_progress"`
}
