//go:build mage

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	packageName = "github.com/sawyer/go-discord-bots"
	binaryName  = "go-discord-bots"
	buildDir    = "bin"
	tmpDir      = "tmp"
)

// Default target to run when none is specified
var Default = Build

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
	defer file.Close()

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
				os.Setenv(key, value)
			}
		}
	}

	return scanner.Err()
}

// Build builds all Discord bot applications
func Build() error {
	fmt.Println("Building Discord bot framework...")

	if err := sh.Run("mkdir", "-p", buildDir); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	// Build the main launcher
	ldflags := "-s -w -X main.version=1.0.0 -X main.buildTime=" + getCurrentTime()
	binaryPath := filepath.Join(buildDir, binaryName)

	// Add .exe extension on Windows
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	fmt.Println("Building main launcher...")
	if err := sh.RunV("go", "build", "-ldflags="+ldflags, "-o", binaryPath, "."); err != nil {
		return fmt.Errorf("failed to build main launcher: %w", err)
	}

	// Build individual apps
	apps := []string{"clippy", "music", "mtg-card-bot"}
	for _, app := range apps {
		fmt.Printf("Building %s app...\n", app)
		appBinaryPath := filepath.Join(buildDir, app)
		if runtime.GOOS == "windows" {
			appBinaryPath += ".exe"
		}

		appDir := filepath.Join("apps", app)
		if err := sh.RunV("go", "build", "-ldflags="+ldflags, "-o", appBinaryPath, "./"+appDir); err != nil {
			return fmt.Errorf("failed to build %s app: %w", app, err)
		}
	}

	fmt.Println("All builds completed successfully!")
	return nil
}

func getCurrentTime() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// getGoBinaryPath finds the path to a Go binary, checking GOBIN, GOPATH/bin, and PATH
func getGoBinaryPath(binaryName string) (string, error) {
	// First check if it's in PATH
	if err := sh.Run("which", binaryName); err == nil {
		return binaryName, nil
	}

	// Check GOBIN first
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		binaryPath := filepath.Join(gobin, binaryName)
		if _, err := os.Stat(binaryPath); err == nil {
			return binaryPath, nil
		}
	}

	// Check GOPATH/bin
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		if home := os.Getenv("HOME"); home != "" {
			gopath = filepath.Join(home, "go")
		}
	}

	if gopath != "" {
		binaryPath := filepath.Join(gopath, "bin", binaryName)
		if _, err := os.Stat(binaryPath); err == nil {
			return binaryPath, nil
		}
	}

	return "", fmt.Errorf("%s not found in PATH, GOBIN, or GOPATH/bin", binaryName)
}

// Fmt formats and tidies code using goimports and standard tooling
func Fmt() error {
	fmt.Println("Formatting and tidying...")

	// Tidy go modules
	if err := sh.RunV("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to tidy modules: %w", err)
	}

	// Use goimports for better import management and formatting
	fmt.Println("  Running goimports...")
	goimportsPath, err := getGoBinaryPath("goimports")
	if err != nil {
		fmt.Printf("Warning: goimports not found, falling back to go fmt: %v\n", err)
		if err := sh.RunV("go", "fmt", "./..."); err != nil {
			return fmt.Errorf("failed to format code: %w", err)
		}
	} else {
		if err := sh.RunV(goimportsPath, "-w", "."); err != nil {
			fmt.Printf("Warning: goimports failed, falling back to go fmt: %v\n", err)
			if err := sh.RunV("go", "fmt", "./..."); err != nil {
				return fmt.Errorf("failed to format code: %w", err)
			}
		}
	}

	return nil
}

// Vet analyzes code for common errors
func Vet() error {
	fmt.Println("Running go vet...")
	return sh.RunV("go", "vet", "./...")
}

// VulnCheck scans for known vulnerabilities
func VulnCheck() error {
	fmt.Println("Running vulnerability check...")
	govulncheckPath, err := getGoBinaryPath("govulncheck")
	if err != nil {
		return fmt.Errorf("govulncheck not found: %w", err)
	}
	return sh.RunV(govulncheckPath, "./...")
}

// Lint runs golangci-lint with comprehensive linting rules
func Lint() error {
	fmt.Println("Running golangci-lint...")

	// Find golangci-lint binary
	lintPath, err := getGoBinaryPath("golangci-lint")
	if err != nil {
		fmt.Println("Installing golangci-lint...")
		if err := sh.RunV("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"); err != nil {
			return fmt.Errorf("failed to install golangci-lint: %w", err)
		}
		// Try to find it again after installation
		lintPath, err = getGoBinaryPath("golangci-lint")
		if err != nil {
			return fmt.Errorf("golangci-lint not found after installation: %w", err)
		}
	}

	return sh.RunV(lintPath, "run", "--disable=whitespace,wsl,nlreturn,wsl_v5", "./...")
}

// RunClipper builds and runs only the Clippy bot
func RunClipper() error {
	if err := loadEnvFile(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	mg.SerialDeps(Build)
	fmt.Println("Starting Clippy bot...")

	binaryPath := filepath.Join(buildDir, "clippy")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	return sh.RunV(binaryPath)
}

// RunMusic builds and runs only the Music bot
func RunMusic() error {
	if err := loadEnvFile(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	mg.SerialDeps(Build)
	fmt.Println("Starting Music bot...")

	binaryPath := filepath.Join(buildDir, "music")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	return sh.RunV(binaryPath)
}

// RunMTG builds and runs only the MTG Card bot
func RunMTG() error {
	if err := loadEnvFile(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	mg.SerialDeps(Build)
	fmt.Println("Starting MTG Card bot...")

	binaryPath := filepath.Join(buildDir, "mtg-card-bot")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	return sh.RunV(binaryPath)
}

// Run builds and runs all Discord bots
func Run() error {
	if err := loadEnvFile(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	mg.SerialDeps(Build)
	fmt.Println("Starting all Discord bots...")

	binaryPath := filepath.Join(buildDir, binaryName)
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	return sh.RunV(binaryPath, "--bot", "all")
}

// Dev starts development with debug logging
func Dev() error {
	if err := loadEnvFile(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	mg.SerialDeps(Build)
	fmt.Println("Starting Discord bot framework in development mode...")

	binaryPath := filepath.Join(buildDir, binaryName)
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	return sh.RunV(binaryPath, "--bot", "all", "--debug")
}

// Clean removes built binaries and generated files
func Clean() error {
	fmt.Println("Cleaning up...")

	// Remove build directory
	if err := sh.Rm(buildDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove build directory: %w", err)
	}

	// Remove tmp directory
	if err := sh.Rm(tmpDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove tmp directory: %w", err)
	}

	// Remove any SQLite databases created during testing
	databases := []string{"music.db", "clippy.db", "bot.db"}
	for _, db := range databases {
		if err := sh.Rm(db); err == nil {
			fmt.Printf("  Removed database file: %s\n", db)
		}
	}

	fmt.Println("Clean complete!")
	return nil
}

// Reset completely resets the repository to a fresh state
func Reset() error {
	fmt.Println("Resetting repository to clean state...")

	// First run clean to remove built artifacts
	if err := Clean(); err != nil {
		return fmt.Errorf("failed to clean build artifacts: %w", err)
	}

	// Remove any log files
	fmt.Println("Removing log files...")
	logFiles := []string{"bot.log", "clippy.log", "music.log", "discord.log"}
	for _, logFile := range logFiles {
		if err := sh.Rm(logFile); err == nil {
			fmt.Printf("  Removed log file: %s\n", logFile)
		}
	}

	fmt.Println("Reset complete! Repository is now in fresh state.")
	fmt.Println("You can now run 'mage dev' or 'mage run' to start the bots.")
	return nil
}

// Setup installs required development tools
func Setup() error {
	fmt.Println("Setting up development environment...")

	tools := map[string]string{
		"govulncheck":   "golang.org/x/vuln/cmd/govulncheck@latest",
		"goimports":     "golang.org/x/tools/cmd/goimports@latest",
		"golangci-lint": "github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
	}

	for tool, pkg := range tools {
		fmt.Printf("  Installing %s...\n", tool)
		if err := sh.RunV("go", "install", pkg); err != nil {
			return fmt.Errorf("failed to install %s: %w", tool, err)
		}
	}

	// Download module dependencies
	fmt.Println("Downloading dependencies...")
	if err := sh.RunV("go", "mod", "download"); err != nil {
		return fmt.Errorf("failed to download dependencies: %w", err)
	}

	// Create example .env file if it doesn't exist
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		fmt.Println("Creating example .env file...")
		envContent := `# Discord Bot Framework Environment Variables

# Global settings (applies to all bots unless overridden)
DISCORD_TOKEN=your_default_discord_token_here
COMMAND_PREFIX=!
LOG_LEVEL=info
JSON_LOGGING=false
DEBUG=false

# Guild ID (can be overridden per bot)
GUILD_ID=your_guild_id_for_testing

# Timeout configurations
SHUTDOWN_TIMEOUT=30s
REQUEST_TIMEOUT=30s
MAX_RETRIES=3

# Clippy Bot Configuration
CLIPPY_DISCORD_TOKEN=your_clippy_bot_token_here
CLIPPY_GUILD_ID=your_guild_id_for_testing
RANDOM_RESPONSES=true
RANDOM_INTERVAL=45m
RANDOM_MESSAGE_DELAY=3s

# Music Bot Configuration  
MUSIC_DISCORD_TOKEN=your_music_bot_token_here
MUSIC_GUILD_ID=your_guild_id_for_testing
MUSIC_DATABASE_URL=music.db
MAX_QUEUE_SIZE=100
INACTIVITY_TIMEOUT=5m
VOLUME_LEVEL=0.5

# MTG Card Bot Configuration
# MTG bot will use DISCORD_TOKEN if no specific token is provided
CACHE_TTL=1h
CACHE_SIZE=1000
`
		if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
			fmt.Printf("Warning: failed to create .env file: %v\n", err)
		} else {
			fmt.Println("  Created .env file with example configuration")
		}
	}

	fmt.Println("Setup complete!")
	fmt.Println("Next steps:")
	fmt.Println("   • Edit .env file with your Discord bot tokens")
	fmt.Println("   • For music bot: ensure yt-dlp and FFmpeg are installed")
	fmt.Println("   • Run 'mage dev' to start development with debug logging")
	fmt.Println("   • Run 'mage run' to start all bots in production mode")

	return nil
}

// Test runs the test suite
func Test() error {
	fmt.Println("Running test suite...")
	return sh.RunV("go", "test", "./...")
}

// CI runs the complete CI pipeline
func CI() error {
	fmt.Println("Running complete CI pipeline...")
	mg.SerialDeps(Fmt, Vet, Lint, Test, Build, showBuildInfo)
	return nil
}

// Quality runs all quality checks
func Quality() error {
	fmt.Println("Running all quality checks...")
	mg.Deps(Vet, Lint, VulnCheck)
	return nil
}

// Help prints a help message with available commands
func Help() {
	fmt.Println(`
Discord Bot Framework Magefile

Available commands:

Development:
  mage setup (s)        Install all development tools and dependencies
  mage dev (d)          Build and run all bots in development mode (with debug)
  mage run (r)          Build and run all bots in production mode
  mage runClipper (rc)  Build and run only Clippy bot
  mage runMusic (rm)    Build and run only Music bot
  mage runMTG (rmtg)    Build and run only MTG Card bot
  mage build (b)        Build production binary

Quality:
  mage fmt (f)          Format code with goimports and tidy modules
  mage vet (v)          Run go vet static analysis
  mage lint (l)         Run golangci-lint comprehensive linting
  mage vulnCheck (vc)   Check for security vulnerabilities
  mage test (t)         Run test suite
  mage quality (q)      Run all quality checks (vet + lint + vulncheck)

Production:
  mage ci               Complete CI pipeline (fmt + vet + lint + test + build)
  mage clean (c)        Clean build artifacts and temporary files
  mage reset            Reset repository to fresh state (clean + remove logs/databases)

Other:
  mage help (h)         Show this help message

Bot Commands (when running):
  --bot clippy          Run only Clippy bot
  --bot music           Run only Music bot  
  --bot all             Run all bots (default)
  --debug               Enable debug mode
  --config <file>       Use custom config file
    `)
}

// showBuildInfo displays information about the built binary
func showBuildInfo() error {
	binaryPath := filepath.Join(buildDir, binaryName)
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("binary not found: %s", binaryPath)
	}

	fmt.Println("\nBuild Information:")

	// Show binary size
	if info, err := os.Stat(binaryPath); err == nil {
		size := info.Size()
		fmt.Printf("   Binary size: %.2f MB\n", float64(size)/1024/1024)
	}

	// Show Go version
	if version, err := sh.Output("go", "version"); err == nil {
		fmt.Printf("   Go version: %s\n", version)
	}

	fmt.Printf("   Binary path: %s\n", binaryPath)

	return nil
}

// Aliases for common commands
var Aliases = map[string]interface{}{
	"b":    Build,
	"f":    Fmt,
	"v":    Vet,
	"l":    Lint,
	"vc":   VulnCheck,
	"t":    Test,
	"r":    Run,
	"rc":   RunClipper,
	"rm":   RunMusic,
	"rmtg": RunMTG,
	"d":    Dev,
	"c":    Clean,
	"s":    Setup,
	"q":    Quality,
	"h":    Help,
}
