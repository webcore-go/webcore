package middleware

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/infra/config"
)

// RateLimitConfig represents the configuration for rate limiting middleware
type RateLimitConfig struct {
	Window time.Duration // Time window (e.g., 1 minute)
	Limit  int64         // Maximum requests per window
}

// RateLimiter represents a rate limiter implementation
type RateLimiter struct {
	config  RateLimitConfig
	clients map[string]*clientData
	mu      sync.RWMutex
}

type clientData struct {
	count       int64
	windowStart time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		config:  config,
		clients: make(map[string]*clientData),
	}
}

// Middleware creates a rate limiting middleware
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client identifier (IP address or API key)
		var apiKey string
		clientID := c.IP()
		if apiKey = c.Get("X-API-Key"); apiKey != "" {
			clientID = apiKey
		} else if apiKey = c.Get("Authorization"); apiKey != "" {
			if strings.HasPrefix(apiKey, "Bearer ") {
				clientID = strings.TrimPrefix(apiKey, "Bearer ")
			}
		}

		// Check rate limit
		allowed, resetTime, err := rl.Allow(clientID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		if !allowed {
			c.Set("X-RateLimit-Limit", strconv.FormatInt(rl.config.Limit, 10))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
		}

		// Set rate limit headers
		// Get fresh count to avoid race conditions
		rl.mu.RLock()
		currentCount := rl.clients[clientID].count
		rl.mu.RUnlock()

		remaining := rl.config.Limit - currentCount
		if remaining < 0 {
			remaining = 0
		}
		c.Set("X-RateLimit-Limit", strconv.FormatInt(rl.config.Limit, 10))
		c.Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Set("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

		return c.Next()
	}
}

// Allow checks if a client is allowed to make a request
func (rl *RateLimiter) Allow(clientID string) (bool, time.Time, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Handle special case: limit is 0 (no requests allowed)
	if rl.config.Limit == 0 {
		return false, now.Add(rl.config.Window), nil
	}

	data, exists := rl.clients[clientID]

	// If client doesn't exist or window has expired, create new entry
	if !exists || now.Sub(data.windowStart) > rl.config.Window {
		rl.clients[clientID] = &clientData{
			count:       1,
			windowStart: now,
		}
		return true, now.Add(rl.config.Window), nil
	}

	// Check if limit exceeded
	if data.count >= rl.config.Limit {
		return false, data.windowStart.Add(rl.config.Window), nil
	}

	// Increment count
	data.count++
	return true, data.windowStart.Add(rl.config.Window), nil
}

// getClientCount gets the current request count for a client
// Note: This method is kept for potential future use or external access,
// but the main middleware now gets the count directly to avoid race conditions
func (rl *RateLimiter) getClientCount(clientID string) int64 {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if clientData, exists := rl.clients[clientID]; exists {
		return clientData.count
	}
	return 0
}

// Reset resets the rate limiter (useful for testing)
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.clients = make(map[string]*clientData)
}

// Cleanup removes expired client entries to prevent memory leaks
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for clientID, data := range rl.clients {
		if now.Sub(data.windowStart) > rl.config.Window {
			delete(rl.clients, clientID)
		}
	}
}

// NewRateLimit creates a rate limiting middleware with the given configuration
func NewRateLimit(config RateLimitConfig) fiber.Handler {
	limiter := NewRateLimiter(config)
	return limiter.Middleware()
}

// DefaultRateLimit creates a default rate limiting middleware
func DefaultRateLimit(config config.RateLimitConfig) fiber.Handler {
	return NewRateLimit(RateLimitConfig{
		Window: time.Minute,
		Limit:  int64(config.Max), // 60 requests per minute
	})
}
