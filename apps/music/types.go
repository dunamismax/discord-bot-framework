// Package main provides types for the Music bot.
package main

import (
	"context"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sawyer/go-discord-bots/pkg/errors"
	"github.com/sawyer/go-discord-bots/pkg/logging"
)

// Queue represents a music queue for a guild - uses MusicQueue from queue.go
type Queue = MusicQueue

// NewQueue creates a new music queue.
func NewQueue() *Queue {
	return NewMusicQueue()
}

// QueueManager manages music queues for multiple guilds.
type QueueManager struct {
	queues map[string]*Queue
	mutex  sync.RWMutex
}

// NewQueueManager creates a new queue manager.
func NewQueueManager() *QueueManager {
	return &QueueManager{
		queues: make(map[string]*Queue),
	}
}

// GetQueue gets or creates a queue for a guild.
func (qm *QueueManager) GetQueue(guildID string) *Queue {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	if queue, exists := qm.queues[guildID]; exists {
		return queue
	}
	queue := NewQueue()
	qm.queues[guildID] = queue
	return queue
}

// ClearQueue clears a guild's queue.
func (qm *QueueManager) ClearQueue(guildID string) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	if queue, exists := qm.queues[guildID]; exists {
		queue.Clear()
	}
}

// Cleanup clears all queues.
func (qm *QueueManager) Cleanup() {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	for guildID := range qm.queues {
		if queue, exists := qm.queues[guildID]; exists {
			queue.Clear()
		}
	}
}

// AudioPlayer manages audio playback for multiple guilds.
type AudioPlayer struct {
	volumes     map[string]float64
	connections map[string]*discordgo.VoiceConnection
	enhanced    *EnhancedAudioPlayer
	mutex       sync.RWMutex
}

// NewAudioPlayer creates a new audio player.
func NewAudioPlayer() *AudioPlayer {
	ap := &AudioPlayer{
		volumes:     make(map[string]float64),
		connections: make(map[string]*discordgo.VoiceConnection),
	}
	// Create enhanced player after base player is created
	ap.enhanced = &EnhancedAudioPlayer{
		AudioPlayer: ap,
		streams:     make(map[string]*AudioStream),
	}
	return ap
}

// GetConnection gets or creates a Discord voice connection.
func (ap *AudioPlayer) GetConnection(session *discordgo.Session, guildID, channelID string) (*discordgo.VoiceConnection, error) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	// Check if we already have a connection for this guild
	if conn, exists := ap.connections[guildID]; exists {
		if conn.ChannelID == channelID {
			return conn, nil
		}
		// Different channel, disconnect and reconnect
		if err := conn.Disconnect(); err != nil {
			logger := logging.WithComponent("audio-player")
			logger.Error("Failed to disconnect from voice channel", "error", err)
		}
		delete(ap.connections, guildID)
	}

	// Create new voice connection (mute=false, deaf=false for audio streaming)
	conn, err := session.ChannelVoiceJoin(guildID, channelID, false, false)
	if err != nil {
		return nil, err
	}

	// Wait for connection to be fully ready with timeout
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			logger := logging.WithComponent("audio-player")
			logger.Error("Voice connection timeout", "guild", guildID, "channel", channelID)
			if err := conn.Disconnect(); err != nil {
				logger.Error("Failed to disconnect after timeout", "error", err)
			}
			return nil, errors.NewDiscordError("voice connection timeout", nil)
		case <-ticker.C:
			if conn.Ready {
				ap.connections[guildID] = conn
				return conn, nil
			}
		}
	}
}

// PlayNext plays the next song in queue.
func (ap *AudioPlayer) PlayNext(session *discordgo.Session, guildID string, connection *discordgo.VoiceConnection, queue *Queue) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	logger := logging.WithComponent("audio-player")

	// Check if queue is empty
	if queue.IsEmpty() {
		queue.SetPlaying(false)
		queue.SetCurrent(nil)
		logger.Info("Queue is empty, stopping playback", "guild", guildID)
		return
	}

	// Get next song
	nextSong := queue.Next()
	if nextSong == nil {
		queue.SetPlaying(false)
		queue.SetCurrent(nil)
		return
	}

	// Set current song and playing state
	queue.SetCurrent(nextSong)
	queue.SetPlaying(true)
	queue.SetSkip(false)

	logger.Info("Playing next song", "guild", guildID, "song", nextSong.Title)

	// Start playing the song using enhanced audio player
	go func() {
		// Use a context without timeout for audio streaming since songs can be long
		ctx := context.Background()
		err := ap.enhanced.PlaySong(ctx, guildID, nextSong, connection)
		if err != nil {
			logger.Error("Failed to play song", "error", err, "song", nextSong.Title)
			// Clear current song and try next song on error
			queue.SetCurrent(nil)
			queue.SetPlaying(false)
			ap.PlayNext(session, guildID, connection, queue)
			return
		}

		// Song finished naturally, play next song if not skipped
		logger.Info("Song playback finished", "guild", guildID, "song", nextSong.Title, "should_skip", queue.ShouldSkip())
		if !queue.ShouldSkip() {
			// Clear current song before playing next
			queue.SetCurrent(nil)
			ap.PlayNext(session, guildID, connection, queue)
		} else {
			// Song was skipped, clear states
			queue.SetCurrent(nil)
			queue.SetSkip(false)
			ap.PlayNext(session, guildID, connection, queue)
		}
	}()
}

// Disconnect disconnects from voice channel.
func (ap *AudioPlayer) Disconnect(guildID string) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if conn, exists := ap.connections[guildID]; exists {
		if err := conn.Disconnect(); err != nil {
			logger := logging.WithComponent("audio-player")
			logger.Error("Failed to disconnect from voice channel", "error", err)
		}
		delete(ap.connections, guildID)
	}
}

// GetVolume gets the volume for a guild.
func (ap *AudioPlayer) GetVolume(guildID string) float64 {
	ap.mutex.RLock()
	defer ap.mutex.RUnlock()

	if volume, exists := ap.volumes[guildID]; exists {
		return volume
	}
	return 0.5 // Default volume
}

// SetVolume sets the volume for a guild.
func (ap *AudioPlayer) SetVolume(guildID string, volume float64) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	ap.volumes[guildID] = volume
}

// Cleanup clears all audio connections.
func (ap *AudioPlayer) Cleanup() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	// Clean up enhanced audio player first
	if ap.enhanced != nil {
		ap.enhanced.Cleanup()
	}

	// Disconnect all voice connections
	for guildID, conn := range ap.connections {
		if err := conn.Disconnect(); err != nil {
			logger := logging.WithComponent("audio-player")
			logger.Error("Failed to disconnect from voice channel", "error", err, "guild_id", guildID)
		}
		delete(ap.connections, guildID)
	}
	ap.volumes = make(map[string]float64)
}

// Playlist represents a music playlist.
type Playlist struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	OwnerID string  `json:"owner_id"`
	GuildID string  `json:"guild_id"`
	Songs   []*Song `json:"songs"`
}

// Database provides playlist storage (stub implementation).
type Database struct {
	// This would contain actual database connections
}

// NewDatabase creates a new database connection.
func NewDatabase(url string) (*Database, error) {
	// This would create a real database connection
	return &Database{}, nil
}

// CreatePlaylist creates a new playlist.
func (db *Database) CreatePlaylist(userID, guildID, name string) (int, error) {
	// This would create a playlist in the database
	return 1, nil // Mock playlist ID
}

// GetUserPlaylists gets playlists for a user.
func (db *Database) GetUserPlaylists(userID, guildID string) ([]*Playlist, error) {
	// This would fetch playlists from the database
	return []*Playlist{}, nil
}

// Close closes the database connection.
func (db *Database) Close() error {
	return nil
}
