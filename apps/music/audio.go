// Package main provides audio streaming functionality.
package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sawyer/go-discord-bots/pkg/errors"
	"github.com/sawyer/go-discord-bots/pkg/logging"
	"layeh.com/gopus"
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

	// Build FFmpeg command for Discord voice streaming with optimized settings
	ffmpegArgs := []string{
		"-reconnect", "1", // Enable reconnection for streams
		"-reconnect_streamed", "1", // Reconnect even if stream seems to be CBR
		"-reconnect_delay_max", "5", // Max delay between reconnection attempts
		"-i", as.song.URL,
		"-f", "s16le", // 16-bit signed little endian PCM
		"-ar", "48000", // 48kHz sample rate for Discord
		"-ac", "2", // Stereo
		"-loglevel", "error", // Show errors but not info
		"-bufsize", "64k", // Larger buffer for smoother streaming
		"-fflags", "+genpts", // Generate presentation timestamps
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

	// Monitor FFmpeg process health in a separate goroutine
	go func() {
		err := as.cmd.Wait()
		if err != nil && streamCtx.Err() == nil {
			logger.Error("FFmpeg process exited unexpectedly", "error", err, "guild", as.guildID)
			// Cancel the stream context to stop audio streaming
			cancel()
		}
	}()

	// Start streaming in a goroutine
	go as.streamAudio(stdout)

	return nil
}

// streamAudio streams audio data to Discord voice connection with proper Opus encoding.
func (as *AudioStream) streamAudio(source io.Reader) {
	defer close(as.done)

	logger := logging.WithComponent("audio-stream")

	// Create Opus encoder for Discord voice (48kHz, stereo, audio application)
	opusEncoder, err := gopus.NewEncoder(48000, 2, gopus.Audio)
	if err != nil {
		logger.Error("Failed to create Opus encoder", "error", err)
		return
	}

	// Set Opus encoder options for better quality
	opusEncoder.SetBitrate(128000) // 128 kbps
	// Note: SetComplexity method not available in this gopus version

	// Wait for voice connection to be ready and properly connected
	maxWaitTime := 10 * time.Second
	waitStart := time.Now()
	for !as.conn.Ready || as.conn.OpusSend == nil {
		if time.Since(waitStart) > maxWaitTime {
			logger.Error("Voice connection failed to become ready within timeout", "guild", as.guildID)
			return
		}
		logger.Info("Waiting for voice connection to be ready", "guild", as.guildID,
			"ready", as.conn.Ready, "opus_send_available", as.conn.OpusSend != nil)
		time.Sleep(100 * time.Millisecond)
	}

	logger.Info("Voice connection is ready, setting speaking state", "guild", as.guildID)

	// Set speaking state - CRITICAL: This must be called before sending audio!
	if err := as.conn.Speaking(true); err != nil {
		logger.Error("Failed to set speaking state", "error", err)
		return
	}
	defer func() {
		logger.Info("Clearing speaking state", "guild", as.guildID)
		if err := as.conn.Speaking(false); err != nil {
			logger.Error("Failed to clear speaking state", "error", err)
		}
	}()

	urlPreview := as.song.URL
	if len(urlPreview) > 50 {
		urlPreview = urlPreview[:50] + "..."
	}
	logger.Info("Audio streaming started with Opus encoding", "guild", as.guildID,
		"volume", as.volume, "song_url", urlPreview)

	// Create large buffered reader for smooth streaming
	reader := bufio.NewReaderSize(source, 64*1024) // 64KB buffer

	// Discord expects 20ms of audio per frame at 48kHz stereo
	// For Opus: 48000 Hz * 0.02 seconds = 960 samples per channel per frame
	// For stereo: 960 samples * 2 channels = 1920 total samples
	// For PCM bytes: 1920 samples * 2 bytes = 3840 bytes per frame
	const frameSize = 3840
	const samplesPerFrame = 1920 // Total samples for stereo (960 per channel)
	const opusFrameSize = 960    // Samples per channel for Opus encoder
	pcmBuffer := make([]byte, frameSize)

	for {
		// Check if context is cancelled
		select {
		case <-as.getContext().Done():
			logger.Info("Audio stream cancelled", "guild", as.guildID)
			return
		default:
		}

		// Check if paused
		as.mutex.RLock()
		paused := as.paused
		as.mutex.RUnlock()

		if paused {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Read PCM audio frame from FFmpeg with retry logic
		n, err := io.ReadFull(reader, pcmBuffer)
		if err != nil {
			if err == io.EOF {
				logger.Info("Audio stream finished normally (EOF)", "guild", as.guildID)
				return
			}
			if err == io.ErrUnexpectedEOF {
				logger.Warn("Unexpected EOF, trying to continue", "guild", as.guildID, "bytes_read", n)
				// Try to continue with partial data if we got some
				if n == 0 {
					return
				}
				// Pad the remaining buffer with silence
				for i := n; i < frameSize; i++ {
					pcmBuffer[i] = 0
				}
				n = frameSize
			} else {
				logger.Error("Error reading audio data", "error", err, "guild", as.guildID)
				return
			}
		}

		// Apply volume adjustment if needed
		if as.volume != 1.0 {
			as.adjustVolume(pcmBuffer[:n])
		}

		// Convert PCM bytes to int16 samples for Opus encoder
		samples := make([]int16, n/2)
		for i := 0; i < len(samples); i++ {
			samples[i] = int16(binary.LittleEndian.Uint16(pcmBuffer[i*2 : i*2+2]))
		}

		// Ensure we have exactly the right number of samples for Opus
		if len(samples) != samplesPerFrame {
			// Pad with silence if needed
			if len(samples) < samplesPerFrame {
				paddedSamples := make([]int16, samplesPerFrame)
				copy(paddedSamples, samples)
				samples = paddedSamples
			} else {
				samples = samples[:samplesPerFrame]
			}
		}

		// Encode PCM to Opus with correct frame size
		// Opus expects 960 samples per channel, so for stereo we pass 1920 samples total
		opusData, err := opusEncoder.Encode(samples, opusFrameSize, len(samples)*2) // *2 for bytes
		if err != nil {
			logger.Error("Failed to encode audio to Opus", "error", err, "samples_len", len(samples), "frame_size", opusFrameSize)
			continue
		}

		// Send Opus-encoded audio to Discord
		select {
		case as.conn.OpusSend <- opusData:
			// Successfully sent audio frame (only log every 50 frames to reduce spam)
			// logger.Debug("Sent audio frame", "guild", as.guildID, "opus_size", len(opusData))
		case <-time.After(time.Millisecond * 100):
			logger.Warn("Audio send timeout, skipping frame", "guild", as.guildID)
		case <-as.getContext().Done():
			logger.Info("Audio stream cancelled during send", "guild", as.guildID)
			return
		}

		// Wait exactly 20ms before sending next frame for proper timing
		// Use a timer for more accurate timing than Sleep
		timer := time.NewTimer(20 * time.Millisecond)
		select {
		case <-timer.C:
			// Continue to next frame
		case <-as.getContext().Done():
			timer.Stop()
			logger.Info("Audio stream cancelled during timing wait", "guild", as.guildID)
			return
		}
		timer.Stop()
	}
}

// adjustVolume applies volume adjustment to audio buffer.
func (as *AudioStream) adjustVolume(buffer []byte) {
	// Simple volume adjustment for 16-bit signed little-endian audio
	for i := 0; i < len(buffer)-1; i += 2 {
		// Convert bytes to 16-bit signed integer (little-endian)
		sample := int16(binary.LittleEndian.Uint16(buffer[i : i+2]))

		// Apply volume with clamping to prevent overflow
		newSample := float64(sample) * as.volume
		if newSample > 32767 {
			newSample = 32767
		} else if newSample < -32768 {
			newSample = -32768
		}

		sample = int16(newSample)

		// Convert back to bytes (little-endian)
		binary.LittleEndian.PutUint16(buffer[i:i+2], uint16(sample))
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
