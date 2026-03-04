package port

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type IRemoteLog interface {
	Connector

	SetLevel(level slog.Level)
	NewHandler() fiber.Handler
	SetTag(key string, value string)
	SetContext(key string, context map[string]any)
	Log(level slog.Level, msg string, args ...any)
}
