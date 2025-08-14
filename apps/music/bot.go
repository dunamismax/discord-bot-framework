// Package main provides the Music Discord bot implementation.
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sawyer/discord-bot-framework/pkg/config"
	"github.com/sawyer/discord-bot-framework/pkg/discord"
	"github.com/sawyer/discord-bot-framework/pkg/errors"
	"github.com/sawyer/discord-bot-framework/pkg/logging"
	"github.com/sawyer/discord-bot-framework/pkg/metrics"
)

// Bot represents the Music Discord bot.
type Bot struct {
	*discord.BaseBot
	database     *Database
	audioPlayer  *AudioPlayer
	queueManager *QueueManager
}

// NewBot creates a new Music bot instance.
func NewBot(cfg *config.Config) (*Bot, error) {
	baseBot, err := discord.NewBaseBot(cfg)
	if err != nil {
		return nil, errors.NewConfigError("failed to create base bot", err)
	}

	bot := &Bot{
		BaseBot:      baseBot,
		queueManager: NewQueueManager(),
		audioPlayer:  NewAudioPlayer(),
	}

	// Initialize database if URL is provided
	if cfg.DatabaseURL != "" {
		database, err := NewDatabase(cfg.DatabaseURL)
		if err != nil {
			return nil, errors.NewDatabaseError("failed to initialize database", err)
		}
		bot.database = database
	}

	// Register commands
	bot.registerCommands()

	return bot, nil
}

// Start starts the Music bot.
func (b *Bot) Start() error {
	if err := b.BaseBot.Start(); err != nil {
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

// registerCommands registers all Music commands.
func (b *Bot) registerCommands() {
	// Basic playback commands
	b.RegisterCommand("play", b.handlePlayCommand)
	b.RegisterCommand("pause", b.handlePauseCommand)
	b.RegisterCommand("resume", b.handleResumeCommand)
	b.RegisterCommand("skip", b.handleSkipCommand)
	b.RegisterCommand("stop", b.handleStopCommand)
	b.RegisterCommand("queue", b.handleQueueCommand)
	b.RegisterCommand("volume", b.handleVolumeCommand)

	// Playlist commands (only if database is available)
	if b.database != nil {
		b.RegisterCommand("playlist", b.handlePlaylistCommand)
	}
}

// handlePlayCommand handles the play command.
func (b *Bot) handlePlayCommand(ctx *discord.CommandContext) error {
	startTime := time.Now()
	
	if len(ctx.Args) == 0 {
		return errors.NewValidationError("Please provide a song name or URL")
	}

	// Validate input
	query := strings.Join(ctx.Args, " ")
	if err := discord.ValidateInput(query, 500); err != nil {
		return err
	}

	// Check if user is in a voice channel
	voiceState, err := b.getUserVoiceState(ctx.Session, ctx.GuildID, ctx.UserID)
	if err != nil {
		return errors.NewDiscordError("failed to get voice state", err)
	}
	if voiceState == nil {
		return errors.NewValidationError("You must be in a voice channel to play music")
	}

	// Send typing indicator
	discord.SendTyping(ctx.Session, ctx.ChannelID)

	// Extract song information
	song, err := b.extractSongInfo(query)
	if err != nil {
		metrics.RecordAPIRequest("youtube", "extract", false, time.Since(startTime))
		return errors.NewAPIError("could not find or load the requested song", err)
	}

	song.RequesterID = ctx.UserID
	song.RequesterName = ctx.Username

	// Get or create audio connection
	audioConn, err := b.audioPlayer.GetConnection(ctx.Session, ctx.GuildID, voiceState.ChannelID)
	if err != nil {
		return errors.NewDiscordError("failed to connect to voice channel", err)
	}

	// Add to queue
	queue := b.queueManager.GetQueue(ctx.GuildID)
	position := queue.Add(song)

	var response string
	if position == 0 && !queue.IsPlaying() {
		response = fmt.Sprintf("🎵 Now playing: **%s**", song.Title)
		go b.audioPlayer.PlayNext(ctx.Session, ctx.GuildID, audioConn, queue)
	} else {
		response = fmt.Sprintf("🎵 Added to queue: **%s**\nPosition in queue: %d", song.Title, position+1)
	}

	_, err = ctx.Session.ChannelMessageSend(ctx.ChannelID, response)
	
	metrics.RecordCommand("play", ctx.UserID, err == nil, time.Since(startTime))
	return err
}

// handlePauseCommand handles the pause command.
func (b *Bot) handlePauseCommand(ctx *discord.CommandContext) error {
	startTime := time.Now()
	
	queue := b.queueManager.GetQueue(ctx.GuildID)
	if !queue.IsPlaying() {
		return errors.NewValidationError("Nothing is currently playing")
	}

	queue.SetPaused(true)
	_, err := ctx.Session.ChannelMessageSend(ctx.ChannelID, "⏸️ Paused the current song")
	
	metrics.RecordCommand("pause", ctx.UserID, err == nil, time.Since(startTime))
	return err
}

// handleResumeCommand handles the resume command.
func (b *Bot) handleResumeCommand(ctx *discord.CommandContext) error {
	startTime := time.Now()
	
	queue := b.queueManager.GetQueue(ctx.GuildID)
	if !queue.IsPaused() {
		return errors.NewValidationError("Nothing is currently paused")
	}

	queue.SetPaused(false)
	_, err := ctx.Session.ChannelMessageSend(ctx.ChannelID, "▶️ Resumed the current song")
	
	metrics.RecordCommand("resume", ctx.UserID, err == nil, time.Since(startTime))
	return err
}

// handleSkipCommand handles the skip command.
func (b *Bot) handleSkipCommand(ctx *discord.CommandContext) error {
	startTime := time.Now()
	
	queue := b.queueManager.GetQueue(ctx.GuildID)
	if !queue.IsPlaying() {
		return errors.NewValidationError("Nothing is currently playing")
	}

	current := queue.Current()
	queue.Skip()
	
	response := "⏭️ Skipped the current song"
	if current != nil {
		response = fmt.Sprintf("⏭️ Skipped **%s**", current.Title)
	}
	
	_, err := ctx.Session.ChannelMessageSend(ctx.ChannelID, response)
	
	metrics.RecordCommand("skip", ctx.UserID, err == nil, time.Since(startTime))
	return err
}

// handleStopCommand handles the stop command.
func (b *Bot) handleStopCommand(ctx *discord.CommandContext) error {
	startTime := time.Now()
	
	// Disconnect from voice and clear queue
	b.audioPlayer.Disconnect(ctx.GuildID)
	b.queueManager.ClearQueue(ctx.GuildID)
	
	_, err := ctx.Session.ChannelMessageSend(ctx.ChannelID, "⏹️ Stopped music and disconnected from voice channel")
	
	metrics.RecordCommand("stop", ctx.UserID, err == nil, time.Since(startTime))
	return err
}

// handleQueueCommand handles the queue command.
func (b *Bot) handleQueueCommand(ctx *discord.CommandContext) error {
	startTime := time.Now()
	
	queue := b.queueManager.GetQueue(ctx.GuildID)
	
	if queue.Current() == nil && queue.IsEmpty() {
		_, err := ctx.Session.ChannelMessageSend(ctx.ChannelID, "📭 The queue is empty")
		metrics.RecordCommand("queue", ctx.UserID, err == nil, time.Since(startTime))
		return err
	}

	embed := b.buildQueueEmbed(queue)
	_, err := ctx.Session.ChannelMessageSendEmbed(ctx.ChannelID, embed)
	
	metrics.RecordCommand("queue", ctx.UserID, err == nil, time.Since(startTime))
	return err
}

// handleVolumeCommand handles the volume command.
func (b *Bot) handleVolumeCommand(ctx *discord.CommandContext) error {
	startTime := time.Now()
	
	if len(ctx.Args) == 0 {
		// Show current volume
		volume := b.audioPlayer.GetVolume(ctx.GuildID)
		_, err := ctx.Session.ChannelMessageSend(ctx.ChannelID, fmt.Sprintf("🔊 Current volume: %d%%", int(volume*100)))
		metrics.RecordCommand("volume", ctx.UserID, err == nil, time.Since(startTime))
		return err
	}

	// Parse volume level
	var volume int
	if _, err := fmt.Sscanf(ctx.Args[0], "%d", &volume); err != nil {
		return errors.NewValidationError("Please provide a valid volume level (0-100)")
	}

	if volume < 0 || volume > 100 {
		return errors.NewValidationError("Volume must be between 0 and 100")
	}

	volumeFloat := float64(volume) / 100.0
	b.audioPlayer.SetVolume(ctx.GuildID, volumeFloat)
	
	_, err := ctx.Session.ChannelMessageSend(ctx.ChannelID, fmt.Sprintf("🔊 Volume set to %d%%", volume))
	
	metrics.RecordCommand("volume", ctx.UserID, err == nil, time.Since(startTime))
	return err
}

// handlePlaylistCommand handles playlist subcommands.
func (b *Bot) handlePlaylistCommand(ctx *discord.CommandContext) error {
	if b.database == nil {
		return errors.NewValidationError("Playlist functionality is not available")
	}

	if len(ctx.Args) == 0 {
		return errors.NewValidationError("Please specify a playlist subcommand: create, list, show, play, add, remove, delete")
	}

	subcommand := strings.ToLower(ctx.Args[0])
	switch subcommand {
	case "create":
		return b.handlePlaylistCreate(ctx)
	case "list":
		return b.handlePlaylistList(ctx)
	case "show":
		return b.handlePlaylistShow(ctx)
	case "play":
		return b.handlePlaylistPlay(ctx)
	case "add":
		return b.handlePlaylistAdd(ctx)
	case "remove":
		return b.handlePlaylistRemove(ctx)
	case "delete":
		return b.handlePlaylistDelete(ctx)
	default:
		return errors.NewValidationError("Unknown playlist subcommand. Available: create, list, show, play, add, remove, delete")
	}
}

// Helper methods

// getUserVoiceState gets the user's voice state.
func (b *Bot) getUserVoiceState(s *discordgo.Session, guildID, userID string) (*discordgo.VoiceState, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return nil, err
	}

	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs, nil
		}
	}

	return nil, nil
}

// extractSongInfo extracts song information from a query.
func (b *Bot) extractSongInfo(query string) (*Song, error) {
	// This would integrate with yt-dlp or similar
	// For now, return a mock song
	return &Song{
		Title:    query,
		URL:      "https://example.com/" + query,
		Duration: new(int),
	}, nil
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

// Playlist command implementations

func (b *Bot) handlePlaylistCreate(ctx *discord.CommandContext) error {
	startTime := time.Now()
	
	if len(ctx.Args) < 2 {
		return errors.NewValidationError("Please provide a playlist name")
	}

	name := strings.Join(ctx.Args[1:], " ")
	if len(name) > 50 {
		return errors.NewValidationError("Playlist name must be 50 characters or less")
	}

	playlistID, err := b.database.CreatePlaylist(ctx.UserID, ctx.GuildID, name)
	if err != nil {
		logging.LogError(logging.WithComponent("playlist"), err, "Failed to create playlist")
		return errors.NewDatabaseError("failed to create playlist", err)
	}

	response := fmt.Sprintf("✅ Created playlist **%s** (ID: %d)", name, playlistID)
	_, err = ctx.Session.ChannelMessageSend(ctx.ChannelID, response)
	
	metrics.RecordCommand("playlist_create", ctx.UserID, err == nil, time.Since(startTime))
	return err
}

func (b *Bot) handlePlaylistList(ctx *discord.CommandContext) error {
	startTime := time.Now()
	
	playlists, err := b.database.GetUserPlaylists(ctx.UserID, ctx.GuildID)
	if err != nil {
		logging.LogError(logging.WithComponent("playlist"), err, "Failed to list playlists")
		return errors.NewDatabaseError("failed to list playlists", err)
	}

	if len(playlists) == 0 {
		_, err := ctx.Session.ChannelMessageSend(ctx.ChannelID, "📝 You don't have any playlists yet. Use `!playlist create <name>` to make one!")
		metrics.RecordCommand("playlist_list", ctx.UserID, err == nil, time.Since(startTime))
		return err
	}

	embed := discord.CreateEmbed(fmt.Sprintf("🎵 %s's Playlists", ctx.Username), "", "info")
	
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

	_, err = ctx.Session.ChannelMessageSendEmbed(ctx.ChannelID, embed)
	
	metrics.RecordCommand("playlist_list", ctx.UserID, err == nil, time.Since(startTime))
	return err
}

func (b *Bot) handlePlaylistShow(ctx *discord.CommandContext) error {
	// Implementation for showing playlist contents
	return errors.NewValidationError("🚧 Playlist show not yet implemented")
}

func (b *Bot) handlePlaylistPlay(ctx *discord.CommandContext) error {
	// Implementation for playing a playlist
	return errors.NewValidationError("🚧 Playlist play not yet implemented")
}

func (b *Bot) handlePlaylistAdd(ctx *discord.CommandContext) error {
	// Implementation for adding to playlist
	return errors.NewValidationError("🚧 Playlist add not yet implemented")
}

func (b *Bot) handlePlaylistRemove(ctx *discord.CommandContext) error {
	// Implementation for removing from playlist
	return errors.NewValidationError("🚧 Playlist remove not yet implemented")
}

func (b *Bot) handlePlaylistDelete(ctx *discord.CommandContext) error {
	// Implementation for deleting playlist
	return errors.NewValidationError("🚧 Playlist delete not yet implemented")
}