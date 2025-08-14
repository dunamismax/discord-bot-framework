// Package main provides music queue functionality.
package main

import (
	"sync"
)

// Song represents a song in the music queue.
type Song struct {
	Title         string `json:"title"`
	URL           string `json:"url"`
	WebpageURL    string `json:"webpage_url"`
	Duration      *int   `json:"duration,omitempty"`
	RequesterID   string `json:"requester_id"`
	RequesterName string `json:"requester_name"`
}

// MusicQueue manages the music queue for a guild.
type MusicQueue struct {
	songs      []*Song
	current    *Song
	isPlaying  bool
	isPaused   bool
	shouldSkip bool
	mutex      sync.RWMutex
}

// NewMusicQueue creates a new music queue.
func NewMusicQueue() *MusicQueue {
	return &MusicQueue{
		songs: make([]*Song, 0),
	}
}

// Add adds a song to the queue.
func (q *MusicQueue) Add(song *Song) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.songs = append(q.songs, song)
}

// Next returns and removes the next song from the queue.
func (q *MusicQueue) Next() *Song {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.songs) == 0 {
		return nil
	}

	song := q.songs[0]
	q.songs = q.songs[1:]
	return song
}

// Current returns the currently playing song.
func (q *MusicQueue) Current() *Song {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.current
}

// SetCurrent sets the currently playing song.
func (q *MusicQueue) SetCurrent(song *Song) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.current = song
}

// IsPlaying returns whether music is currently playing.
func (q *MusicQueue) IsPlaying() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.isPlaying
}

// SetPlaying sets the playing status.
func (q *MusicQueue) SetPlaying(playing bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.isPlaying = playing
}

// IsPaused returns whether music is currently paused.
func (q *MusicQueue) IsPaused() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.isPaused
}

// SetPaused sets the paused status.
func (q *MusicQueue) SetPaused(paused bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.isPaused = paused
}

// ShouldSkip returns whether the current song should be skipped.
func (q *MusicQueue) ShouldSkip() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return q.shouldSkip
}

// Skip marks the current song to be skipped.
func (q *MusicQueue) Skip() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.shouldSkip = true
}

// SetSkip sets the skip flag.
func (q *MusicQueue) SetSkip(skip bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.shouldSkip = skip
}

// IsEmpty returns whether the queue is empty.
func (q *MusicQueue) IsEmpty() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.songs) == 0
}

// Size returns the number of songs in the queue.
func (q *MusicQueue) Size() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.songs)
}

// GetSongs returns a copy of all songs in the queue.
func (q *MusicQueue) GetSongs() []*Song {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	songs := make([]*Song, len(q.songs))
	copy(songs, q.songs)
	return songs
}

// Clear clears the entire queue and resets all state.
func (q *MusicQueue) Clear() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.songs = q.songs[:0]
	q.current = nil
	q.isPlaying = false
	q.isPaused = false
	q.shouldSkip = false
}
