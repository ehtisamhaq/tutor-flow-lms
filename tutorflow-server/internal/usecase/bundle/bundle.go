package bundle

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

type bundleUseCase struct {
	bundleRepo     repository.BundleRepository
	courseRepo     repository.CourseRepository
	orderRepo      repository.OrderRepository
	enrollmentRepo repository.EnrollmentRepository
}

// NewUseCase creates a new bundle use case
func NewUseCase(
	bundleRepo repository.BundleRepository,
	courseRepo repository.CourseRepository,
	orderRepo repository.OrderRepository,
	enrollmentRepo repository.EnrollmentRepository,
) domain.BundleUseCase {
	return &bundleUseCase{
		bundleRepo:     bundleRepo,
		courseRepo:     courseRepo,
		orderRepo:      orderRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

// CreateBundle creates a new course bundle
func (uc *bundleUseCase) CreateBundle(
	ctx context.Context,
	bundle *domain.Bundle,
	courseIDs []uuid.UUID,
) (*domain.Bundle, error) {
	if bundle.Title == "" {
		return nil, errors.New("bundle title is required")
	}

	if len(courseIDs) == 0 {
		return nil, errors.New("at least one course is required")
	}

	// Generate slug if not provided
	if bundle.Slug == "" {
		bundle.Slug = slug.Make(bundle.Title)
	}

	// Calculate original price from courses
	var originalPrice float64
	for _, courseID := range courseIDs {
		course, err := uc.courseRepo.GetByID(ctx, courseID)
		if err != nil {
			return nil, errors.New("course not found: " + courseID.String())
		}
		if course.DiscountPrice != nil && *course.DiscountPrice > 0 {
			originalPrice += *course.DiscountPrice
		} else {
			originalPrice += course.Price
		}
	}

	bundle.OriginalPrice = originalPrice
	bundle.BundlePrice = originalPrice * (1 - bundle.DiscountPercent/100)
	bundle.Type = domain.BundleTypeFixed

	if err := uc.bundleRepo.Create(ctx, bundle); err != nil {
		return nil, err
	}

	// Add courses to bundle
	for i, courseID := range courseIDs {
		if err := uc.bundleRepo.AddCourse(ctx, bundle.ID, courseID, i+1); err != nil {
			return nil, err
		}
	}

	// Reload with courses
	return uc.bundleRepo.GetByID(ctx, bundle.ID)
}

// GetBundleBySlug returns a bundle by slug
func (uc *bundleUseCase) GetBundleBySlug(ctx context.Context, bundleSlug string) (*domain.Bundle, error) {
	bundle, err := uc.bundleRepo.GetBySlug(ctx, bundleSlug)
	if err != nil {
		return nil, err
	}

	if !bundle.IsAvailable() {
		return nil, errors.New("bundle is not available")
	}

	return bundle, nil
}

// GetActiveBundles returns all active bundles
func (uc *bundleUseCase) GetActiveBundles(ctx context.Context, page, limit int) ([]domain.Bundle, int64, error) {
	return uc.bundleRepo.GetActive(ctx, page, limit)
}

// UpdateBundle updates an existing bundle
func (uc *bundleUseCase) UpdateBundle(ctx context.Context, bundle *domain.Bundle) error {
	existing, err := uc.bundleRepo.GetByID(ctx, bundle.ID)
	if err != nil {
		return errors.New("bundle not found")
	}

	existing.Title = bundle.Title
	existing.Description = bundle.Description
	existing.ThumbnailURL = bundle.ThumbnailURL
	existing.DiscountPercent = bundle.DiscountPercent
	existing.BundlePrice = existing.OriginalPrice * (1 - bundle.DiscountPercent/100)
	existing.IsActive = bundle.IsActive
	existing.StartDate = bundle.StartDate
	existing.EndDate = bundle.EndDate
	existing.MaxPurchases = bundle.MaxPurchases
	existing.UpdatedAt = time.Now()

	return uc.bundleRepo.Update(ctx, existing)
}

// DeleteBundle deletes a bundle
func (uc *bundleUseCase) DeleteBundle(ctx context.Context, id uuid.UUID) error {
	return uc.bundleRepo.Delete(ctx, id)
}

// AddCourseToBundle adds a course to a bundle
func (uc *bundleUseCase) AddCourseToBundle(ctx context.Context, bundleID, courseID uuid.UUID) error {
	bundle, err := uc.bundleRepo.GetByID(ctx, bundleID)
	if err != nil {
		return errors.New("bundle not found")
	}

	course, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return errors.New("course not found")
	}

	// Update original price
	price := course.Price
	if course.DiscountPrice != nil && *course.DiscountPrice > 0 {
		price = *course.DiscountPrice
	}
	bundle.OriginalPrice += price
	bundle.BundlePrice = bundle.OriginalPrice * (1 - bundle.DiscountPercent/100)

	if err := uc.bundleRepo.Update(ctx, bundle); err != nil {
		return err
	}

	order := len(bundle.Courses) + 1
	return uc.bundleRepo.AddCourse(ctx, bundleID, courseID, order)
}

// RemoveCourseFromBundle removes a course from a bundle
func (uc *bundleUseCase) RemoveCourseFromBundle(ctx context.Context, bundleID, courseID uuid.UUID) error {
	bundle, err := uc.bundleRepo.GetByID(ctx, bundleID)
	if err != nil {
		return errors.New("bundle not found")
	}

	course, err := uc.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return errors.New("course not found")
	}

	// Update original price
	price := course.Price
	if course.DiscountPrice != nil && *course.DiscountPrice > 0 {
		price = *course.DiscountPrice
	}
	bundle.OriginalPrice -= price
	if bundle.OriginalPrice < 0 {
		bundle.OriginalPrice = 0
	}
	bundle.BundlePrice = bundle.OriginalPrice * (1 - bundle.DiscountPercent/100)

	if err := uc.bundleRepo.Update(ctx, bundle); err != nil {
		return err
	}

	return uc.bundleRepo.RemoveCourse(ctx, bundleID, courseID)
}

// PurchaseBundle purchases a bundle and enrolls user in all courses
func (uc *bundleUseCase) PurchaseBundle(
	ctx context.Context,
	userID, bundleID uuid.UUID,
) (*domain.BundlePurchase, error) {
	bundle, err := uc.bundleRepo.GetByID(ctx, bundleID)
	if err != nil {
		return nil, errors.New("bundle not found")
	}

	if !bundle.IsAvailable() {
		return nil, errors.New("bundle is not available")
	}

	// Get bundle courses
	courses, err := uc.bundleRepo.GetCourses(ctx, bundleID)
	if err != nil {
		return nil, err
	}

	// Create order for the bundle
	order := &domain.Order{
		UserID:      userID,
		OrderNumber: domain.GenerateOrderNumber(),
		Subtotal:    bundle.OriginalPrice,
		Discount:    bundle.Savings(),
		Total:       bundle.BundlePrice,
		Status:      domain.OrderStatusCompleted,
	}

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Create enrollments for each course
	for _, course := range courses {
		enrollment := &domain.Enrollment{
			UserID:   userID,
			CourseID: course.ID,
		}
		uc.enrollmentRepo.Create(ctx, enrollment)
	}

	// Record purchase
	purchase := &domain.BundlePurchase{
		BundleID: bundleID,
		UserID:   userID,
		OrderID:  order.ID,
		Price:    bundle.BundlePrice,
	}

	if err := uc.bundleRepo.RecordPurchase(ctx, purchase); err != nil {
		return nil, err
	}

	// Increment purchase count
	uc.bundleRepo.IncrementPurchaseCount(ctx, bundleID)

	purchase.Bundle = bundle
	return purchase, nil
}

// GetUserBundles returns user's purchased bundles
func (uc *bundleUseCase) GetUserBundles(ctx context.Context, userID uuid.UUID) ([]domain.BundlePurchase, error) {
	return uc.bundleRepo.GetUserPurchases(ctx, userID)
}

// CalculateBundlePrice calculates bundle pricing
func (uc *bundleUseCase) CalculateBundlePrice(
	ctx context.Context,
	courseIDs []uuid.UUID,
	discountPercent float64,
) (original, discounted float64, err error) {
	for _, courseID := range courseIDs {
		course, err := uc.courseRepo.GetByID(ctx, courseID)
		if err != nil {
			return 0, 0, errors.New("course not found")
		}
		if course.DiscountPrice != nil && *course.DiscountPrice > 0 {
			original += *course.DiscountPrice
		} else {
			original += course.Price
		}
	}

	discounted = original * (1 - discountPercent/100)
	return original, discounted, nil
}
