package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/semanggilab/webcore-go/app/config"
	"github.com/semanggilab/webcore-go/app/helper"
	"github.com/semanggilab/webcore-go/app/out"
)

// SetupGlobalMiddleware sets up all global middleware
func SetupGlobalMiddleware(app *fiber.App, cfg *config.Config) {
	// Ignore favicon
	app.Use(favicon.New())

	// Recovery middleware
	if cfg.App.Features.Recovery {
		app.Use(recover.New(recover.Config{
			EnableStackTrace: true,
		}))
	}
	// Logger middleware
	app.Use(flogger.New(flogger.Config{
		Output:     helper.FiberLoggerOutput(cfg.App.Logging.Output),
		Format:     cfg.App.Logging.Format, //  "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat: "02-Jan-2006 15:04:05",
	}))

	corsConfig := cors.Config{
		AllowOrigins:     strings.Join(cfg.App.CORS.AllowOrigins, ","),
		AllowMethods:     strings.Join(cfg.App.CORS.AllowMethods, ","),
		AllowHeaders:     strings.Join(cfg.App.CORS.AllowHeaders, ","),
		ExposeHeaders:    strings.Join(cfg.App.CORS.ExposeHeaders, ","),
		AllowCredentials: cfg.App.CORS.AllowCredentials,
		MaxAge:           int(cfg.App.CORS.MaxAge.Seconds()),
	}
	app.Use(cors.New(corsConfig))

	// Remove Trailing Slash middleware
	app.Use(RemoveTrailingSlash())

	// Request ID middleware
	if cfg.App.Features.Tracing {
		app.Use(RequestID())
	}

	// Request metrics middleware
	if cfg.App.Features.Metrics {
		app.Use(Metrics())
	}

	// Custom request logger middleware
	if cfg.App.Features.Profiling {
		app.Use(RequestLogger())
	}

	// Security headers middleware
	if cfg.App.SecurityHeaders {
		app.Use(SecurityHeadersMiddleware())
	}

	if cfg.App.RateLimit.Enabled {
		app.Use(DefaultRateLimit(cfg.App.RateLimit))
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Add security headers
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Set("Content-Security-Policy", "default-src 'self'")

		return c.Next()
	}
}

// CompressMiddleware enables response compression
func CompressMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Fiber has built-in compression middleware
		// This is a placeholder for custom compression logic if needed
		return c.Next()
	}
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError

	// Retrieve the custom status code if it's a *fiber.Error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	// Send custom error page
	return c.Status(code).JSON(out.Error(code, 1, "UNKNOWN", err.Error()))
}
