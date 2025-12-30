package cart

import (
	"context"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// UseCase defines cart business logic
type UseCase struct {
	cartRepo       repository.CartRepository
	wishlistRepo   repository.WishlistRepository
	courseRepo     repository.CourseRepository
	enrollmentRepo repository.EnrollmentRepository
}

// NewUseCase creates a new cart use case
func NewUseCase(
	cartRepo repository.CartRepository,
	wishlistRepo repository.WishlistRepository,
	courseRepo repository.CourseRepository,
	enrollmentRepo repository.EnrollmentRepository,
) *UseCase {
	return &UseCase{
		cartRepo:       cartRepo,
		wishlistRepo:   wishlistRepo,
		courseRepo:     courseRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

// GetCart returns user's cart
func (uc *UseCase) GetCart(ctx context.Context, userID *uuid.UUID, sessionID *string) (*domain.Cart, error) {
	return uc.cartRepo.GetOrCreate(ctx, userID, sessionID)
}

// AddToCartInput for adding item
type AddToCartInput struct {
	CourseID uuid.UUID `json:"course_id" validate:"required"`
}

// AddToCart adds a course to cart
func (uc *UseCase) AddToCart(ctx context.Context, userID *uuid.UUID, sessionID *string, input AddToCartInput) (*domain.Cart, error) {
	// Get or create cart
	cart, err := uc.cartRepo.GetOrCreate(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify course exists and is published
	course, err := uc.courseRepo.GetByID(ctx, input.CourseID)
	if err != nil {
		return nil, domain.ErrCourseNotFound
	}
	if course.Status != domain.CourseStatusPublished {
		return nil, domain.ErrCourseNotPublished
	}

	// Check if user is already enrolled
	if userID != nil {
		enrollment, _ := uc.enrollmentRepo.GetByUserAndCourse(ctx, *userID, input.CourseID)
		if enrollment != nil && enrollment.CanAccess() {
			return nil, domain.ErrAlreadyEnrolled
		}
	}

	// Add to cart
	if err := uc.cartRepo.AddItem(ctx, cart.ID, input.CourseID); err != nil {
		return nil, err
	}

	// Return updated cart
	return uc.cartRepo.GetOrCreate(ctx, userID, sessionID)
}

// RemoveFromCart removes a course from cart
func (uc *UseCase) RemoveFromCart(ctx context.Context, userID *uuid.UUID, sessionID *string, courseID uuid.UUID) (*domain.Cart, error) {
	cart, err := uc.cartRepo.GetOrCreate(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if err := uc.cartRepo.RemoveItem(ctx, cart.ID, courseID); err != nil {
		return nil, err
	}

	return uc.cartRepo.GetOrCreate(ctx, userID, sessionID)
}

// ClearCart removes all items from cart
func (uc *UseCase) ClearCart(ctx context.Context, userID *uuid.UUID, sessionID *string) error {
	cart, err := uc.cartRepo.GetOrCreate(ctx, userID, sessionID)
	if err != nil {
		return err
	}

	return uc.cartRepo.Clear(ctx, cart.ID)
}

// MergeGuestCart merges guest cart into user cart
func (uc *UseCase) MergeGuestCart(ctx context.Context, sessionID string, userID uuid.UUID) error {
	return uc.cartRepo.MergeGuestCart(ctx, sessionID, userID)
}

// CartSummary contains cart totals
type CartSummary struct {
	Items         []CartItemSummary `json:"items"`
	Subtotal      float64           `json:"subtotal"`
	Discount      float64           `json:"discount"`
	Total         float64           `json:"total"`
	TotalItems    int               `json:"total_items"`
	CouponApplied *string           `json:"coupon_applied,omitempty"`
}

type CartItemSummary struct {
	CourseID       uuid.UUID `json:"course_id"`
	Title          string    `json:"title"`
	Instructor     string    `json:"instructor"`
	ThumbnailURL   *string   `json:"thumbnail_url,omitempty"`
	Price          float64   `json:"price"`
	DiscountPrice  *float64  `json:"discount_price,omitempty"`
	EffectivePrice float64   `json:"effective_price"`
}

// GetCartSummary returns cart with totals
func (uc *UseCase) GetCartSummary(ctx context.Context, userID *uuid.UUID, sessionID *string) (*CartSummary, error) {
	cart, err := uc.cartRepo.GetOrCreate(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	var items []CartItemSummary
	var subtotal float64

	for _, item := range cart.Items {
		if item.Course == nil {
			continue
		}

		effectivePrice := item.Course.GetEffectivePrice()
		instructor := ""
		if item.Course.Instructor != nil {
			instructor = item.Course.Instructor.FirstName + " " + item.Course.Instructor.LastName
		}

		items = append(items, CartItemSummary{
			CourseID:       item.CourseID,
			Title:          item.Course.Title,
			Instructor:     instructor,
			ThumbnailURL:   item.Course.ThumbnailURL,
			Price:          item.Course.Price,
			DiscountPrice:  item.Course.DiscountPrice,
			EffectivePrice: effectivePrice,
		})

		subtotal += effectivePrice
	}

	return &CartSummary{
		Items:      items,
		Subtotal:   subtotal,
		Discount:   0,
		Total:      subtotal,
		TotalItems: len(items),
	}, nil
}

// --- Wishlist ---

// AddToWishlist adds a course to wishlist
func (uc *UseCase) AddToWishlist(ctx context.Context, userID, courseID uuid.UUID) error {
	// Verify course exists
	_, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return domain.ErrCourseNotFound
	}

	return uc.wishlistRepo.Add(ctx, userID, courseID)
}

// RemoveFromWishlist removes a course from wishlist
func (uc *UseCase) RemoveFromWishlist(ctx context.Context, userID, courseID uuid.UUID) error {
	return uc.wishlistRepo.Remove(ctx, userID, courseID)
}

// GetWishlist returns user's wishlist
func (uc *UseCase) GetWishlist(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Wishlist, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 12
	}
	return uc.wishlistRepo.GetByUser(ctx, userID, page, limit)
}

// IsInWishlist checks if course is in wishlist
func (uc *UseCase) IsInWishlist(ctx context.Context, userID, courseID uuid.UUID) (bool, error) {
	return uc.wishlistRepo.Exists(ctx, userID, courseID)
}

// MoveToCart moves item from wishlist to cart
func (uc *UseCase) MoveToCart(ctx context.Context, userID, courseID uuid.UUID) error {
	// Add to cart
	_, err := uc.AddToCart(ctx, &userID, nil, AddToCartInput{CourseID: courseID})
	if err != nil {
		return err
	}

	// Remove from wishlist
	return uc.wishlistRepo.Remove(ctx, userID, courseID)
}
