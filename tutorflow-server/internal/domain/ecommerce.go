package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// OrderStatus enum
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusRefunded  OrderStatus = "refunded"
	OrderStatusFailed    OrderStatus = "failed"
)

// PaymentMethod enum
type PaymentMethod string

const (
	PaymentMethodStripe       PaymentMethod = "stripe"
	PaymentMethodPaypal       PaymentMethod = "paypal"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

// CouponType enum
type CouponType string

const (
	CouponTypePercentage CouponType = "percentage"
	CouponTypeFixed      CouponType = "fixed"
	CouponTypeFree       CouponType = "free"
)

// Cart represents a shopping cart
type Cart struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	SessionID *string    `gorm:"type:varchar(100)" json:"session_id,omitempty"`
	CreatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	User  *User      `gorm:"foreignKey:UserID" json:"-"`
	Items []CartItem `gorm:"foreignKey:CartID" json:"items,omitempty"`
}

func (c *Cart) Total() float64 {
	var total float64
	for _, item := range c.Items {
		if item.Course != nil {
			total += item.Course.GetEffectivePrice()
		}
	}
	return total
}

// CartItem represents an item in the cart
type CartItem struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CartID   uuid.UUID `gorm:"type:uuid;index;not null" json:"cart_id"`
	CourseID uuid.UUID `gorm:"type:uuid;not null" json:"course_id"`
	AddedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"added_at"`

	Cart   *Cart   `gorm:"foreignKey:CartID" json:"-"`
	Course *Course `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}

// Wishlist represents a user's wishlist item
type Wishlist struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID   uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	CourseID uuid.UUID `gorm:"type:uuid;not null" json:"course_id"`
	AddedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"added_at"`

	User   *User   `gorm:"foreignKey:UserID" json:"-"`
	Course *Course `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}

// Coupon represents a discount coupon
type Coupon struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code              string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	CouponType        CouponType     `gorm:"type:coupon_type;not null" json:"coupon_type"`
	Value             float64        `gorm:"type:decimal(10,2);not null" json:"value"`
	MinPurchase       float64        `gorm:"type:decimal(10,2);default:0" json:"min_purchase"`
	MaxDiscount       *float64       `gorm:"type:decimal(10,2)" json:"max_discount,omitempty"`
	UsageLimit        *int           `json:"usage_limit,omitempty"`
	UsedCount         int            `gorm:"default:0" json:"used_count"`
	PerUserLimit      int            `gorm:"default:1" json:"per_user_limit"`
	ApplicableCourses pq.StringArray `gorm:"type:uuid[]" json:"applicable_courses,omitempty"`
	StartsAt          *time.Time     `json:"starts_at,omitempty"`
	ExpiresAt         *time.Time     `json:"expires_at,omitempty"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	CreatedBy         *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt         time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Creator *User `gorm:"foreignKey:CreatedBy" json:"-"`
}

func (c *Coupon) IsValid() bool {
	if !c.IsActive {
		return false
	}
	now := time.Now()
	if c.StartsAt != nil && now.Before(*c.StartsAt) {
		return false
	}
	if c.ExpiresAt != nil && now.After(*c.ExpiresAt) {
		return false
	}
	if c.UsageLimit != nil && c.UsedCount >= *c.UsageLimit {
		return false
	}
	return true
}

func (c *Coupon) CalculateDiscount(subtotal float64) float64 {
	if subtotal < c.MinPurchase {
		return 0
	}
	var discount float64
	switch c.CouponType {
	case CouponTypePercentage:
		discount = subtotal * (c.Value / 100)
	case CouponTypeFixed:
		discount = c.Value
	case CouponTypeFree:
		discount = subtotal
	}
	if c.MaxDiscount != nil && discount > *c.MaxDiscount {
		discount = *c.MaxDiscount
	}
	if discount > subtotal {
		discount = subtotal
	}
	return discount
}

// Order represents a purchase order
type Order struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderNumber     string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"order_number"`
	UserID          uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	Status          OrderStatus    `gorm:"type:order_status;not null;default:'pending'" json:"status"`
	Subtotal        float64        `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	Discount        float64        `gorm:"type:decimal(10,2);default:0" json:"discount"`
	Tax             float64        `gorm:"type:decimal(10,2);default:0" json:"tax"`
	Total           float64        `gorm:"type:decimal(10,2);not null" json:"total"`
	Currency        string         `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	CouponID        *uuid.UUID     `gorm:"type:uuid" json:"coupon_id,omitempty"`
	PaymentMethod   *PaymentMethod `gorm:"type:payment_method" json:"payment_method,omitempty"`
	PaymentIntentID *string        `gorm:"type:varchar(255)" json:"payment_intent_id,omitempty"`
	PaidAt          *time.Time     `json:"paid_at,omitempty"`
	RefundedAt      *time.Time     `json:"refunded_at,omitempty"`
	RefundReason    *string        `gorm:"type:text" json:"refund_reason,omitempty"`
	CreatedAt       time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	User   *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Coupon *Coupon     `gorm:"foreignKey:CouponID" json:"coupon,omitempty"`
	Items  []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusCompleted && o.PaidAt != nil
}

func GenerateOrderNumber() string {
	return "ORD-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:8]
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID         uuid.UUID `gorm:"type:uuid;index;not null" json:"order_id"`
	CourseID        uuid.UUID `gorm:"type:uuid;not null" json:"course_id"`
	Price           float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Discount        float64   `gorm:"type:decimal(10,2);default:0" json:"discount"`
	InstructorShare float64   `gorm:"type:decimal(10,2)" json:"instructor_share"`

	Order  *Order  `gorm:"foreignKey:OrderID" json:"-"`
	Course *Course `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}

// InstructorEarning tracks earnings for instructors
type InstructorEarning struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	InstructorID uuid.UUID  `gorm:"type:uuid;index;not null" json:"instructor_id"`
	OrderItemID  uuid.UUID  `gorm:"type:uuid;not null" json:"order_item_id"`
	Amount       float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	PlatformFee  float64    `gorm:"type:decimal(10,2);not null" json:"platform_fee"`
	Status       string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	PayoutID     *uuid.UUID `gorm:"type:uuid" json:"payout_id,omitempty"`
	CreatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Instructor *User      `gorm:"foreignKey:InstructorID" json:"-"`
	OrderItem  *OrderItem `gorm:"foreignKey:OrderItemID" json:"-"`
}

// Payout represents a payout to an instructor
type Payout struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	InstructorID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"instructor_id"`
	Amount        float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	Currency      string     `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	Method        *string    `gorm:"type:varchar(50)" json:"method,omitempty"`
	Status        string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	TransactionID *string    `gorm:"type:varchar(255)" json:"transaction_id,omitempty"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	CreatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Instructor *User `gorm:"foreignKey:InstructorID" json:"-"`
}

// OrderRepository interface
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]Order, int64, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*Order, error)
	GetAll(ctx context.Context, status *OrderStatus, page, limit int) ([]Order, int64, error)
	GetByPaymentIntent(ctx context.Context, paymentIntentID string) (*Order, error)
	Update(ctx context.Context, order *Order) error
}

// EarningRepository interface
type EarningRepository interface {
	Create(ctx context.Context, earning *InstructorEarning) error
	GetByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]InstructorEarning, int64, error)
	GetStats(ctx context.Context, instructorID uuid.UUID) (*InstructorStats, error)
	UpdateStatus(ctx context.Context, instructorID uuid.UUID, fromStatus, toStatus string) error
}

// PayoutRepository interface
type PayoutRepository interface {
	Create(ctx context.Context, payout *Payout) error
	GetByID(ctx context.Context, id uuid.UUID) (*Payout, error)
	GetByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]Payout, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// InstructorStats for earning overview
type InstructorStats struct {
	TotalEarnings     float64 `json:"total_earnings"`
	AvailableEarnings float64 `json:"available_earnings"`
	PendingEarnings   float64 `json:"pending_earnings"`
	WithdrawnAmount   float64 `json:"withdrawn_amount"`
}

// RevenueUseCase interface
type RevenueUseCase interface {
	GetInstructorStats(ctx context.Context, instructorID uuid.UUID) (*InstructorStats, error)
	GetEarnings(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]InstructorEarning, int64, error)
	GetPayouts(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]Payout, int64, error)
	RequestPayout(ctx context.Context, instructorID uuid.UUID, amount float64) (*Payout, error)
	HandlePaymentSuccess(ctx context.Context, order *Order) error
}

// OrderUseCase interface
type OrderUseCase interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*Order, error)
	GetMyOrders(ctx context.Context, userID uuid.UUID, page, limit int) ([]Order, int64, error)
	CreateOrder(ctx context.Context, userID uuid.UUID, email string, input CreateOrderInput) (*CreateOrderOutput, error)
	CreateCheckout(ctx context.Context, userID uuid.UUID, email string, input CreateOrderInput) (*CreateOrderOutput, error)
	ConfirmPayment(ctx context.Context, paymentIntentID string) (*Order, error)
	HandleWebhook(ctx context.Context, eventType string, paymentIntentID string) error
	ValidateCoupon(ctx context.Context, code string, subtotal float64) (*Coupon, float64, error)
	ListCoupons(ctx context.Context, page, limit int) ([]Coupon, int64, error)
	CreateCoupon(ctx context.Context, input CreateCouponInput, createdBy uuid.UUID) (*Coupon, error)
	DeleteCoupon(ctx context.Context, id uuid.UUID) error
	ToggleCoupon(ctx context.Context, id uuid.UUID, isActive bool) error
}

// These inputs/outputs should ideally be in domain as well if shared
type CreateOrderInput struct {
	CouponCode *string `json:"coupon_code"`
}

type CreateOrderOutput struct {
	Order           *Order `json:"order"`
	CheckoutURL     string `json:"checkout_url,omitempty"`
	ClientSecret    string `json:"client_secret,omitempty"`
	PaymentIntentID string `json:"payment_intent_id,omitempty"`
}

type CreateCouponInput struct {
	Code        string     `json:"code" validate:"required,min=3,max=50"`
	CouponType  CouponType `json:"coupon_type" validate:"required,oneof=percentage fixed free"`
	Value       float64    `json:"value" validate:"required,gte=0"`
	MinPurchase float64    `json:"min_purchase"`
	MaxDiscount *float64   `json:"max_discount"`
	UsageLimit  *int       `json:"usage_limit"`
	ExpiresAt   *time.Time `json:"expires_at"`
}
