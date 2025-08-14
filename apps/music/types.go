// Package main provides types for the Music bot.
package main

import "time"

// Song represents a music track.
type Song struct {
	Title         string `json:"title"`
	URL           string `json:"url"`
	Duration      *int   `json:"duration,omitempty"` // Duration in seconds
	RequesterID   string `json:"requester_id"`
	RequesterName string `json:"requester_name"`
}

// Queue represents a music queue for a guild.
type Queue struct {
	current *Song
	songs   []*Song
	playing bool
	paused  bool
	skip    bool
}

// NewQueue creates a new music queue.
func NewQueue() *Queue {
	return &Queue{
		songs: make([]*Song, 0),
	}
}

// Add adds a song to the queue and returns its position.
func (q *Queue) Add(song *Song) int {
	q.songs = append(q.songs, song)
	return len(q.songs) - 1
}

// Next returns the next song in the queue.
func (q *Queue) Next() *Song {
	if len(q.songs) == 0 {
		return nil
	}
	song := q.songs[0]
	q.songs = q.songs[1:]
	return song
}

// Current returns the currently playing song.
func (q *Queue) Current() *Song {
	return q.current
}

// SetCurrent sets the currently playing song.
func (q *Queue) SetCurrent(song *Song) {
	q.current = song
}

// IsEmpty returns true if the queue is empty.
func (q *Queue) IsEmpty() bool {
	return len(q.songs) == 0
}

// IsPlaying returns true if music is playing.
func (q *Queue) IsPlaying() bool {
	return q.playing
}

// SetPlaying sets the playing state.
func (q *Queue) SetPlaying(playing bool) {
	q.playing = playing
}

// IsPaused returns true if music is paused.
func (q *Queue) IsPaused() bool {
	return q.paused
}

// SetPaused sets the paused state.
func (q *Queue) SetPaused(paused bool) {
	q.paused = paused
}

// Skip skips the current song.
func (q *Queue) Skip() {
	q.skip = true
}

// ShouldSkip returns true if the current song should be skipped.
func (q *Queue) ShouldSkip() bool {
	return q.skip
}

// SetSkip sets the skip state.
func (q *Queue) SetSkip(skip bool) {
	q.skip = skip
}

// GetSongs returns a copy of the queue songs.
func (q *Queue) GetSongs() []*Song {
	songs := make([]*Song, len(q.songs))
	copy(songs, q.songs)
	return songs
}

// Clear clears the queue.
func (q *Queue) Clear() {
	q.songs = q.songs[:0]
	q.current = nil
	q.playing = false
	q.paused = false
	q.skip = false
}

// QueueManager manages music queues for multiple guilds.
type QueueManager struct {
	queues map[string]*Queue
}

// NewQueueManager creates a new queue manager.
func NewQueueManager() *QueueManager {
	return &QueueManager{
		queues: make(map[string]*Queue),
	}
}

// GetQueue gets or creates a queue for a guild.
func (qm *QueueManager) GetQueue(guildID string) *Queue {
	if queue, exists := qm.queues[guildID]; exists {
		return queue
	}
	queue := NewQueue()
	qm.queues[guildID] = queue
	return queue
}

// ClearQueue clears a guild's queue.
func (qm *QueueManager) ClearQueue(guildID string) {
	if queue, exists := qm.queues[guildID]; exists {
		queue.Clear()
	}
}

// Cleanup clears all queues.
func (qm *QueueManager) Cleanup() {
	for guildID := range qm.queues {
		qm.ClearQueue(guildID)
	}
}

// AudioPlayer manages audio playback for multiple guilds.
type AudioPlayer struct {
	volumes map[string]float64
}

// NewAudioPlayer creates a new audio player.
func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{
		volumes: make(map[string]float64),
	}
}

// GetConnection gets or creates an audio connection (stub implementation).
func (ap *AudioPlayer) GetConnection(session interface{}, guildID, channelID string) (interface{}, error) {
	// This would create a real Discord voice connection
	return &struct{}{}, nil
}

// PlayNext plays the next song in queue (stub implementation).
func (ap *AudioPlayer) PlayNext(session interface{}, guildID string, connection interface{}, queue *Queue) {
	// This would implement actual audio playback
}

// Disconnect disconnects from voice channel (stub implementation).
func (ap *AudioPlayer) Disconnect(guildID string) {
	// This would disconnect from voice
}

// GetVolume gets the volume for a guild.
func (ap *AudioPlayer) GetVolume(guildID string) float64 {
	if volume, exists := ap.volumes[guildID]; exists {
		return volume
	}
	return 0.5 // Default volume
}

// SetVolume sets the volume for a guild.
func (ap *AudioPlayer) SetVolume(guildID string, volume float64) {
	ap.volumes[guildID] = volume
}

// Cleanup clears all audio connections.
func (ap *AudioPlayer) Cleanup() {
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