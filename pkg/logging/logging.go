// Package logging provides unified structured logging functionality for all Discord bots.
package logging

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"
)

// DefaultLogger is the global logger instance.
var DefaultLogger *slog.Logger

// LogLevel represents the logging level.
type LogLevel string

const (
	// LevelDebug represents the debug logging level.
	LevelDebug LogLevel = "debug"
	// LevelInfo represents the info logging level.
	LevelInfo LogLevel = "info"
	// LevelWarn represents the warning logging level.
	LevelWarn LogLevel = "warn"
	// LevelError represents the error logging level.
	LevelError LogLevel = "error"
)

// BotError represents a bot-specific error with categorization.
type BotError interface {
	error
	Type() string
	Context() map[string]interface{}
}

// InitializeLogger initializes the global logger with the specified level and format.
func InitializeLogger(level string, jsonFormat bool) {
	var slogLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02T15:04:05.000Z07:00"))
			}
			return a
		},
	}

	var handler slog.Handler
	if jsonFormat {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	DefaultLogger = slog.New(handler)
	slog.SetDefault(DefaultLogger)
}

// WithContext returns a logger with context values.
func WithContext(_ context.Context) *slog.Logger {
	return DefaultLogger.With()
}

// WithComponent returns a logger with a component field.
func WithComponent(component string) *slog.Logger {
	return DefaultLogger.With("component", component)
}

// WithBot returns a logger with bot-specific information.
func WithBot(botName, botType string) *slog.Logger {
	return DefaultLogger.With("bot_name", botName, "bot_type", botType)
}

// WithUser returns a logger with user information.
func WithUser(userID, username string) *slog.Logger {
	return DefaultLogger.With("user_id", userID, "username", username)
}

// WithCommand returns a logger with command information.
func WithCommand(command string) *slog.Logger {
	return DefaultLogger.With("command", command)
}

// WithDuration returns a logger with duration information.
func WithDuration(operation string, duration time.Duration) *slog.Logger {
	return DefaultLogger.With("operation", operation, "duration_ms", duration.Milliseconds())
}

// LogError logs a BotError with appropriate structured fields.
func LogError(logger *slog.Logger, err error, message string) {
	var botErr BotError
	if errors.As(err, &botErr) {
		attrs := []slog.Attr{
			slog.String("error_type", botErr.Type()),
			slog.String("error_message", err.Error()),
		}

		for key, value := range botErr.Context() {
			attrs = append(attrs, slog.Any(key, value))
		}

		logger.LogAttrs(context.Background(), slog.LevelError, message, attrs...)
	} else {
		logger.Error(message, "error", err)
	}
}

// Debug logs a debug message with optional attributes.
func Debug(msg string, args ...any) {
	DefaultLogger.Debug(msg, args...)
}

// Info logs an info message with optional attributes.
func Info(msg string, args ...any) {
	DefaultLogger.Info(msg, args...)
}

// Warn logs a warning message with optional attributes.
func Warn(msg string, args ...any) {
	DefaultLogger.Warn(msg, args...)
}

// Error logs an error message with optional attributes.
func Error(msg string, args ...any) {
	DefaultLogger.Error(msg, args...)
}

// DebugWithContext logs a debug message with context.
func DebugWithContext(ctx context.Context, msg string, args ...any) {
	DefaultLogger.DebugContext(ctx, msg, args...)
}

// InfoWithContext logs an info message with context.
func InfoWithContext(ctx context.Context, msg string, args ...any) {
	DefaultLogger.InfoContext(ctx, msg, args...)
}

// WarnWithContext logs a warning message with context.
func WarnWithContext(ctx context.Context, msg string, args ...any) {
	DefaultLogger.WarnContext(ctx, msg, args...)
}

// ErrorWithContext logs an error message with context.
func ErrorWithContext(ctx context.Context, msg string, args ...any) {
	DefaultLogger.ErrorContext(ctx, msg, args...)
}

// LogStartup logs application startup information.
func LogStartup(botName, botType, prefix, logLevel string, debugMode bool) {
	logger := WithBot(botName, botType)
	logger.Info("Starting Discord bot",
		"command_prefix", prefix,
		"log_level", logLevel,
		"debug_mode", debugMode,
		"timestamp", time.Now().Format(time.RFC3339),
	)
}

// LogShutdown logs application shutdown information.
func LogShutdown(botName, botType string) {
	logger := WithBot(botName, botType)
	logger.Info("Bot shutdown complete", "timestamp", time.Now().Format(time.RFC3339))
}

// LogAPIRequest logs API request information.
func LogAPIRequest(service, endpoint string, duration time.Duration, success bool) {
	logger := WithComponent(service)
	logger.Debug("API request completed",
		"endpoint", endpoint,
		"duration_ms", duration.Milliseconds(),
		"success", success,
	)
}

// LogDiscordCommand logs Discord command execution.
func LogDiscordCommand(userID, username, command string, success bool) {
	logger := WithComponent("discord").With(
		"user_id", userID,
		"username", username,
		"command", command,
		"success", success,
	)

	if success {
		logger.Info("Command executed successfully")
	} else {
		logger.Warn("Command execution failed")
	}
}

// LogCacheOperation logs cache operations with performance metrics.
func LogCacheOperation(operation, key string, hit bool, duration time.Duration) {
	logger := WithComponent("cache")
	logger.Debug("Cache operation",
		"operation", operation,
		"key", key,
		"hit", hit,
		"duration_ns", duration.Nanoseconds(),
		"duration_ms", duration.Milliseconds(),
	)
}

// LogPerformanceMetric logs performance-related metrics.
func LogPerformanceMetric(component, metric string, value interface{}, unit string) {
	logger := WithComponent("metrics")
	logger.Info("Performance metric",
		"component", component,
		"metric", metric,
		"value", value,
		"unit", unit,
		"timestamp", time.Now().UnixNano(),
	)
}

// LogSecurityEvent logs security-related events.
func LogSecurityEvent(event, userID, reason string, severity string) {
	logger := WithComponent("security")

	level := slog.LevelInfo
	switch strings.ToLower(severity) {
	case "critical", "high":
		level = slog.LevelError
	case "medium", "warn":
		level = slog.LevelWarn
	case "low", "info":
		level = slog.LevelInfo
	}

	logger.Log(context.Background(), level, "Security event",
		"event", event,
		"user_id", userID,
		"reason", reason,
		"severity", severity,
		"timestamp", time.Now().Format(time.RFC3339),
	)
}
