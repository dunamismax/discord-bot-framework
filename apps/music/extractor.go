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

	// Prepare the command
	var cmd *exec.Cmd
	if strings.HasPrefix(query, "http://") || strings.HasPrefix(query, "https://") {
		// Direct URL
		cmd = exec.Command("yt-dlp", "--dump-json", "--no-playlist", query)
	} else {
		// Search query
		cmd = exec.Command("yt-dlp", "--dump-json", "--no-playlist", fmt.Sprintf("ytsearch:%s", query))
	}

	// Execute the command
	output, err := cmd.Output()
	if err != nil {
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
		URL:        getStringFromMap(info, "url", ""),
		WebpageURL: getStringFromMap(info, "webpage_url", ""),
	}

	// Extract duration if available
	if duration, ok := info["duration"].(float64); ok {
		durationInt := int(duration)
		song.Duration = &durationInt
	}

	if song.URL == "" || song.Title == "" {
		return nil, errors.NewAudioError("incomplete song information", nil)
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
