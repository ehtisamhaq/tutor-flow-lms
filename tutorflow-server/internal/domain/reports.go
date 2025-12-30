package domain

import (
	"time"

	"github.com/google/uuid"
)

// RecentlyViewed tracks user's recently viewed courses
type RecentlyViewed struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID   uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	CourseID uuid.UUID `gorm:"type:uuid;index;not null" json:"course_id"`
	ViewedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"viewed_at"`

	User   *User   `gorm:"foreignKey:UserID" json:"-"`
	Course *Course `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}

// ScheduledReport represents a scheduled report configuration
type ScheduledReport struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Name           string     `gorm:"type:varchar(255);not null" json:"name"`
	ReportType     string     `gorm:"type:varchar(50);not null" json:"report_type"`          // revenue, enrollments, users, courses
	Format         string     `gorm:"type:varchar(10);not null;default:'pdf'" json:"format"` // pdf, excel, csv
	Schedule       string     `gorm:"type:varchar(50);not null" json:"schedule"`             // daily, weekly, monthly
	Filters        string     `gorm:"type:text" json:"filters,omitempty"`                    // JSON filters
	RecipientEmail *string    `gorm:"type:varchar(255)" json:"recipient_email,omitempty"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time `json:"next_run_at,omitempty"`
	CreatedAt      time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Report types
const (
	ReportTypeRevenue     = "revenue"
	ReportTypeEnrollments = "enrollments"
	ReportTypeUsers       = "users"
	ReportTypeCourses     = "courses"
	ReportTypeQuizzes     = "quizzes"
	ReportTypeInstructors = "instructors"
)

// Schedule types
const (
	ScheduleDaily   = "daily"
	ScheduleWeekly  = "weekly"
	ScheduleMonthly = "monthly"
)

// Export formats
const (
	FormatPDF   = "pdf"
	FormatExcel = "excel"
	FormatCSV   = "csv"
)

// ExportRequest for on-demand exports
type ExportRequest struct {
	ReportType string                 `json:"report_type"`
	Format     string                 `json:"format"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
	DateFrom   *time.Time             `json:"date_from,omitempty"`
	DateTo     *time.Time             `json:"date_to,omitempty"`
}
