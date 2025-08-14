package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sawyer/discord-bot-framework/internal/config"
	"github.com/sawyer/discord-bot-framework/internal/logging"
)

func main() {
	// Initialize logging
	logging.InitializeLogger("INFO", false)
	logger := logging.WithComponent("clippy-main")

	// Load configuration (defaults + environment variables)
	cfg, err := config.Load("")
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting Clippy bot")

	// Create bot
	bot, err := NewBot(cfg.Clippy)
	if err != nil {
		logger.Error("Failed to create Clippy bot", "error", err)
		os.Exit(1)
	}

	// Start bot
	if err := bot.Start(); err != nil {
		logger.Error("Failed to start Clippy bot", "error", err)
		os.Exit(1)
	}

	logger.Info("Clippy bot is running")

	// Wait for interrupt signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		logger.Info("Received signal, shutting down", "signal", sig.String())
		cancel()
	}()

	<-ctx.Done()

	// Graceful shutdown
	logger.Info("Shutting down Clippy bot")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := bot.Stop(shutdownCtx); err != nil {
		logger.Error("Error stopping Clippy bot", "error", err)
	} else {
		logger.Info("Clippy bot stopped successfully")
	}
}