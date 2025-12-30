package domain

import (
	"time"

	"github.com/google/uuid"
)

// LearningPath represents a curated sequence of courses
type LearningPath struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title            string     `gorm:"type:varchar(255);not null" json:"title"`
	Slug             string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Description      string     `gorm:"type:text" json:"description"`
	ShortDescription string     `gorm:"type:varchar(500)" json:"short_description"`
	ThumbnailURL     *string    `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`
	CategoryID       *uuid.UUID `gorm:"type:uuid;index" json:"category_id,omitempty"`
	Level            string     `gorm:"type:varchar(50);default:'beginner'" json:"level"`
	EstimatedHours   int        `gorm:"default:0" json:"estimated_hours"`
	IsPublished      bool       `gorm:"default:false" json:"is_published"`
	IsFeatured       bool       `gorm:"default:false" json:"is_featured"`
	CreatedBy        uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt        time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Stats
	TotalCourses  int `gorm:"default:0" json:"total_courses"`
	TotalStudents int `gorm:"default:0" json:"total_students"`

	// Relationships
	Category *Category            `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Creator  *User                `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Courses  []LearningPathCourse `gorm:"foreignKey:PathID" json:"courses,omitempty"`
}

// LearningPathCourse links a course to a learning path with ordering
type LearningPathCourse struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PathID      uuid.UUID `gorm:"type:uuid;index;not null" json:"path_id"`
	CourseID    uuid.UUID `gorm:"type:uuid;not null" json:"course_id"`
	Position    int       `gorm:"not null;default:0" json:"position"`
	IsRequired  bool      `gorm:"default:true" json:"is_required"`
	Description *string   `gorm:"type:text" json:"description,omitempty"` // Why this course is in the path
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// Relationships
	Path   *LearningPath `gorm:"foreignKey:PathID" json:"-"`
	Course *Course       `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}

// LearningPathEnrollment tracks user enrollment in a learning path
type LearningPathEnrollment struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PathID        uuid.UUID  `gorm:"type:uuid;index;not null" json:"path_id"`
	UserID        uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Status        string     `gorm:"type:varchar(50);default:'active'" json:"status"` // active, completed, paused
	Progress      float64    `gorm:"default:0" json:"progress"`                       // 0-100%
	EnrolledAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"enrolled_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CertificateID *uuid.UUID `gorm:"type:uuid" json:"certificate_id,omitempty"`

	// Relationships
	Path        *LearningPath `gorm:"foreignKey:PathID" json:"path,omitempty"`
	User        *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Certificate *Certificate  `gorm:"foreignKey:CertificateID" json:"certificate,omitempty"`
}

// LearningPathEnrollmentStatus constants
const (
	PathEnrollmentActive    = "active"
	PathEnrollmentCompleted = "completed"
	PathEnrollmentPaused    = "paused"
)

// LearningPathProgress for detailed progress tracking
type LearningPathProgress struct {
	PathID           uuid.UUID            `json:"path_id"`
	UserID           uuid.UUID            `json:"user_id"`
	TotalCourses     int                  `json:"total_courses"`
	CompletedCourses int                  `json:"completed_courses"`
	Progress         float64              `json:"progress"`
	CourseProgress   []CourseProgressItem `json:"course_progress"`
}

type CourseProgressItem struct {
	CourseID    uuid.UUID `json:"course_id"`
	CourseTitle string    `json:"course_title"`
	Position    int       `json:"position"`
	IsRequired  bool      `json:"is_required"`
	IsCompleted bool      `json:"is_completed"`
	Progress    float64   `json:"progress"`
}
