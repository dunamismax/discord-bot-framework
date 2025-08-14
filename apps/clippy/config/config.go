// Package config handles application configuration loading from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration settings.
type Config struct {
	DiscordToken       string
	BotName            string
	LogLevel           string
	JSONLogging        bool
	DebugMode          bool
	ShutdownTimeout    time.Duration
	RequestTimeout     time.Duration
	MaxRetries         int
	RandomResponses    bool
	RandomMessageDelay time.Duration
	RandomInterval     time.Duration
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		BotName:            getEnv("BOT_NAME", "clippy-bot"),
		LogLevel:           "info", // default log level
		JSONLogging:        false,  // default to text logging
		DebugMode:          false,  // default debug mode
		ShutdownTimeout:    30 * time.Second,
		RequestTimeout:     30 * time.Second,
		MaxRetries:         3,
		RandomResponses:    true,
		RandomMessageDelay: 3 * time.Second,  // max delay for random responses
		RandomInterval:     45 * time.Minute, // average interval between random messages
	}

	// Discord token is required
	cfg.DiscordToken = os.Getenv("DISCORD_TOKEN")
	if cfg.DiscordToken == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN environment variable is required")
	}

	// Optional configurations
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = strings.ToLower(logLevel)
	}

	// Parse timeout configurations
	if timeout := os.Getenv("SHUTDOWN_TIMEOUT"); timeout != "" {
		if parsed, err := time.ParseDuration(timeout); err == nil {
			cfg.ShutdownTimeout = parsed
		}
	}

	if timeout := os.Getenv("REQUEST_TIMEOUT"); timeout != "" {
		if parsed, err := time.ParseDuration(timeout); err == nil {
			cfg.RequestTimeout = parsed
		}
	}

	// Parse other configurations
	cfg.MaxRetries = GetInt("MAX_RETRIES", cfg.MaxRetries)
	cfg.DebugMode = GetBool("DEBUG", cfg.DebugMode)
	cfg.JSONLogging = GetBool("JSON_LOGGING", cfg.JSONLogging)
	cfg.RandomResponses = GetBool("RANDOM_RESPONSES", cfg.RandomResponses)

	// Parse timing configurations
	if delay := os.Getenv("RANDOM_MESSAGE_DELAY"); delay != "" {
		if parsed, err := time.ParseDuration(delay); err == nil {
			cfg.RandomMessageDelay = parsed
		}
	}

	if interval := os.Getenv("RANDOM_INTERVAL"); interval != "" {
		if parsed, err := time.ParseDuration(interval); err == nil {
			cfg.RandomInterval = parsed
		}
	}

	return cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.DiscordToken == "" {
		return fmt.Errorf("discord token is required")
	}

	if c.BotName == "" {
		return fmt.Errorf("bot name cannot be empty")
	}

	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.LogLevel) {
		return fmt.Errorf("invalid log level: %s (valid: %s)", c.LogLevel, strings.Join(validLogLevels, ", "))
	}

	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive")
	}

	if c.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if c.RandomMessageDelay <= 0 {
		return fmt.Errorf("random message delay must be positive")
	}

	if c.RandomInterval <= 0 {
		return fmt.Errorf("random interval must be positive")
	}

	return nil
}

// GetBool returns a boolean environment variable with a default value.
func GetBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolVal
}

// GetInt returns an integer environment variable with a default value.
func GetInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intVal
}

// getEnv returns an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// contains checks if a slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}