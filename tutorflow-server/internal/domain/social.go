package domain

import (
	"time"

	"github.com/google/uuid"
)

// CourseReview represents a course review
type CourseReview struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CourseID           uuid.UUID  `gorm:"type:uuid;index;not null" json:"course_id"`
	UserID             uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Rating             float64    `gorm:"type:decimal(2,1);not null" json:"rating"` // 1.0 - 5.0
	Title              *string    `gorm:"type:varchar(200)" json:"title,omitempty"`
	Content            *string    `gorm:"type:text" json:"content,omitempty"`
	HelpfulCount       int        `gorm:"default:0" json:"helpful_count"`
	UnhelpfulCount     int        `gorm:"default:0" json:"unhelpful_count"`
	InstructorReply    *string    `gorm:"type:text" json:"instructor_reply,omitempty"`
	InstructorReplyAt  *time.Time `json:"instructor_reply_at,omitempty"`
	IsVerifiedPurchase bool       `gorm:"default:false" json:"is_verified_purchase"`
	IsFeatured         bool       `gorm:"default:false" json:"is_featured"`
	Status             string     `gorm:"type:varchar(20);default:'published'" json:"status"`
	CreatedAt          time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Course *Course      `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	User   *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Votes  []ReviewVote `gorm:"foreignKey:ReviewID" json:"-"`
}

// ReviewVote represents a vote on a review
type ReviewVote struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ReviewID  uuid.UUID `gorm:"type:uuid;index;not null" json:"review_id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	IsHelpful bool      `gorm:"not null" json:"is_helpful"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Review *CourseReview `gorm:"foreignKey:ReviewID" json:"-"`
	User   *User         `gorm:"foreignKey:UserID" json:"-"`
}

// NotificationType enum
type NotificationType string

const (
	NotificationEnrollmentApproved NotificationType = "enrollment_approved"
	NotificationNewLesson          NotificationType = "new_lesson"
	NotificationAssignmentDue      NotificationType = "assignment_due"
	NotificationGradePosted        NotificationType = "grade_posted"
	NotificationAnnouncement       NotificationType = "announcement"
	NotificationMessage            NotificationType = "message"
	NotificationCourseUpdate       NotificationType = "course_update"
	NotificationPaymentReceived    NotificationType = "payment_received"
	NotificationReviewReceived     NotificationType = "review_received"
)

// Announcement represents a course or global announcement
type Announcement struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CourseID    *uuid.UUID `gorm:"type:uuid;index" json:"course_id,omitempty"`
	AuthorID    uuid.UUID  `gorm:"type:uuid;not null" json:"author_id"`
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	IsPinned    bool       `gorm:"default:false" json:"is_pinned"`
	PublishedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"published_at"`
	CreatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Course *Course `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Author *User   `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

// Discussion represents a discussion thread
type Discussion struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CourseID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"course_id"`
	LessonID   *uuid.UUID `gorm:"type:uuid;index" json:"lesson_id,omitempty"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	ParentID   *uuid.UUID `gorm:"type:uuid" json:"parent_id,omitempty"`
	Content    string     `gorm:"type:text;not null" json:"content"`
	IsPinned   bool       `gorm:"default:false" json:"is_pinned"`
	IsResolved bool       `gorm:"default:false" json:"is_resolved"`
	Upvotes    int        `gorm:"default:0" json:"upvotes"`
	CreatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Course  *Course      `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Lesson  *Lesson      `gorm:"foreignKey:LessonID" json:"lesson,omitempty"`
	User    *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Parent  *Discussion  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies []Discussion `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// Notification represents a user notification
type Notification struct {
	ID        uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID        `gorm:"type:uuid;index;not null" json:"user_id"`
	Type      NotificationType `gorm:"type:notification_type;not null" json:"type"`
	Title     string           `gorm:"type:varchar(255);not null" json:"title"`
	Message   *string          `gorm:"type:text" json:"message,omitempty"`
	Data      *string          `gorm:"type:jsonb;default:'{}'" json:"data,omitempty"`
	ReadAt    *time.Time       `json:"read_at,omitempty"`
	CreatedAt time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"-"`
}

func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}
