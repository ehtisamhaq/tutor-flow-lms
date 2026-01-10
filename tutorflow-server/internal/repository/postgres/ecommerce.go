package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// CartRepository
type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) repository.CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetOrCreate(ctx context.Context, userID *uuid.UUID, sessionID *string) (*domain.Cart, error) {
	var cart domain.Cart

	query := r.db.WithContext(ctx)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else if sessionID != nil {
		query = query.Where("session_id = ?", *sessionID)
	} else {
		return nil, errors.New("either user_id or session_id is required")
	}

	err := query.Preload("Items").Preload("Items.Course").First(&cart).Error
	if err == nil {
		return &cart, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new cart
	cart = domain.Cart{
		UserID:    userID,
		SessionID: sessionID,
	}
	if err := r.db.WithContext(ctx).Create(&cart).Error; err != nil {
		return nil, err
	}

	return &cart, nil
}

func (r *cartRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cart, error) {
	var cart domain.Cart
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Course").
		Preload("Items.Course.Instructor").
		Where("id = ?", id).
		First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) AddItem(ctx context.Context, cartID, courseID uuid.UUID) error {
	// Check if item already exists
	var count int64
	r.db.WithContext(ctx).Model(&domain.CartItem{}).
		Where("cart_id = ? AND course_id = ?", cartID, courseID).
		Count(&count)

	if count > 0 {
		return nil // Already in cart
	}

	item := &domain.CartItem{
		CartID:   cartID,
		CourseID: courseID,
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartID, courseID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("cart_id = ? AND course_id = ?", cartID, courseID).
		Delete(&domain.CartItem{}).Error
}

func (r *cartRepository) Clear(ctx context.Context, cartID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("cart_id = ?", cartID).
		Delete(&domain.CartItem{}).Error
}

func (r *cartRepository) MergeGuestCart(ctx context.Context, sessionID string, userID uuid.UUID) error {
	// Find guest cart
	var guestCart domain.Cart
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("session_id = ?", sessionID).
		First(&guestCart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // No guest cart to merge
		}
		return err
	}

	// Get or create user cart
	userCart, err := r.GetOrCreate(ctx, &userID, nil)
	if err != nil {
		return err
	}

	// Merge items
	for _, item := range guestCart.Items {
		_ = r.AddItem(ctx, userCart.ID, item.CourseID)
	}

	// Delete guest cart
	r.db.WithContext(ctx).Where("cart_id = ?", guestCart.ID).Delete(&domain.CartItem{})
	return r.db.WithContext(ctx).Delete(&guestCart).Error
}

// WishlistRepository
type wishlistRepository struct {
	db *gorm.DB
}

func NewWishlistRepository(db *gorm.DB) repository.WishlistRepository {
	return &wishlistRepository{db: db}
}

func (r *wishlistRepository) Add(ctx context.Context, userID, courseID uuid.UUID) error {
	var count int64
	r.db.WithContext(ctx).Model(&domain.Wishlist{}).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Count(&count)

	if count > 0 {
		return nil
	}

	item := &domain.Wishlist{
		UserID:   userID,
		CourseID: courseID,
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *wishlistRepository) Remove(ctx context.Context, userID, courseID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Delete(&domain.Wishlist{}).Error
}

func (r *wishlistRepository) GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Wishlist, int64, error) {
	var items []domain.Wishlist
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Wishlist{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Course").
		Preload("Course.Instructor").
		Order("added_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error

	return items, total, err
}

func (r *wishlistRepository) Exists(ctx context.Context, userID, courseID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Wishlist{}).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Count(&count).Error
	return count > 0, err
}

// CouponRepository
type couponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) repository.CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) Create(ctx context.Context, coupon *domain.Coupon) error {
	return r.db.WithContext(ctx).Create(coupon).Error
}

func (r *couponRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Coupon, error) {
	var coupon domain.Coupon
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) GetByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	var coupon domain.Coupon
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCouponInvalid
		}
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) Update(ctx context.Context, coupon *domain.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}

func (r *couponRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Coupon{}, "id = ?", id).Error
}

func (r *couponRepository) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Coupon{}).
		Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}

func (r *couponRepository) List(ctx context.Context, page, limit int) ([]domain.Coupon, int64, error) {
	var coupons []domain.Coupon
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Coupon{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&coupons).Error
	return coupons, total, err
}

// OrderRepository
type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repository.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Items").
		Preload("Items.Course").
		Preload("Coupon").
		Where("id = ?", id).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Course").
		Where("order_number = ?", orderNumber).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Update(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepository) GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Order, int64, error) {
	var orders []domain.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Order{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("Items").
		Preload("Items.Course").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepository) GetByPaymentIntent(ctx context.Context, paymentIntentID string) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("payment_intent_id = ?", paymentIntentID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// EarningRepository
type earningRepository struct {
	db *gorm.DB
}

func NewEarningRepository(db *gorm.DB) repository.EarningRepository {
	return &earningRepository{db: db}
}

func (r *earningRepository) Create(ctx context.Context, earning *domain.InstructorEarning) error {
	return r.db.WithContext(ctx).Create(earning).Error
}

func (r *earningRepository) GetByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]domain.InstructorEarning, int64, error) {
	var earnings []domain.InstructorEarning
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.InstructorEarning{}).Where("instructor_id = ?", instructorID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&earnings).Error
	return earnings, total, err
}

func (r *earningRepository) GetStats(ctx context.Context, instructorID uuid.UUID) (*domain.InstructorStats, error) {
	var stats domain.InstructorStats

	// Calculate Total Earnings
	r.db.WithContext(ctx).Model(&domain.InstructorEarning{}).
		Where("instructor_id = ?", instructorID).
		Select("SUM(amount)").Scan(&stats.TotalEarnings)

	// Calculate Pending Earnings
	r.db.WithContext(ctx).Model(&domain.InstructorEarning{}).
		Where("instructor_id = ? AND status = ?", instructorID, "pending").
		Select("SUM(amount)").Scan(&stats.PendingEarnings)

	// Calculate Available Earnings
	r.db.WithContext(ctx).Model(&domain.InstructorEarning{}).
		Where("instructor_id = ? AND status = ?", instructorID, "available").
		Select("SUM(amount)").Scan(&stats.AvailableEarnings)

	// Calculate Withdrawn Amount
	r.db.WithContext(ctx).Model(&domain.Payout{}).
		Where("instructor_id = ? AND status = ?", instructorID, "processed").
		Select("SUM(amount)").Scan(&stats.WithdrawnAmount)

	return &stats, nil
}

func (r *earningRepository) UpdateStatus(ctx context.Context, instructorID uuid.UUID, fromStatus, toStatus string) error {
	return r.db.WithContext(ctx).Model(&domain.InstructorEarning{}).
		Where("instructor_id = ? AND status = ?", instructorID, fromStatus).
		Update("status", toStatus).Error
}

// PayoutRepository
type payoutRepository struct {
	db *gorm.DB
}

func NewPayoutRepository(db *gorm.DB) repository.PayoutRepository {
	return &payoutRepository{db: db}
}

func (r *payoutRepository) Create(ctx context.Context, payout *domain.Payout) error {
	return r.db.WithContext(ctx).Create(payout).Error
}

func (r *payoutRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payout, error) {
	var payout domain.Payout
	if err := r.db.WithContext(ctx).First(&payout, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &payout, nil
}

func (r *payoutRepository) GetByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]domain.Payout, int64, error) {
	var payouts []domain.Payout
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Payout{}).Where("instructor_id = ?", instructorID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&payouts).Error
	return payouts, total, err
}

func (r *payoutRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	now := time.Now()
	update := map[string]interface{}{
		"status": status,
	}
	if status == "processed" {
		update["processed_at"] = &now
	}
	return r.db.WithContext(ctx).Model(&domain.Payout{}).Where("id = ?", id).Updates(update).Error
}
