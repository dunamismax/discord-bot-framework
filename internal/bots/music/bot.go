// Package music provides the Music Discord bot implementation.
package music

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sawyer/discord-bot-framework/internal/config"
	"github.com/sawyer/discord-bot-framework/internal/database"
	"github.com/sawyer/discord-bot-framework/internal/errors"
	"github.com/sawyer/discord-bot-framework/internal/framework"
	"github.com/sawyer/discord-bot-framework/internal/logging"
)

// Bot represents the Music Discord bot.
type Bot struct {
	*framework.Bot
	logger             *logging.Logger
	db                 *database.DB
	queues             map[string]*MusicQueue
	voiceConnections   map[string]*discordgo.VoiceConnection
	inactivityTimers   map[string]*time.Timer
	playerMutex        sync.RWMutex
	audioExtractor     *AudioExtractor
}

// NewBot creates a new Music bot instance.
func NewBot(cfg *config.BotConfig) (*Bot, error) {
	frameworkBot, err := framework.NewBot(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize database if URL is provided
	var db *database.DB
	if cfg.DatabaseURL != "" {
		db, err = database.NewDB(cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
	}

	bot := &Bot{
		Bot:              frameworkBot,
		logger:           logging.WithComponent("music"),
		db:               db,
		queues:           make(map[string]*MusicQueue),
		voiceConnections: make(map[string]*discordgo.VoiceConnection),
		inactivityTimers: make(map[string]*time.Timer),
		audioExtractor:   NewAudioExtractor(),
	}

	// Register commands
	bot.registerCommands()

	return bot, nil
}

// Start starts the Music bot.
func (b *Bot) Start() error {
	if err := b.Bot.Start(); err != nil {
		return err
	}

	b.logger.Info("Music bot started successfully")
	return nil
}

// Stop stops the Music bot.
func (b *Bot) Stop(ctx context.Context) error {
	// Clean up voice connections
	b.playerMutex.Lock()
	for guildID, vc := range b.voiceConnections {
		if vc != nil {
			vc.Disconnect()
		}
		delete(b.voiceConnections, guildID)
	}

	// Clear queues
	for guildID := range b.queues {
		delete(b.queues, guildID)
	}

	// Stop inactivity timers
	for guildID, timer := range b.inactivityTimers {
		if timer != nil {
			timer.Stop()
		}
		delete(b.inactivityTimers, guildID)
	}
	b.playerMutex.Unlock()

	// Close database
	if b.db != nil {
		if err := b.db.Close(); err != nil {
			b.logger.Error("Error closing database", "error", err)
		}
	}

	return b.Bot.Stop(ctx)
}

// registerCommands registers all Music commands.
func (b *Bot) registerCommands() {
	// Basic playback commands
	playCommand := &discordgo.ApplicationCommand{
		Name:        "play",
		Description: "Play music from YouTube",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "query",
				Description: "YouTube URL or search query",
				Required:    true,
			},
		},
	}
	b.RegisterCommand(playCommand, b.handlePlayCommand)

	pauseCommand := &discordgo.ApplicationCommand{
		Name:        "pause",
		Description: "Pause the current song",
	}
	b.RegisterCommand(pauseCommand, b.handlePauseCommand)

	resumeCommand := &discordgo.ApplicationCommand{
		Name:        "resume",
		Description: "Resume the current song",
	}
	b.RegisterCommand(resumeCommand, b.handleResumeCommand)

	skipCommand := &discordgo.ApplicationCommand{
		Name:        "skip",
		Description: "Skip the current song",
	}
	b.RegisterCommand(skipCommand, b.handleSkipCommand)

	stopCommand := &discordgo.ApplicationCommand{
		Name:        "stop",
		Description: "Stop music and clear the queue",
	}
	b.RegisterCommand(stopCommand, b.handleStopCommand)

	queueCommand := &discordgo.ApplicationCommand{
		Name:        "queue",
		Description: "Show the current music queue",
	}
	b.RegisterCommand(queueCommand, b.handleQueueCommand)

	// Playlist commands (only if database is available)
	if b.db != nil {
		b.registerPlaylistCommands()
	}
}

// registerPlaylistCommands registers playlist-related commands.
func (b *Bot) registerPlaylistCommands() {
	// Create playlist command
	createPlaylistCommand := &discordgo.ApplicationCommand{
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
	}
	b.RegisterCommand(createPlaylistCommand, b.handleCreatePlaylistCommand)

	// List playlists command
	listPlaylistsCommand := &discordgo.ApplicationCommand{
		Name:        "playlist_list",
		Description: "List your playlists",
	}
	b.RegisterCommand(listPlaylistsCommand, b.handleListPlaylistsCommand)

	// Show playlist command
	showPlaylistCommand := &discordgo.ApplicationCommand{
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
	}
	b.RegisterCommand(showPlaylistCommand, b.handleShowPlaylistCommand)

	// Play playlist command
	playPlaylistCommand := &discordgo.ApplicationCommand{
		Name:        "playlist_play",
		Description: "Play all songs from a playlist",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "playlist_id",
				Description: "Playlist ID",
				Required:    true,
			},
		},
	}
	b.RegisterCommand(playPlaylistCommand, b.handlePlayPlaylistCommand)

	// Add to playlist command
	addToPlaylistCommand := &discordgo.ApplicationCommand{
		Name:        "playlist_add",
		Description: "Add current song to a playlist",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "playlist_id",
				Description: "Playlist ID",
				Required:    true,
			},
		},
	}
	b.RegisterCommand(addToPlaylistCommand, b.handleAddToPlaylistCommand)

	// Remove from playlist command
	removeFromPlaylistCommand := &discordgo.ApplicationCommand{
		Name:        "playlist_remove",
		Description: "Remove a song from a playlist",
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
				Description: "Song number in playlist",
				Required:    true,
			},
		},
	}
	b.RegisterCommand(removeFromPlaylistCommand, b.handleRemoveFromPlaylistCommand)

	// Delete playlist command
	deletePlaylistCommand := &discordgo.ApplicationCommand{
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
	}
	b.RegisterCommand(deletePlaylistCommand, b.handleDeletePlaylistCommand)
}

// getQueue gets or creates a music queue for a guild.
func (b *Bot) getQueue(guildID string) *MusicQueue {
	b.playerMutex.Lock()
	defer b.playerMutex.Unlock()

	if queue, exists := b.queues[guildID]; exists {
		return queue
	}

	queue := NewMusicQueue()
	b.queues[guildID] = queue
	return queue
}

// handlePlayCommand handles the /play command.
func (b *Bot) handlePlayCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Check if user is in a voice channel
	if i.Member.VoiceState == nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You need to be in a voice channel to use this command!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Defer the response since this might take a while
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return errors.NewDiscordError("failed to defer interaction", err)
	}

	guildID := i.GuildID
	channelID := i.Member.VoiceState.ChannelID
	query := i.ApplicationCommandData().Options[0].StringValue()

	// Get or create voice connection
	vc, err := b.getVoiceConnection(s, guildID, channelID)
	if err != nil {
		return b.followupError(s, i.Interaction, fmt.Sprintf("❌ Failed to connect to voice channel: %v", err))
	}

	// Extract song information
	song, err := b.audioExtractor.ExtractSongInfo(query)
	if err != nil {
		return b.followupError(s, i.Interaction, "❌ Could not find or load the requested song.")
	}

	song.RequesterID = i.Member.User.ID
	song.RequesterName = i.Member.User.Username

	queue := b.getQueue(guildID)
	queue.Add(song)

	if !queue.IsPlaying() {
		_, err = s.FollowupMessageCreate(i.Interaction, &discordgo.WebhookParams{
			Content: fmt.Sprintf("🎵 Now playing: **%s**", song.Title),
		})
		if err != nil {
			b.logger.Error("Failed to send followup message", "error", err)
		}

		// Start playing
		go b.playNext(s, guildID, vc)
	} else {
		_, err = s.FollowupMessageCreate(i.Interaction, &discordgo.WebhookParams{
			Content: fmt.Sprintf("🎵 Added to queue: **%s**\nPosition in queue: %d", song.Title, queue.Size()),
		})
		if err != nil {
			b.logger.Error("Failed to send followup message", "error", err)
		}
	}

	return nil
}

// handlePauseCommand handles the /pause command.
func (b *Bot) handlePauseCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guildID := i.GuildID
	queue := b.getQueue(guildID)

	b.playerMutex.RLock()
	vc, exists := b.voiceConnections[guildID]
	b.playerMutex.RUnlock()

	if !exists || vc == nil || !queue.IsPlaying() {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Nothing is currently playing.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	queue.SetPaused(true)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏸️ Paused the current song.",
		},
	})
}

// handleResumeCommand handles the /resume command.
func (b *Bot) handleResumeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guildID := i.GuildID
	queue := b.getQueue(guildID)

	if !queue.IsPaused() {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Nothing is currently paused.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	queue.SetPaused(false)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "▶️ Resumed the current song.",
		},
	})
}

// handleSkipCommand handles the /skip command.
func (b *Bot) handleSkipCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guildID := i.GuildID
	queue := b.getQueue(guildID)

	if !queue.IsPlaying() {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Nothing is currently playing.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	queue.Skip()

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏭️ Skipped the current song.",
		},
	})
}

// handleStopCommand handles the /stop command.
func (b *Bot) handleStopCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guildID := i.GuildID

	b.playerMutex.Lock()
	vc, exists := b.voiceConnections[guildID]
	if exists && vc != nil {
		vc.Disconnect()
		delete(b.voiceConnections, guildID)
	}

	queue := b.getQueue(guildID)
	queue.Clear()

	// Stop inactivity timer
	if timer, exists := b.inactivityTimers[guildID]; exists && timer != nil {
		timer.Stop()
		delete(b.inactivityTimers, guildID)
	}
	b.playerMutex.Unlock()

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⏹️ Stopped music and cleared the queue. Disconnected from voice channel.",
		},
	})
}

// handleQueueCommand handles the /queue command.
func (b *Bot) handleQueueCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guildID := i.GuildID
	queue := b.getQueue(guildID)

	if queue.Current() == nil && queue.IsEmpty() {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "📭 The queue is empty.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	embed := &discordgo.MessageEmbed{
		Title: "🎵 Music Queue",
		Color: 0x00ff00,
	}

	if current := queue.Current(); current != nil {
		status := "⏸️ Paused" if queue.IsPaused() else "▶️ Playing"
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Now", status),
			Value:  fmt.Sprintf("**%s**\nRequested by: <@%s>", current.Title, current.RequesterID),
			Inline: false,
		})
	}

	songs := queue.GetSongs()
	if len(songs) > 0 {
		var queueList []string
		for i, song := range songs {
			if i >= 10 { // Show first 10 songs
				break
			}
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

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// Database-related command handlers (only available if database is configured)

// handleCreatePlaylistCommand handles the /playlist_create command.
func (b *Bot) handleCreatePlaylistCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if b.db == nil {
		return b.respondError(s, i, "❌ Playlist functionality is not available.")
	}

	name := i.ApplicationCommandData().Options[0].StringValue()
	if len(name) > 50 {
		return b.respondError(s, i, "❌ Playlist name must be 50 characters or less.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	playlistID, err := b.db.CreatePlaylist(ctx, name, i.Member.User.ID, i.GuildID)
	if err != nil {
		b.logger.Error("Error creating playlist", "error", err)
		return b.respondError(s, i, "❌ Failed to create playlist.")
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ Created playlist **%s** (ID: %d)", name, playlistID),
		},
	})
}

// handleListPlaylistsCommand handles the /playlist_list command.
func (b *Bot) handleListPlaylistsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if b.db == nil {
		return b.respondError(s, i, "❌ Playlist functionality is not available.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	playlists, err := b.db.GetUserPlaylists(ctx, i.Member.User.ID, i.GuildID)
	if err != nil {
		b.logger.Error("Error listing playlists", "error", err)
		return b.respondError(s, i, "❌ Failed to list playlists.")
	}

	if len(playlists) == 0 {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "📝 You don't have any playlists yet. Use `/playlist_create` to make one!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🎵 %s's Playlists", i.Member.User.Username),
		Color: 0x00ff00,
	}

	for _, playlist := range playlists {
		if len(embed.Fields) >= 10 { // Discord embed field limit
			break
		}
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

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// handleShowPlaylistCommand handles the /playlist_show command.
func (b *Bot) handleShowPlaylistCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if b.db == nil {
		return b.respondError(s, i, "❌ Playlist functionality is not available.")
	}

	playlistID := int(i.ApplicationCommandData().Options[0].IntValue())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	playlist, err := b.db.GetPlaylist(ctx, playlistID)
	if err != nil {
		b.logger.Error("Error getting playlist", "error", err)
		return b.respondError(s, i, "❌ Failed to get playlist.")
	}

	if playlist == nil {
		return b.respondError(s, i, "❌ Playlist not found.")
	}

	if playlist.GuildID != i.GuildID {
		return b.respondError(s, i, "❌ That playlist belongs to a different server.")
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎵 %s", playlist.Name),
		Description: fmt.Sprintf("Created by <@%s>", playlist.OwnerID),
		Color:       0x00ff00,
	}

	if len(playlist.Songs) == 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Empty Playlist",
			Value:  "No songs added yet.",
			Inline: false,
		})
	} else {
		var songList []string
		for i, song := range playlist.Songs {
			if i >= 10 { // Show first 10 songs
				break
			}
			songList = append(songList, fmt.Sprintf("%d. **%s**", i+1, song.Title))
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Songs (%d total)", len(playlist.Songs)),
			Value:  strings.Join(songList, "\n"),
			Inline: false,
		})

		if len(playlist.Songs) > 10 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "",
				Value:  fmt.Sprintf("... and %d more songs", len(playlist.Songs)-10),
				Inline: false,
			})
		}
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// Helper methods

// getVoiceConnection gets or creates a voice connection for a guild.
func (b *Bot) getVoiceConnection(s *discordgo.Session, guildID, channelID string) (*discordgo.VoiceConnection, error) {
	b.playerMutex.Lock()
	defer b.playerMutex.Unlock()

	if vc, exists := b.voiceConnections[guildID]; exists && vc != nil {
		if vc.ChannelID != channelID {
			// Move to new channel
			if err := vc.ChangeChannel(channelID, false, false); err != nil {
				return nil, err
			}
		}
		return vc, nil
	}

	// Create new connection
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, err
	}

	b.voiceConnections[guildID] = vc
	return vc, nil
}

// playNext plays the next song in the queue.
func (b *Bot) playNext(s *discordgo.Session, guildID string, vc *discordgo.VoiceConnection) {
	queue := b.getQueue(guildID)

	song := queue.Next()
	if song == nil {
		queue.SetPlaying(false)
		// Start inactivity timer
		b.startInactivityTimer(guildID)
		return
	}

	// Stop inactivity timer
	b.stopInactivityTimer(guildID)

	queue.SetCurrent(song)
	queue.SetPlaying(true)

	b.logger.Info("Playing song", "title", song.Title, "guild", guildID)

	// Here you would implement actual audio playback using FFmpeg
	// For this example, we'll simulate playback
	go func() {
		// Simulate song duration (in a real implementation, this would be handled by FFmpeg)
		duration := 3 * time.Minute
		if song.Duration != nil {
			duration = time.Duration(*song.Duration) * time.Second
		}

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		elapsed := time.Duration(0)
		for {
			select {
			case <-ticker.C:
				if queue.IsPaused() {
					continue // Don't increment elapsed time when paused
				}

				elapsed += time.Second
				if elapsed >= duration || queue.ShouldSkip() {
					// Song finished or was skipped
					queue.SetSkip(false)
					b.playNext(s, guildID, vc)
					return
				}
			}
		}
	}()
}

// startInactivityTimer starts the inactivity timer for a guild.
func (b *Bot) startInactivityTimer(guildID string) {
	b.playerMutex.Lock()
	defer b.playerMutex.Unlock()

	// Stop existing timer
	if timer, exists := b.inactivityTimers[guildID]; exists && timer != nil {
		timer.Stop()
	}

	// Start new timer
	timer := time.AfterFunc(b.GetConfig().InactivityTimeout, func() {
		b.playerMutex.Lock()
		if vc, exists := b.voiceConnections[guildID]; exists && vc != nil {
			vc.Disconnect()
			delete(b.voiceConnections, guildID)
		}
		delete(b.inactivityTimers, guildID)
		b.playerMutex.Unlock()

		b.logger.Info("Disconnected due to inactivity", "guild", guildID)
	})

	b.inactivityTimers[guildID] = timer
}

// stopInactivityTimer stops the inactivity timer for a guild.
func (b *Bot) stopInactivityTimer(guildID string) {
	b.playerMutex.Lock()
	defer b.playerMutex.Unlock()

	if timer, exists := b.inactivityTimers[guildID]; exists && timer != nil {
		timer.Stop()
		delete(b.inactivityTimers, guildID)
	}
}

// respondError responds to an interaction with an error message.
func (b *Bot) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// followupError sends a followup error message.
func (b *Bot) followupError(s *discordgo.Session, interaction *discordgo.Interaction, message string) error {
	_, err := s.FollowupMessageCreate(interaction, &discordgo.WebhookParams{
		Content: message,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	return err
}

// Additional playlist command handlers would be implemented here
// handlePlayPlaylistCommand, handleAddToPlaylistCommand, etc.
// These follow similar patterns to the above commands

// Placeholder implementations for the remaining playlist commands
func (b *Bot) handlePlayPlaylistCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Implementation would go here
	return b.respondError(s, i, "🚧 Playlist playback not yet implemented")
}

func (b *Bot) handleAddToPlaylistCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Implementation would go here
	return b.respondError(s, i, "🚧 Add to playlist not yet implemented")
}

func (b *Bot) handleRemoveFromPlaylistCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Implementation would go here
	return b.respondError(s, i, "🚧 Remove from playlist not yet implemented")
}

func (b *Bot) handleDeletePlaylistCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Implementation would go here
	return b.respondError(s, i, "🚧 Delete playlist not yet implemented")
}