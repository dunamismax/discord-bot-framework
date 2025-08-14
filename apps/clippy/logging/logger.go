// Package logging provides structured logging functionality using slog.
package logging

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	clippyErrors "github.com/sawyer/go-discord-bots/apps/clippy/errors"
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

// WithUser returns a logger with user information.
func WithUser(userID, username string) *slog.Logger {
	return DefaultLogger.With("user_id", userID, "username", username)
}

// WithCommand returns a logger with command information.
func WithCommand(command string) *slog.Logger {
	return DefaultLogger.With("command", command)
}

// LogError logs a ClippyError with appropriate structured fields.
func LogError(logger *slog.Logger, err error, message string) {
	var clippyErr *clippyErrors.ClippyError
	if errors.As(err, &clippyErr) {
		attrs := []slog.Attr{
			slog.String("error_type", string(clippyErr.Type)),
			slog.String("error_message", clippyErr.Message),
		}

		if clippyErr.StatusCode != 0 {
			attrs = append(attrs, slog.Int("status_code", clippyErr.StatusCode))
		}

		if clippyErr.Cause != nil {
			attrs = append(attrs, slog.String("cause", clippyErr.Cause.Error()))
		}

		for key, value := range clippyErr.Context {
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

// LogStartup logs application startup information.
func LogStartup(botName, logLevel string, debugMode bool) {
	logger := WithComponent("startup")
	logger.Info("Starting Clippy Bot",
		"bot_name", botName,
		"log_level", logLevel,
		"debug_mode", debugMode,
	)
}

// LogShutdown logs application shutdown information.
func LogShutdown() {
	logger := WithComponent("shutdown")
	logger.Info("Clippy bot shutdown complete")
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
