package middleware

import (
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
)

// RateLimiter configuration
type RateLimiterConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// Simple in-memory rate limiter (use Redis in production)
type RateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*clientLimiter
	config  RateLimiterConfig
	cleanup time.Duration
}

type clientLimiter struct {
	tokens    int
	lastCheck time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientLimiter),
		config:  cfg,
		cleanup: 5 * time.Minute,
	}

	// Periodic cleanup of old entries
	go func() {
		ticker := time.NewTicker(rl.cleanup)
		for range ticker.C {
			rl.cleanupOldEntries()
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanupOldEntries() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	threshold := time.Now().Add(-rl.cleanup)
	for key, client := range rl.clients {
		if client.lastCheck.Before(threshold) {
			delete(rl.clients, key)
		}
	}
}

func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	client, exists := rl.clients[clientIP]
	now := time.Now()

	if !exists {
		rl.clients[clientIP] = &clientLimiter{
			tokens:    rl.config.BurstSize - 1,
			lastCheck: now,
		}
		return true
	}

	// Calculate tokens to add based on time passed
	elapsed := now.Sub(client.lastCheck)
	tokensToAdd := int(elapsed.Minutes() * float64(rl.config.RequestsPerMinute))

	client.tokens += tokensToAdd
	if client.tokens > rl.config.BurstSize {
		client.tokens = rl.config.BurstSize
	}
	client.lastCheck = now

	if client.tokens > 0 {
		client.tokens--
		return true
	}

	return false
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientIP := c.RealIP()

			if !limiter.Allow(clientIP) {
				return response.Error(c, 429, "Too many requests. Please try again later.")
			}

			return next(c)
		}
	}
}

// DefaultRateLimiter creates a rate limiter with default settings
func DefaultRateLimiter() *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 60,
		BurstSize:         20,
	})
}

// StrictRateLimiter for sensitive endpoints (login, register)
func StrictRateLimiter() *RateLimiter {
	return NewRateLimiter(RateLimiterConfig{
		RequestsPerMinute: 10,
		BurstSize:         5,
	})
}
