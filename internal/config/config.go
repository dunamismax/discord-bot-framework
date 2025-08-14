// Package config provides configuration loading and validation for the Discord bot framework.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config represents the main configuration for the bot framework.
type Config struct {
	Clippy *BotConfig `json:"clippy"`
	Music  *BotConfig `json:"music"`
}

// BotConfig represents configuration for a specific bot instance.
type BotConfig struct {
	// Bot identification
	BotName       string `json:"bot_name"`
	DiscordToken  string `json:"discord_token"`
	CommandPrefix string `json:"command_prefix"`

	// Server configuration
	GuildID string `json:"guild_id,omitempty"`

	// Behavior settings
	DebugMode       bool          `json:"debug_mode"`
	LogLevel        string        `json:"log_level"`
	JSONLogging     bool          `json:"json_logging"`
	CommandCooldown time.Duration `json:"command_cooldown"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	// Feature flags
	RandomResponses bool `json:"random_responses"`

	// Music bot specific settings
	MaxQueueSize      int           `json:"max_queue_size,omitempty"`
	InactivityTimeout time.Duration `json:"inactivity_timeout,omitempty"`
	VolumeLevel       float64       `json:"volume_level,omitempty"`

	// Database settings (for music bot)
	DatabaseURL string `json:"database_url,omitempty"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Clippy: &BotConfig{
			BotName:         "Clippy Bot",
			CommandPrefix:   "!",
			LogLevel:        "INFO",
			JSONLogging:     false,
			DebugMode:       false,
			CommandCooldown: 5 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			RandomResponses: true,
		},
		Music: &BotConfig{
			BotName:           "Music Bot",
			CommandPrefix:     "!",
			LogLevel:          "INFO",
			JSONLogging:       false,
			DebugMode:         false,
			CommandCooldown:   3 * time.Second,
			ShutdownTimeout:   10 * time.Second,
			MaxQueueSize:      100,
			InactivityTimeout: 5 * time.Minute,
			VolumeLevel:       0.5,
		},
	}
}

// Load loads configuration from a JSON file.
func Load(configPath string) (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil // Return defaults if no config file
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if set
	cfg.loadFromEnvironment()

	return cfg, nil
}

// loadFromEnvironment loads configuration from environment variables.
func (c *Config) loadFromEnvironment() {
	// Clippy bot environment variables
	if token := os.Getenv("CLIPPY_DISCORD_TOKEN"); token != "" {
		c.Clippy.DiscordToken = token
	}
	if guildID := os.Getenv("CLIPPY_GUILD_ID"); guildID != "" {
		c.Clippy.GuildID = guildID
	}
	if prefix := os.Getenv("CLIPPY_COMMAND_PREFIX"); prefix != "" {
		c.Clippy.CommandPrefix = prefix
	}
	if os.Getenv("CLIPPY_DEBUG") == "true" {
		c.Clippy.DebugMode = true
		c.Clippy.LogLevel = "DEBUG"
	}

	// Music bot environment variables
	if token := os.Getenv("MUSIC_DISCORD_TOKEN"); token != "" {
		c.Music.DiscordToken = token
	}
	if guildID := os.Getenv("MUSIC_GUILD_ID"); guildID != "" {
		c.Music.GuildID = guildID
	}
	if prefix := os.Getenv("MUSIC_COMMAND_PREFIX"); prefix != "" {
		c.Music.CommandPrefix = prefix
	}
	if os.Getenv("MUSIC_DEBUG") == "true" {
		c.Music.DebugMode = true
		c.Music.LogLevel = "DEBUG"
	}
	if dbURL := os.Getenv("MUSIC_DATABASE_URL"); dbURL != "" {
		c.Music.DatabaseURL = dbURL
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Clippy != nil {
		if err := c.Clippy.validate("Clippy"); err != nil {
			return err
		}
	}

	if c.Music != nil {
		if err := c.Music.validate("Music"); err != nil {
			return err
		}
	}

	return nil
}

// validate validates a bot configuration.
func (bc *BotConfig) validate(botName string) error {
	if bc.DiscordToken == "" {
		return fmt.Errorf("%s bot: discord_token is required", botName)
	}

	if bc.BotName == "" {
		return fmt.Errorf("%s bot: bot_name is required", botName)
	}

	if bc.CommandPrefix == "" {
		return fmt.Errorf("%s bot: command_prefix is required", botName)
	}

	if bc.LogLevel == "" {
		bc.LogLevel = "INFO"
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
	}

	if !validLogLevels[bc.LogLevel] {
		return fmt.Errorf("%s bot: invalid log_level '%s', must be one of: DEBUG, INFO, WARN, ERROR", botName, bc.LogLevel)
	}

	// Set defaults for durations if not set
	if bc.CommandCooldown == 0 {
		bc.CommandCooldown = 5 * time.Second
	}

	if bc.ShutdownTimeout == 0 {
		bc.ShutdownTimeout = 10 * time.Second
	}

	// Music bot specific validation
	if botName == "Music" {
		if bc.MaxQueueSize <= 0 {
			bc.MaxQueueSize = 100
		}

		if bc.InactivityTimeout == 0 {
			bc.InactivityTimeout = 5 * time.Minute
		}

		if bc.VolumeLevel <= 0 || bc.VolumeLevel > 1 {
			bc.VolumeLevel = 0.5
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
