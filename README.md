<p align="center">
  <img src="https://github.com/dunamismax/images/blob/main/discord-bot-framework/discord-bot-framework.png" alt="Discord Bot Framework Logo" width="300" />
</p>

<div align="center">

</div>

<p align="center">
  <a href="https://github.com/dunamismax/discord-bot-framework">
    <img src="https://readme-typing-svg.demolab.com/?font=Inter&weight=600&size=28&pause=1000&color=5865F2&center=true&vCenter=true&width=1200&height=90&lines=Modern+Discord+Bots+with+Python+3.13%2B;Self-Hosted+on+Ubuntu+%2B+WSL+Support;High-Performance+py-cord+%2B+SQLite+Database;Unhinged+Clippy+Bot+with+Chaotic+Responses;Music+Bot+with+YouTube+%26+Playlist+Support;Advanced+Error+Handling+%26+Logging;Caddy+Reverse+Proxy+Integration;Database+Persistence+%26+Command+Analytics;Slash+Commands+%26+Modern+Discord+Features;Complete+Help+System+%26+User+Experience;Lightning+Fast+uv+Package+Management;Open+Source+MIT+Licensed+Framework" alt="Typing SVG" />
  </a>
</p>

<div align="center">

A production-ready framework for building modern Discord bots with Python
Designed for self-hosting on Ubuntu Linux with enterprise-grade features

</div>

<p align="center">
  <a href="https://discord-bots.dunamismax.com/">
    <img src="https://img.shields.io/badge/Live-Demo-5865F2.svg?style=for-the-badge&logo=discord&logoColor=white" alt="Live Demo">
  </a>
</p>

<p align="center">
  <a href="https://python.org/"><img src="https://img.shields.io/badge/Python-3.13+-5865F2.svg?logo=python&logoColor=white&style=for-the-badge" alt="Python Version"></a>
  <a href="https://docs.pycord.dev/"><img src="https://img.shields.io/badge/py--cord-2.6.1+-5865F2.svg?logo=discord&logoColor=white&style=for-the-badge" alt="py-cord Version"></a>
  <a href="https://docs.astral.sh/uv/"><img src="https://img.shields.io/badge/uv-Package_Manager-5865F2.svg?style=for-the-badge&logoColor=white" alt="uv"></a>
  <a href="https://sqlite.org/"><img src="https://img.shields.io/badge/SQLite-3.x-5865F2.svg?logo=sqlite&logoColor=white&style=for-the-badge" alt="SQLite"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-5865F2.svg?style=for-the-badge&logoColor=white" alt="MIT License"></a>
</p>

---

## What Makes This Different

This framework goes beyond basic Discord bot tutorials. Built with production deployment in mind, it includes enterprise-grade monitoring, security hardening, and modern Discord API integration. The architecture uses py-cord 2.6.1 with async/await patterns, comprehensive error handling with automatic retries, and a sophisticated health monitoring system that tracks both system performance and bot metrics.

The framework implements user-level cooldowns, input validation with XSS prevention, and rate limiting to protect against abuse. Database operations use SQLite with automatic validation, and the entire stack is designed for self-hosting with Caddy providing automatic HTTPS and security headers including Content Security Policy and Permissions Policy.

## Core Features

**Production-Ready Bots**

- **Advanced Music Bot** - YouTube integration with yt-dlp, database-backed playlist management, auto-disconnect when alone, and high-quality audio processing
- **Unhinged Clippy Bot** - Interactive polls, Discord UI buttons, 30+ contextual responses, and modern meme integration with smart permission detection

**Enterprise Infrastructure**

- **System Monitoring** - Real-time CPU, memory, and disk metrics with process-level tracking and uptime analytics
- **Security Hardening** - Rate limiting (30 req/min per IP), input sanitization, XSS prevention, and comprehensive security headers
- **Error Resilience** - Exponential backoff retry mechanisms for Discord API calls, user-friendly error messages, and detailed logging
- **Configuration Validation** - Automatic validation of tokens, ports, limits, and file formats with descriptive error messages

**Modern Discord Integration**

- **Latest py-cord 2.6.1** - Slash commands, interactive buttons, reaction polls, rich embeds, and ephemeral messages
- **Smart Cooldowns** - Per-user, per-command cooldowns with remaining time display and automatic cleanup
- **Database Analytics** - Command usage tracking, user engagement metrics, and playlist analytics

## Quick Start

### Installation (Ubuntu/WSL)

```bash
# System setup
sudo apt update && sudo apt upgrade -y
sudo apt install git curl ffmpeg -y

# Install uv package manager (ultra-fast Python package management)
curl -LsSf https://astral.sh/uv/install.sh | sh
source ~/.bashrc

# Clone and setup
git clone https://github.com/dunamismax/discord-bot-framework.git
cd discord-bot-framework

# Install Python and project dependencies with uv
uv python install 3.13
uv sync --all-extras

# Configure bot tokens (create applications at Discord Developer Portal)
cp .env.example .env
# Edit .env with your Discord bot tokens

# Start bots
./scripts/start-all.sh

# Validate installation
uv run python validate.py
```

### UV Environment Management

This project relies entirely on [uv](https://docs.astral.sh/uv/) for Python workflow:

- **Python Versions:** The repository pins Python 3.13 via `.python-version`. Run `uv python install` to download and manage the interpreter.
- **Virtual Environments:** `uv venv` creates an isolated `.venv` directory. `uv sync` will also create it automatically if missing.
- **Dependencies:** Use `uv sync --all-extras` to install project and workspace dependencies from `uv.lock`.
- **Running Bots:** All scripts invoke `uv run` so commands execute within the managed environment.
- **Validation:** Utilities such as `validate.py` run through `uv run python ...` to ensure consistent tooling.

### Production Deployment

The framework includes systemd service files for professional deployment with automatic startup, crash recovery, and proper logging integration.

```bash
# Install as systemd services
sudo ./systemd/install-services.sh

# Install Caddy reverse proxy for automatic HTTPS
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update && sudo apt install caddy

# Configure and start Caddy
sudo cp Caddyfile /etc/caddy/Caddyfile
sudo nano /etc/caddy/Caddyfile  # Edit domain configuration
sudo systemctl enable caddy && sudo systemctl start caddy
```

### Health Check Endpoints

The framework provides comprehensive monitoring endpoints that work with load balancers and monitoring systems:

```bash
curl http://localhost:8081/health   # Basic health check
curl http://localhost:8081/metrics  # Detailed system metrics (CPU, memory, disk, bot stats)
curl http://localhost:8081/status   # Simple OK/NOT OK for load balancers
```

## Project Architecture

```
discord-bot-framework/
├── apps/
│   ├── clippy_bot/              # Chaotic Microsoft Clippy recreation
│   │   ├── cogs/unhinged_responses.py    # Interactive commands and UI components
│   │   └── main.py              # Bot initialization with database integration
│   └── music_bot/               # YouTube music bot with playlists
│       ├── cogs/music_player.py          # Queue management and playback control
│       └── main.py              # Bot initialization with voice state handling
├── libs/shared_utils/           # Framework core components
│   ├── base_bot.py             # Enhanced bot class with retry logic and cooldowns
│   ├── config_loader.py        # Configuration with automatic validation
│   ├── database.py             # SQLite operations with connection pooling
│   ├── health_check.py         # HTTP server with rate limiting and metrics
│   └── help_system.py          # Universal help command system
├── scripts/                     # Management automation
├── systemd/                     # Production service configurations
├── tests/                       # Integration and unit test suite
├── Caddyfile                    # Reverse proxy with security headers
└── validate.py                  # Code validation and syntax checking
```

## Bot Applications

### Clippy Bot Features

The Clippy bot goes beyond simple responses with modern Discord UI integration and intelligent behavior patterns.

**Interactive Commands:**

- `/clippy` - On-demand chaotic assistance with user cooldowns
- `/clippy_wisdom` - Questionable life advice with personality
- `/clippy_poll` - Creates polls with Clippy's chaotic answer options
- `/clippy_help` - Interactive help with Discord UI buttons ("More Chaos", "I Regret This")

**Smart Behavior:**

- Random responses every 15-45 minutes with permission checking
- 3% chance to respond to any message with contextual advice
- 30+ unique responses inspired by 2024-2025 internet culture
- Input validation prevents abuse and dangerous content

### Music Bot Features

Built for reliability and user experience with comprehensive playlist management and audio optimization.

**Playback Control:**

- `/play` - YouTube URL or search query support with error handling
- `/pause`, `/resume`, `/skip`, `/stop` - All with visual feedback and confirmation
- Auto-disconnect when alone in voice channel (configurable timeout)
- High-quality audio processing with FFmpeg optimization

**Playlist System:**

- `/playlist create/list/play/add/remove` - Full CRUD operations
- Database persistence with automatic backup
- User-specific playlists with ownership validation
- Configurable size limits (default: 100 songs, max: 1000)

**Advanced Features:**

- 5-minute inactivity timeout with graceful disconnect
- Rich embed displays with song information and duration
- Network error handling with automatic retries
- Queue management with visual feedback

## Configuration System

The framework supports both environment variables and JSON configuration with automatic validation.

**Environment Variables:**

```env
# Clippy Bot Configuration
CLIPPY_BOT_TOKEN=your_bot_token_here
CLIPPY_GUILD_ID=guild_id_for_development_testing
CLIPPY_DEBUG=false
CLIPPY_HEALTH_CHECK_PORT=8081

# Music Bot Configuration
MUSIC_BOT_TOKEN=your_bot_token_here
MUSIC_GUILD_ID=guild_id_for_development_testing
MUSIC_DEBUG=false
MUSIC_MAX_PLAYLIST_SIZE=100
MUSIC_AUTO_DISCONNECT_TIMEOUT=300
MUSIC_HEALTH_CHECK_PORT=8082
```

**JSON Configuration Alternative:**

```json
{
  "clippy": {
    "command_prefix": "!",
    "debug": false,
    "health_check_port": 8081,
    "command_cooldown": 1.0
  },
  "music": {
    "max_playlist_size": 100,
    "max_queue_size": 50,
    "auto_disconnect_timeout": 300,
    "health_check_port": 8082,
    "allowed_file_formats": ["mp3", "mp4", "webm", "ogg"],
    "max_song_duration": 3600
  }
}
```

The configuration system validates token formats, port ranges, file limits, and provides descriptive error messages for common mistakes.

## Development & Service Management

**Development Workflow:**

```bash
uv sync --all-extras             # Install all dependencies with extras
python validate.py               # Syntax validation and import checking
uv run ruff check .              # Code linting with modern Python standards
uv run ruff format .             # Automatic code formatting
python -m pytest tests/          # Run integration and unit tests
```

**Service Management:**

```bash
# Systemd services (production)
sudo systemctl start clippy-bot music-bot
sudo systemctl stop clippy-bot music-bot
sudo systemctl enable clippy-bot music-bot    # Auto-start on boot
sudo journalctl -u clippy-bot -f              # Real-time log viewing

# Manual management (development)
./scripts/start-all.sh           # Start all bots with proper environment
./scripts/stop-all.sh            # Graceful shutdown with cleanup
```

## Technology Stack & Architecture

<table align="center">
<tr>
<td align="center">
<img src="https://img.shields.io/badge/Language-Python_3.13+-5865F2?style=for-the-badge&logo=python&logoColor=white" alt="Python"><br>
<sub>Modern async/await with type hints</sub>
</td>
<td align="center">
<img src="https://img.shields.io/badge/Discord-py--cord_2.6.1-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="py-cord"><br>
<sub>Latest Discord API with UI components</sub>
</td>
<td align="center">
<img src="https://img.shields.io/badge/Database-SQLite_3-5865F2?style=for-the-badge&logo=sqlite&logoColor=white" alt="SQLite"><br>
<sub>ACID compliance with WAL mode</sub>
</td>
</tr>
<tr>
<td align="center">
<img src="https://img.shields.io/badge/Package-uv-5865F2?style=for-the-badge&logoColor=white" alt="uv"><br>
<sub>Rust-powered package management</sub>
</td>
<td align="center">
<img src="https://img.shields.io/badge/Proxy-Caddy-5865F2?style=for-the-badge&logoColor=white" alt="Caddy"><br>
<sub>HTTP/2, HTTP/3, automatic HTTPS</sub>
</td>
<td align="center">
<img src="https://img.shields.io/badge/OS-Ubuntu_Linux-5865F2?style=for-the-badge&logo=ubuntu&logoColor=white" alt="Ubuntu"><br>
<sub>LTS stability with WSL2 support</sub>
</td>
</tr>
</table>

**Data Flow:** Discord API ↔ py-cord Bot ↔ SQLite Database ↔ Health Monitoring ↔ Caddy Proxy

The architecture uses dependency injection for configuration, database connection pooling for performance, and event-driven design for scalability. Error handling includes circuit breakers for external services and graceful degradation when components fail.

## Security & Performance

**Security Measures:**

- Content Security Policy and Permissions Policy headers via Caddy
- Input validation with XSS prevention and content sanitization
- Rate limiting on health endpoints (30 requests/minute per IP)
- User cooldowns to prevent command spam and abuse
- Secure token management with environment variable isolation

**Performance Optimizations:**

- Connection pooling for database operations
- Async/await throughout for non-blocking I/O
- Memory-efficient queue management with automatic cleanup
- System monitoring with CPU, memory, and disk tracking
- Exponential backoff for API retries to prevent rate limiting

## Self-Hosting Philosophy

**Why Choose Self-Hosting:**

This framework embraces self-hosting for privacy, control, and learning. Your bot data stays on your infrastructure, you pay no monthly cloud fees, and you gain deep understanding of the entire technology stack. The Ubuntu/WSL design provides stability through LTS releases while maintaining compatibility with Windows development environments.

Caddy handles the complexity of HTTPS certificates and security headers automatically, while systemd provides professional service management with logging integration. The result is a production-ready deployment that rivals cloud-hosted solutions while remaining fully under your control.

**Technical Benefits:**

- **Ubuntu LTS** - Long-term stability with excellent Python ecosystem
- **WSL2 Integration** - Native Linux experience on Windows with full compatibility
- **Systemd Management** - Professional service lifecycle with automatic restart policies
- **Caddy Automation** - Zero-configuration HTTPS with Let's Encrypt integration

## Documentation & Resources

**Configuration Reference:**

- [config.example.json](config.example.json) - Complete configuration options with validation rules

**External Documentation:**

- [py-cord Documentation](https://docs.pycord.dev/) - Modern Discord API library with examples
- [Discord Developer Portal](https://discord.com/developers/docs) - Official Discord API reference
- [Caddy Documentation](https://caddyserver.com/docs/) - Web server configuration and security

## License

<div align="center">

<img src="https://img.shields.io/badge/License-MIT-5865F2?style=for-the-badge&logo=opensource&logoColor=white" alt="MIT License">

**This project is licensed under the MIT License**
*Feel free to use, modify, and distribute*

[View License Details](LICENSE)

</div>

---

<p align="center">
  <a href="https://www.buymeacoffee.com/dunamismax">
    <img src="https://cdn.buymeacoffee.com/buttons/v2/default-blue.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" >
  </a>
</p>

<p align="center">
  <a href="https://twitter.com/dunamismax" target="_blank"><img src="https://img.shields.io/badge/Twitter-5865F2.svg?&style=for-the-badge&logo=twitter&logoColor=white" alt="Twitter"></a>
  <a href="https://bsky.app/profile/dunamismax.bsky.social" target="_blank"><img src="https://img.shields.io/badge/Bluesky-5865F2?style=for-the-badge&logo=bluesky&logoColor=white" alt="Bluesky"></a>
  <a href="https://reddit.com/user/dunamismax" target="_blank"><img src="https://img.shields.io/badge/Reddit-5865F2.svg?&style=for-the-badge&logo=reddit&logoColor=white" alt="Reddit"></a>
  <a href="https://discord.com/users/dunamismax" target="_blank"><img src="https://img.shields.io/badge/Discord-5865F2.svg?style=for-the-badge&logo=discord&logoColor=white" alt="Discord"></a>
  <a href="https://signal.me/#p/+dunamismax.66" target="_blank"><img src="https://img.shields.io/badge/Signal-5865F2.svg?style=for-the-badge&logo=signal&logoColor=white" alt="Signal"></a>
</p>

---

<p align="center">
  <strong>Discord Bot Framework</strong><br>
  <sub>Production-Ready • Enterprise Security • Modern Python • Self-Hosted • Open Source</sub>
</p>
