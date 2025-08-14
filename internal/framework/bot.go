// Package framework provides the core Discord bot framework functionality.
package framework

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sawyer/discord-bot-framework/internal/config"
	"github.com/sawyer/discord-bot-framework/internal/errors"
	"github.com/sawyer/discord-bot-framework/internal/logging"
)

// CommandHandler represents a function that handles Discord bot commands.
type CommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate) error

// MessageHandler represents a function that handles Discord messages.
type MessageHandler func(s *discordgo.Session, m *discordgo.MessageCreate) error

// Bot represents a Discord bot instance with core functionality.
type Bot struct {
	session         *discordgo.Session
	config          *config.BotConfig
	logger          *logging.Logger
	commandHandlers map[string]CommandHandler
	messageHandlers []MessageHandler
	commands        []*discordgo.ApplicationCommand
	cooldowns       map[string]map[string]time.Time
	cooldownMutex   sync.RWMutex
}

// NewBot creates a new Discord bot instance with the framework.
func NewBot(cfg *config.BotConfig) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, errors.NewDiscordError("failed to create Discord session", err)
	}

	bot := &Bot{
		session:         session,
		config:          cfg,
		logger:          logging.WithComponent("framework"),
		commandHandlers: make(map[string]CommandHandler),
		messageHandlers: make([]MessageHandler, 0),
		commands:        make([]*discordgo.ApplicationCommand, 0),
		cooldowns:       make(map[string]map[string]time.Time),
	}

	// Set Discord session intents
	session.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildVoiceStates

	// Add event handlers
	session.AddHandler(bot.interactionCreate)
	session.AddHandler(bot.messageCreate)
	session.AddHandler(bot.ready)

	return bot, nil
}

// Start starts the Discord bot.
func (b *Bot) Start() error {
	b.logger.Info("Starting bot", "bot_name", b.config.BotName)

	err := b.session.Open()
	if err != nil {
		return errors.NewDiscordError("failed to open Discord session", err)
	}

	b.logger.Info("Bot is now running", "username", b.session.State.User.Username)
	return nil
}

// Stop stops the Discord bot.
func (b *Bot) Stop(ctx context.Context) error {
	b.logger.Info("Stopping bot", "bot_name", b.config.BotName)

	// Remove registered commands
	if err := b.removeCommands(ctx); err != nil {
		b.logger.Error("Error removing commands", "error", err)
	}

	if err := b.session.Close(); err != nil {
		return errors.NewDiscordError("failed to close Discord session", err)
	}

	b.logger.Info("Bot stopped successfully")
	return nil
}

// RegisterCommand registers a slash command with its handler.
func (b *Bot) RegisterCommand(command *discordgo.ApplicationCommand, handler CommandHandler) {
	b.commands = append(b.commands, command)
	b.commandHandlers[command.Name] = handler
}

// RegisterMessageHandler registers a message handler.
func (b *Bot) RegisterMessageHandler(handler MessageHandler) {
	b.messageHandlers = append(b.messageHandlers, handler)
}

// ready handles the ready event.
func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	b.logger.Info("Bot is ready", "username", event.User.Username)

	// Register slash commands
	if err := b.registerCommands(); err != nil {
		b.logger.Error("Failed to register commands", "error", err)
	}
}

// registerCommands registers all slash commands with Discord.
func (b *Bot) registerCommands() error {
	for _, command := range b.commands {
		var err error
		if b.config.GuildID != "" {
			// Register for specific guild (faster for development)
			_, err = b.session.ApplicationCommandCreate(b.session.State.User.ID, b.config.GuildID, command)
		} else {
			// Register globally
			_, err = b.session.ApplicationCommandCreate(b.session.State.User.ID, "", command)
		}

		if err != nil {
			return errors.NewDiscordError(fmt.Sprintf("failed to register command %s", command.Name), err)
		}

		b.logger.Info("Registered command", "command", command.Name)
	}

	return nil
}

// removeCommands removes all registered commands.
func (b *Bot) removeCommands(ctx context.Context) error {
	var commands []*discordgo.ApplicationCommand
	var err error

	if b.config.GuildID != "" {
		commands, err = b.session.ApplicationCommands(b.session.State.User.ID, b.config.GuildID)
	} else {
		commands, err = b.session.ApplicationCommands(b.session.State.User.ID, "")
	}

	if err != nil {
		return errors.NewDiscordError("failed to fetch commands", err)
	}

	for _, command := range commands {
		var deleteErr error
		if b.config.GuildID != "" {
			deleteErr = b.session.ApplicationCommandDelete(b.session.State.User.ID, b.config.GuildID, command.ID)
		} else {
			deleteErr = b.session.ApplicationCommandDelete(b.session.State.User.ID, "", command.ID)
		}

		if deleteErr != nil {
			b.logger.Error("Failed to delete command", "command", command.Name, "error", deleteErr)
		} else {
			b.logger.Info("Deleted command", "command", command.Name)
		}
	}

	return nil
}

// interactionCreate handles interaction events (slash commands).
func (b *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name == "" {
		return
	}

	commandName := i.ApplicationCommandData().Name
	handler, exists := b.commandHandlers[commandName]
	if !exists {
		b.logger.Warn("Unknown command", "command", commandName)
		return
	}

	// Check cooldown
	if b.isOnCooldown(i.Member.User.ID, commandName) {
		remaining := b.getCooldownRemaining(i.Member.User.ID, commandName)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Command is on cooldown. Try again in %.1f seconds.", remaining.Seconds()),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			b.logger.Error("Failed to respond to interaction", "error", err)
		}
		return
	}

	// Set cooldown
	b.setCooldown(i.Member.User.ID, commandName)

	// Execute command
	if err := handler(s, i); err != nil {
		b.logger.Error("Command execution failed",
			"command", commandName,
			"user_id", i.Member.User.ID,
			"username", i.Member.User.Username,
			"error", err,
		)

		// Try to respond with error message
		content := "Sorry, something went wrong processing your command."
		if errors.IsErrorType(err, errors.ErrorTypeValidation) {
			content = err.Error()
		}

		respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if respondErr != nil {
			b.logger.Error("Failed to respond to interaction with error", "error", respondErr)
		}
	} else {
		logging.LogDiscordCommand(i.Member.User.ID, i.Member.User.Username, commandName, true)
	}
}

// messageCreate handles message events.
func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from bots
	if m.Author.Bot {
		return
	}

	// Check if message starts with command prefix for prefix commands
	if strings.HasPrefix(m.Content, b.config.CommandPrefix) {
		// Handle prefix commands if needed (most functionality moved to slash commands)
		return
	}

	// Process message handlers
	for _, handler := range b.messageHandlers {
		if err := handler(s, m); err != nil {
			b.logger.Error("Message handler failed",
				"user_id", m.Author.ID,
				"username", m.Author.Username,
				"error", err,
			)
		}
	}
}

// isOnCooldown checks if a user is on cooldown for a command.
func (b *Bot) isOnCooldown(userID, commandName string) bool {
	b.cooldownMutex.RLock()
	defer b.cooldownMutex.RUnlock()

	userCooldowns, exists := b.cooldowns[userID]
	if !exists {
		return false
	}

	lastUsed, exists := userCooldowns[commandName]
	if !exists {
		return false
	}

	return time.Since(lastUsed) < b.config.CommandCooldown
}

// setCooldown sets a cooldown for a user and command.
func (b *Bot) setCooldown(userID, commandName string) {
	b.cooldownMutex.Lock()
	defer b.cooldownMutex.Unlock()

	if b.cooldowns[userID] == nil {
		b.cooldowns[userID] = make(map[string]time.Time)
	}

	b.cooldowns[userID][commandName] = time.Now()
}

// getCooldownRemaining gets the remaining cooldown time for a user and command.
func (b *Bot) getCooldownRemaining(userID, commandName string) time.Duration {
	b.cooldownMutex.RLock()
	defer b.cooldownMutex.RUnlock()

	userCooldowns, exists := b.cooldowns[userID]
	if !exists {
		return 0
	}

	lastUsed, exists := userCooldowns[commandName]
	if !exists {
		return 0
	}

	elapsed := time.Since(lastUsed)
	remaining := b.config.CommandCooldown - elapsed

	if remaining < 0 {
		return 0
	}

	return remaining
}

// GetSession returns the Discord session for advanced usage.
func (b *Bot) GetSession() *discordgo.Session {
	return b.session
}

// GetConfig returns the bot configuration.
func (b *Bot) GetConfig() *config.BotConfig {
	return b.config
}

// GetLogger returns the bot logger.
func (b *Bot) GetLogger() *logging.Logger {
	return b.logger
}