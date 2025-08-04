<p align="center">
  <img src="https://github.com/dunamismax/images/blob/main/python/Python-Discord-Bot-Logo.png" alt="Discord Bot Logo" width="300" />
</p>

<div align="center">

# Discord Bot Monorepo

</div>

<p align="center">
  <a href="https://github.com/dunamismax/discord-bot-framework">
    <img src="https://readme-typing-svg.demolab.com/?font=Inter&weight=600&size=28&pause=1000&color=5865F2&center=true&vCenter=true&width=1200&height=90&lines=Modern+Discord+Bots+with+Python+3.8%2B;High-Performance+py-cord+%2B+SQLite+Database;Unhinged+Clippy+Bot+with+Chaotic+Responses;Music+Bot+with+YouTube+%26+Playlist+Support;Advanced+Error+Handling+%26+Logging;Production+Ready+with+Docker+%26+Health+Checks;Database+Persistence+%26+Command+Analytics;Slash+Commands+%26+Modern+Discord+Features;Complete+Help+System+%26+User+Experience;Lightning+Fast+uv+Package+Management;Open+Source+MIT+Licensed+Framework" alt="Typing SVG" />
  </a>
</p>

<div align="center">

<p style="color: #5865F2; font-size: 18px; font-weight: 500;">
A comprehensive monorepo for building modern Discord bots with Python<br>
Featuring database persistence, playlist management, and production-ready deployment
</p>

</div>

<p align="center">
  <a href="https://discord-bots.dunamismax.com/">
    <img src="https://img.shields.io/badge/Live-Demo-5865F2.svg?style=for-the-badge&logo=discord&logoColor=white" alt="Live Demo">
  </a>
</p>

<details>
<summary style="cursor: pointer; padding: 12px; background: linear-gradient(135deg, #5865F2, #404eed); color: white; border-radius: 8px; text-align: center; font-weight: bold; margin: 8px 0; list-style: none;">
  Click to explore the bots
</summary>
<div style="margin-top: 16px;">
<p align="center">
  <strong>Clippy Bot:</strong> <code>python -m apps.clippy_bot.main</code><br>
  <strong>Music Bot:</strong> <code>python -m apps.music_bot.main</code><br>
  <strong>Docker:</strong> <code>docker-compose up</code><br>
  <strong>Health Checks:</strong> <code>http://localhost:8081/health</code><br>
  <strong>Run Tests:</strong> <code>python validate.py</code>
</p>
</div>
</details>

---

<p align="center">
  <a href="https://python.org/"><img src="https://img.shields.io/badge/Python-3.8+-5865F2.svg?logo=python&logoColor=white&style=for-the-badge" alt="Python Version"></a>
  <a href="https://docs.pycord.dev/"><img src="https://img.shields.io/badge/py--cord-2.4+-5865F2.svg?logo=discord&logoColor=white&style=for-the-badge" alt="py-cord Version"></a>
  <a href="https://docs.astral.sh/uv/"><img src="https://img.shields.io/badge/uv-Package_Manager-5865F2.svg?style=for-the-badge&logoColor=white" alt="uv"></a>
  <a href="https://sqlite.org/"><img src="https://img.shields.io/badge/SQLite-3.x-5865F2.svg?logo=sqlite&logoColor=white&style=for-the-badge" alt="SQLite"></a>
</p>

<p align="center">
  <a href="https://docs.astral.sh/ruff/"><img src="https://img.shields.io/badge/Ruff-Linting-5865F2.svg?style=for-the-badge&logoColor=white" alt="Ruff"></a>
  <a href="https://docs.docker.com/"><img src="https://img.shields.io/badge/Docker-Ready-5865F2.svg?logo=docker&logoColor=white&style=for-the-badge" alt="Docker"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-5865F2.svg?style=for-the-badge&logoColor=white" alt="MIT License"></a>
  <img src="https://img.shields.io/badge/Production-Ready-5865F2.svg?style=for-the-badge&logoColor=white" alt="Production Ready">
</p>

---

## Core Features

<table align="center">
<tr>
<td align="center" width="25%">
<img src="https://img.shields.io/badge/-Playlist_Support-5865F2?style=for-the-badge&logo=spotify&logoColor=white" alt="Playlist Support"><br>
<sub><b>Music Playlists<br>CRUD Operations</b></sub>
</td>
<td align="center" width="25%">
<img src="https://img.shields.io/badge/-Modern_Python-5865F2?style=for-the-badge&logo=python&logoColor=white" alt="Modern Python"><br>
<sub><b>Python 3.8+<br>Modern Async/Await</b></sub>
</td>
<td align="center" width="25%">
<img src="https://img.shields.io/badge/-Database_Driven-5865F2?style=for-the-badge&logo=sqlite&logoColor=white" alt="Database Driven"><br>
<sub><b>SQLite Persistence<br>Analytics & Stats</b></sub>
</td>
<td align="center" width="25%">
<img src="https://img.shields.io/badge/-Health_Monitoring-5865F2?style=for-the-badge&logo=heart&logoColor=white" alt="Health Monitoring"><br>
<sub><b>HTTP Endpoints<br>Real-time Metrics</b></sub>
</td>
</tr>
</table>

### Bot Features & Capabilities

- **Advanced music bot** with YouTube integration, playlist management, and queue control
- **Unhinged Clippy bot** with chaotic responses and nostalgic Microsoft Office humor
- **Database persistence** for playlists, user data, and command analytics
- **Comprehensive error handling** with user-friendly messages and detailed logging
- **Health monitoring** with HTTP endpoints for uptime and performance metrics
- **Slash command integration** with modern Discord API features

### Developer Experience

- **Monorepo architecture** with uv workspaces for efficient dependency management
- **Code quality tools** with Ruff for formatting, linting, and import sorting
- **Docker deployment** ready with multi-stage builds and health checks
- **Comprehensive testing** with validation scripts and unit tests

### Production Features

- **Docker containerization** with optimized images and compose orchestration
- **Health check endpoints** for monitoring and load balancer integration
- **Configuration management** supporting environment variables and JSON configs
- **Logging and analytics** with command usage tracking and performance metrics

---

<p align="center">
  <img src="https://github.com/dunamismax/images/blob/main/python/python-discord-banner.jpg" alt="Python Discord Banner" width="500" />
</p>

## Project Structure

```sh
discord-bot-framework/
├── apps/
│   ├── clippy_bot/            # Unhinged Microsoft Clippy bot
│   │   ├── cogs/
│   │   │   └── unhinged_responses.py
│   │   └── main.py
│   └── music_bot/             # YouTube music bot with playlists
│       ├── cogs/
│       │   └── music_player.py
│       └── main.py
├── libs/
│   └── shared_utils/          # Shared utilities and base classes
│       ├── base_bot.py        # Enhanced bot base class
│       ├── config_loader.py   # Configuration management
│       ├── database.py        # SQLite database utilities
│       ├── health_check.py    # HTTP health monitoring
│       └── help_system.py     # Universal help commands
├── tests/                     # Comprehensive test suite
├── docker-compose.yml         # Production deployment
├── Dockerfile                 # Multi-stage container build
└── validate.py               # Code validation script
```

## Documentation

**Guides:** [Setup](docs/SETUP.md) - [Configuration](docs/CONFIG.md) - [Deployment](docs/DEPLOYMENT.md)

**Resources:** [py-cord Docs](https://docs.pycord.dev/) - [Discord Developer Portal](https://discord.com/developers/docs) - [SQLite Docs](https://sqlite.org/docs.html) - [Docker Guide](https://docs.docker.com/)

---

## Quick Start

**Prerequisites:** Python 3.8+ and uv package manager installed

```bash
# Clone and setup
git clone https://github.com/dunamismax/discord-bot-framework.git && cd discord-bot-framework

# Ensure correct Python version (3.8+ required)
python --version  # Should show Python 3.8.x or higher

# Install uv package manager (if not installed)
curl -LsSf https://astral.sh/uv/install.sh | sh

# Install dependencies
uv sync --all-extras

# Configure bot tokens (required)
cp .env.example .env
# Edit .env with your Discord bot tokens

# Install FFmpeg for music bot (required)
# macOS: brew install ffmpeg
# Ubuntu: sudo apt install ffmpeg
# Windows: Download from https://ffmpeg.org/

# Start individual bots
cd apps/clippy_bot && uv run python main.py    # Clippy bot
cd apps/music_bot && uv run python main.py     # Music bot

# Or use Docker (recommended for production)
docker-compose up -d

# Validate installation
python validate.py

# Check health endpoints
curl http://localhost:8081/health  # Clippy bot health
curl http://localhost:8082/health  # Music bot health
```

<div align="center">

**Clippy Bot:** Chaotic responses and wisdom - **Music Bot:** YouTube playlists and queue management

> **Note:** Both bots require Discord bot tokens. Create applications at [Discord Developer Portal](https://discord.com/developers/applications)

<img src="https://img.shields.io/badge/Status-Production_Ready-5865F2?style=for-the-badge&logoColor=white" alt="Production Ready">
<img src="https://img.shields.io/badge/Deployment-Docker_Optimized-5865F2?style=for-the-badge&logoColor=white" alt="Docker Optimized">

</div>

---

## Development & Build

```bash
# Development
uv sync --all-extras             # Install all dependencies
python validate.py               # Validate code syntax
uv run ruff check .              # Run linting
uv run ruff format .             # Format code

# Testing
python -m pytest tests/          # Run test suite
python validate.py               # Syntax validation

# Production deployment
docker-compose up -d             # Start with Docker
docker-compose logs -f           # View logs

# Individual bot management
cd apps/clippy_bot
uv run python main.py            # Run Clippy bot

cd apps/music_bot  
uv run python main.py            # Run Music bot
```

---

## Docker Deployment

```bash
# Production deployment
docker-compose up -d

# View logs
docker-compose logs -f clippy-bot
docker-compose logs -f music-bot

# Health checks
curl http://localhost:8081/health
curl http://localhost:8082/metrics

# Scale and manage
docker-compose restart music-bot
docker-compose stop
```

**Architecture:** Load Balancer → Docker Containers → SQLite Database → Discord API

---

<p align="center">
  <img src="https://github.com/dunamismax/images/blob/main/python/Python-logo.png" alt="Python Logo" width="200" />
</p>

## Technology Stack

<table align="center">
<tr>
<td align="center">
<img src="https://img.shields.io/badge/Language-Python_3.8+-5865F2?style=for-the-badge&logo=python&logoColor=white" alt="Python"><br>
<sub>Modern Python with async/await</sub>
</td>
<td align="center">
<img src="https://img.shields.io/badge/Discord-py--cord_2.4-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="py-cord"><br>
<sub>Modern Discord API library</sub>
</td>
<td align="center">
<img src="https://img.shields.io/badge/Database-SQLite_3-5865F2?style=for-the-badge&logo=sqlite&logoColor=white" alt="SQLite"><br>
<sub>Embedded relational database</sub>
</td>
</tr>
<tr>
<td align="center">
<img src="https://img.shields.io/badge/Package-uv-5865F2?style=for-the-badge&logoColor=white" alt="uv"><br>
<sub>Ultra-fast Python package manager</sub>
</td>
<td align="center">
<img src="https://img.shields.io/badge/Code_Quality-Ruff-5865F2?style=for-the-badge&logoColor=white" alt="Ruff"><br>
<sub>Extremely fast Python linter</sub>
</td>
<td align="center">
<img src="https://img.shields.io/badge/Deployment-Docker-5865F2?style=for-the-badge&logo=docker&logoColor=white" alt="Docker"><br>
<sub>Containerized production deployment</sub>
</td>
</tr>
</table>

**Stack highlights:** Modern Python - Discord slash commands - Database persistence - Health monitoring - Production ready

## Philosophy

### Why This Architecture?

1. **Monorepo Structure** - Shared utilities and consistent development experience across multiple bots
2. **Database Persistence** - SQLite for reliable data storage, playlists, and analytics
3. **Modern Discord API** - Slash commands, embeds, and contemporary Discord features
4. **Production Ready** - Docker, health checks, monitoring, and comprehensive error handling

### Why These Technologies?

- **py-cord** - Modern Discord library with excellent slash command support
- **SQLite** - Zero-configuration database perfect for Discord bot data
- **uv Package Manager** - Extremely fast dependency resolution and installation
- **Ruff** - Lightning-fast Python linting and formatting
- **Docker** - Consistent deployment environment with health monitoring

### Development Philosophy

- **Error Resilience** - Comprehensive error handling with user-friendly messages
- **User Experience** - Intuitive slash commands with helpful feedback
- **Maintainability** - Clean architecture with shared utilities and consistent patterns
- **Observability** - Health endpoints, logging, and command analytics
- **Modern Standards** - Python 3.8+, async/await, type hints, and contemporary patterns

## Bot Applications

### Unhinged Clippy Bot

A chaotic recreation of Microsoft Clippy with modern internet humor and Discord integration.

**Features:**

- Random unhinged responses every 15-45 minutes
- 3% chance to respond to any message with chaotic advice
- `/clippy` command for on-demand unhelpful assistance
- `/clippy_wisdom` for questionable life advice
- 30+ unique responses inspired by 2024-2025 memes
- Smart channel permission detection
- Integrated help system and error handling

### Advanced Music Bot

A comprehensive YouTube music bot with playlist management and advanced queue features.

**Features:**

- **Music Playback**: `/play` with YouTube URL or search query support
- **Queue Management**: `/pause`, `/resume`, `/skip`, `/stop` with visual feedback
- **Playlist System**: Create, manage, and play custom playlists
  - `/playlist create` - Create new playlists
  - `/playlist list` - View your playlists
  - `/playlist play` - Play entire playlists
  - `/playlist add` - Add current song to playlist
  - `/playlist remove` - Remove songs from playlists
- **Advanced Features**:
  - Auto-disconnect when alone in voice channel
  - 5-minute inactivity timeout
  - Database persistence for playlists
  - Rich embed displays with song information
  - Error handling for invalid URLs and network issues
  - High-quality audio with FFmpeg optimization

## Production Features

**[Health Monitoring & Analytics](docs/MONITORING.md)**

Both bots include comprehensive monitoring and analytics:

- **HTTP Health Endpoints** - `/health`, `/metrics`, `/status` for monitoring
- **Command Analytics** - Track usage patterns and popular commands
- **Real-time Metrics** - Bot status, latency, guild count, and performance data
- **Database Analytics** - Playlist usage, user engagement, and growth metrics

**[Docker Deployment](docker-compose.yml)**

Production-ready containerization with:

- **Multi-stage builds** for optimized image sizes
- **Health checks** and automatic restart policies
- **Volume persistence** for database storage
- **Nginx reverse proxy** for load balancing and SSL termination
- **Environment-based configuration** for secure token management

## Configuration

All bots support both environment variables and JSON configuration files:

**Environment Variables:**
```env
# Clippy Bot
CLIPPY_BOT_TOKEN=your_bot_token
CLIPPY_GUILD_ID=guild_id_for_testing
CLIPPY_DEBUG=false

# Music Bot  
MUSIC_BOT_TOKEN=your_bot_token
MUSIC_GUILD_ID=guild_id_for_testing
MUSIC_DEBUG=false
MUSIC_MAX_PLAYLIST_SIZE=100
MUSIC_AUTO_DISCONNECT_TIMEOUT=300
```

**JSON Configuration:**
```json
{
  "music": {
    "max_playlist_size": 100,
    "max_queue_size": 50,
    "auto_disconnect_timeout": 300,
    "health_check_port": 8082
  }
}
```

See [config.example.json](config.example.json) for full configuration options.

## License

<div align="center">

<img src="https://img.shields.io/badge/License-MIT-5865F2?style=for-the-badge&logo=opensource&logoColor=white" alt="MIT License">

**This project is licensed under the MIT License**
_Feel free to use, modify, and distribute_

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
  <strong style="color: #5865F2; font-size: 18px;">Discord Bot Monorepo</strong><br>
  <sub style="color: #5865F2;">Modern Python - Discord API - Database Persistence - Playlist Management - Health Monitoring - Docker Ready - Production Optimized - Open Source</sub>
</p>

<p align="center">
  <img src="https://github.com/dunamismax/images/blob/main/python/python-discord-community.jpg" alt="Python Discord Community" width="500" />
</p>
