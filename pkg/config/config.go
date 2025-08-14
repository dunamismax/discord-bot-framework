// Package config provides unified configuration management for all Discord bots.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// BotType represents the type of bot being configured.
type BotType string

const (
	// BotTypeClipper represents the Clippy bot.
	BotTypeClipper BotType = "clippy"
	// BotTypeMusic represents the Music bot.
	BotTypeMusic BotType = "music"
	// BotTypeMTG represents the MTG Card bot.
	BotTypeMTG BotType = "mtg"
)

// Config represents the unified configuration structure for all bots.
type Config struct {
	// Bot identification
	BotType       BotType `json:"bot_type"`
	BotName       string  `json:"bot_name"`
	DiscordToken  string  `json:"discord_token"`
	CommandPrefix string  `json:"command_prefix"`

	// Server configuration
	GuildID string `json:"guild_id,omitempty"`

	// Behavior settings
	DebugMode       bool          `json:"debug_mode"`
	LogLevel        string        `json:"log_level"`
	JSONLogging     bool          `json:"json_logging"`
	CommandCooldown time.Duration `json:"command_cooldown"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	RequestTimeout  time.Duration `json:"request_timeout"`
	MaxRetries      int           `json:"max_retries"`

	// Feature flags
	RandomResponses    bool          `json:"random_responses,omitempty"`
	RandomInterval     time.Duration `json:"random_interval,omitempty"`
	RandomMessageDelay time.Duration `json:"random_message_delay,omitempty"`

	// Cache settings
	CacheTTL  time.Duration `json:"cache_ttl,omitempty"`
	CacheSize int           `json:"cache_size,omitempty"`

	// Music bot specific settings
	MaxQueueSize      int           `json:"max_queue_size,omitempty"`
	InactivityTimeout time.Duration `json:"inactivity_timeout,omitempty"`
	VolumeLevel       float64       `json:"volume_level,omitempty"`

	// Database settings
	DatabaseURL string `json:"database_url,omitempty"`
}

// Load loads configuration for a specific bot type.
func Load(botType BotType) (*Config, error) {
	cfg := getDefaultConfig(botType)

	// Load from environment variables
	if err := cfg.loadFromEnvironment(); err != nil {
		return nil, fmt.Errorf("failed to load environment configuration: %w", err)
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a JSON file and environment variables.
func LoadFromFile(configPath string, botType BotType) (*Config, error) {
	cfg := getDefaultConfig(botType)

	// Load from config file if provided and exists
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}

			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Override with environment variables
	if err := cfg.loadFromEnvironment(); err != nil {
		return nil, fmt.Errorf("failed to load environment configuration: %w", err)
	}

	return cfg, nil
}

// getDefaultConfig returns default configuration for a bot type.
func getDefaultConfig(botType BotType) *Config {
	base := &Config{
		BotType:         botType,
		CommandPrefix:   "!",
		LogLevel:        "info",
		JSONLogging:     false,
		DebugMode:       false,
		CommandCooldown: 3 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		RequestTimeout:  30 * time.Second,
		MaxRetries:      3,
	}

	switch botType {
	case BotTypeClipper:
		base.BotName = "Clippy Bot"
		base.CommandCooldown = 5 * time.Second
		base.RandomResponses = true
		base.RandomInterval = 45 * time.Minute
		base.RandomMessageDelay = 3 * time.Second

	case BotTypeMusic:
		base.BotName = "Music Bot"
		base.CommandCooldown = 3 * time.Second
		base.MaxQueueSize = 100
		base.InactivityTimeout = 5 * time.Minute
		base.VolumeLevel = 0.5
		base.DatabaseURL = "music.db"

	case BotTypeMTG:
		base.BotName = "MTG Card Bot"
		base.CommandCooldown = 2 * time.Second
		base.CacheTTL = 1 * time.Hour
		base.CacheSize = 1000

	default:
		base.BotName = "Discord Bot"
	}

	return base
}

// loadFromEnvironment loads configuration from environment variables.
func (c *Config) loadFromEnvironment() error {
	// Common environment variables
	if token := os.Getenv("DISCORD_TOKEN"); token != "" {
		c.DiscordToken = token
	}
	if prefix := os.Getenv("COMMAND_PREFIX"); prefix != "" {
		c.CommandPrefix = prefix
	}
	if guildID := os.Getenv("GUILD_ID"); guildID != "" {
		c.GuildID = guildID
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		c.LogLevel = strings.ToLower(logLevel)
	}

	// Debug mode
	c.DebugMode = GetBool("DEBUG", c.DebugMode)
	c.JSONLogging = GetBool("JSON_LOGGING", c.JSONLogging)

	// Timeouts
	if timeout := os.Getenv("SHUTDOWN_TIMEOUT"); timeout != "" {
		if parsed, err := time.ParseDuration(timeout); err == nil {
			c.ShutdownTimeout = parsed
		}
	}
	if timeout := os.Getenv("REQUEST_TIMEOUT"); timeout != "" {
		if parsed, err := time.ParseDuration(timeout); err == nil {
			c.RequestTimeout = parsed
		}
	}

	// Retry configuration
	c.MaxRetries = GetInt("MAX_RETRIES", c.MaxRetries)

	// Bot-specific environment variables
	switch c.BotType {
	case BotTypeClipper:
		if token := os.Getenv("CLIPPY_DISCORD_TOKEN"); token != "" {
			c.DiscordToken = token
		}
		if guildID := os.Getenv("CLIPPY_GUILD_ID"); guildID != "" {
			c.GuildID = guildID
		}
		c.RandomResponses = GetBool("RANDOM_RESPONSES", c.RandomResponses)
		if interval := os.Getenv("RANDOM_INTERVAL"); interval != "" {
			if parsed, err := time.ParseDuration(interval); err == nil {
				c.RandomInterval = parsed
			}
		}
		if delay := os.Getenv("RANDOM_MESSAGE_DELAY"); delay != "" {
			if parsed, err := time.ParseDuration(delay); err == nil {
				c.RandomMessageDelay = parsed
			}
		}

	case BotTypeMusic:
		if token := os.Getenv("MUSIC_DISCORD_TOKEN"); token != "" {
			c.DiscordToken = token
		}
		if guildID := os.Getenv("MUSIC_GUILD_ID"); guildID != "" {
			c.GuildID = guildID
		}
		if dbURL := os.Getenv("MUSIC_DATABASE_URL"); dbURL != "" {
			c.DatabaseURL = dbURL
		}
		c.MaxQueueSize = GetInt("MAX_QUEUE_SIZE", c.MaxQueueSize)
		if timeout := os.Getenv("INACTIVITY_TIMEOUT"); timeout != "" {
			if parsed, err := time.ParseDuration(timeout); err == nil {
				c.InactivityTimeout = parsed
			}
		}
		if volume := os.Getenv("VOLUME_LEVEL"); volume != "" {
			if parsed, err := strconv.ParseFloat(volume, 64); err == nil && parsed >= 0 && parsed <= 1 {
				c.VolumeLevel = parsed
			}
		}

	case BotTypeMTG:
		// MTG bot uses DISCORD_TOKEN by default, already handled above
		if ttl := os.Getenv("CACHE_TTL"); ttl != "" {
			if parsed, err := time.ParseDuration(ttl); err == nil {
				c.CacheTTL = parsed
			}
		}
		c.CacheSize = GetInt("CACHE_SIZE", c.CacheSize)
	}

	return nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.DiscordToken == "" {
		return fmt.Errorf("%s bot: discord_token is required", c.BotType)
	}

	if c.BotName == "" {
		return fmt.Errorf("%s bot: bot_name is required", c.BotType)
	}

	if c.CommandPrefix == "" {
		return fmt.Errorf("%s bot: command_prefix is required", c.BotType)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		return fmt.Errorf("%s bot: invalid log_level '%s', must be one of: debug, info, warn, error", c.BotType, c.LogLevel)
	}

	// Validate timeouts
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("%s bot: shutdown_timeout must be positive", c.BotType)
	}
	if c.RequestTimeout <= 0 {
		return fmt.Errorf("%s bot: request_timeout must be positive", c.BotType)
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("%s bot: max_retries cannot be negative", c.BotType)
	}

	// Bot-specific validation
	switch c.BotType {
	case BotTypeMusic:
		if c.MaxQueueSize <= 0 {
			c.MaxQueueSize = 100
		}
		if c.InactivityTimeout <= 0 {
			c.InactivityTimeout = 5 * time.Minute
		}
		if c.VolumeLevel <= 0 || c.VolumeLevel > 1 {
			c.VolumeLevel = 0.5
		}

	case BotTypeMTG:
		if c.CacheTTL <= 0 {
			return fmt.Errorf("MTG bot: cache_ttl must be positive")
		}
		if c.CacheSize <= 0 {
			return fmt.Errorf("MTG bot: cache_size must be positive")
		}
	}

	return nil
}

// Save saves the configuration to a JSON file.
func (c *Config) Save(configPath string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
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

// GetString returns a string environment variable with a default value.
func GetString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDuration returns a duration environment variable with a default value.
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}

	return duration
}
