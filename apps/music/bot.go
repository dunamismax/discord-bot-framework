// Package main provides the Music Discord bot implementation.
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sawyer/go-discord-bots/pkg/config"
	"github.com/sawyer/go-discord-bots/pkg/discord"
	"github.com/sawyer/go-discord-bots/pkg/errors"
	"github.com/sawyer/go-discord-bots/pkg/logging"
	"github.com/sawyer/go-discord-bots/pkg/metrics"
)

// Bot represents the Music Discord bot.
type Bot struct {
	*discord.BaseBot
	database        *Database
	audioPlayer     *AudioPlayer
	queueManager    *QueueManager
	audioExtractor  *AudioExtractor
	commandHandlers map[string]SlashCommandHandler
}

// SlashCommandHandler represents a function that handles slash commands.
type SlashCommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate) error

// NewBot creates a new Music bot instance.
func NewBot(cfg *config.Config) (*Bot, error) {
	// Change command prefix to avoid conflicts since music bot only uses slash commands
	cfg.CommandPrefix = "/music_disabled" // Use a prefix that won't conflict

	// Create normal BaseBot - we'll override the message handler behavior
	baseBot, err := discord.NewBaseBot(cfg)
	if err != nil {
		return nil, errors.NewConfigError("failed to create base bot", err)
	}

	bot := &Bot{
		BaseBot:         baseBot,
		queueManager:    NewQueueManager(),
		audioPlayer:     NewAudioPlayer(),
		audioExtractor:  NewAudioExtractor(),
		commandHandlers: make(map[string]SlashCommandHandler),
	}

	// Initialize database if URL is provided
	if cfg.DatabaseURL != "" {
		database, err := NewDatabase(cfg.DatabaseURL)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to initialize database", err)
		}
		bot.database = database
	}

	// Register both slash commands and event handlers
	bot.registerSlashCommands()
	bot.BaseBot.GetSession().AddHandler(bot.onInteractionCreate)

	// Remove the BaseBot's message handler and replace with empty handler
	// The music bot should only respond to slash commands, not text commands
	bot.BaseBot.GetSession().AddHandler(bot.onMessageCreate)

	// Set voice intents for audio functionality
	bot.GetSession().Identify.Intents |= discordgo.IntentsGuildVoiceStates

	// Enable voice state tracking for the music bot
	bot.GetSession().State.TrackVoice = true

	return bot, nil
}

// Start starts the Music bot.
func (b *Bot) Start() error {
	if err := b.BaseBot.Start(); err != nil {
		return err
	}

	// Register slash commands with Discord
	if err := b.registerSlashCommandsWithDiscord(); err != nil {
		return err
	}

	logger := logging.WithComponent("music-bot")
	logger.Info("Music bot started successfully")

	return nil
}

// Stop stops the Music bot.
func (b *Bot) Stop() error {
	logger := logging.WithComponent("music-bot")
	logger.Info("Stopping Music bot")

	// Clean up audio connections
	if b.audioPlayer != nil {
		b.audioPlayer.Cleanup()
	}

	// Clean up queues
	if b.queueManager != nil {
		b.queueManager.Cleanup()
	}

	// Close database
	if b.database != nil {
		if err := b.database.Close(); err != nil {
			logger.Error("Error closing database", "error", err)
		}
	}

	return b.BaseBot.Stop()
}

// registerSlashCommands registers all slash command handlers.
func (b *Bot) registerSlashCommands() {
	// Basic playback commands
	b.commandHandlers["play"] = b.handlePlaySlashCommand
	b.commandHandlers["pause"] = b.handlePauseSlashCommand
	b.commandHandlers["resume"] = b.handleResumeSlashCommand
	b.commandHandlers["skip"] = b.handleSkipSlashCommand
	b.commandHandlers["stop"] = b.handleStopSlashCommand
	b.commandHandlers["queue"] = b.handleQueueSlashCommand
	b.commandHandlers["volume"] = b.handleVolumeSlashCommand

	// Playlist commands (only if database is available)
	if b.database != nil {
		b.commandHandlers["playlist_create"] = b.handlePlaylistCreateSlashCommand
		b.commandHandlers["playlist_list"] = b.handlePlaylistListSlashCommand
		b.commandHandlers["playlist_show"] = b.handlePlaylistShowSlashCommand
		b.commandHandlers["playlist_play"] = b.handlePlaylistPlaySlashCommand
		b.commandHandlers["playlist_add"] = b.handlePlaylistAddSlashCommand
		b.commandHandlers["playlist_remove"] = b.handlePlaylistRemoveSlashCommand
		b.commandHandlers["playlist_delete"] = b.handlePlaylistDeleteSlashCommand
	}
}

// registerSlashCommandsWithDiscord registers slash commands with Discord API.
func (b *Bot) registerSlashCommandsWithDiscord() error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "play",
			Description: "Play music from YouTube (auto-joins your voice channel)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "YouTube URL or search query",
					Required:    true,
				},
			},
		},
		{
			Name:        "pause",
			Description: "Pause the current song",
		},
		{
			Name:        "resume",
			Description: "Resume playback",
		},
		{
			Name:        "skip",
			Description: "Skip the current song",
		},
		{
			Name:        "stop",
			Description: "Stop music and disconnect",
		},
		{
			Name:        "queue",
			Description: "Show the music queue",
		},
		{
			Name:        "volume",
			Description: "Set or show volume level",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "level",
					Description: "Volume level (0-100)",
					Required:    false,
					MinValue:    func() *float64 { v := 0.0; return &v }(),
					MaxValue:    100,
				},
			},
		},
	}

	// Add playlist commands if database is available
	if b.database != nil {
		playlistCommands := []*discordgo.ApplicationCommand{
			{
				Name:        "playlist_create",
				Description: "Create a new playlist",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "name",
						Description: "Playlist name",
						Required:    true,
					},
				},
			},
			{
				Name:        "playlist_list",
				Description: "List your playlists",
			},
			{
				Name:        "playlist_show",
				Description: "Show songs in a playlist",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "playlist_id",
						Description: "Playlist ID",
						Required:    true,
					},
				},
			},
			{
				Name:        "playlist_play",
				Description: "Queue an entire playlist",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "playlist_id",
						Description: "Playlist ID",
						Required:    true,
					},
				},
			},
			{
				Name:        "playlist_add",
				Description: "Add current song to playlist",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "playlist_id",
						Description: "Playlist ID",
						Required:    true,
					},
				},
			},
			{
				Name:        "playlist_remove",
				Description: "Remove a song from playlist",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "playlist_id",
						Description: "Playlist ID",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "song_number",
						Description: "Song position in playlist",
						Required:    true,
					},
				},
			},
			{
				Name:        "playlist_delete",
				Description: "Delete a playlist",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "playlist_id",
						Description: "Playlist ID",
						Required:    true,
					},
				},
			},
		}
		commands = append(commands, playlistCommands...)
	}

	logger := logging.WithComponent("music-bot")
	var guildID string
	if b.isValidGuildID(b.GetConfig().GuildID) {
		guildID = b.GetConfig().GuildID
		logger.Info("Registering guild-specific commands", "guild_id", guildID)
	} else {
		if b.GetConfig().GuildID != "" {
			logger.Warn("Invalid guild ID provided, falling back to global commands", "invalid_guild_id", b.GetConfig().GuildID)
		} else {
			logger.Info("No guild ID provided, registering global commands")
		}
		guildID = ""
	}

	for _, command := range commands {
		_, err := b.GetSession().ApplicationCommandCreate(b.GetSession().State.User.ID, guildID, command)
		if err != nil {
			return errors.NewDiscordError(fmt.Sprintf("failed to register slash command %s", command.Name), err)
		}
		logger.Info("Registered slash command", "command", command.Name)
	}

	return nil
}

// isValidGuildID checks if the guild ID is a valid Discord snowflake.
func (b *Bot) isValidGuildID(guildID string) bool {
	if guildID == "" {
		return false
	}

	// Discord snowflakes are 17-19 digit numbers
	if len(guildID) < 17 || len(guildID) > 19 {
		return false
	}

	// Check if all characters are digits
	for _, char := range guildID {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// onInteractionCreate handles slash command interactions.
func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name == "" {
		return
	}

	startTime := time.Now()
	commandName := i.ApplicationCommandData().Name

	logger := logging.WithComponent("music-bot")
	logger.Debug("Received slash command", "command", commandName, "user", getUserID(i))

	handler, exists := b.commandHandlers[commandName]
	if !exists {
		b.respondWithError(s, i, "Unknown command")
		return
	}

	// Execute command
	err := handler(s, i)
	success := err == nil

	// Record metrics
	metrics.RecordCommand(commandName, getUserID(i), success, time.Since(startTime))

	// Handle errors
	if err != nil {
		logger.Error("Slash command failed", "command", commandName, "error", err)
		b.respondWithError(s, i, "Command failed: "+err.Error())
	}
}

// Helper functions for interaction handling
func getUserID(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		return i.Member.User.ID
	}
	return i.User.ID
}

func getUsername(i *discordgo.InteractionCreate) string {
	if i.Member != nil {
		return i.Member.User.Username
	}
	return i.User.Username
}

func (b *Bot) respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ " + message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger := logging.WithComponent("music-bot")
		logger.Error("Failed to send error response", "error", err)
	}
}

// handlePlaySlashCommand handles the /play slash command.
// Automatically joins the user's voice channel and plays the requested music.
func (b *Bot) handlePlaySlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	startTime := time.Now()

	// Get command options
	options := i.ApplicationCommandData().Options
	if len(options) == 0 || options[0].Name != "query" {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Please provide a song name or URL",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	query := options[0].StringValue()
	if err := discord.ValidateInput(query, 500); err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Invalid input: %s", err.Error()),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	userID := getUserID(i)
	username := getUsername(i)
	guildID := i.GuildID

	// Check if user is in a voice channel
	voiceState, err := b.getUserVoiceState(s, guildID, userID)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Failed to check your voice channel status",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	if voiceState == nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You must be in a voice channel to play music! Please join a voice channel and try again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Check if bot is already in a different voice channel
	botVoiceState, err := b.getUserVoiceState(s, guildID, s.State.User.ID)
	if err == nil && botVoiceState != nil && botVoiceState.ChannelID != voiceState.ChannelID {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ I'm already playing music in another voice channel! Please join that channel or wait for the current session to end.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Send immediate response to avoid timeout
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🎵 Searching and loading your song...",
		},
	})
	if err != nil {
		return errors.NewDiscordError("failed to respond to interaction", err)
	}

	// Extract song information
	song, err := b.extractSongInfo(query)
	if err != nil {
		metrics.RecordAPIRequest("youtube", "extract", false, time.Since(startTime))
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &[]string{fmt.Sprintf("❌ Could not find or load the requested song: %s", err.Error())}[0],
		})
		if editErr != nil {
			return editErr
		}
		return err
	}

	song.RequesterID = userID
	song.RequesterName = username

	// Automatically join the user's voice channel
	audioConn, err := b.audioPlayer.GetConnection(s, guildID, voiceState.ChannelID)
	if err != nil {
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &[]string{fmt.Sprintf("❌ Failed to join your voice channel: %s\n\nPlease check that I have permission to connect and speak in this channel.", err.Error())}[0],
		})
		if editErr != nil {
			return editErr
		}
		metrics.RecordCommand("play", userID, false, time.Since(startTime))
		return err
	}

	// Add to queue
	queue := b.queueManager.GetQueue(guildID)
	position := queue.Add(song)

	var response string
	if position == 0 && !queue.IsPlaying() {
		response = fmt.Sprintf("🔊 Joined your voice channel and now playing: **%s**", song.Title)
		go b.audioPlayer.PlayNext(s, guildID, audioConn, queue)
	} else {
		response = fmt.Sprintf("🎵 Added to queue: **%s**\nPosition in queue: %d", song.Title, position+1)
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &response,
	})

	metrics.RecordCommand("play", userID, err == nil, time.Since(startTime))
	return err
}

// handlePauseSlashCommand handles the /pause slash command.
func (b *Bot) handlePauseSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	startTime := time.Now()
	userID := getUserID(i)
	guildID := i.GuildID

	// Validate user is in a voice channel with the bot
	if err := b.validateUserInBotVoiceChannel(s, i); err != nil {
		return err
	}

	queue := b.queueManager.GetQueue(guildID)
	if !queue.IsPlaying() {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Nothing is currently playing",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		metrics.RecordCommand("pause", userID, false, time.Since(startTime))
		return err
	}

	// Pause both queue and audio stream
	queue.SetPaused(true)
	b.audioPlayer.enhanced.PauseStream(guildID)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏸️ Paused the current song",
		},
	})

	metrics.RecordCommand("pause", userID, err == nil, time.Since(startTime))
	return err
}

// handleResumeSlashCommand handles the /resume slash command.
func (b *Bot) handleResumeSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	startTime := time.Now()
	userID := getUserID(i)
	guildID := i.GuildID

	// Validate user is in a voice channel with the bot
	if err := b.validateUserInBotVoiceChannel(s, i); err != nil {
		return err
	}

	queue := b.queueManager.GetQueue(guildID)
	if !queue.IsPaused() {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Nothing is currently paused",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		metrics.RecordCommand("resume", userID, false, time.Since(startTime))
		return err
	}

	// Resume both queue and audio stream
	queue.SetPaused(false)
	b.audioPlayer.enhanced.ResumeStream(guildID)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "▶️ Resumed the current song",
		},
	})

	metrics.RecordCommand("resume", userID, err == nil, time.Since(startTime))
	return err
}

// handleSkipSlashCommand handles the /skip slash command.
func (b *Bot) handleSkipSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	startTime := time.Now()
	userID := getUserID(i)
	guildID := i.GuildID

	// Validate user is in a voice channel with the bot
	if err := b.validateUserInBotVoiceChannel(s, i); err != nil {
		return err
	}

	queue := b.queueManager.GetQueue(guildID)
	if !queue.IsPlaying() {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Nothing is currently playing",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		metrics.RecordCommand("skip", userID, false, time.Since(startTime))
		return err
	}

	current := queue.Current()
	queue.Skip()

	// Stop current audio stream to trigger next song
	b.audioPlayer.enhanced.StopStream(guildID)

	response := "⏭️ Skipped the current song"
	if current != nil {
		response = fmt.Sprintf("⏭️ Skipped **%s**", current.Title)
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})

	metrics.RecordCommand("skip", userID, err == nil, time.Since(startTime))
	return err
}

// handleStopSlashCommand handles the /stop slash command.
func (b *Bot) handleStopSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	startTime := time.Now()
	userID := getUserID(i)
	guildID := i.GuildID

	// Validate user is in a voice channel with the bot
	if err := b.validateUserInBotVoiceChannel(s, i); err != nil {
		return err
	}

	// Stop audio stream, disconnect from voice and clear queue
	b.audioPlayer.enhanced.StopStream(guildID)
	b.audioPlayer.Disconnect(guildID)
	b.queueManager.ClearQueue(guildID)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹️ Stopped music and disconnected from voice channel",
		},
	})

	metrics.RecordCommand("stop", userID, err == nil, time.Since(startTime))
	return err
}

// handleQueueSlashCommand handles the /queue slash command.
func (b *Bot) handleQueueSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	startTime := time.Now()
	userID := getUserID(i)
	guildID := i.GuildID

	queue := b.queueManager.GetQueue(guildID)

	if queue.Current() == nil && queue.IsEmpty() {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "📭 The queue is empty",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		metrics.RecordCommand("queue", userID, err == nil, time.Since(startTime))
		return err
	}

	embed := b.buildQueueEmbed(queue)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	metrics.RecordCommand("queue", userID, err == nil, time.Since(startTime))
	return err
}

// handleVolumeSlashCommand handles the /volume slash command.
func (b *Bot) handleVolumeSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	startTime := time.Now()
	userID := getUserID(i)
	guildID := i.GuildID

	// Validate user is in a voice channel with the bot
	if err := b.validateUserInBotVoiceChannel(s, i); err != nil {
		return err
	}

	options := i.ApplicationCommandData().Options

	if len(options) == 0 {
		// Show current volume
		volume := b.audioPlayer.GetVolume(guildID)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("🔊 Current volume: %d%%", int(volume*100)),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		metrics.RecordCommand("volume", userID, err == nil, time.Since(startTime))
		return err
	}

	// Get volume level from options
	volume := int(options[0].IntValue())

	if volume < 0 || volume > 100 {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Volume must be between 0 and 100",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		metrics.RecordCommand("volume", userID, false, time.Since(startTime))
		return err
	}

	volumeFloat := float64(volume) / 100.0
	// Set volume for both base player and active stream
	b.audioPlayer.enhanced.SetStreamVolume(guildID, volumeFloat)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🔊 Volume set to %d%%", volume),
		},
	})

	metrics.RecordCommand("volume", userID, err == nil, time.Since(startTime))
	return err
}

// Playlist slash command handlers

// handlePlaylistCreateSlashCommand handles the /playlist_create slash command.
func (b *Bot) handlePlaylistCreateSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if b.database == nil {
		return b.respondPlaylistNotAvailable(s, i)
	}

	startTime := time.Now()
	userID := getUserID(i)
	guildID := i.GuildID

	options := i.ApplicationCommandData().Options
	if len(options) == 0 || options[0].Name != "name" {
		return b.respondWithSlashError(s, i, "Please provide a playlist name")
	}

	name := options[0].StringValue()
	if len(name) > 50 {
		return b.respondWithSlashError(s, i, "Playlist name must be 50 characters or less")
	}

	playlistID, err := b.database.CreatePlaylist(userID, guildID, name)
	if err != nil {
		logger := logging.WithComponent("playlist")
		logging.LogError(logger, err, "Failed to create playlist")
		return b.respondWithSlashError(s, i, "Failed to create playlist")
	}

	response := fmt.Sprintf("✅ Created playlist **%s** (ID: %d)", name, playlistID)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})

	metrics.RecordCommand("playlist_create", userID, err == nil, time.Since(startTime))
	return err
}

// handlePlaylistListSlashCommand handles the /playlist_list slash command.
func (b *Bot) handlePlaylistListSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if b.database == nil {
		return b.respondPlaylistNotAvailable(s, i)
	}

	startTime := time.Now()
	userID := getUserID(i)
	username := getUsername(i)
	guildID := i.GuildID

	playlists, err := b.database.GetUserPlaylists(userID, guildID)
	if err != nil {
		logger := logging.WithComponent("playlist")
		logging.LogError(logger, err, "Failed to list playlists")
		return b.respondWithSlashError(s, i, "Failed to list playlists")
	}

	if len(playlists) == 0 {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "📝 You don't have any playlists yet. Use `/playlist_create` to make one!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		metrics.RecordCommand("playlist_list", userID, err == nil, time.Since(startTime))
		return err
	}

	embed := discord.CreateEmbed(fmt.Sprintf("🎵 %s's Playlists", username), "", "info")

	displayCount := len(playlists)
	if displayCount > 10 {
		displayCount = 10
	}

	for i := 0; i < displayCount; i++ {
		playlist := playlists[i]
		songCount := len(playlist.Songs)
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s (ID: %d)", playlist.Name, playlist.ID),
			Value:  fmt.Sprintf("%d song%s", songCount, map[bool]string{true: "", false: "s"}[songCount == 1]),
			Inline: true,
		})
	}

	if len(playlists) > 10 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "",
			Value:  fmt.Sprintf("... and %d more playlists", len(playlists)-10),
			Inline: false,
		})
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	metrics.RecordCommand("playlist_list", userID, err == nil, time.Since(startTime))
	return err
}

// handlePlaylistShowSlashCommand handles the /playlist_show slash command.
func (b *Bot) handlePlaylistShowSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if b.database == nil {
		return b.respondPlaylistNotAvailable(s, i)
	}

	return b.respondWithSlashError(s, i, "🚧 Playlist show not yet implemented")
}

// handlePlaylistPlaySlashCommand handles the /playlist_play slash command.
func (b *Bot) handlePlaylistPlaySlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if b.database == nil {
		return b.respondPlaylistNotAvailable(s, i)
	}

	return b.respondWithSlashError(s, i, "🚧 Playlist play not yet implemented")
}

// handlePlaylistAddSlashCommand handles the /playlist_add slash command.
func (b *Bot) handlePlaylistAddSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if b.database == nil {
		return b.respondPlaylistNotAvailable(s, i)
	}

	return b.respondWithSlashError(s, i, "🚧 Playlist add not yet implemented")
}

// handlePlaylistRemoveSlashCommand handles the /playlist_remove slash command.
func (b *Bot) handlePlaylistRemoveSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if b.database == nil {
		return b.respondPlaylistNotAvailable(s, i)
	}

	return b.respondWithSlashError(s, i, "🚧 Playlist remove not yet implemented")
}

// handlePlaylistDeleteSlashCommand handles the /playlist_delete slash command.
func (b *Bot) handlePlaylistDeleteSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if b.database == nil {
		return b.respondPlaylistNotAvailable(s, i)
	}

	return b.respondWithSlashError(s, i, "🚧 Playlist delete not yet implemented")
}

// Helper methods for slash command responses
func (b *Bot) respondPlaylistNotAvailable(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ Playlist functionality is not available",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *Bot) respondWithSlashError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// Helper methods

// getUserVoiceState gets the user's voice state.
func (b *Bot) getUserVoiceState(s *discordgo.Session, guildID, userID string) (*discordgo.VoiceState, error) {
	logger := logging.WithComponent("music-bot")

	guild, err := s.State.Guild(guildID)
	if err != nil {
		logger.Error("Failed to get guild", "guild_id", guildID, "error", err)
		return nil, err
	}

	logger.Info("Checking voice states", "guild_id", guildID, "user_id", userID, "total_voice_states", len(guild.VoiceStates))

	for _, vs := range guild.VoiceStates {
		logger.Info("Checking voice state", "voice_user_id", vs.UserID, "target_user_id", userID, "channel_id", vs.ChannelID)
		if vs.UserID == userID {
			logger.Info("Found user voice state", "user_id", userID, "channel_id", vs.ChannelID)
			return vs, nil
		}
	}

	logger.Info("User not found in any voice channel", "user_id", userID)
	return nil, nil
}

// validateUserInBotVoiceChannel validates that the user is in the same voice channel as the bot.
func (b *Bot) validateUserInBotVoiceChannel(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	userID := getUserID(i)
	guildID := i.GuildID

	// Check if user is in a voice channel
	userVoiceState, err := b.getUserVoiceState(s, guildID, userID)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Failed to check your voice channel status",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	if userVoiceState == nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You must be in a voice channel to use this command",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Check if bot is in a voice channel
	botVoiceState, err := b.getUserVoiceState(s, guildID, s.State.User.ID)
	if err != nil || botVoiceState == nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ I'm not currently in a voice channel. Use `/play` to start playing music first",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Check if user and bot are in the same voice channel
	if userVoiceState.ChannelID != botVoiceState.ChannelID {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You must be in the same voice channel as me to use this command",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	return nil
}

// extractSongInfo extracts song information from a query using yt-dlp.
func (b *Bot) extractSongInfo(query string) (*Song, error) {
	return b.audioExtractor.ExtractSongInfo(query)
}

// onMessageCreate overrides BaseBot's message handler to prevent interference.
// The music bot should only respond to slash commands, not text commands.
func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Explicitly do nothing - music bot only handles slash commands
	// This prevents interference with other bots that use text commands
	// Debug: if this function is being called with ! commands, there's still interference
	logger := logging.WithComponent("music-bot")
	if strings.HasPrefix(m.Content, "!") && !m.Author.Bot {
		logger.Debug("Music bot received text command but ignoring", "content", m.Content[:min(20, len(m.Content))], "author", m.Author.Username)
	}
	// Explicitly do nothing - no return needed
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildQueueEmbed builds an embed showing the current queue.
func (b *Bot) buildQueueEmbed(queue *Queue) *discordgo.MessageEmbed {
	embed := discord.CreateEmbed("🎵 Music Queue", "", "info")

	if current := queue.Current(); current != nil {
		status := "▶️ Playing"
		if queue.IsPaused() {
			status = "⏸️ Paused"
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Now", status),
			Value:  fmt.Sprintf("**%s**\nRequested by: <@%s>", current.Title, current.RequesterID),
			Inline: false,
		})
	}

	songs := queue.GetSongs()
	if len(songs) > 0 {
		var queueList []string
		displayCount := len(songs)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			song := songs[i]
			queueList = append(queueList, fmt.Sprintf("%d. **%s** - <@%s>", i+1, song.Title, song.RequesterID))
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Up Next",
			Value:  strings.Join(queueList, "\n"),
			Inline: false,
		})

		if len(songs) > 10 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "",
				Value:  fmt.Sprintf("... and %d more songs", len(songs)-10),
				Inline: false,
			})
		}
	}

	return embed
}
