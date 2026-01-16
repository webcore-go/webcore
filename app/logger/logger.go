package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"

	"github.com/webcore-go/webcore/app/helper"
)

var defaultLogger atomic.Pointer[Logger]

// Logger represents shared logger
type Logger struct {
	context context.Context
	logger  *slog.Logger
}

func PrepareLogger(ctx context.Context, level string) *Logger {
	if defaultLogger.Load() == nil {
		var logLevel slog.Level

		// Set log level based on input
		switch strings.ToLower(level) {
		case "debug":
			logLevel = slog.LevelDebug
		case "info":
			logLevel = slog.LevelInfo
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		default:
			logLevel = slog.LevelInfo
		}

		slog.SetLogLoggerLevel(logLevel)
		logger := slog.Default()
		// handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		// 	Level: logLevel,
		// 	ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		// 		if a.Key == slog.TimeKey {
		// 			return slog.Attr{}
		// 		}
		// 		return a
		// 	},
		// })
		// logger := slog.New(handler)

		log := &Logger{
			context: ctx,
			logger:  logger,
		}
		defaultLogger.Store(log)

		return log
	}
	return defaultLogger.Load()
}

// Default returns the default [Logger].
func logDefault() *Logger { return defaultLogger.Load() }

// Log logs a message with the given level
func (l *Logger) Log(level slog.Level, msg string, args ...any) {
	l.logger.Log(l.context, level, msg, args...)
}

// Log logs a message with the given level
func (l *Logger) LogJson(level slog.Level, msg string, obj any) {
	l.logger.Log(l.context, level, msg+helper.ToLogJSON(obj))
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...any) {
	l.Log(slog.LevelDebug, msg, args...)
}

// Debug logs a debug message
func (l *Logger) DebugJson(msg string, obj any) {
	l.LogJson(slog.LevelDebug, msg, obj)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...any) {
	l.Log(slog.LevelInfo, msg, args...)
}

// Info logs an info message
func (l *Logger) InfoJson(msg string, obj any) {
	l.LogJson(slog.LevelInfo, msg, obj)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...any) {
	l.Log(slog.LevelWarn, msg, args...)
}

// Warn logs a warning message
func (l *Logger) WarnJson(msg string, obj any) {
	l.LogJson(slog.LevelWarn, msg, obj)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...any) {
	l.Log(slog.LevelError, msg, args...)
}

// Error logs a error message
func (l *Logger) ErrorJson(msg string, obj any) {
	l.LogJson(slog.LevelError, msg, obj)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...any) {
	l.Log(slog.LevelError, msg, args...)
	os.Exit(1)
}

// Fatal logs a fatal message and exits
func (l *Logger) FatalJson(msg string, obj any) {
	l.LogJson(slog.LevelError, msg, obj)
	os.Exit(1)
}

// Log logs a message with the given level
func Log(level slog.Level, msg string, args ...any) {
	logDefault().Log(level, msg, args...)
}

// Log logs a message with the given level
func LogJson(level slog.Level, msg string, obj any) {
	logDefault().LogJson(level, msg, obj)
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	logDefault().Debug(msg, args...)
}

// Debug logs a debug message
func DebugJson(msg string, obj any) {
	logDefault().DebugJson(msg, obj)
}

// Info logs an info message
func Info(msg string, args ...any) {
	logDefault().Info(msg, args...)
}

// Info logs a info message
func InfoJson(msg string, obj any) {
	logDefault().InfoJson(msg, obj)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	logDefault().Warn(msg, args...)
}

// Warn logs a warning message
func WarnJson(msg string, obj any) {
	logDefault().WarnJson(msg, obj)
}

// Error logs an error message
func Error(msg string, args ...any) {
	logDefault().Error(msg, args...)
}

// Error logs a error message
func ErrorJson(msg string, obj any) {
	logDefault().ErrorJson(msg, obj)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, args ...any) {
	logDefault().Log(slog.LevelError, msg, args...)
	os.Exit(1)
}

// Fatal logs a fatal message and exits
func FatalJson(msg string, obj any) {
	logDefault().LogJson(slog.LevelError, msg, obj)
	os.Exit(1)
}
