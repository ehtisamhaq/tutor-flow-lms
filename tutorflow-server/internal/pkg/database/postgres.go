package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/pkg/config"
)

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	// Create custom types first
	if err := createEnumTypes(db); err != nil {
		return err
	}

	// Auto migrate all domain models
	return db.AutoMigrate(
		// Users
		&domain.User{},
		&domain.TutorProfile{},
		&domain.RefreshToken{},
		&domain.UserDevice{},

		// Courses
		&domain.Category{},
		&domain.Course{},
		&domain.CourseCategory{},
		&domain.Module{},
		&domain.Lesson{},
		&domain.VideoAsset{},

		// Enrollments
		&domain.Enrollment{},
		&domain.LessonProgress{},

		// Assessments
		&domain.Quiz{},
		&domain.QuizQuestion{},
		&domain.QuizOption{},
		&domain.QuizAttempt{},
		&domain.Assignment{},
		&domain.Submission{},

		// E-Commerce
		&domain.Cart{},
		&domain.CartItem{},
		&domain.Wishlist{},
		&domain.Coupon{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.InstructorEarning{},
		&domain.Payout{},

		// Reviews
		&domain.CourseReview{},
		&domain.ReviewVote{},

		// Learning Paths
		&domain.LearningPath{},
		&domain.LearningPathCourse{},
		&domain.LearningPathEnrollment{},

		// Notes & Bookmarks
		&domain.CourseNote{},
		&domain.VideoBookmark{},

		// Communication
		&domain.Announcement{},
		&domain.Discussion{},
		&domain.Notification{},

		// Certificates
		&domain.Certificate{},

		// Playback
		&domain.PlaybackSession{},
	)
}

func createEnumTypes(db *gorm.DB) error {
	enums := []struct {
		name   string
		values []string
	}{
		{"user_role", []string{"admin", "manager", "tutor", "student"}},
		{"user_status", []string{"active", "inactive", "suspended", "pending"}},
		{"course_status", []string{"draft", "published", "archived"}},
		{"course_level", []string{"beginner", "intermediate", "advanced"}},
		{"lesson_type", []string{"video", "text", "quiz", "assignment", "resource"}},
		{"content_access", []string{"free", "enrolled", "premium"}},
		{"enrollment_status", []string{"pending", "active", "completed", "cancelled", "expired"}},
		{"question_type", []string{"single_choice", "multiple_choice", "true_false", "short_answer", "essay"}},
		{"submission_status", []string{"pending", "submitted", "graded", "returned"}},
		{"order_status", []string{"pending", "completed", "refunded", "failed"}},
		{"payment_method", []string{"stripe", "paypal", "bank_transfer"}},
		{"coupon_type", []string{"percentage", "fixed", "free"}},
		{"notification_type", []string{"enrollment_approved", "new_lesson", "assignment_due", "grade_posted", "announcement", "message", "course_update", "payment_received", "review_received"}},
	}

	for _, e := range enums {
		// Check if type exists
		var exists bool
		db.Raw("SELECT EXISTS(SELECT 1 FROM pg_type WHERE typname = ?)", e.name).Scan(&exists)
		if !exists {
			values := ""
			for i, v := range e.values {
				if i > 0 {
					values += ", "
				}
				values += fmt.Sprintf("'%s'", v)
			}
			sql := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s)", e.name, values)
			if err := db.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create enum type %s: %w", e.name, err)
			}
		}
	}

	return nil
}
