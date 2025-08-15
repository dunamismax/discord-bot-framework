package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sawyer/go-discord-bots/pkg/config"
	"github.com/sawyer/go-discord-bots/pkg/logging"
	"github.com/sawyer/go-discord-bots/pkg/metrics"
)

// loadEnvFile loads environment variables from .env file if it exists
func loadEnvFile() error {
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		// .env file doesn't exist, that's okay
		return nil
	}

	file, err := os.Open(envFile)
	if err != nil {
		return fmt.Errorf("failed to open .env file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close .env file: %v", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes if present
			if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
				(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
				value = value[1 : len(value)-1]
			}

			// Only set if not already set by system environment
			if os.Getenv(key) == "" {
				if err := os.Setenv(key, value); err != nil {
					log.Printf("Warning: failed to set environment variable %s: %v", key, err)
				}
			}
		}
	}

	return scanner.Err()
}

func main() {
	// Load environment variables from .env file
	if err := loadEnvFile(); err != nil {
		log.Printf("Warning: failed to load .env file: %v", err)
	}

	// Load configuration
	cfg, err := config.Load(config.BotTypeMusic)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize logging
	logging.InitializeLogger(cfg.LogLevel, cfg.JSONLogging)
	logger := logging.WithComponent("main")

	logger.Info("Starting Music Bot", "version", "2.0.0")

	// Initialize metrics
	metrics.Initialize(cfg.BotName, string(cfg.BotType))

	// Create bot
	bot, err := NewBot(cfg)
	if err != nil {
		logger.Error("Failed to create Music bot", "error", err)
		os.Exit(1)
	}

	// Start bot
	if err := bot.Start(); err != nil {
		logger.Error("Failed to start Music bot", "error", err)
		os.Exit(1)
	}

	logger.Info("Music Bot is running. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Received shutdown signal. Gracefully shutting down...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Stop the bot
	if err := bot.Stop(); err != nil {
		logger.Error("Error during bot shutdown", "error", err)
	}

	// Wait for shutdown to complete or timeout
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Additional cleanup can be added here if needed
		time.Sleep(1 * time.Second) // Give some time for cleanup
	}()

	select {
	case <-done:
		logger.Info("Music Bot shutdown completed successfully")
	case <-ctx.Done():
		logger.Warn("Shutdown timeout exceeded, forcing exit")
	}
}
