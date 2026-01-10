package order

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
	"github.com/tutorflow/tutorflow-server/internal/service/payment"
)

// UseCase defines order business logic
type UseCase struct {
	orderRepo      repository.OrderRepository
	cartRepo       repository.CartRepository
	couponRepo     repository.CouponRepository
	enrollmentRepo repository.EnrollmentRepository
	courseRepo     repository.CourseRepository
	paymentSvc     *payment.Service
}

// NewUseCase creates a new order use case
func NewUseCase(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	couponRepo repository.CouponRepository,
	enrollmentRepo repository.EnrollmentRepository,
	courseRepo repository.CourseRepository,
	paymentSvc *payment.Service,
) *UseCase {
	return &UseCase{
		orderRepo:      orderRepo,
		cartRepo:       cartRepo,
		couponRepo:     couponRepo,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
		paymentSvc:     paymentSvc,
	}
}

// GetByID returns order by ID
func (uc *UseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return uc.orderRepo.GetByID(ctx, id)
}

// GetByOrderNumber returns order by number
func (uc *UseCase) GetByOrderNumber(ctx context.Context, orderNumber string) (*domain.Order, error) {
	return uc.orderRepo.GetByOrderNumber(ctx, orderNumber)
}

// GetMyOrders returns user's orders
func (uc *UseCase) GetMyOrders(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return uc.orderRepo.GetByUser(ctx, userID, page, limit)
}

// CreateOrder creates an order from cart
func (uc *UseCase) CreateOrder(ctx context.Context, userID uuid.UUID, email string, input domain.CreateOrderInput) (*domain.CreateOrderOutput, error) {
	// Get user's cart
	cart, err := uc.cartRepo.GetOrCreate(ctx, &userID, nil)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// Calculate totals
	var subtotal float64
	var orderItems []domain.OrderItem
	var lineItems []payment.LineItem

	platformFeePercent := 0.30 // 30% platform fee

	for _, item := range cart.Items {
		course, err := uc.courseRepo.GetByID(ctx, item.CourseID)
		if err != nil {
			continue
		}

		price := course.GetEffectivePrice()
		instructorShare := price * (1 - platformFeePercent)

		orderItems = append(orderItems, domain.OrderItem{
			CourseID:        course.ID,
			Price:           price,
			InstructorShare: instructorShare,
		})

		lineItems = append(lineItems, payment.LineItem{
			Name:        course.Title,
			Description: fmt.Sprintf("Course by %s", course.Instructor.FirstName),
			Amount:      int64(price * 100), // Convert to cents
			Quantity:    1,
		})

		subtotal += price
	}

	// Apply coupon if provided
	var discount float64
	var coupon *domain.Coupon
	var couponID *uuid.UUID

	if input.CouponCode != nil && *input.CouponCode != "" {
		coupon, err = uc.couponRepo.GetByCode(ctx, *input.CouponCode)
		if err == nil && coupon.IsValid() {
			discount = coupon.CalculateDiscount(subtotal)
			couponID = &coupon.ID
		}
	}

	total := subtotal - discount

	// Create order
	orderNumber := domain.GenerateOrderNumber()
	order := &domain.Order{
		OrderNumber: orderNumber,
		UserID:      userID,
		Status:      domain.OrderStatusPending,
		Subtotal:    subtotal,
		Discount:    discount,
		Total:       total,
		CouponID:    couponID,
		Items:       orderItems,
	}

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Handle free orders
	if total == 0 {
		return uc.completeOrder(ctx, order)
	}

	// Create Stripe payment intent
	pi, err := uc.paymentSvc.CreatePaymentIntent(ctx, payment.CreatePaymentIntentInput{
		Amount:   int64(total * 100),
		Currency: "usd",
		OrderID:  order.ID.String(),
		Email:    email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Update order with payment intent
	order.PaymentIntentID = &pi.ID
	paymentMethod := domain.PaymentMethodStripe
	order.PaymentMethod = &paymentMethod
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	return &domain.CreateOrderOutput{
		Order:           order,
		ClientSecret:    pi.ClientSecret,
		PaymentIntentID: pi.ID,
	}, nil
}

// CreateCheckout creates a Stripe Checkout session from cart
func (uc *UseCase) CreateCheckout(ctx context.Context, userID uuid.UUID, email string, input domain.CreateOrderInput) (*domain.CreateOrderOutput, error) {
	// Get user's cart
	cart, err := uc.cartRepo.GetOrCreate(ctx, &userID, nil)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// Calculate totals and prepare line items
	var subtotal float64
	var orderItems []domain.OrderItem
	var lineItems []payment.LineItem

	platformFeePercent := 0.30

	for _, item := range cart.Items {
		course, err := uc.courseRepo.GetByID(ctx, item.CourseID)
		if err != nil {
			continue
		}

		price := course.GetEffectivePrice()
		instructorShare := price * (1 - platformFeePercent)

		orderItems = append(orderItems, domain.OrderItem{
			CourseID:        course.ID,
			Price:           price,
			InstructorShare: instructorShare,
		})

		imageURL := ""
		if course.ThumbnailURL != nil {
			imageURL = *course.ThumbnailURL
		}

		lineItems = append(lineItems, payment.LineItem{
			Name:        course.Title,
			Description: fmt.Sprintf("Course by %s", course.Instructor.FirstName),
			Amount:      int64(price * 100),
			Quantity:    1,
			ImageURL:    imageURL,
		})

		subtotal += price
	}

	// Apply coupon
	var discount float64
	var couponID *uuid.UUID
	if input.CouponCode != nil && *input.CouponCode != "" {
		coupon, err := uc.couponRepo.GetByCode(ctx, *input.CouponCode)
		if err == nil && coupon.IsValid() {
			discount = coupon.CalculateDiscount(subtotal)
			couponID = &coupon.ID
		}
	}

	total := subtotal - discount

	// Create order
	order := &domain.Order{
		OrderNumber: domain.GenerateOrderNumber(),
		UserID:      userID,
		Status:      domain.OrderStatusPending,
		Subtotal:    subtotal,
		Discount:    discount,
		Total:       total,
		CouponID:    couponID,
		Items:       orderItems,
	}

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Handle free orders
	if total == 0 {
		return uc.completeOrder(ctx, order)
	}

	// Create Stripe Checkout session
	session, err := uc.paymentSvc.CreateCheckoutSession(ctx, payment.CreateCheckoutSessionInput{
		CustomerEmail: email,
		OrderID:       order.ID.String(),
		Items:         lineItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	// Update order with payment intent if available (Stripe Checkout sessions might not have PI immediately or it's different)
	// For now, we mainly need the checkout URL.
	paymentMethod := domain.PaymentMethodStripe
	order.PaymentMethod = &paymentMethod
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	return &domain.CreateOrderOutput{
		Order:       order,
		CheckoutURL: session.URL,
	}, nil
}

// ConfirmPayment confirms payment and completes order
func (uc *UseCase) ConfirmPayment(ctx context.Context, paymentIntentID string) (*domain.Order, error) {
	// Get order by payment intent
	order, err := uc.orderRepo.GetByPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return nil, err
	}

	// Verify payment with Stripe
	pi, err := uc.paymentSvc.GetPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return nil, err
	}

	if pi.Status != "succeeded" {
		return nil, fmt.Errorf("payment not completed: %s", pi.Status)
	}

	// Complete the order
	output, err := uc.completeOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	return output.Order, nil
}

// completeOrder marks order as complete and creates enrollments
func (uc *UseCase) completeOrder(ctx context.Context, order *domain.Order) (*domain.CreateOrderOutput, error) {
	now := time.Now()
	order.Status = domain.OrderStatusCompleted
	order.PaidAt = &now

	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	// Create enrollments for each course
	for _, item := range order.Items {
		enrollment := &domain.Enrollment{
			UserID:    order.UserID,
			CourseID:  item.CourseID,
			Status:    domain.EnrollmentStatusActive,
			StartedAt: &now,
			OrderID:   &order.ID,
		}
		_ = uc.enrollmentRepo.Create(ctx, enrollment)

		// Update course student count
		_ = uc.courseRepo.IncrementStudentCount(ctx, item.CourseID)
	}

	// Clear cart
	cart, _ := uc.cartRepo.GetOrCreate(ctx, &order.UserID, nil)
	if cart != nil {
		_ = uc.cartRepo.Clear(ctx, cart.ID)
	}

	// Increment coupon usage
	if order.CouponID != nil {
		_ = uc.couponRepo.IncrementUsage(ctx, *order.CouponID)
	}

	return &domain.CreateOrderOutput{Order: order}, nil
}

// HandleWebhook handles Stripe webhook events
func (uc *UseCase) HandleWebhook(ctx context.Context, eventType string, paymentIntentID string) error {
	switch eventType {
	case "payment_intent.succeeded":
		_, err := uc.ConfirmPayment(ctx, paymentIntentID)
		return err
	case "payment_intent.payment_failed":
		order, err := uc.orderRepo.GetByPaymentIntent(ctx, paymentIntentID)
		if err != nil {
			return err
		}
		order.Status = domain.OrderStatusFailed
		return uc.orderRepo.Update(ctx, order)
	}
	return nil
}

// --- Coupons ---

// ValidateCoupon validates a coupon code
func (uc *UseCase) ValidateCoupon(ctx context.Context, code string, subtotal float64) (*domain.Coupon, float64, error) {
	coupon, err := uc.couponRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, 0, domain.ErrCouponInvalid
	}

	if !coupon.IsValid() {
		return nil, 0, domain.ErrCouponInvalid
	}

	discount := coupon.CalculateDiscount(subtotal)
	return coupon, discount, nil
}

// CreateCouponInput for creating coupon
type CreateCouponInput struct {
	Code        string            `json:"code" validate:"required,min=3,max=50"`
	CouponType  domain.CouponType `json:"coupon_type" validate:"required,oneof=percentage fixed free"`
	Value       float64           `json:"value" validate:"required,gte=0"`
	MinPurchase float64           `json:"min_purchase"`
	MaxDiscount *float64          `json:"max_discount"`
	UsageLimit  *int              `json:"usage_limit"`
	ExpiresAt   *time.Time        `json:"expires_at"`
}

// CreateCoupon creates a new coupon (admin)
func (uc *UseCase) CreateCoupon(ctx context.Context, input domain.CreateCouponInput, createdBy uuid.UUID) (*domain.Coupon, error) {
	coupon := &domain.Coupon{
		Code:        input.Code,
		CouponType:  input.CouponType,
		Value:       input.Value,
		MinPurchase: input.MinPurchase,
		MaxDiscount: input.MaxDiscount,
		UsageLimit:  input.UsageLimit,
		ExpiresAt:   input.ExpiresAt,
		CreatedBy:   &createdBy,
		IsActive:    true,
	}

	if err := uc.couponRepo.Create(ctx, coupon); err != nil {
		return nil, err
	}

	return coupon, nil
}

// ListCoupons lists all coupons
func (uc *UseCase) ListCoupons(ctx context.Context, page, limit int) ([]domain.Coupon, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	return uc.couponRepo.List(ctx, page, limit)
}

// DeleteCoupon deletes a coupon
func (uc *UseCase) DeleteCoupon(ctx context.Context, id uuid.UUID) error {
	return uc.couponRepo.Delete(ctx, id)
}

// ToggleCoupon enables/disables a coupon
func (uc *UseCase) ToggleCoupon(ctx context.Context, id uuid.UUID, isActive bool) error {
	coupon, err := uc.couponRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	coupon.IsActive = isActive
	return uc.couponRepo.Update(ctx, coupon)
}
