// Package logging provides structured logging for the Discord bot framework.
package logging

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

var logger *slog.Logger

// InitializeLogger initializes the global logger with the specified level and format.
func InitializeLogger(level string, jsonFormat bool) {
	var logLevel slog.Level

	switch strings.ToUpper(level) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize timestamp format
			if a.Key == slog.TimeKey {
				return slog.String("time", a.Value.Time().Format(time.RFC3339))
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

	logger = slog.New(handler)
}

// WithComponent returns a logger with a component field added.
func WithComponent(component string) *slog.Logger {
	if logger == nil {
		InitializeLogger("INFO", false)
	}
	return logger.With("component", component)
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	if logger == nil {
		InitializeLogger("INFO", false)
	}
	logger.Debug(msg, args...)
}

// Info logs an info message.
func Info(msg string, args ...any) {
	if logger == nil {
		InitializeLogger("INFO", false)
	}
	logger.Info(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	if logger == nil {
		InitializeLogger("INFO", false)
	}
	logger.Warn(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	if logger == nil {
		InitializeLogger("INFO", false)
	}
	logger.Error(msg, args...)
}

// LogStartup logs startup information.
func LogStartup(botName, commandPrefix, logLevel string, debugMode bool) {
	logger := WithComponent("startup")
	logger.Info("Bot starting up",
		"bot_name", botName,
		"command_prefix", commandPrefix,
		"log_level", logLevel,
		"debug_mode", debugMode,
	)
}

// LogShutdown logs shutdown information.
func LogShutdown() {
	logger := WithComponent("shutdown")
	logger.Info("Bot shutdown complete")
}

// LogDiscordCommand logs Discord command execution.
func LogDiscordCommand(userID, username, command string, success bool) {
	logger := WithComponent("discord")
	logger.Info("Command executed",
		"user_id", userID,
		"username", username,
		"command", command,
		"success", success,
	)
}

// LogError logs an error with additional context.
func LogError(logger *slog.Logger, err error, msg string) {
	logger.Error(msg, "error", err)
}
