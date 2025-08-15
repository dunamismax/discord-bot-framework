// Package main provides audio extraction functionality.
package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sawyer/go-discord-bots/pkg/errors"
)

// AudioExtractor handles extracting audio information from various sources.
type AudioExtractor struct {
	// Add any configuration or state here
}

// NewAudioExtractor creates a new audio extractor.
func NewAudioExtractor() *AudioExtractor {
	return &AudioExtractor{}
}

// ExtractSongInfo extracts song information from a URL or search query.
func (e *AudioExtractor) ExtractSongInfo(query string) (*Song, error) {
	// Check if yt-dlp is available
	if !e.isYtDlpAvailable() {
		return nil, errors.NewAudioError("yt-dlp is not available", nil)
	}

	// Prepare the command with audio format selection
	var cmd *exec.Cmd
	// Common yt-dlp arguments for better reliability and YouTube compatibility
	baseArgs := []string{
		"--dump-json",
		"--no-playlist",
		"--format", "bestaudio/best", // Get the best audio available
		"--socket-timeout", "30", // Shorter timeout to fail faster
		"--retries", "2", // Fewer retries to avoid long waits
		"--fragment-retries", "2", // Fewer fragment retries
		"--extractor-retries", "2", // Extractor retries
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", // Better user agent
		"--referer", "https://www.youtube.com/", // Add referer for YouTube
	}

	if strings.HasPrefix(query, "http://") || strings.HasPrefix(query, "https://") {
		// Direct URL with audio format preference
		args := append(baseArgs, query)
		cmd = exec.Command("yt-dlp", args...)
	} else {
		// Search query with audio format preference
		args := append(baseArgs, fmt.Sprintf("ytsearch:%s", query))
		cmd = exec.Command("yt-dlp", args...)
	}

	// Execute the command with better error capture
	output, err := cmd.Output()
	if err != nil {
		// Try to get stderr for better error information
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr := string(exitError.Stderr)
			return nil, errors.NewAudioError(fmt.Sprintf("yt-dlp failed: %s", stderr), err)
		}
		return nil, errors.NewAudioError("failed to extract song info", err)
	}

	// Parse the JSON output
	var info map[string]interface{}
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, errors.NewAudioError("failed to parse song info", err)
	}

	// Extract relevant information
	song := &Song{
		Title:      getStringFromMap(info, "title", "Unknown"),
		WebpageURL: getStringFromMap(info, "webpage_url", ""),
	}

	// Extract duration if available
	if duration, ok := info["duration"].(float64); ok {
		durationInt := int(duration)
		song.Duration = &durationInt
	}

	// Try to get the best audio URL from formats
	song.URL = extractBestAudioURL(info)

	if song.URL == "" || song.Title == "" {
		return nil, errors.NewAudioError("incomplete song information or no audio URL found", nil)
	}

	return song, nil
}

// isYtDlpAvailable checks if yt-dlp is available in the system.
func (e *AudioExtractor) isYtDlpAvailable() bool {
	_, err := exec.LookPath("yt-dlp")
	return err == nil
}

// getStringFromMap safely extracts a string value from a map.
func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}

// extractBestAudioURL finds the best audio URL from available formats
func extractBestAudioURL(info map[string]interface{}) string {
	// First try the direct URL field
	if url := getStringFromMap(info, "url", ""); url != "" {
		return url
	}

	// Look through formats for audio-only streams
	formats, ok := info["formats"].([]interface{})
	if !ok {
		return ""
	}

	// Priority order: audio-only > lowest video quality with audio
	var bestAudioURL, fallbackURL string

	for _, format := range formats {
		fmt, ok := format.(map[string]interface{})
		if !ok {
			continue
		}

		url := getStringFromMap(fmt, "url", "")
		if url == "" {
			continue
		}

		// Check if this is an audio-only format
		vcodec := getStringFromMap(fmt, "vcodec", "")
		acodec := getStringFromMap(fmt, "acodec", "")
		ext := getStringFromMap(fmt, "ext", "")

		// Prefer audio-only formats (vcodec=none but acodec exists)
		if vcodec == "none" && acodec != "none" && acodec != "" {
			bestAudioURL = url
			break // Found audio-only, this is best
		}

		// Fallback to formats with audio (even if they have video)
		if acodec != "none" && acodec != "" && ext != "mhtml" {
			fallbackURL = url
		}
	}

	if bestAudioURL != "" {
		return bestAudioURL
	}
	return fallbackURL
}
