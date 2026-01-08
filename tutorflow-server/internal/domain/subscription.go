package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// SubscriptionInterval defines billing intervals
type SubscriptionInterval string

const (
	SubscriptionIntervalMonthly SubscriptionInterval = "monthly"
	SubscriptionIntervalYearly  SubscriptionInterval = "yearly"
)

// SubscriptionStatus defines subscription states
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
)

// SubscriptionPlan represents a subscription tier
type SubscriptionPlan struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name                 string         `gorm:"size:100;not null" json:"name"`
	Slug                 string         `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Description          string         `gorm:"type:text" json:"description"`
	PriceMonthly         float64        `gorm:"type:decimal(10,2);not null" json:"price_monthly"`
	PriceYearly          float64        `gorm:"type:decimal(10,2);not null" json:"price_yearly"`
	Features             pq.StringArray `gorm:"type:text[]" json:"features"`
	MaxCourses           *int           `gorm:"" json:"max_courses,omitempty"`   // nil = unlimited
	MaxDownloads         *int           `gorm:"" json:"max_downloads,omitempty"` // nil = unlimited
	OfflineAccess        bool           `gorm:"default:false" json:"offline_access"`
	CertificateAccess    bool           `gorm:"default:true" json:"certificate_access"`
	Priority             int            `gorm:"default:0" json:"priority"`
	IsActive             bool           `gorm:"default:true" json:"is_active"`
	StripeProductID      string         `gorm:"size:100" json:"-"`
	StripePriceMonthlyID string         `gorm:"size:100" json:"-"`
	StripePriceYearlyID  string         `gorm:"size:100" json:"-"`
	CreatedAt            time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// Subscription represents a user's subscription
type Subscription struct {
	ID                   uuid.UUID            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID               uuid.UUID            `gorm:"type:uuid;index;not null" json:"user_id"`
	PlanID               uuid.UUID            `gorm:"type:uuid;not null" json:"plan_id"`
	Status               SubscriptionStatus   `gorm:"size:20;not null;default:'active'" json:"status"`
	Interval             SubscriptionInterval `gorm:"size:20;not null" json:"interval"`
	CurrentPeriodStart   time.Time            `gorm:"not null" json:"current_period_start"`
	CurrentPeriodEnd     time.Time            `gorm:"not null" json:"current_period_end"`
	CancelAtPeriodEnd    bool                 `gorm:"default:false" json:"cancel_at_period_end"`
	CanceledAt           *time.Time           `gorm:"" json:"canceled_at,omitempty"`
	TrialStart           *time.Time           `gorm:"" json:"trial_start,omitempty"`
	TrialEnd             *time.Time           `gorm:"" json:"trial_end,omitempty"`
	StripeSubscriptionID string               `gorm:"size:100" json:"-"`
	StripeCustomerID     string               `gorm:"size:100" json:"-"`
	CreatedAt            time.Time            `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time            `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	User *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Plan *SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

// IsActive checks if subscription is currently active
func (s *Subscription) IsActive() bool {
	now := time.Now()
	return s.Status == SubscriptionStatusActive &&
		s.CurrentPeriodEnd.After(now)
}

// IsTrialing checks if subscription is in trial period
func (s *Subscription) IsTrialing() bool {
	if s.TrialEnd == nil {
		return false
	}
	return s.Status == SubscriptionStatusTrialing &&
		s.TrialEnd.After(time.Now())
}

// DaysRemaining returns days until subscription ends
func (s *Subscription) DaysRemaining() int {
	if s.CurrentPeriodEnd.Before(time.Now()) {
		return 0
	}
	return int(time.Until(s.CurrentPeriodEnd).Hours() / 24)
}

// SubscriptionRepository interface
type SubscriptionRepository interface {
	// Plans
	CreatePlan(plan *SubscriptionPlan) error
	GetPlanByID(id uuid.UUID) (*SubscriptionPlan, error)
	GetPlanBySlug(slug string) (*SubscriptionPlan, error)
	GetActivePlans() ([]SubscriptionPlan, error)
	UpdatePlan(plan *SubscriptionPlan) error

	// Subscriptions
	Create(subscription *Subscription) error
	GetByID(id uuid.UUID) (*Subscription, error)
	GetByUserID(userID uuid.UUID) (*Subscription, error)
	GetActiveByUserID(userID uuid.UUID) (*Subscription, error)
	Update(subscription *Subscription) error
	Cancel(id uuid.UUID) error
	GetExpiringSubscriptions(days int) ([]Subscription, error)
}

// SubscriptionUseCase interface
type SubscriptionUseCase interface {
	// Plans
	CreatePlan(plan *SubscriptionPlan) error
	GetPlans() ([]SubscriptionPlan, error)
	GetPlanBySlug(slug string) (*SubscriptionPlan, error)
	UpdatePlan(plan *SubscriptionPlan) error

	// Subscriptions
	Subscribe(userID uuid.UUID, planSlug string, interval SubscriptionInterval) (*Subscription, error)
	GetUserSubscription(userID uuid.UUID) (*Subscription, error)
	CancelSubscription(userID uuid.UUID) error
	ResumeSubscription(userID uuid.UUID) error
	ChangeSubscription(userID uuid.UUID, newPlanSlug string) (*Subscription, error)
	HandleWebhook(event string, data map[string]interface{}) error
}
