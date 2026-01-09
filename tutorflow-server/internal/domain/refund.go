package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RefundStatus defines refund states
type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusApproved  RefundStatus = "approved"
	RefundStatusRejected  RefundStatus = "rejected"
	RefundStatusProcessed RefundStatus = "processed"
)

// RefundReason defines refund reasons
type RefundReason string

const (
	RefundReasonNotAsDescribed RefundReason = "not_as_described"
	RefundReasonDuplicate      RefundReason = "duplicate"
	RefundReasonTechnicalIssue RefundReason = "technical_issue"
	RefundReasonNoLongerNeeded RefundReason = "no_longer_needed"
	RefundReasonOther          RefundReason = "other"
)

// Refund represents a refund request
type Refund struct {
	ID             uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID        uuid.UUID    `gorm:"type:uuid;index;not null" json:"order_id"`
	UserID         uuid.UUID    `gorm:"type:uuid;index;not null" json:"user_id"`
	Amount         float64      `gorm:"type:decimal(10,2);not null" json:"amount"`
	Reason         RefundReason `gorm:"size:50;not null" json:"reason"`
	Description    string       `gorm:"type:text" json:"description"`
	Status         RefundStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	AdminNotes     string       `gorm:"type:text" json:"admin_notes,omitempty"`
	ProcessedBy    *uuid.UUID   `gorm:"type:uuid" json:"processed_by,omitempty"`
	ProcessedAt    *time.Time   `gorm:"" json:"processed_at,omitempty"`
	StripeRefundID string       `gorm:"size:100" json:"-"`
	CreatedAt      time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Order     *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	User      *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Processor *User  `gorm:"foreignKey:ProcessedBy" json:"processor,omitempty"`
}

// IsPending checks if refund is pending
func (r *Refund) IsPending() bool {
	return r.Status == RefundStatusPending
}

// CanProcess checks if refund can be processed
func (r *Refund) CanProcess() bool {
	return r.Status == RefundStatusPending
}

// RefundUseCase interface
type RefundUseCase interface {
	RequestRefund(ctx context.Context, userID, orderID uuid.UUID, reason RefundReason, description string) (*Refund, error)
	ApproveRefund(ctx context.Context, refundID, adminID uuid.UUID, notes string) (*Refund, error)
	RejectRefund(ctx context.Context, refundID, adminID uuid.UUID, notes string) (*Refund, error)
	GetUserRefunds(ctx context.Context, userID uuid.UUID, page, limit int) ([]Refund, int64, error)
	GetPendingRefunds(ctx context.Context, page, limit int) ([]Refund, int64, error)
	GetAllRefunds(ctx context.Context, status *RefundStatus, page, limit int) ([]Refund, int64, error)
	GetRefundByID(ctx context.Context, id uuid.UUID) (*Refund, error)
}

// RefundPolicy represents refund rules
type RefundPolicy struct {
	MaxDaysAfterPurchase int     `json:"max_days_after_purchase"` // e.g., 30 days
	MaxProgressPercent   float64 `json:"max_progress_percent"`    // e.g., 30% watched
	RequiresApproval     bool    `json:"requires_approval"`
	AutoApproveUnder     float64 `json:"auto_approve_under"` // Auto-approve if amount under this
}

// DefaultRefundPolicy returns default refund policy
func DefaultRefundPolicy() RefundPolicy {
	return RefundPolicy{
		MaxDaysAfterPurchase: 30,
		MaxProgressPercent:   30.0,
		RequiresApproval:     true,
		AutoApproveUnder:     10.0,
	}
}
