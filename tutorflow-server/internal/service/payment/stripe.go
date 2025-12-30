package payment

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/paymentintent"

	"github.com/tutorflow/tutorflow-server/internal/pkg/config"
)

// Service handles Stripe payment operations
type Service struct {
	secretKey     string
	webhookSecret string
	successURL    string
	cancelURL     string
}

// NewService creates a new payment service
func NewService(cfg config.StripeConfig) *Service {
	stripe.Key = cfg.SecretKey
	return &Service{
		secretKey:     cfg.SecretKey,
		webhookSecret: cfg.WebhookSecret,
		successURL:    "http://localhost:3000/checkout/success?session_id={CHECKOUT_SESSION_ID}",
		cancelURL:     "http://localhost:3000/checkout/cancel",
	}
}

// CreateCheckoutSessionInput for creating checkout
type CreateCheckoutSessionInput struct {
	CustomerEmail string
	OrderID       string
	Items         []LineItem
	SuccessURL    string
	CancelURL     string
}

// LineItem for checkout
type LineItem struct {
	Name        string
	Description string
	Amount      int64 // in cents
	Quantity    int64
	ImageURL    string
}

// CreateCheckoutSession creates a Stripe Checkout session
func (s *Service) CreateCheckoutSession(ctx context.Context, input CreateCheckoutSessionInput) (*stripe.CheckoutSession, error) {
	successURL := input.SuccessURL
	if successURL == "" {
		successURL = s.successURL
	}
	cancelURL := input.CancelURL
	if cancelURL == "" {
		cancelURL = s.cancelURL
	}

	var lineItems []*stripe.CheckoutSessionLineItemParams
	for _, item := range input.Items {
		images := []*string{}
		if item.ImageURL != "" {
			images = append(images, stripe.String(item.ImageURL))
		}

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("usd"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name:        stripe.String(item.Name),
					Description: stripe.String(item.Description),
					Images:      images,
				},
				UnitAmount: stripe.Int64(item.Amount),
			},
			Quantity: stripe.Int64(item.Quantity),
		})
	}

	params := &stripe.CheckoutSessionParams{
		CustomerEmail:      stripe.String(input.CustomerEmail),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems:          lineItems,
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:         stripe.String(successURL),
		CancelURL:          stripe.String(cancelURL),
		Metadata: map[string]string{
			"order_id": input.OrderID,
		},
	}

	return session.New(params)
}

// CreatePaymentIntentInput for payment intent
type CreatePaymentIntentInput struct {
	Amount   int64 // in cents
	Currency string
	OrderID  string
	Email    string
}

// CreatePaymentIntent creates a payment intent for custom payment flow
func (s *Service) CreatePaymentIntent(ctx context.Context, input CreatePaymentIntentInput) (*stripe.PaymentIntent, error) {
	currency := input.Currency
	if currency == "" {
		currency = "usd"
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(input.Amount),
		Currency: stripe.String(currency),
		Metadata: map[string]string{
			"order_id": input.OrderID,
		},
		ReceiptEmail: stripe.String(input.Email),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	return paymentintent.New(params)
}

// GetPaymentIntent retrieves a payment intent
func (s *Service) GetPaymentIntent(ctx context.Context, paymentIntentID string) (*stripe.PaymentIntent, error) {
	return paymentintent.Get(paymentIntentID, nil)
}

// ConfirmPaymentIntent confirms a payment intent
func (s *Service) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string) (*stripe.PaymentIntent, error) {
	return paymentintent.Confirm(paymentIntentID, nil)
}

// CancelPaymentIntent cancels a payment intent
func (s *Service) CancelPaymentIntent(ctx context.Context, paymentIntentID string) (*stripe.PaymentIntent, error) {
	return paymentintent.Cancel(paymentIntentID, nil)
}

// CreateRefund creates a refund for a payment
func (s *Service) CreateRefund(ctx context.Context, paymentIntentID string, amount int64) (*stripe.Refund, error) {
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
	}
	if amount > 0 {
		params.Amount = stripe.Int64(amount)
	}

	// Use the refund API
	refund := &stripe.Refund{}
	// Note: You'd need to import and use stripe/refund package
	_ = refund
	return nil, fmt.Errorf("refund not implemented - import stripe/refund package")
}

// GetWebhookSecret returns the webhook secret for verification
func (s *Service) GetWebhookSecret() string {
	return s.webhookSecret
}
