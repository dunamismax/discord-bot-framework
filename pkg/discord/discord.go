// Package discord provides shared Discord utilities and interfaces for all bot implementations.
package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sawyer/discord-bot-framework/pkg/config"
	"github.com/sawyer/discord-bot-framework/pkg/errors"
	"github.com/sawyer/discord-bot-framework/pkg/logging"
	"github.com/sawyer/discord-bot-framework/pkg/metrics"
)

// BotInterface defines the common interface that all bots must implement.
type BotInterface interface {
	// Start starts the bot
	Start() error
	
	// Stop stops the bot
	Stop() error
	
	// GetConfig returns the bot configuration
	GetConfig() *config.Config
	
	// GetSession returns the Discord session
	GetSession() *discordgo.Session
	
	// GetBotInfo returns information about the bot
	GetBotInfo() BotInfo
}

// BotInfo contains basic information about a bot.
type BotInfo struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Version     string    `json:"version"`
	StartTime   time.Time `json:"start_time"`
	IsConnected bool      `json:"is_connected"`
}

// CommandContext provides context for command execution.
type CommandContext struct {
	Session   *discordgo.Session
	Message   *discordgo.MessageCreate
	Args      []string
	Command   string
	UserID    string
	Username  string
	ChannelID string
	GuildID   string
	BotConfig *config.Config
}

// CommandHandler represents a function that handles a Discord command.
type CommandHandler func(ctx *CommandContext) error

// EventHandler represents a function that handles Discord events.
type EventHandler func(s *discordgo.Session, event interface{})

// BaseBot provides common functionality for all Discord bots.
type BaseBot struct {
	config      *config.Config
	session     *discordgo.Session
	handlers    map[string]CommandHandler
	eventHandlers []EventHandler
	startTime   time.Time
	isConnected bool
}

// NewBaseBot creates a new base bot instance.
func NewBaseBot(cfg *config.Config) (*BaseBot, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.NewConfigError("invalid configuration", err)
	}

	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, errors.NewDiscordError("failed to create Discord session", err)
	}

	// Configure session settings for optimal performance
	session.StateEnabled = true
	session.State.TrackChannels = false
	session.State.TrackEmojis = false
	session.State.TrackMembers = false
	session.State.TrackRoles = false
	session.State.TrackVoice = false
	session.State.TrackPresences = false

	bot := &BaseBot{
		config:   cfg,
		session:  session,
		handlers: make(map[string]CommandHandler),
	}

	// Register default event handlers
	bot.session.AddHandler(bot.onReady)
	bot.session.AddHandler(bot.onDisconnect)
	bot.session.AddHandler(bot.onMessageCreate)

	return bot, nil
}

// RegisterCommand registers a command handler.
func (b *BaseBot) RegisterCommand(command string, handler CommandHandler) {
	b.handlers[command] = handler
	logger := logging.WithBot(b.config.BotName, string(b.config.BotType))
	logger.Debug("Registered command", "command", command)
}

// RegisterEventHandler registers a custom event handler.
func (b *BaseBot) RegisterEventHandler(handler EventHandler) {
	b.eventHandlers = append(b.eventHandlers, handler)
}

// Start starts the Discord bot.
func (b *BaseBot) Start() error {
	logger := logging.WithBot(b.config.BotName, string(b.config.BotType))
	logger.Info("Starting Discord bot connection")
	
	if err := b.session.Open(); err != nil {
		return errors.NewDiscordError("failed to open Discord connection", err)
	}

	b.startTime = time.Now()
	b.isConnected = true
	
	logging.LogStartup(b.config.BotName, string(b.config.BotType), b.config.CommandPrefix, b.config.LogLevel, b.config.DebugMode)
	
	return nil
}

// Stop stops the Discord bot.
func (b *BaseBot) Stop() error {
	logger := logging.WithBot(b.config.BotName, string(b.config.BotType))
	logger.Info("Stopping Discord bot")
	
	b.isConnected = false
	
	if b.session != nil {
		if err := b.session.Close(); err != nil {
			return errors.NewDiscordError("failed to close Discord connection", err)
		}
	}
	
	logging.LogShutdown(b.config.BotName, string(b.config.BotType))
	
	return nil
}

// GetConfig returns the bot configuration.
func (b *BaseBot) GetConfig() *config.Config {
	return b.config
}

// GetSession returns the Discord session.
func (b *BaseBot) GetSession() *discordgo.Session {
	return b.session
}

// GetBotInfo returns information about the bot.
func (b *BaseBot) GetBotInfo() BotInfo {
	return BotInfo{
		Name:        b.config.BotName,
		Type:        string(b.config.BotType),
		Version:     "2.0.0",
		StartTime:   b.startTime,
		IsConnected: b.isConnected,
	}
}

// onReady handles the ready event.
func (b *BaseBot) onReady(s *discordgo.Session, event *discordgo.Ready) {
	logger := logging.WithBot(b.config.BotName, string(b.config.BotType))
	logger.Info("Bot is ready",
		"username", event.User.Username,
		"discriminator", event.User.Discriminator,
		"guilds", len(event.Guilds),
	)
	
	// Set bot status
	status := fmt.Sprintf("Ready | %s", b.config.CommandPrefix+"help")
	err := s.UpdateGameStatus(0, status)
	if err != nil {
		logger.Warn("Failed to set bot status", "error", err)
	}
}

// onDisconnect handles disconnect events.
func (b *BaseBot) onDisconnect(s *discordgo.Session, event *discordgo.Disconnect) {
	logger := logging.WithBot(b.config.BotName, string(b.config.BotType))
	logger.Warn("Bot disconnected", "reason", event.CloseCode)
	b.isConnected = false
	
	metrics.RecordPerformanceMetric("discord", "disconnections", 1, "count")
}

// onMessageCreate handles message creation events.
func (b *BaseBot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from bots
	if m.Author.Bot {
		return
	}

	// Check if message starts with command prefix
	if !strings.HasPrefix(m.Content, b.config.CommandPrefix) {
		// Call custom event handlers for non-command messages
		for _, handler := range b.eventHandlers {
			handler(s, m)
		}
		return
	}

	// Parse command and arguments
	content := strings.TrimPrefix(m.Content, b.config.CommandPrefix)
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	// Create command context
	ctx := &CommandContext{
		Session:   s,
		Message:   m,
		Args:      args,
		Command:   command,
		UserID:    m.Author.ID,
		Username:  m.Author.Username,
		ChannelID: m.ChannelID,
		GuildID:   m.GuildID,
		BotConfig: b.config,
	}

	// Execute command
	startTime := time.Now()
	var success bool
	var err error

	if handler, exists := b.handlers[command]; exists {
		err = handler(ctx)
		success = err == nil
	} else {
		// Command not found
		success = false
		err = errors.NewValidationError(fmt.Sprintf("Unknown command: %s", command))
	}

	duration := time.Since(startTime)

	// Record metrics
	metrics.RecordCommand(command, m.Author.ID, success, duration)
	logging.LogDiscordCommand(m.Author.ID, m.Author.Username, command, success)

	// Handle errors
	if err != nil {
		b.handleCommandError(ctx, err)
	}
}

// handleCommandError handles command execution errors.
func (b *BaseBot) handleCommandError(ctx *CommandContext, err error) {
	logger := logging.WithBot(b.config.BotName, string(b.config.BotType))
	// Log the error
	logging.LogError(logger, err, "Command execution failed")

	// Don't send error messages to users for security errors
	if errors.IsErrorType(err, errors.ErrorTypeSecurity) {
		logging.LogSecurityEvent("command_error", ctx.UserID, err.Error(), "medium")
		return
	}

	// Send user-friendly error message
	var message string
	switch {
	case errors.IsErrorType(err, errors.ErrorTypeNotFound):
		message = "❌ Not found. Try a different search."
	case errors.IsErrorType(err, errors.ErrorTypeValidation):
		message = "❌ Invalid command or parameters. Use `" + ctx.BotConfig.CommandPrefix + "help` for usage."
	case errors.IsErrorType(err, errors.ErrorTypeRateLimit):
		message = "⏱️ Rate limited. Please wait a moment before trying again."
	case errors.IsErrorType(err, errors.ErrorTypePermission):
		message = "🚫 Permission denied."
	default:
		message = "❌ An error occurred. Please try again later."
	}

	// Send error message to channel
	_, sendErr := ctx.Session.ChannelMessageSend(ctx.ChannelID, message)
	if sendErr != nil {
		logger := logging.WithBot(b.config.BotName, string(b.config.BotType))
		logger.Error("Failed to send error message", "error", sendErr)
	}
}

// ValidateInput provides basic input validation and sanitization.
func ValidateInput(input string, maxLength int) error {
	if len(input) == 0 {
		return errors.NewValidationError("input cannot be empty")
	}
	
	if len(input) > maxLength {
		return errors.NewValidationError(fmt.Sprintf("input too long (max %d characters)", maxLength))
	}
	
	// Basic security checks
	if containsSuspiciousContent(input) {
		return errors.NewSecurityError("suspicious input detected", nil)
	}
	
	return nil
}

// containsSuspiciousContent performs basic security checks on user input.
func containsSuspiciousContent(input string) bool {
	suspicious := []string{
		"<script",
		"javascript:",
		"data:",
		"file://",
		"vbscript:",
		"onload=",
		"onerror=",
		"@everyone",
		"@here",
	}
	
	lower := strings.ToLower(input)
	for _, pattern := range suspicious {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	
	return false
}

// CreateEmbed creates a standardized Discord embed.
func CreateEmbed(title, description, color string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	
	// Set color based on string input
	switch strings.ToLower(color) {
	case "red", "error":
		embed.Color = 0xFF0000
	case "green", "success":
		embed.Color = 0x00FF00
	case "blue", "info":
		embed.Color = 0x0000FF
	case "yellow", "warning":
		embed.Color = 0xFFFF00
	case "purple", "magic":
		embed.Color = 0x9932CC
	default:
		embed.Color = 0x7289DA // Discord blurple
	}
	
	return embed
}

// SendTyping sends a typing indicator to the channel.
func SendTyping(s *discordgo.Session, channelID string) {
	err := s.ChannelTyping(channelID)
	if err != nil {
		logging.Debug("Failed to send typing indicator", "error", err)
	}
}