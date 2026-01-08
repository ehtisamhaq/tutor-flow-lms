package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tutorflow/tutorflow-server/internal/domain"
)

type subscriptionUseCase struct {
	subscriptionRepo domain.SubscriptionRepository
	userRepo         domain.UserRepository
}

// NewSubscriptionUseCase creates a new subscription use case
func NewSubscriptionUseCase(
	subscriptionRepo domain.SubscriptionRepository,
	userRepo domain.UserRepository,
) domain.SubscriptionUseCase {
	return &subscriptionUseCase{
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
	}
}

// CreatePlan creates a new subscription plan
func (uc *subscriptionUseCase) CreatePlan(plan *domain.SubscriptionPlan) error {
	if plan.Name == "" {
		return errors.New("plan name is required")
	}
	if plan.Slug == "" {
		return errors.New("plan slug is required")
	}
	if plan.PriceMonthly < 0 || plan.PriceYearly < 0 {
		return errors.New("prices must be non-negative")
	}
	return uc.subscriptionRepo.CreatePlan(plan)
}

// GetPlans returns all active subscription plans
func (uc *subscriptionUseCase) GetPlans() ([]domain.SubscriptionPlan, error) {
	return uc.subscriptionRepo.GetActivePlans()
}

// GetPlanBySlug returns a plan by slug
func (uc *subscriptionUseCase) GetPlanBySlug(slug string) (*domain.SubscriptionPlan, error) {
	return uc.subscriptionRepo.GetPlanBySlug(slug)
}

// UpdatePlan updates an existing plan
func (uc *subscriptionUseCase) UpdatePlan(plan *domain.SubscriptionPlan) error {
	existing, err := uc.subscriptionRepo.GetPlanByID(plan.ID)
	if err != nil {
		return errors.New("plan not found")
	}

	existing.Name = plan.Name
	existing.Description = plan.Description
	existing.PriceMonthly = plan.PriceMonthly
	existing.PriceYearly = plan.PriceYearly
	existing.Features = plan.Features
	existing.MaxCourses = plan.MaxCourses
	existing.MaxDownloads = plan.MaxDownloads
	existing.OfflineAccess = plan.OfflineAccess
	existing.CertificateAccess = plan.CertificateAccess
	existing.IsActive = plan.IsActive
	existing.UpdatedAt = time.Now()

	return uc.subscriptionRepo.UpdatePlan(existing)
}

// Subscribe creates a new subscription for a user
func (uc *subscriptionUseCase) Subscribe(
	userID uuid.UUID,
	planSlug string,
	interval domain.SubscriptionInterval,
) (*domain.Subscription, error) {
	// Check if user already has an active subscription
	existing, _ := uc.subscriptionRepo.GetActiveByUserID(userID)
	if existing != nil && existing.IsActive() {
		return nil, errors.New("user already has an active subscription")
	}

	// Get the plan
	plan, err := uc.subscriptionRepo.GetPlanBySlug(planSlug)
	if err != nil {
		return nil, errors.New("plan not found")
	}

	if !plan.IsActive {
		return nil, errors.New("plan is not available")
	}

	// Calculate period dates
	now := time.Now()
	var periodEnd time.Time
	if interval == domain.SubscriptionIntervalMonthly {
		periodEnd = now.AddDate(0, 1, 0)
	} else {
		periodEnd = now.AddDate(1, 0, 0)
	}

	// Create subscription
	subscription := &domain.Subscription{
		UserID:             userID,
		PlanID:             plan.ID,
		Status:             domain.SubscriptionStatusActive,
		Interval:           interval,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
		CancelAtPeriodEnd:  false,
	}

	if err := uc.subscriptionRepo.Create(subscription); err != nil {
		return nil, err
	}

	// Load the plan for response
	subscription.Plan = plan

	return subscription, nil
}

// GetUserSubscription returns the user's current subscription
func (uc *subscriptionUseCase) GetUserSubscription(userID uuid.UUID) (*domain.Subscription, error) {
	return uc.subscriptionRepo.GetActiveByUserID(userID)
}

// CancelSubscription cancels the user's subscription at period end
func (uc *subscriptionUseCase) CancelSubscription(userID uuid.UUID) error {
	sub, err := uc.subscriptionRepo.GetActiveByUserID(userID)
	if err != nil {
		return errors.New("no active subscription found")
	}

	now := time.Now()
	sub.CancelAtPeriodEnd = true
	sub.CanceledAt = &now
	sub.UpdatedAt = now

	return uc.subscriptionRepo.Update(sub)
}

// ResumeSubscription resumes a canceled subscription
func (uc *subscriptionUseCase) ResumeSubscription(userID uuid.UUID) error {
	sub, err := uc.subscriptionRepo.GetActiveByUserID(userID)
	if err != nil {
		return errors.New("no active subscription found")
	}

	if !sub.CancelAtPeriodEnd {
		return errors.New("subscription is not scheduled for cancellation")
	}

	sub.CancelAtPeriodEnd = false
	sub.CanceledAt = nil
	sub.UpdatedAt = time.Now()

	return uc.subscriptionRepo.Update(sub)
}

// ChangeSubscription changes the user's subscription to a new plan
func (uc *subscriptionUseCase) ChangeSubscription(
	userID uuid.UUID,
	newPlanSlug string,
) (*domain.Subscription, error) {
	sub, err := uc.subscriptionRepo.GetActiveByUserID(userID)
	if err != nil {
		return nil, errors.New("no active subscription found")
	}

	newPlan, err := uc.subscriptionRepo.GetPlanBySlug(newPlanSlug)
	if err != nil {
		return nil, errors.New("plan not found")
	}

	if !newPlan.IsActive {
		return nil, errors.New("plan is not available")
	}

	// Update to new plan (takes effect immediately)
	sub.PlanID = newPlan.ID
	sub.UpdatedAt = time.Now()

	if err := uc.subscriptionRepo.Update(sub); err != nil {
		return nil, err
	}

	sub.Plan = newPlan
	return sub, nil
}

// HandleWebhook handles Stripe webhook events
func (uc *subscriptionUseCase) HandleWebhook(event string, data map[string]interface{}) error {
	switch event {
	case "customer.subscription.updated":
		// Handle subscription updates from Stripe
		return uc.handleSubscriptionUpdated(data)
	case "customer.subscription.deleted":
		// Handle subscription deletion
		return uc.handleSubscriptionDeleted(data)
	case "invoice.payment_failed":
		// Handle failed payment
		return uc.handlePaymentFailed(data)
	case "invoice.paid":
		// Handle successful payment
		return uc.handlePaymentSucceeded(data)
	default:
		return nil
	}
}

func (uc *subscriptionUseCase) handleSubscriptionUpdated(data map[string]interface{}) error {
	object, ok := data["object"].(map[string]interface{})
	if !ok {
		return nil
	}

	stripeSubID, _ := object["id"].(string)
	if stripeSubID == "" {
		return nil
	}

	// Find subscription by Stripe ID and update
	// In production, you'd look up by StripeSubscriptionID
	return nil
}

func (uc *subscriptionUseCase) handleSubscriptionDeleted(data map[string]interface{}) error {
	// Mark subscription as expired
	return nil
}

func (uc *subscriptionUseCase) handlePaymentFailed(data map[string]interface{}) error {
	// Mark subscription as past_due
	return nil
}

func (uc *subscriptionUseCase) handlePaymentSucceeded(data map[string]interface{}) error {
	// Extend subscription period
	return nil
}
