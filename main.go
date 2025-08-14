package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sawyer/discord-bot-framework/internal/bots/clippy"
	"github.com/sawyer/discord-bot-framework/internal/bots/music"
	"github.com/sawyer/discord-bot-framework/internal/config"
	"github.com/sawyer/discord-bot-framework/internal/logging"
	"github.com/spf13/cobra"
)

var (
	botFlag    string
	configFlag string
	debugFlag  bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "discord-bot-framework",
		Short: "A multi-bot Discord framework",
		Long:  "Discord Bot Framework - Run multiple Discord bots from a single application",
		Run:   runBot,
	}

	rootCmd.Flags().StringVarP(&botFlag, "bot", "b", "", "Bot to run (clippy, music, all)")
	rootCmd.Flags().StringVarP(&configFlag, "config", "c", "config.json", "Configuration file path")
	rootCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debug mode")

	rootCmd.MarkFlagRequired("bot")

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runBot(cmd *cobra.Command, args []string) {
	// Initialize logging
	logging.InitializeLogger("INFO", false)
	if debugFlag {
		logging.InitializeLogger("DEBUG", false)
	}

	logger := logging.WithComponent("main")

	// Load configuration
	cfg, err := config.Load(configFlag)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting Discord Bot Framework", "bot", botFlag, "config", configFlag)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigChan
		logger.Info("Received signal, shutting down", "signal", sig.String())
		cancel()
	}()

	switch strings.ToLower(botFlag) {
	case "clippy":
		if err := runClippyBot(ctx, cfg); err != nil {
			logger.Error("Clippy bot failed", "error", err)
			os.Exit(1)
		}
	case "music":
		if err := runMusicBot(ctx, cfg); err != nil {
			logger.Error("Music bot failed", "error", err)
			os.Exit(1)
		}
	case "all":
		if err := runAllBots(ctx, cfg); err != nil {
			logger.Error("Failed to run bots", "error", err)
			os.Exit(1)
		}
	default:
		logger.Error("Invalid bot specified", "bot", botFlag)
		os.Exit(1)
	}

	logger.Info("Bot framework shutdown complete")
}

func runClippyBot(ctx context.Context, cfg *config.Config) error {
	logger := logging.WithComponent("clippy-runner")
	logger.Info("Starting Clippy bot")

	bot, err := clippy.NewBot(cfg.Clippy)
	if err != nil {
		return fmt.Errorf("failed to create Clippy bot: %w", err)
	}

	if err := bot.Start(); err != nil {
		return fmt.Errorf("failed to start Clippy bot: %w", err)
	}

	logger.Info("Clippy bot is running")

	<-ctx.Done()

	logger.Info("Shutting down Clippy bot")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := bot.Stop(shutdownCtx); err != nil {
		logger.Error("Error stopping Clippy bot", "error", err)
	} else {
		logger.Info("Clippy bot stopped successfully")
	}

	return nil
}

func runMusicBot(ctx context.Context, cfg *config.Config) error {
	logger := logging.WithComponent("music-runner")
	logger.Info("Starting Music bot")

	bot, err := music.NewBot(cfg.Music)
	if err != nil {
		return fmt.Errorf("failed to create Music bot: %w", err)
	}

	if err := bot.Start(); err != nil {
		return fmt.Errorf("failed to start Music bot: %w", err)
	}

	logger.Info("Music bot is running")

	<-ctx.Done()

	logger.Info("Shutting down Music bot")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := bot.Stop(shutdownCtx); err != nil {
		logger.Error("Error stopping Music bot", "error", err)
	} else {
		logger.Info("Music bot stopped successfully")
	}

	return nil
}

func runAllBots(ctx context.Context, cfg *config.Config) error {
	logger := logging.WithComponent("all-bots-runner")
	logger.Info("Starting all bots")

	errChan := make(chan error, 2)

	// Start Clippy bot
	go func() {
		if err := runClippyBot(ctx, cfg); err != nil {
			errChan <- fmt.Errorf("clippy bot error: %w", err)
		}
	}()

	// Start Music bot
	go func() {
		if err := runMusicBot(ctx, cfg); err != nil {
			errChan <- fmt.Errorf("music bot error: %w", err)
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, all bots will shut down")
		return nil
	case err := <-errChan:
		return err
	}
}