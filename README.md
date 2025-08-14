# Discord Bot Framework (Go)

A modern, scalable Discord bot framework written in Go, supporting multiple bot instances with shared infrastructure. This is a complete rewrite of the original Python-based framework, following patterns from the MTG Card Bot.

<p align="center">
  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.23+-00ADD8.svg?logo=go&logoColor=white&style=for-the-badge" alt="Go Version"></a>
  <a href="https://github.com/bwmarrin/discordgo"><img src="https://img.shields.io/badge/DiscordGo-0.27.1+-5865F2.svg?logo=discord&logoColor=white&style=for-the-badge" alt="DiscordGo Version"></a>
  <a href="https://sqlite.org/"><img src="https://img.shields.io/badge/SQLite-3.x-5865F2.svg?logo=sqlite&logoColor=white&style=for-the-badge" alt="SQLite"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-5865F2.svg?style=for-the-badge&logoColor=white" alt="MIT License"></a>
</p>

## Features

### Core Framework

- **Multi-bot architecture** - Run multiple Discord bots from a single application
- **Structured logging** - Comprehensive logging with configurable levels and formats
- **Configuration management** - JSON-based configuration with environment variable overrides
- **Error handling** - Comprehensive error types and handling
- **Graceful shutdown** - Proper cleanup and shutdown handling
- **Command cooldowns** - Per-user, per-command cooldown system

### Clippy Bot

- **Unhinged responses** - Classic Microsoft Clippy with modern chaotic energy
- **Random messages** - Periodic random responses in channels
- **Slash commands** - `/clippy`, `/clippy_wisdom`, `/clippy_help`
- **Interactive elements** - Buttons and embeds for enhanced user experience

### Music Bot

- **YouTube playback** - Play music from YouTube URLs or search queries
- **Queue management** - Add, skip, pause, resume, and view queue
- **Playlist system** - Create, manage, and play custom playlists (with database)
- **Voice management** - Automatic connection handling and inactivity timeouts
- **Database support** - SQLite database for persistent playlist storage

## Project Structure

```
discord-bot-framework/
├── main.go                           # Main application entry point
├── go.mod                           # Go module definition
├── config.example.json              # Example configuration file
├── internal/
│   ├── config/
│   │   └── config.go                # Configuration loading and validation
│   ├── logging/
│   │   └── logger.go                # Structured logging functionality
│   ├── errors/
│   │   └── errors.go                # Error types and handling
│   ├── framework/
│   │   └── bot.go                   # Core Discord bot framework
│   ├── database/
│   │   └── database.go              # Database functionality (SQLite)
│   └── bots/
│       ├── clippy/
│       │   └── bot.go               # Clippy bot implementation
│       └── music/
│           ├── bot.go               # Music bot implementation
│           ├── queue.go             # Music queue management
│           └── extractor.go         # Audio extraction (yt-dlp)
└── README.md                        # This file
```

## Quick Start

### Prerequisites

1. **Go 1.23+** - [Download Go](https://golang.org/dl/)
2. **Discord Bot Token(s)** - Create bot applications at [Discord Developer Portal](https://discord.com/developers/applications)
3. **yt-dlp** (for music bot) - `pip install yt-dlp` or download from [yt-dlp releases](https://github.com/yt-dlp/yt-dlp/releases)
4. **FFmpeg** (for music bot) - [Download FFmpeg](https://ffmpeg.org/download.html)

### Installation

1. **Clone the repository:**

   ```bash
   git clone <repository-url>
   cd discord-bot-framework
   ```

2. **Install Mage (if not already installed):**

   ```bash
   go install github.com/magefile/mage@latest
   ```

3. **Setup development environment:**

   ```bash
   # This will install all tools, download dependencies, and create .env file
   mage setup
   ```

4. **Configure your bots:**

   ```bash
   # Edit the .env file with your Discord bot tokens
   nano .env
   ```

### Configuration

The framework uses environment variables for configuration. The `.env` file is created automatically during setup:

```bash
# Discord Bot Framework Environment Variables

# Clippy Bot Configuration
CLIPPY_DISCORD_TOKEN=your_clippy_bot_token_here
CLIPPY_GUILD_ID=your_guild_id_for_testing
CLIPPY_DEBUG=false

# Music Bot Configuration  
MUSIC_DISCORD_TOKEN=your_music_bot_token_here
MUSIC_GUILD_ID=your_guild_id_for_testing
MUSIC_DEBUG=false
MUSIC_DATABASE_URL=music.db

# Optional: Global settings
BOT_LOG_LEVEL=INFO
BOT_JSON_LOGGING=false
```

You can also use JSON configuration by creating a `config.json` file - see `config.example.json` for the format.

### Environment Variables

You can override configuration with environment variables:

```bash
# Clippy Bot
export CLIPPY_DISCORD_TOKEN="your_token_here"
export CLIPPY_GUILD_ID="your_guild_id"
export CLIPPY_DEBUG="true"

# Music Bot
export MUSIC_DISCORD_TOKEN="your_token_here"
export MUSIC_GUILD_ID="your_guild_id"
export MUSIC_DATABASE_URL="./music.db"
export MUSIC_DEBUG="true"
```

### Building and Running

The project uses [Mage](https://magefile.org/) for build automation and task management.

1. **Setup development environment:**

   ```bash
   # Install development tools and create .env file
   mage setup
   ```

2. **Build the application:**

   ```bash
   mage build
   ```

3. **Run the bots:**

   ```bash
   # Run all bots in production mode
   mage run
   
   # Run all bots in development mode (with debug logging)
   mage dev
   
   # Run only Clippy bot
   mage runClipper
   
   # Run only Music bot
   mage runMusic
   ```

4. **Quality checks:**

   ```bash
   # Format code and tidy modules
   mage fmt
   
   # Run all quality checks (vet, lint, vulnerability scan)
   mage quality
   
   # Complete CI pipeline
   mage ci
   ```

5. **Additional commands:**

   ```bash
   # Clean build artifacts
   mage clean
   
   # Show all available commands
   mage help
   ```

## Bot Commands

### Clippy Bot Commands

- `/clippy` - Get an unhinged Clippy response
- `/clippy_wisdom` - Receive Clippy's questionable wisdom
- `/clippy_help` - Get help from Clippy (if you dare)

### Music Bot Commands

**Basic Playback:**

- `/play <query>` - Play music from YouTube URL or search query
- `/pause` - Pause the current song
- `/resume` - Resume the current song
- `/skip` - Skip the current song
- `/stop` - Stop music and clear the queue
- `/queue` - Show the current music queue

**Playlist Management** (requires database):

- `/playlist_create <name>` - Create a new playlist
- `/playlist_list` - List your playlists
- `/playlist_show <id>` - Show songs in a playlist
- `/playlist_play <id>` - Play all songs from a playlist
- `/playlist_add <id>` - Add current song to a playlist
- `/playlist_remove <id> <song_number>` - Remove a song from a playlist
- `/playlist_delete <id>` - Delete a playlist

## Development

### Adding New Features

1. **For framework features:** Modify files in `internal/framework/`, `internal/config/`, etc.
2. **For bot-specific features:** Modify files in `internal/bots/clippy/` or `internal/bots/music/`
3. **For new bots:** Create a new directory under `internal/bots/` and implement the bot interface

### Code Style

- Follow Go conventions and best practices
- Use structured logging with the provided logger
- Handle errors properly using the custom error types
- Write comprehensive comments for public functions
- Use contexts for cancellation and timeouts

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/config
```

## Architecture

### Design Principles

This framework follows the architectural patterns established in the MTG Card Bot:

- **Separation of concerns** - Clear boundaries between framework, bots, and utilities
- **Configuration-driven** - Behavior controlled through configuration files
- **Structured logging** - Comprehensive logging with structured data
- **Error handling** - Typed errors with proper propagation
- **Resource management** - Proper cleanup and resource lifecycle management

### Framework Components

1. **Core Framework** (`internal/framework/`)
   - Bot lifecycle management
   - Command registration and handling
   - Event dispatching
   - Cooldown management

2. **Configuration** (`internal/config/`)
   - JSON configuration loading
   - Environment variable overrides
   - Validation and defaults

3. **Logging** (`internal/logging/`)
   - Structured logging with slog
   - Component-based loggers
   - Configurable levels and formats

4. **Database** (`internal/database/`)
   - SQLite database abstraction
   - Playlist and song management
   - Migration handling

5. **Error Handling** (`internal/errors/`)
   - Typed error system
   - Error categorization
   - Proper error propagation

## Migration from Python

This Go implementation provides equivalent functionality to the original Python bots:

### Key Differences

1. **Performance** - Go's compiled nature provides better performance
2. **Type Safety** - Static typing prevents many runtime errors
3. **Concurrency** - Native goroutines for better concurrency handling
4. **Single Binary** - No dependency management or virtual environments
5. **Memory Usage** - Lower memory footprint compared to Python

### Migration Benefits

- **Improved Reliability** - Static typing and compile-time checks
- **Better Performance** - Faster startup and lower resource usage
- **Easier Deployment** - Single binary with no external dependencies
- **Enhanced Maintainability** - Clear structure and type safety

## Troubleshooting

### Common Issues

1. **yt-dlp not found** - Ensure yt-dlp is installed and in PATH
2. **FFmpeg not found** - Install FFmpeg and ensure it's in PATH
3. **Database errors** - Check file permissions for SQLite database
4. **Voice connection issues** - Verify bot has voice permissions in Discord

### Debug Mode

Enable debug mode for verbose logging:

```bash
./discord-bot-framework --bot all --debug
```

Or set in config:

```json
{
  "clippy": {
    "debug_mode": true,
    "log_level": "DEBUG"
  }
}
```

### Logs

The application uses structured logging. Key log fields:

- `component` - Which part of the system generated the log
- `error` - Error details when applicable
- `user_id` - Discord user ID for command logs
- `guild` - Guild/server information

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the original project license for details.

## Support

For issues and questions:

1. Check the troubleshooting section
2. Review the logs with debug mode enabled
3. Create an issue with detailed information about the problem

## Roadmap

- [ ] Complete playlist functionality implementation
- [ ] Add metrics and monitoring
- [ ] Implement health checks
- [ ] Add configuration validation UI
- [ ] Support for additional audio sources
- [ ] Distributed deployment support
- [ ] Web dashboard for management
