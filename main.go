package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	_ = rootCmd.MarkFlagRequired("bot")

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runBot(cmd *cobra.Command, args []string) {
	fmt.Printf("Starting Discord Bot Framework - Bot: %s\n", botFlag)

	// Get the directory where the main binary is located
	binaryDir := filepath.Dir(os.Args[0])
	if binaryDir == "." {
		// If running with 'go run', use current directory
		binaryDir, _ = os.Getwd()
	}

	switch strings.ToLower(botFlag) {
	case "clippy":
		if err := runSingleApp("clippy", binaryDir); err != nil {
			fmt.Printf("Error running Clippy bot: %v\n", err)
			os.Exit(1)
		}
	case "music":
		if err := runSingleApp("music", binaryDir); err != nil {
			fmt.Printf("Error running Music bot: %v\n", err)
			os.Exit(1)
		}
	case "mtg":
		if err := runSingleApp("mtg-card-bot", binaryDir); err != nil {
			fmt.Printf("Error running MTG Card bot: %v\n", err)
			os.Exit(1)
		}
	case "all":
		if err := runAllApps(binaryDir); err != nil {
			fmt.Printf("Error running all bots: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Invalid bot specified: %s\n", botFlag)
		fmt.Println("Valid options: clippy, music, mtg, all")
		os.Exit(1)
	}
}

func runSingleApp(appName, binaryDir string) error {
	appPath := filepath.Join("bin", appName)
	
	// Check if binary exists
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return fmt.Errorf("app binary not found: %s (run 'mage build' first)", appPath)
	}

	fmt.Printf("Starting %s bot...\n", appName)
	
	cmd := exec.Command(appPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func runAllApps(binaryDir string) error {
	apps := []string{"clippy", "music", "mtg-card-bot"}
	
	fmt.Println("Starting all bots...")
	
	// Start all apps concurrently
	for _, app := range apps {
		appPath := filepath.Join("bin", app)
		
		// Check if binary exists
		if _, err := os.Stat(appPath); os.IsNotExist(err) {
			fmt.Printf("Warning: %s binary not found, skipping\n", app)
			continue
		}

		go func(appName string) {
			fmt.Printf("Starting %s...\n", appName)
			cmd := exec.Command(filepath.Join("bin", appName))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Error running %s: %v\n", appName, err)
			}
		}(app)
	}

	// Keep the main process running
	select {}
}
