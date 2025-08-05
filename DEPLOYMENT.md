# Discord Bot Deployment Guide

This guide walks you through setting up and deploying the Clippy Bot and Music Bot to your Discord server.

## Prerequisites

- Python 3.12+ installed (⚠️ Python 3.13+ not supported due to py-cord voice dependencies)
- Discord account with server admin permissions
- Linux/macOS environment (Ubuntu recommended for production)
- FFmpeg installed (for music bot voice functionality)

## Step 1: Create Discord Applications

### 1.1 Access Discord Developer Portal

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Log in with your Discord account

### 1.2 Create Bot Applications

Create **two separate applications** (one for each bot):

**For Clippy Bot:**

1. Click "New Application"
2. Name it "Clippy Bot" (or your preferred name)
3. Click "Create"

**For Music Bot:**

1. Click "New Application"
2. Name it "Music Bot" (or your preferred name)
3. Click "Create"

### 1.3 Configure Bot Settings

For **each application**:

1. Go to the "Bot" section in the left sidebar
2. Click "Add Bot" if not already created
3. **Copy the bot token** (click "Reset Token" if needed, then copy)
   - ⚠️ **Keep these tokens secure** - treat them like passwords
4. Under "Privileged Gateway Intents":
   - ✅ Enable "Message Content Intent"
   - ✅ Enable "Server Members Intent" (required for Music Bot voice features)

## Step 2: Get Your Server ID

1. In Discord, enable Developer Mode:
   - User Settings → App Settings → Advanced → Developer Mode (toggle on)
2. Right-click your server name
3. Click "Copy Server ID"
4. Save this ID - you'll need it for configuration

## Step 3: Install Dependencies

### 3.1 Install uv Package Manager

```bash
# Install uv (ultra-fast Python package manager)
curl -LsSf https://astral.sh/uv/install.sh | sh
source ~/.bashrc

# Verify installation
uv --version
```

### 3.2 Install Project Dependencies

```bash
# Navigate to project directory
cd discord-bot-framework

# Install Python 3.12 and all dependencies
uv python install 3.12
uv sync --all-extras

# Verify installation
uv run python validate.py
```

## Step 4: Configure Bot Tokens

### Option A: Environment Variables (Recommended)

Create a `.env` file in the project root:

```bash
# Copy example configuration
cp config.example.json .env.example

# Create your .env file
cat > .env << 'EOF'
# Clippy Bot Configuration
CLIPPY_BOT_TOKEN=your_clippy_bot_token_here
CLIPPY_GUILD_ID=your_server_id_here
CLIPPY_DEBUG=false
CLIPPY_HEALTH_CHECK_PORT=8081

# Music Bot Configuration
MUSIC_BOT_TOKEN=your_music_bot_token_here
MUSIC_GUILD_ID=your_server_id_here
MUSIC_DEBUG=false
MUSIC_HEALTH_CHECK_PORT=8082
MUSIC_MAX_PLAYLIST_SIZE=100
MUSIC_AUTO_DISCONNECT_TIMEOUT=300
EOF
```

Replace the placeholder values:

- `your_clippy_bot_token_here` → Your Clippy Bot token from Step 1.3
- `your_music_bot_token_here` → Your Music Bot token from Step 1.3  
- `your_server_id_here` → Your server ID from Step 2

### Option B: JSON Configuration

Alternatively, copy and edit the JSON config:

```bash
cp config.example.json config.json
# Edit config.json with your tokens and server ID
```

## Step 5: Invite Bots to Your Server

### 5.1 Generate Invite URLs

For **each bot application** in the Developer Portal:

1. Go to OAuth2 → URL Generator
2. Select Scopes:
   - ✅ `bot`
   - ✅ `applications.commands`

3. Select Bot Permissions:

   **For Clippy Bot:**
   - ✅ Send Messages
   - ✅ Use Slash Commands
   - ✅ Read Message History
   - ✅ Add Reactions
   - ✅ Embed Links

   **For Music Bot:**
   - ✅ Send Messages
   - ✅ Use Slash Commands
   - ✅ Read Message History
   - ✅ Connect (to voice channels)
   - ✅ Speak (in voice channels)
   - ✅ Use Voice Activity

4. Copy the generated URL
5. Open the URL in your browser
6. Select your server and authorize the bot

### 5.2 Verify Bot Presence

Check that both bots appear in your server's member list (they'll show as offline until started).

## Step 6: Start the Bots

### Development Mode

Start both bots for testing:

```bash
# Start all bots with management script
./scripts/start-all.sh

# Or start individually for debugging
uv run python apps/clippy_bot/main.py
uv run python apps/music_bot/main.py
```

### Production Mode (systemd services)

For production deployment on Ubuntu:

```bash
# Install as system services
sudo ./systemd/install-services.sh

# Start services
sudo systemctl start clippy-bot music-bot

# Enable auto-start on boot
sudo systemctl enable clippy-bot music-bot

# Check service status
sudo systemctl status clippy-bot music-bot

# View logs
sudo journalctl -u clippy-bot -f
sudo journalctl -u music-bot -f
```

## Step 7: Test Bot Functionality

### 7.1 Test Clippy Bot

In any text channel:

- `/clippy` - Get chaotic assistance
- `/clippy_wisdom` - Receive questionable life advice
- `/clippy_poll` - Create polls with Clippy's options
- `/clippy_help` - Interactive help with buttons

### 7.2 Test Music Bot

Join a voice channel, then in any text channel:

- `/play Never Gonna Give You Up` - Play a song
- `/pause` / `/resume` - Control playback
- `/skip` - Skip current song
- `/playlist create MyPlaylist` - Create a playlist
- `/playlist add MyPlaylist [song]` - Add songs to playlist

## Step 8: Monitor and Maintain

### Health Check Endpoints

The bots provide monitoring endpoints:

```bash
# Check bot health
curl http://localhost:8081/health  # Clippy Bot
curl http://localhost:8082/health  # Music Bot

# Get detailed metrics
curl http://localhost:8081/metrics  # System and bot metrics
```

### Log Management

```bash
# View real-time logs (development)
tail -f clippy_bot.log
tail -f music_bot.log

# View service logs (production)
sudo journalctl -u clippy-bot -f
sudo journalctl -u music-bot -f
```

### Updates and Maintenance

```bash
# Pull latest changes
git pull origin main

# Update dependencies
uv sync --all-extras

# Restart services (production)
sudo systemctl restart clippy-bot music-bot

# Restart bots (development)
./scripts/stop-all.sh && ./scripts/start-all.sh
```

## Troubleshooting

### Common Issues

**Bot appears offline:**

- Verify token is correct in configuration
- Check that bot has proper permissions in Discord
- Review logs for connection errors

**Slash commands not appearing:**

- Wait up to 1 hour for global command sync
- Use guild-specific sync for faster testing (set `GUILD_ID`)
- Verify bot has "Use Slash Commands" permission

**Music bot can't join voice:**

- Ensure bot has "Connect" and "Speak" permissions
- Check voice channel user limit
- Verify you're in a voice channel when using commands

**Permission errors:**

- Bot needs "Send Messages" in channels where it responds
- Music bot needs voice permissions in target channels
- Check role hierarchy (bot role should be above roles it needs to manage)

### Debug Mode

Enable debug logging for troubleshooting:

```env
CLIPPY_DEBUG=true
MUSIC_DEBUG=true
```

### Validation

Run the validation script to check configuration:

```bash
uv run python validate.py
```

## Security Best Practices

1. **Never commit tokens** to version control
2. **Use environment variables** for sensitive configuration
3. **Keep tokens secure** - treat them like passwords
4. **Regularly update dependencies**: `uv sync --all-extras`
5. **Monitor logs** for suspicious activity
6. **Use least-privilege permissions** when inviting bots

## Production Considerations

### Reverse Proxy Setup

For production with HTTPS:

```bash
# Install Caddy
sudo apt install caddy

# Copy Caddyfile configuration
sudo cp Caddyfile /etc/caddy/Caddyfile

# Edit domain configuration
sudo nano /etc/caddy/Caddyfile

# Start Caddy
sudo systemctl enable caddy && sudo systemctl start caddy
```

### Backup Strategy

```bash
# Backup database and configuration
tar -czf bot-backup-$(date +%Y%m%d).tar.gz data/ config.json .env

# Automated daily backups
echo "0 2 * * * cd /path/to/discord-bot-framework && tar -czf backup-\$(date +\%Y\%m\%d).tar.gz data/ config.json .env" | crontab -
```

## Need Help?

- Check the [main README](README.md) for architecture details
- Review logs for specific error messages
- Ensure all permissions are correctly configured
- Verify network connectivity and firewall settings

Your bots should now be running and responding to commands in your Discord server!
