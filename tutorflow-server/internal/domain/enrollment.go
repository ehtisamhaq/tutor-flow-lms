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
	ID              uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID        `gorm:"type:uuid;index;not null" json:"user_id"`
	CourseID        uuid.UUID        `gorm:"type:uuid;index;not null" json:"course_id"`
	Status          EnrollmentStatus `gorm:"type:enrollment_status;not null;default:'pending'" json:"status"`
	EnrolledAt      time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"enrolled_at"`
	StartedAt       *time.Time       `json:"started_at,omitempty"`
	CompletedAt     *time.Time       `json:"completed_at,omitempty"`
	ExpiresAt       *time.Time       `json:"expires_at,omitempty"`
	ProgressPercent float64          `gorm:"type:decimal(5,2);default:0" json:"progress_percent"`
	LastAccessedAt  *time.Time       `json:"last_accessed_at,omitempty"`
	CertificateID   *uuid.UUID       `gorm:"type:uuid" json:"certificate_id,omitempty"`
	OrderID         *uuid.UUID       `gorm:"type:uuid" json:"order_id,omitempty"` // Link to purchase

	// Relationships
	User        *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Course      *Course          `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Certificate *Certificate     `gorm:"foreignKey:CertificateID" json:"certificate,omitempty"`
	Progress    []LessonProgress `gorm:"foreignKey:EnrollmentID" json:"progress,omitempty"`
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
	Lesson     *Lesson     `gorm:"foreignKey:LessonID" json:"lesson,omitempty"`
}

// Certificate represents a completion certificate
type Certificate struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EnrollmentID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"enrollment_id"`
	CertificateNumber string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"certificate_number"`
	IssuedAt          time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"issued_at"`
	PDFURL            *string   `gorm:"type:varchar(500)" json:"pdf_url,omitempty"`

	Enrollment *Enrollment `gorm:"foreignKey:EnrollmentID" json:"enrollment,omitempty"`
}

// GenerateCertificateNumber creates a unique certificate number
func GenerateCertificateNumber() string {
	return "TF-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:8]
}
