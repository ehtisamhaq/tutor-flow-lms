package push

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/repository"
)

// Config for push notifications
type Config struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string // mailto: or https:// URL
}

// Service handles web push notifications
type Service struct {
	cfg      Config
	pushRepo repository.PushSubscriptionRepository
}

// NewService creates a new push notification service
func NewService(cfg Config, pushRepo repository.PushSubscriptionRepository) *Service {
	return &Service{
		cfg:      cfg,
		pushRepo: pushRepo,
	}
}

// GetVAPIDPublicKey returns the public key for client registration
func (s *Service) GetVAPIDPublicKey() string {
	return s.cfg.VAPIDPublicKey
}

// Subscribe registers a push subscription
func (s *Service) Subscribe(ctx context.Context, userID uuid.UUID, sub SubscriptionInput) (*domain.PushSubscription, error) {
	// Check for existing subscription with same endpoint
	existing, _ := s.pushRepo.GetByEndpoint(ctx, sub.Endpoint)
	if existing != nil {
		// Update existing
		existing.UserID = userID
		existing.P256dh = sub.Keys.P256dh
		existing.Auth = sub.Keys.Auth
		existing.UserAgent = sub.UserAgent
		if err := s.pushRepo.Update(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Create new subscription
	pushSub := &domain.PushSubscription{
		UserID:    userID,
		Endpoint:  sub.Endpoint,
		P256dh:    sub.Keys.P256dh,
		Auth:      sub.Keys.Auth,
		UserAgent: sub.UserAgent,
	}

	if err := s.pushRepo.Create(ctx, pushSub); err != nil {
		return nil, err
	}

	return pushSub, nil
}

// SubscriptionInput from client
type SubscriptionInput struct {
	Endpoint  string           `json:"endpoint" validate:"required"`
	Keys      SubscriptionKeys `json:"keys" validate:"required"`
	UserAgent string           `json:"user_agent,omitempty"`
}

type SubscriptionKeys struct {
	P256dh string `json:"p256dh" validate:"required"`
	Auth   string `json:"auth" validate:"required"`
}

// Unsubscribe removes a push subscription
func (s *Service) Unsubscribe(ctx context.Context, userID uuid.UUID, endpoint string) error {
	sub, err := s.pushRepo.GetByEndpoint(ctx, endpoint)
	if err != nil {
		return err
	}
	if sub == nil {
		return nil
	}
	if sub.UserID != userID {
		return fmt.Errorf("unauthorized")
	}
	return s.pushRepo.Delete(ctx, sub.ID)
}

// UnsubscribeAll removes all subscriptions for a user
func (s *Service) UnsubscribeAll(ctx context.Context, userID uuid.UUID) error {
	return s.pushRepo.DeleteByUser(ctx, userID)
}

// GetUserSubscriptions returns all subscriptions for a user
func (s *Service) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]domain.PushSubscription, error) {
	return s.pushRepo.GetByUser(ctx, userID)
}

// SendToUser sends a push notification to all user's devices
func (s *Service) SendToUser(ctx context.Context, userID uuid.UUID, notification domain.PushNotification) error {
	subs, err := s.pushRepo.GetByUser(ctx, userID)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		go func(sub domain.PushSubscription) {
			if err := s.sendNotification(&sub, notification); err != nil {
				// Remove invalid subscription
				if isSubscriptionGone(err) {
					_ = s.pushRepo.Delete(context.Background(), sub.ID)
				}
			}
		}(sub)
	}

	return nil
}

// SendToUsers sends a push notification to multiple users
func (s *Service) SendToUsers(ctx context.Context, userIDs []uuid.UUID, notification domain.PushNotification) error {
	for _, userID := range userIDs {
		go func(uid uuid.UUID) {
			_ = s.SendToUser(ctx, uid, notification)
		}(userID)
	}
	return nil
}

// sendNotification sends a push notification to a single subscription
func (s *Service) sendNotification(sub *domain.PushSubscription, notification domain.PushNotification) error {
	payload, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	// Create the push message (simplified - in production use a proper web-push library)
	req, err := http.NewRequest("POST", sub.Endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("TTL", "86400") // 24 hours

	// In production, add VAPID authentication headers here
	// This is a simplified implementation

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("push failed: status %d", resp.StatusCode)
	}

	return nil
}

// isSubscriptionGone checks if the error indicates the subscription is no longer valid
func isSubscriptionGone(err error) bool {
	// In production, check for 404/410 status codes
	return false
}

// GenerateVAPIDKeys generates a new VAPID key pair (utility function)
func GenerateVAPIDKeys() (publicKey, privateKey string, err error) {
	curve := elliptic.P256()
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return "", "", err
	}

	// Encode public key
	pubBytes := elliptic.Marshal(curve, key.PublicKey.X, key.PublicKey.Y)
	publicKey = base64.RawURLEncoding.EncodeToString(pubBytes)

	// Encode private key
	privBytes := key.D.Bytes()
	privateKey = base64.RawURLEncoding.EncodeToString(privBytes)

	return publicKey, privateKey, nil
}

// --- Pre-built Notification Methods ---

// NotifyNewMessage sends a notification for a new message
func (s *Service) NotifyNewMessage(ctx context.Context, userID uuid.UUID, senderName, preview string) error {
	return s.SendToUser(ctx, userID, domain.PushNotification{
		Title: fmt.Sprintf("New message from %s", senderName),
		Body:  preview,
		Icon:  "/icons/message.png",
		Tag:   "message",
		URL:   "/messages",
	})
}

// NotifyNewEnrollment sends a notification for a new enrollment
func (s *Service) NotifyNewEnrollment(ctx context.Context, userID uuid.UUID, courseName string) error {
	return s.SendToUser(ctx, userID, domain.PushNotification{
		Title: "Enrollment Confirmed!",
		Body:  fmt.Sprintf("You're now enrolled in %s", courseName),
		Icon:  "/icons/course.png",
		Tag:   "enrollment",
		URL:   "/my-courses",
	})
}

// NotifyNewGrade sends a notification for a new grade
func (s *Service) NotifyNewGrade(ctx context.Context, userID uuid.UUID, itemTitle string, score float64) error {
	return s.SendToUser(ctx, userID, domain.PushNotification{
		Title: "Grade Posted",
		Body:  fmt.Sprintf("%s: %.1f points", itemTitle, score),
		Icon:  "/icons/grade.png",
		Tag:   "grade",
		URL:   "/grades",
	})
}

// NotifyAnnouncement sends a notification for a new announcement
func (s *Service) NotifyAnnouncement(ctx context.Context, userID uuid.UUID, title, courseName string) error {
	return s.SendToUser(ctx, userID, domain.PushNotification{
		Title: "New Announcement",
		Body:  fmt.Sprintf("%s: %s", courseName, title),
		Icon:  "/icons/announcement.png",
		Tag:   "announcement",
		URL:   "/announcements",
	})
}

// NotifyCertificate sends a notification for a new certificate
func (s *Service) NotifyCertificate(ctx context.Context, userID uuid.UUID, courseName string) error {
	return s.SendToUser(ctx, userID, domain.PushNotification{
		Title: "🎉 Certificate Earned!",
		Body:  fmt.Sprintf("Congratulations! You completed %s", courseName),
		Icon:  "/icons/certificate.png",
		Tag:   "certificate",
		URL:   "/certificates",
	})
}
