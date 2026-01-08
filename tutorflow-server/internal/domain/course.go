package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// CourseStatus enum
type CourseStatus string

const (
	CourseStatusDraft     CourseStatus = "draft"
	CourseStatusPublished CourseStatus = "published"
	CourseStatusArchived  CourseStatus = "archived"
)

// CourseLevel enum
type CourseLevel string

const (
	CourseLevelBeginner     CourseLevel = "beginner"
	CourseLevelIntermediate CourseLevel = "intermediate"
	CourseLevelAdvanced     CourseLevel = "advanced"
)

// Category represents a course category
type Category struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Slug        string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Description *string    `gorm:"type:text" json:"description,omitempty"`
	Icon        *string    `gorm:"type:varchar(50)" json:"icon,omitempty"`
	ParentID    *uuid.UUID `gorm:"type:uuid" json:"parent_id,omitempty"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Parent        *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Subcategories []Category `gorm:"foreignKey:ParentID" json:"subcategories,omitempty"`
}

// Course represents a course
type Course struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title            string         `gorm:"type:varchar(255);not null" json:"title"`
	Slug             string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Description      *string        `gorm:"type:text" json:"description,omitempty"`
	ShortDescription *string        `gorm:"type:varchar(500)" json:"short_description,omitempty"`
	ThumbnailURL     *string        `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`
	PreviewVideoURL  *string        `gorm:"type:varchar(500)" json:"preview_video_url,omitempty"`
	InstructorID     uuid.UUID      `gorm:"type:uuid;index;not null" json:"instructor_id"`
	Status           CourseStatus   `gorm:"type:course_status;not null;default:'draft'" json:"status"`
	Level            CourseLevel    `gorm:"type:course_level;not null;default:'beginner'" json:"level"`
	Price            float64        `gorm:"type:decimal(10,2);default:0" json:"price"`
	DiscountPrice    *float64       `gorm:"type:decimal(10,2)" json:"discount_price,omitempty"`
	DurationHours    *int           `json:"duration_hours,omitempty"`
	TotalLessons     int            `gorm:"default:0" json:"total_lessons"`
	TotalStudents    int            `gorm:"default:0" json:"total_students"`
	Rating           float64        `gorm:"type:decimal(3,2);default:0" json:"rating"`
	TotalReviews     int            `gorm:"default:0" json:"total_reviews"`
	IsFeatured       bool           `gorm:"default:false" json:"is_featured"`
	Requirements     pq.StringArray `gorm:"type:text[]" json:"requirements,omitempty" swaggertype:"array,string"`
	WhatYouLearn     pq.StringArray `gorm:"type:text[]" json:"what_you_learn,omitempty" swaggertype:"array,string"`
	Language         string         `gorm:"type:varchar(50);default:'English'" json:"language"`
	PublishedAt      *time.Time     `json:"published_at,omitempty"`
	CreatedAt        time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Instructor  *User          `gorm:"foreignKey:InstructorID" json:"instructor,omitempty"`
	Categories  []Category     `gorm:"many2many:course_categories" json:"categories,omitempty"`
	Modules     []Module       `gorm:"foreignKey:CourseID" json:"modules,omitempty"`
	Enrollments []Enrollment   `gorm:"foreignKey:CourseID" json:"enrollments,omitempty"`
	Reviews     []CourseReview `gorm:"foreignKey:CourseID" json:"reviews,omitempty"`
}

func (c *Course) IsPublished() bool {
	return c.Status == CourseStatusPublished
}

func (c *Course) IsFree() bool {
	return c.Price == 0
}

func (c *Course) GetEffectivePrice() float64 {
	if c.DiscountPrice != nil && *c.DiscountPrice < c.Price {
		return *c.DiscountPrice
	}
	return c.Price
}

// CourseCategory join table
type CourseCategory struct {
	CourseID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"course_id"`
	CategoryID uuid.UUID `gorm:"type:uuid;primaryKey" json:"category_id"`
}

func (CourseCategory) TableName() string {
	return "course_categories"
}

// Module represents a course module/section
type Module struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CourseID    uuid.UUID `gorm:"type:uuid;index;not null" json:"course_id"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	SortOrder   int       `gorm:"not null;default:0" json:"sort_order"`
	IsPublished bool      `gorm:"default:false" json:"is_published"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Course  *Course  `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Lessons []Lesson `gorm:"foreignKey:ModuleID" json:"lessons,omitempty"`
}

// CourseRepository interface
type CourseRepository interface {
	Create(course *Course) error
	GetByID(id uuid.UUID) (*Course, error)
	GetBySlug(slug string) (*Course, error)
	GetAll(page, limit int) ([]Course, int64, error)
	GetByInstructorID(instructorID uuid.UUID, page, limit int) ([]Course, int64, error)
	Update(course *Course) error
	Delete(id uuid.UUID) error
	Search(query string, page, limit int) ([]Course, int64, error)
}
