package domain

import (
	"time"

	"github.com/google/uuid"
)

// BundleType defines bundle types
type BundleType string

const (
	BundleTypeFixed    BundleType = "fixed"    // Fixed set of courses
	BundleTypeCategory BundleType = "category" // All courses in a category
	BundleTypePath     BundleType = "path"     // Learning path bundle
)

// Bundle represents a course bundle with discounted pricing
type Bundle struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title           string     `gorm:"size:200;not null" json:"title"`
	Slug            string     `gorm:"size:200;uniqueIndex;not null" json:"slug"`
	Description     string     `gorm:"type:text" json:"description"`
	ThumbnailURL    string     `gorm:"size:500" json:"thumbnail_url,omitempty"`
	Type            BundleType `gorm:"size:20;not null;default:'fixed'" json:"type"`
	OriginalPrice   float64    `gorm:"type:decimal(10,2);not null" json:"original_price"`
	BundlePrice     float64    `gorm:"type:decimal(10,2);not null" json:"bundle_price"`
	DiscountPercent float64    `gorm:"type:decimal(5,2)" json:"discount_percent"`
	CategoryID      *uuid.UUID `gorm:"type:uuid" json:"category_id,omitempty"`
	LearningPathID  *uuid.UUID `gorm:"type:uuid" json:"learning_path_id,omitempty"`
	IsActive        bool       `gorm:"default:true" json:"is_active"`
	StartDate       *time.Time `gorm:"" json:"start_date,omitempty"`
	EndDate         *time.Time `gorm:"" json:"end_date,omitempty"`
	MaxPurchases    *int       `gorm:"" json:"max_purchases,omitempty"`
	PurchaseCount   int        `gorm:"default:0" json:"purchase_count"`
	CreatedBy       uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt       time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Category     *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	LearningPath *LearningPath  `gorm:"foreignKey:LearningPathID" json:"learning_path,omitempty"`
	Creator      *User          `gorm:"foreignKey:CreatedBy" json:"-"`
	Courses      []BundleCourse `gorm:"foreignKey:BundleID" json:"courses,omitempty"`
}

// IsAvailable checks if bundle is available for purchase
func (b *Bundle) IsAvailable() bool {
	if !b.IsActive {
		return false
	}
	now := time.Now()
	if b.StartDate != nil && now.Before(*b.StartDate) {
		return false
	}
	if b.EndDate != nil && now.After(*b.EndDate) {
		return false
	}
	if b.MaxPurchases != nil && b.PurchaseCount >= *b.MaxPurchases {
		return false
	}
	return true
}

// Savings returns the amount saved
func (b *Bundle) Savings() float64 {
	return b.OriginalPrice - b.BundlePrice
}

// BundleCourse links courses to bundles
type BundleCourse struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BundleID uuid.UUID `gorm:"type:uuid;index;not null" json:"bundle_id"`
	CourseID uuid.UUID `gorm:"type:uuid;not null" json:"course_id"`
	Order    int       `gorm:"default:0" json:"order"`

	Bundle *Bundle `gorm:"foreignKey:BundleID" json:"-"`
	Course *Course `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}

// BundlePurchase tracks bundle purchases
type BundlePurchase struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BundleID  uuid.UUID `gorm:"type:uuid;index;not null" json:"bundle_id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null" json:"order_id"`
	Price     float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Bundle *Bundle `gorm:"foreignKey:BundleID" json:"bundle,omitempty"`
	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Order  *Order  `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// BundleRepository interface
type BundleRepository interface {
	Create(bundle *Bundle) error
	GetByID(id uuid.UUID) (*Bundle, error)
	GetBySlug(slug string) (*Bundle, error)
	GetActive(page, limit int) ([]Bundle, int64, error)
	GetByCategory(categoryID uuid.UUID) ([]Bundle, error)
	Update(bundle *Bundle) error
	Delete(id uuid.UUID) error
	AddCourse(bundleID, courseID uuid.UUID, order int) error
	RemoveCourse(bundleID, courseID uuid.UUID) error
	GetCourses(bundleID uuid.UUID) ([]Course, error)
	RecordPurchase(purchase *BundlePurchase) error
	GetUserPurchases(userID uuid.UUID) ([]BundlePurchase, error)
	IncrementPurchaseCount(bundleID uuid.UUID) error
}

// BundleUseCase interface
type BundleUseCase interface {
	CreateBundle(bundle *Bundle, courseIDs []uuid.UUID) (*Bundle, error)
	GetBundleBySlug(slug string) (*Bundle, error)
	GetActiveBundles(page, limit int) ([]Bundle, int64, error)
	UpdateBundle(bundle *Bundle) error
	DeleteBundle(id uuid.UUID) error
	AddCourseToBundle(bundleID, courseID uuid.UUID) error
	RemoveCourseFromBundle(bundleID, courseID uuid.UUID) error
	PurchaseBundle(userID, bundleID uuid.UUID) (*BundlePurchase, error)
	GetUserBundles(userID uuid.UUID) ([]BundlePurchase, error)
	CalculateBundlePrice(courseIDs []uuid.UUID, discountPercent float64) (original, discounted float64, err error)
}
