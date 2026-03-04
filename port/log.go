package port

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type IRemoteLog interface {
	Connector

	SetMinimumLevelLog(level slog.Level)
	SetMinimumLevelCapture(level slog.Level)
	SetDefaultTags(tags map[string]string)
	SetDefaultContexts(contexts map[string]map[string]any)
	NewHandler() fiber.Handler
	SetTag(key string, value string)
	SetContext(key string, context map[string]any)
	Log(level slog.Level, msg string, args ...any)
	CaptureMessage(msg string)
	CaptureError(err error)
}
