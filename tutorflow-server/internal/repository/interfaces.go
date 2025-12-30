package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/tutorflow/tutorflow-server/internal/domain"
)

// UserRepository interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters UserFilters) ([]domain.User, int64, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	VerifyEmail(ctx context.Context, id uuid.UUID) error
}

type UserFilters struct {
	Role   *domain.UserRole
	Status *domain.UserStatus
	Search string
	Page   int
	Limit  int
}

// TutorProfileRepository interface
type TutorProfileRepository interface {
	Create(ctx context.Context, profile *domain.TutorProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.TutorProfile, error)
	Update(ctx context.Context, profile *domain.TutorProfile) error
	List(ctx context.Context, page, limit int) ([]domain.TutorProfile, int64, error)
}

// RefreshTokenRepository interface
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// CourseRepository interface
type CourseRepository interface {
	Create(ctx context.Context, course *domain.Course) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Course, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Course, error)
	Update(ctx context.Context, course *domain.Course) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters CourseFilters) ([]domain.Course, int64, error)
	GetByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]domain.Course, int64, error)
	UpdateStats(ctx context.Context, id uuid.UUID) error
	IncrementStudentCount(ctx context.Context, id uuid.UUID) error
}

type CourseFilters struct {
	Status       *domain.CourseStatus
	Level        *domain.CourseLevel
	InstructorID *uuid.UUID
	CategoryID   *uuid.UUID
	IsFeatured   *bool
	Search       string
	MinPrice     *float64
	MaxPrice     *float64
	MinRating    *float64
	SortBy       string // "created_at", "price", "rating", "students"
	SortOrder    string // "asc", "desc"
	Page         int
	Limit        int
}

// CategoryRepository interface
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]domain.Category, error)
	GetWithSubcategories(ctx context.Context) ([]domain.Category, error)
}

// ModuleRepository interface
type ModuleRepository interface {
	Create(ctx context.Context, module *domain.Module) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Module, error)
	Update(ctx context.Context, module *domain.Module) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByCourse(ctx context.Context, courseID uuid.UUID) ([]domain.Module, error)
	Reorder(ctx context.Context, courseID uuid.UUID, moduleIDs []uuid.UUID) error
}

// LessonRepository interface
type LessonRepository interface {
	Create(ctx context.Context, lesson *domain.Lesson) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Lesson, error)
	Update(ctx context.Context, lesson *domain.Lesson) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByModule(ctx context.Context, moduleID uuid.UUID) ([]domain.Lesson, error)
	Reorder(ctx context.Context, moduleID uuid.UUID, lessonIDs []uuid.UUID) error
}

// EnrollmentRepository interface
type EnrollmentRepository interface {
	Create(ctx context.Context, enrollment *domain.Enrollment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Enrollment, error)
	GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (*domain.Enrollment, error)
	Update(ctx context.Context, enrollment *domain.Enrollment) error
	List(ctx context.Context, filters EnrollmentFilters) ([]domain.Enrollment, int64, error)
	GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Enrollment, int64, error)
	GetByCourse(ctx context.Context, courseID uuid.UUID, page, limit int) ([]domain.Enrollment, int64, error)
	UpdateProgress(ctx context.Context, id uuid.UUID, progress float64) error
}

type EnrollmentFilters struct {
	UserID   *uuid.UUID
	CourseID *uuid.UUID
	Status   *domain.EnrollmentStatus
	Page     int
	Limit    int
}

// LessonProgressRepository interface
type LessonProgressRepository interface {
	Upsert(ctx context.Context, progress *domain.LessonProgress) error
	GetByEnrollmentAndLesson(ctx context.Context, enrollmentID, lessonID uuid.UUID) (*domain.LessonProgress, error)
	GetByEnrollment(ctx context.Context, enrollmentID uuid.UUID) ([]domain.LessonProgress, error)
	MarkComplete(ctx context.Context, enrollmentID, lessonID uuid.UUID) error
	UpdateVideoPosition(ctx context.Context, enrollmentID, lessonID uuid.UUID, position int) error
}

// OrderRepository interface
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Order, int64, error)
	GetByPaymentIntent(ctx context.Context, paymentIntentID string) (*domain.Order, error)
}

// CartRepository interface
type CartRepository interface {
	GetOrCreate(ctx context.Context, userID *uuid.UUID, sessionID *string) (*domain.Cart, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Cart, error)
	AddItem(ctx context.Context, cartID, courseID uuid.UUID) error
	RemoveItem(ctx context.Context, cartID, courseID uuid.UUID) error
	Clear(ctx context.Context, cartID uuid.UUID) error
	MergeGuestCart(ctx context.Context, sessionID string, userID uuid.UUID) error
}

// WishlistRepository interface
type WishlistRepository interface {
	Add(ctx context.Context, userID, courseID uuid.UUID) error
	Remove(ctx context.Context, userID, courseID uuid.UUID) error
	GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Wishlist, int64, error)
	Exists(ctx context.Context, userID, courseID uuid.UUID) (bool, error)
}

// CouponRepository interface
type CouponRepository interface {
	Create(ctx context.Context, coupon *domain.Coupon) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Coupon, error)
	GetByCode(ctx context.Context, code string) (*domain.Coupon, error)
	Update(ctx context.Context, coupon *domain.Coupon) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementUsage(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, limit int) ([]domain.Coupon, int64, error)
}

// ReviewRepository interface
type ReviewRepository interface {
	Create(ctx context.Context, review *domain.CourseReview) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CourseReview, error)
	Update(ctx context.Context, review *domain.CourseReview) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByCourse(ctx context.Context, courseID uuid.UUID, page, limit int) ([]domain.CourseReview, int64, error)
	GetByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (*domain.CourseReview, error)
	Vote(ctx context.Context, reviewID, userID uuid.UUID, isHelpful bool) error
}

// NotificationRepository interface
type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Notification, int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
}

// QuizRepository interface
type QuizRepository interface {
	Create(ctx context.Context, quiz *domain.Quiz) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error)
	GetByLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Quiz, error)
	Update(ctx context.Context, quiz *domain.Quiz) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddQuestion(ctx context.Context, question *domain.QuizQuestion) error
	UpdateQuestion(ctx context.Context, question *domain.QuizQuestion) error
	DeleteQuestion(ctx context.Context, id uuid.UUID) error
	GetQuestion(ctx context.Context, id uuid.UUID) (*domain.QuizQuestion, error)
	AddOption(ctx context.Context, option *domain.QuizOption) error
	UpdateOption(ctx context.Context, option *domain.QuizOption) error
	DeleteOption(ctx context.Context, id uuid.UUID) error
}

// QuizAttemptRepository interface
type QuizAttemptRepository interface {
	Create(ctx context.Context, attempt *domain.QuizAttempt) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.QuizAttempt, error)
	Update(ctx context.Context, attempt *domain.QuizAttempt) error
	GetByUserAndQuiz(ctx context.Context, userID, quizID uuid.UUID) ([]domain.QuizAttempt, error)
	CountByUserAndQuiz(ctx context.Context, userID, quizID uuid.UUID) (int, error)
	GetLatestByUserAndQuiz(ctx context.Context, userID, quizID uuid.UUID) (*domain.QuizAttempt, error)
}

// AssignmentRepository interface
type AssignmentRepository interface {
	Create(ctx context.Context, assignment *domain.Assignment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Assignment, error)
	GetByLesson(ctx context.Context, lessonID uuid.UUID) (*domain.Assignment, error)
	Update(ctx context.Context, assignment *domain.Assignment) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SubmissionRepository interface
type SubmissionRepository interface {
	Create(ctx context.Context, submission *domain.Submission) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error)
	Update(ctx context.Context, submission *domain.Submission) error
	GetByUserAndAssignment(ctx context.Context, userID, assignmentID uuid.UUID) (*domain.Submission, error)
	GetByAssignment(ctx context.Context, assignmentID uuid.UUID, page, limit int) ([]domain.Submission, int64, error)
	GetPendingByAssignment(ctx context.Context, assignmentID uuid.UUID) ([]domain.Submission, error)
}

// DiscussionRepository interface
type DiscussionRepository interface {
	Create(ctx context.Context, discussion *domain.Discussion) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Discussion, error)
	Update(ctx context.Context, discussion *domain.Discussion) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByCourse(ctx context.Context, courseID uuid.UUID, page, limit int) ([]domain.Discussion, int64, error)
	GetByLesson(ctx context.Context, lessonID uuid.UUID, page, limit int) ([]domain.Discussion, int64, error)
	GetReplies(ctx context.Context, parentID uuid.UUID, page, limit int) ([]domain.Discussion, int64, error)
	Upvote(ctx context.Context, id uuid.UUID) error
	RemoveUpvote(ctx context.Context, id uuid.UUID) error
	MarkResolved(ctx context.Context, id uuid.UUID, resolved bool) error
	Pin(ctx context.Context, id uuid.UUID, pinned bool) error
	CountByCourse(ctx context.Context, courseID uuid.UUID) (int64, error)
	CountByLesson(ctx context.Context, lessonID uuid.UUID) (int64, error)
}

// CertificateRepository interface
type CertificateRepository interface {
	Create(ctx context.Context, cert *domain.Certificate) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Certificate, error)
	GetByNumber(ctx context.Context, number string) (*domain.Certificate, error)
	GetByEnrollment(ctx context.Context, enrollmentID uuid.UUID) (*domain.Certificate, error)
	Update(ctx context.Context, cert *domain.Certificate) error
	GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Certificate, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// SearchFilters for course search
type SearchFilters struct {
	Query      string
	CategoryID *uuid.UUID
	Level      *string
	MinPrice   *float64
	MaxPrice   *float64
	MinRating  *float64
	IsFree     *bool
	ExcludeID  *uuid.UUID
	SortBy     string
	SortOrder  string
	Page       int
	Limit      int
}

// FacetItem for search facets
type FacetItem struct {
	Label string
	Value string
	Count int64
}

// SearchFacets for filter UI
type SearchFacets struct {
	Categories  []FacetItem
	Levels      []FacetItem
	PriceRanges []FacetItem
}

// SearchRepository interface
type SearchRepository interface {
	SearchCourses(ctx context.Context, filters SearchFilters) ([]domain.Course, int64, error)
	GetFacets(ctx context.Context, query string) (*SearchFacets, error)
	GetSuggestions(ctx context.Context, query string, limit int) ([]string, error)
	GetTrendingSearches(ctx context.Context, limit int) ([]string, error)
	RecordSearch(ctx context.Context, query string, userID *uuid.UUID, resultCount int64) error
}

// AnnouncementRepository interface
type AnnouncementRepository interface {
	Create(ctx context.Context, announcement *domain.Announcement) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error)
	Update(ctx context.Context, announcement *domain.Announcement) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByCourse(ctx context.Context, courseID uuid.UUID, page, limit int) ([]domain.Announcement, int64, error)
	GetGlobal(ctx context.Context, page, limit int) ([]domain.Announcement, int64, error)
	GetForUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Announcement, int64, error)
	GetByAuthor(ctx context.Context, authorID uuid.UUID, page, limit int) ([]domain.Announcement, int64, error)
	Pin(ctx context.Context, id uuid.UUID, pinned bool) error
}

// MessageRepository interface
type MessageRepository interface {
	// Conversations
	CreateConversation(ctx context.Context, conv *domain.Conversation) error
	GetConversationByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error)
	GetConversationBetween(ctx context.Context, user1, user2 uuid.UUID) (*domain.Conversation, error)
	GetUserConversations(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.ConversationWithUnread, int64, error)
	UpdateConversation(ctx context.Context, conv *domain.Conversation) error

	// Messages
	CreateMessage(ctx context.Context, msg *domain.Message) error
	GetMessageByID(ctx context.Context, id uuid.UUID) (*domain.Message, error)
	GetConversationMessages(ctx context.Context, convID uuid.UUID, page, limit int) ([]domain.Message, int64, error)
	MarkAsRead(ctx context.Context, msgID uuid.UUID) error
	MarkConversationAsRead(ctx context.Context, convID, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	DeleteMessage(ctx context.Context, id uuid.UUID) error
}
