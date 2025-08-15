// Package main provides audio streaming functionality.
package main

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sawyer/go-discord-bots/pkg/errors"
	"github.com/sawyer/go-discord-bots/pkg/logging"
)

// AudioStream represents an active audio stream.
type AudioStream struct {
	guildID string
	song    *Song
	conn    *discordgo.VoiceConnection
	cmd     *exec.Cmd
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	volume  float64
	paused  bool
	mutex   sync.RWMutex
}

// NewAudioStream creates a new audio stream.
func NewAudioStream(guildID string, song *Song, conn *discordgo.VoiceConnection, volume float64) *AudioStream {
	return &AudioStream{
		guildID: guildID,
		song:    song,
		conn:    conn,
		done:    make(chan struct{}),
		volume:  volume,
	}
}

// Start starts streaming audio from the song URL.
func (as *AudioStream) Start(ctx context.Context) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	logger := logging.WithComponent("audio-stream")
	logger.Info("Starting audio stream", "guild", as.guildID, "song", as.song.Title)

	// Create cancellable context
	streamCtx, cancel := context.WithCancel(ctx)
	as.ctx = streamCtx
	as.cancel = cancel

	// Check if FFmpeg is available
	if !as.isFFmpegAvailable() {
		return errors.NewAudioError("FFmpeg is not available", nil)
	}

	// Build FFmpeg command for Discord voice streaming
	ffmpegArgs := []string{
		"-i", as.song.URL,
		"-f", "s16le", // 16-bit signed little endian
		"-ar", "48000", // 48kHz sample rate for Discord
		"-ac", "2", // Stereo
		"-loglevel", "panic", // Suppress FFmpeg output
		"-", // Output to stdout
	}

	as.cmd = exec.CommandContext(streamCtx, "ffmpeg", ffmpegArgs...)

	// Get stdout pipe for audio data
	stdout, err := as.cmd.StdoutPipe()
	if err != nil {
		return errors.NewAudioError("failed to create stdout pipe", err)
	}

	// Start FFmpeg process
	if err := as.cmd.Start(); err != nil {
		return errors.NewAudioError("failed to start FFmpeg", err)
	}

	// Start streaming in a goroutine
	go as.streamAudio(stdout)

	return nil
}

// streamAudio streams audio data to Discord voice connection.
func (as *AudioStream) streamAudio(source io.Reader) {
	defer close(as.done)

	logger := logging.WithComponent("audio-stream")

	// Create buffered reader for efficient reading
	reader := bufio.NewReader(source)

	// Discord voice expects 20ms frames (960 samples * 2 channels * 2 bytes = 3840 bytes per frame)
	frameSize := 3840
	buffer := make([]byte, frameSize)

	// Wait for voice connection to be ready
	if err := as.conn.Speaking(true); err != nil {
		logger.Error("Failed to set speaking state", "error", err)
		return
	}
	defer func() {
		if err := as.conn.Speaking(false); err != nil {
			logger := logging.WithComponent("audio-stream")
			logger.Error("Failed to set speaking to false", "error", err)
		}
	}()

	logger.Info("Audio streaming started", "guild", as.guildID)

	for {
		// Check if context is cancelled
		if as.cancel != nil {
			select {
			case <-as.getContext().Done():
				logger.Info("Audio stream cancelled", "guild", as.guildID)
				return
			default:
			}
		}

		// Check if paused
		as.mutex.RLock()
		if as.paused {
			as.mutex.RUnlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		as.mutex.RUnlock()

		// Read audio frame
		n, err := io.ReadFull(reader, buffer)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				logger.Info("Audio stream finished", "guild", as.guildID)
			} else {
				logger.Error("Error reading audio data", "error", err)
			}
			return
		}

		// Apply volume adjustment if needed
		if as.volume != 1.0 {
			as.adjustVolume(buffer[:n])
		}

		// Send audio frame to Discord
		select {
		case as.conn.OpusSend <- buffer[:n]:
			// Frame sent successfully
		case <-time.After(time.Second):
			logger.Warn("Audio send timeout", "guild", as.guildID)
			return
		}

		// Discord expects 20ms between frames
		time.Sleep(20 * time.Millisecond)
	}
}

// adjustVolume applies volume adjustment to audio buffer.
func (as *AudioStream) adjustVolume(buffer []byte) {
	// Simple volume adjustment for 16-bit signed audio
	for i := 0; i < len(buffer)-1; i += 2 {
		// Convert bytes to 16-bit signed integer
		sample := int16(buffer[i]) | int16(buffer[i+1])<<8

		// Apply volume
		sample = int16(float64(sample) * as.volume)

		// Convert back to bytes
		buffer[i] = byte(sample)
		buffer[i+1] = byte(sample >> 8)
	}
}

// Pause pauses the audio stream.
func (as *AudioStream) Pause() {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	as.paused = true
}

// Resume resumes the audio stream.
func (as *AudioStream) Resume() {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	as.paused = false
}

// IsPaused returns whether the stream is paused.
func (as *AudioStream) IsPaused() bool {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	return as.paused
}

// SetVolume sets the volume for the stream.
func (as *AudioStream) SetVolume(volume float64) {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	as.volume = volume
}

// Stop stops the audio stream.
func (as *AudioStream) Stop() {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	logger := logging.WithComponent("audio-stream")
	logger.Info("Stopping audio stream", "guild", as.guildID)

	// Cancel context
	if as.cancel != nil {
		as.cancel()
	}

	// Kill FFmpeg process
	if as.cmd != nil && as.cmd.Process != nil {
		if err := as.cmd.Process.Kill(); err != nil {
			logger := logging.WithComponent("audio-stream")
			logger.Error("Failed to kill FFmpeg process", "error", err)
		}
	}

	// Wait for stream to finish
	select {
	case <-as.done:
		// Stream finished
	case <-time.After(5 * time.Second):
		logger.Warn("Audio stream stop timeout", "guild", as.guildID)
	}
}

// Wait waits for the audio stream to finish.
func (as *AudioStream) Wait() {
	<-as.done
}

// getContext returns the stream context.
func (as *AudioStream) getContext() context.Context {
	if as.ctx == nil {
		return context.Background()
	}
	return as.ctx
}

// isFFmpegAvailable checks if FFmpeg is available.
func (as *AudioStream) isFFmpegAvailable() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// Enhanced AudioPlayer with real streaming support.
type EnhancedAudioPlayer struct {
	*AudioPlayer
	streams map[string]*AudioStream
	mutex   sync.RWMutex
}

// NewEnhancedAudioPlayer creates a new enhanced audio player.
func NewEnhancedAudioPlayer() *EnhancedAudioPlayer {
	return &EnhancedAudioPlayer{
		AudioPlayer: NewAudioPlayer(),
		streams:     make(map[string]*AudioStream),
	}
}

// PlaySong starts playing a song in the specified guild.
func (eap *EnhancedAudioPlayer) PlaySong(ctx context.Context, guildID string, song *Song, conn *discordgo.VoiceConnection) error {
	eap.mutex.Lock()
	defer eap.mutex.Unlock()

	logger := logging.WithComponent("enhanced-audio-player")

	// Stop any existing stream for this guild
	if existingStream, exists := eap.streams[guildID]; exists {
		existingStream.Stop()
		delete(eap.streams, guildID)
	}

	// Get volume for this guild
	volume := eap.GetVolume(guildID)

	// Create new audio stream
	stream := NewAudioStream(guildID, song, conn, volume)
	eap.streams[guildID] = stream

	// Start streaming
	if err := stream.Start(ctx); err != nil {
		delete(eap.streams, guildID)
		return err
	}

	logger.Info("Started playing song", "guild", guildID, "song", song.Title)

	// Wait for stream to finish in a goroutine
	go func() {
		stream.Wait()
		eap.mutex.Lock()
		delete(eap.streams, guildID)
		eap.mutex.Unlock()
		logger.Info("Song finished", "guild", guildID, "song", song.Title)
	}()

	return nil
}

// PauseStream pauses the audio stream for a guild.
func (eap *EnhancedAudioPlayer) PauseStream(guildID string) bool {
	eap.mutex.RLock()
	defer eap.mutex.RUnlock()

	if stream, exists := eap.streams[guildID]; exists {
		stream.Pause()
		return true
	}
	return false
}

// ResumeStream resumes the audio stream for a guild.
func (eap *EnhancedAudioPlayer) ResumeStream(guildID string) bool {
	eap.mutex.RLock()
	defer eap.mutex.RUnlock()

	if stream, exists := eap.streams[guildID]; exists {
		stream.Resume()
		return true
	}
	return false
}

// IsStreamPaused returns whether the stream is paused.
func (eap *EnhancedAudioPlayer) IsStreamPaused(guildID string) bool {
	eap.mutex.RLock()
	defer eap.mutex.RUnlock()

	if stream, exists := eap.streams[guildID]; exists {
		return stream.IsPaused()
	}
	return false
}

// StopStream stops the audio stream for a guild.
func (eap *EnhancedAudioPlayer) StopStream(guildID string) {
	eap.mutex.Lock()
	defer eap.mutex.Unlock()

	if stream, exists := eap.streams[guildID]; exists {
		stream.Stop()
		delete(eap.streams, guildID)
	}
}

// SetStreamVolume sets the volume for a guild's stream.
func (eap *EnhancedAudioPlayer) SetStreamVolume(guildID string, volume float64) {
	// Set the base volume
	eap.SetVolume(guildID, volume)

	// Update active stream volume
	eap.mutex.RLock()
	defer eap.mutex.RUnlock()

	if stream, exists := eap.streams[guildID]; exists {
		stream.SetVolume(volume)
	}
}

// Cleanup cleans up all streams and connections.
func (eap *EnhancedAudioPlayer) Cleanup() {
	eap.mutex.Lock()
	defer eap.mutex.Unlock()

	// Stop all active streams
	for guildID, stream := range eap.streams {
		stream.Stop()
		delete(eap.streams, guildID)
	}

	// Clean up the base audio player
	eap.AudioPlayer.Cleanup()
}
