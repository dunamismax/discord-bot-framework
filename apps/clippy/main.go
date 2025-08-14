package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sawyer/discord-bot-framework/apps/clippy/config"
	"github.com/sawyer/discord-bot-framework/apps/clippy/discord"
	"github.com/sawyer/discord-bot-framework/apps/clippy/logging"
	"github.com/sawyer/discord-bot-framework/apps/clippy/metrics"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize logging
	logging.InitializeLogger(cfg.LogLevel, cfg.JSONLogging)
	logger := logging.WithComponent("main")

	logger.Info("Starting Clippy Bot", "version", "2.0.0")

	// Initialize metrics
	metrics.Initialize()

	// Create Discord bot
	bot, err := discord.NewBot(cfg)
	if err != nil {
		logger.Error("Failed to create Discord bot", "error", err)
		os.Exit(1)
	}

	// Start the bot
	if err := bot.Start(); err != nil {
		logger.Error("Failed to start bot", "error", err)
		os.Exit(1)
	}

	logger.Info("Clippy Bot is running. Press Ctrl+C to stop.")

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
		logger.Info("Clippy Bot shutdown completed successfully")
	case <-ctx.Done():
		logger.Warn("Shutdown timeout exceeded, forcing exit")
	}
}