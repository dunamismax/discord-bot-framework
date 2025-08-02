# Discord Bot Monorepo

A modern Python monorepo containing multiple Discord bots built with modern 2025 best practices.

## 🤖 Bots Included

### 1. Clippy Bot (`apps/clippy_bot/`)

An unhinged version of Microsoft Clippy that randomly appears in your Discord channels with hilariously unhelpful advice and nostalgic chaos.

**Features:**

- Random unhinged responses at intervals (15-45 minutes)
- 3% chance to respond to any message
- Slash commands: `/clippy`, `/clippy_wisdom`
- 30+ unique unhinged quotes inspired by 2024-2025 memes

### 2. Music Bot (`apps/music_bot/`)

A high-quality YouTube music bot with full queue management and voice channel support.

**Features:**

- `/play [query/url]` - Play music from YouTube or search
- `/pause` / `/resume` - Control playback
- `/skip` - Skip current song
- `/stop` - Stop music and clear queue
- `/queue` - View current queue
- Automatic queue management
- 5-minute inactivity timeout
- Auto-disconnect when alone in voice channel

## 🏗️ Architecture

This monorepo uses modern Python tooling and follows 2025 best practices:

- **Package Management**: `uv` with workspace support
- **Code Quality**: `ruff` (replaces Black, isort, flake8, etc.)
- **Discord Library**: `py-cord` (modern, slash command focused)
- **Audio Processing**: `yt-dlp` + `FFmpeg` for high-quality streaming
- **Configuration**: Environment-based with `python-dotenv`

### Directory Structure

```
/pycord
├── apps/
│   ├── clippy_bot/
│   │   ├── cogs/
│   │   │   └── unhinged_responses.py
│   │   ├── main.py
│   │   └── pyproject.toml
│   └── music_bot/
│       ├── cogs/
│       │   └── music_player.py
│       ├── main.py
│       └── pyproject.toml
├── libs/
│   └── shared_utils/
│       ├── __init__.py
│       ├── config_loader.py
│       ├── base_bot.py
│       └── pyproject.toml
├── .gitignore
├── pyproject.toml (workspace root)
└── README.md
```

## 🚀 Setup

### Prerequisites

- Python 3.8+
- `uv` package manager
- FFmpeg (for music bot)

### Installation

1. **Install uv** (if not already installed):

   ```bash
   curl -LsSf https://astral.sh/uv/install.sh | sh
   ```

2. **Clone and install dependencies**:

   ```bash
   git clone https://github.com/dunamismax/pycord.git
   cd pycord
   uv sync --all-extras
   ```

3. **Install FFmpeg** (for music bot):

   ```bash
   # macOS
   brew install ffmpeg
   
   # Ubuntu/Debian
   sudo apt update && sudo apt install ffmpeg
   
   # Windows
   # Download from https://ffmpeg.org/download.html
   ```

### Configuration

Create a `.env` file in the root directory:

```env
# Clippy Bot
CLIPPY_BOT_TOKEN=your_clippy_bot_token_here
CLIPPY_GUILD_ID=your_guild_id_for_testing  # Optional
CLIPPY_DEBUG=false

# Music Bot
MUSIC_BOT_TOKEN=your_music_bot_token_here
MUSIC_GUILD_ID=your_guild_id_for_testing   # Optional
MUSIC_DEBUG=false
```

## 🎮 Running the Bots

### Run Individual Bots

```bash
# Run Clippy Bot
cd apps/clippy_bot
uv run python main.py

# Run Music Bot
cd apps/music_bot
uv run python main.py
```

### Development Mode

```bash
# Install development dependencies
uv sync --group dev

# Run linting and formatting
uv run ruff check .
uv run ruff format .

# Run tests (when available)
uv run pytest
```

## 🔧 Bot Configuration

### Clippy Bot Environment Variables

- `CLIPPY_BOT_TOKEN`: Discord bot token
- `CLIPPY_GUILD_ID`: Guild ID for testing (optional)
- `CLIPPY_COMMAND_PREFIX`: Command prefix (default: "!")
- `CLIPPY_DEBUG`: Enable debug logging (default: false)

### Music Bot Environment Variables

- `MUSIC_BOT_TOKEN`: Discord bot token
- `MUSIC_GUILD_ID`: Guild ID for testing (optional)
- `MUSIC_COMMAND_PREFIX`: Command prefix (default: "!")
- `MUSIC_DEBUG`: Enable debug logging (default: false)

## 🎵 Music Bot Usage

1. Join a voice channel
2. Use `/play <YouTube URL or search term>` to start music
3. Manage playback with `/pause`, `/resume`, `/skip`, `/stop`
4. View the queue with `/queue`

**Note**: The bot requires FFmpeg to be installed and accessible in your system PATH.

## 📝 Development

### Adding New Bots

1. Create new directory in `apps/`
2. Add `pyproject.toml` with workspace dependencies
3. Implement using the `shared_utils.BaseBot` class
4. Update root `pyproject.toml` workspace members

### Shared Utilities

The `libs/shared_utils` package provides:

- `BaseBot`: Base class with common Discord bot functionality
- `load_config()`: Environment-based configuration loading
- `BotConfig`: Configuration dataclass

### Code Quality

This project uses `ruff` for all code quality needs:

- Formatting (replaces Black)
- Import sorting (replaces isort)
- Linting (replaces flake8, pyupgrade, etc.)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run `uv run ruff check .` and `uv run ruff format .`
5. Submit a pull request

## 📄 License

This project is open source and available under the MIT License.

## 🆘 Support

- **Issues**: Report bugs and request features via GitHub Issues
- **Documentation**: See inline code documentation and docstrings
- **Discord.py Docs**: <https://docs.pycord.dev/>

---

Built with ❤️ using modern Python tooling and practices for 2025.
