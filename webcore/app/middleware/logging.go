package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/semanggilab/webcore-go/app/logger"
)

// RequestLogger creates a middleware for logging HTTP requests
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate processing time
		latency := time.Since(start)

		// Get response status
		status := c.Response().StatusCode()

		// Log the request
		logger.Debug("HTTP Request",
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"latency", latency,
			"ip", c.IP(),
			"user_agent", c.Get("User-Agent"),
		)

		return err
	}
}

// Middleware to remove trailing slash
func RemoveTrailingSlash() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.HasSuffix(path, "/") && path != "/" {
			return c.Redirect(strings.TrimSuffix(path, "/"), fiber.StatusMovedPermanently)
		}
		return c.Next()
	}
}

// RequestID creates a middleware for generating and tracking request IDs
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Generate or get existing request ID
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			// In a real implementation, you would generate a proper UUID
			requestID = "req-" + time.Now().Format("20060102150405")
		}

		// Set request ID in response header
		c.Set("X-Request-ID", requestID)

		// Store in context for use in other handlers/middleware
		c.Locals("request_id", requestID)

		return c.Next()
	}
}

// Metrics creates a middleware for collecting request metrics
func Metrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate metrics
		latency := time.Since(start)
		status := c.Response().StatusCode()

		// Store metrics in context for collection
		c.Locals("metrics", map[string]any{
			"latency":   latency,
			"status":    status,
			"timestamp": start,
		})

		return err
	}
}
