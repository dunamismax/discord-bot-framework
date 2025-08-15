# Discord Bot Deployment Guide

This guide walks you through setting up and deploying the Discord Bot Framework to your Discord server using Go and mage build tools.

## Prerequisites

- Go 1.23+ installed - [Download Go](https://golang.org/dl/)
- Discord account with server admin permissions
- Linux/macOS/Windows environment (Ubuntu recommended for production)
- FFmpeg installed (for music bot voice functionality)
- mage build tool - `go install github.com/magefile/mage@latest`

## Step 1: Create Discord Applications

### 1.1 Access Discord Developer Portal

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Log in with your Discord account

### 1.2 Create Bot Applications

Create **separate applications** for each bot you plan to use:

**For MTG Card Bot:**
1. Click "New Application"
2. Name it "MTG Card Bot" (or your preferred name)
3. Click "Create"

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
   - **Keep these tokens secure** - treat them like passwords
4. Under "Privileged Gateway Intents":
   - Enable "Message Content Intent" (required for MTG Card Bot prefix commands)
   - Enable "Server Members Intent" (recommended for Music Bot voice features)

## Step 2: Get Your Server ID

1. In Discord, enable Developer Mode:
   - User Settings → App Settings → Advanced → Developer Mode (toggle on)
2. Right-click your server name
3. Click "Copy Server ID"
4. Save this ID - you'll need it for configuration

## Step 3: Install and Setup

### 3.1 Clone and Setup Development Environment

```bash
# Clone repository
git clone <repository-url>
cd go-discord-bots

# Setup development environment (installs tools, creates configs)
mage setup
```

### 3.2 Configure Environment Variables

Create a `.env` file in the project root:

```bash
# Create your .env file from the example
cp .env.example .env
```

Edit the `.env` file and replace the placeholder values:
- `your_default_MTG_DISCORD_TOKEN_here` → Can be any of your bot tokens as a fallback
- `your_clippy_bot_token_here` → Your Clippy Bot token from Step 1.3
- `your_music_bot_token_here` → Your Music Bot token from Step 1.3
- `your_guild_id_for_testing` → Your server ID from Step 2

Note: The MTG Card Bot will use the global `MTG_DISCORD_TOKEN` unless you specify a specific token. You can add `MTG_MTG_DISCORD_TOKEN=your_mtg_token_here` if you want to use a separate token for the MTG bot.

## Step 4: Invite Bots to Your Server

### 4.1 Generate Invite URLs

For **each bot application** in the Developer Portal:

1. Go to OAuth2 → URL Generator
2. Select Scopes based on bot type:

   **For MTG Card Bot:**
   - `bot` (prefix commands only)
   
   **For Clippy Bot and Music Bot:**
   - `bot`
   - `applications.commands` (for slash commands)

3. Select Bot Permissions:

   **For MTG Card Bot:**
   - Send Messages
   - Embed Links
   - Attach Files
   - Read Message History

   **For Clippy Bot:**
   - Send Messages
   - Use Slash Commands
   - Read Message History
   - Add Reactions
   - Embed Links

   **For Music Bot:**
   - Send Messages
   - Use Slash Commands
   - Read Message History
   - Connect (to voice channels)
   - Speak (in voice channels)
   - Use Voice Activity

4. Copy the generated URL
5. Open the URL in your browser
6. Select your server and authorize the bot

### 4.2 Verify Bot Presence

Check that your bots appear in your server's member list (they'll show as offline until started).

## Step 5: Build and Run the Bots

### Development Mode

```bash
# Build all applications
mage build

# Run individual bots for testing
mage runMTG        # MTG Card Bot
mage runClipper    # Clippy Bot  
mage runMusic      # Music Bot

# Run all bots together (not recommended for production)
mage run

# Run with debug logging
mage dev
```

### Production Mode

For production deployment on Ubuntu using systemd:

1. **Create systemd service files** (example for MTG Card Bot):

```bash
# Create service file
sudo tee /etc/systemd/system/mtg-card-bot.service > /dev/null << 'EOF'
[Unit]
Description=MTG Card Bot
After=network.target

[Service]
Type=simple
User=botuser
WorkingDirectory=/path/to/go-discord-bots
Environment=MTG_DISCORD_TOKEN=your_token_here
ExecStart=/path/to/go-discord-bots/bin/mtg-card-bot
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
```

2. **Start and enable services:**

```bash
# Reload systemd
sudo systemctl daemon-reload

# Start services
sudo systemctl start mtg-card-bot

# Enable auto-start on boot
sudo systemctl enable mtg-card-bot

# Check service status
sudo systemctl status mtg-card-bot

# View logs
sudo journalctl -u mtg-card-bot -f
```

## Step 6: Test Bot Functionality

### 6.1 Test MTG Card Bot

In any text channel:
- `!lightning bolt` - Look up a Magic card
- `!random` - Get a random card
- `!help` - Show help and examples
- `!stats` - Show bot performance metrics

### 6.2 Test Clippy Bot

In any text channel:
- `/clippy` - Get chaotic assistance
- `/clippy_wisdom` - Receive questionable life advice
- `/clippy_help` - Interactive help with buttons

### 6.3 Test Music Bot

Join a voice channel, then in any text channel:
- `/play Never Gonna Give You Up` - Play a song
- `/pause` / `/resume` - Control playback
- `/skip` - Skip current song
- `/playlist_create MyPlaylist` - Create a playlist

## Step 7: Monitor and Maintain

### Available Mage Commands

```bash
# Development
mage setup                # Install dev tools, create configs
mage dev                  # Run with debug logging
mage build                # Build all applications
mage clean                # Clean build artifacts
mage reset                # Reset to fresh state

# Quality Assurance
mage fmt                  # Format code
mage vet                  # Static analysis
mage lint                 # Comprehensive linting
mage vulnCheck           # Security vulnerability scan
mage test                # Run test suite
mage ci                   # Complete CI pipeline

# View all available commands
mage help
```

### Log Management

```bash
# View application logs (adjust path as needed)
tail -f bot.log

# View service logs (production)
sudo journalctl -u mtg-card-bot -f
sudo journalctl -u clippy-bot -f
sudo journalctl -u music-bot -f
```

### Updates and Maintenance

```bash
# Pull latest changes
git pull origin main

# Rebuild applications
mage clean && mage build

# Restart services (production)
sudo systemctl restart mtg-card-bot clippy-bot music-bot

# Quality checks before deployment
mage ci
```

## Troubleshooting

### Common Issues

**Bot appears offline:**
- Verify token is correct in configuration
- Check that bot has proper permissions in Discord
- Review logs for connection errors

**Slash commands not appearing:**
- Wait up to 1 hour for global command sync
- Use guild-specific sync for faster testing (set `GUILD_ID` variables)
- Verify bot has "Use Slash Commands" permission

**MTG Card Bot prefix commands not working:**
- Ensure "Message Content Intent" is enabled for the bot
- Check that bot has "Send Messages" permission
- Verify the correct prefix is being used (default: `!`)

**Music bot can't join voice:**
- Ensure bot has "Connect" and "Speak" permissions
- Check voice channel user limit
- Verify FFmpeg is installed and accessible
- Make sure you're in a voice channel when using commands

**Build errors:**
- Ensure Go 1.23+ is installed: `go version`
- Update dependencies: `go mod tidy`
- Clean and rebuild: `mage clean && mage build`

### Debug Mode

Enable debug logging for troubleshooting:

```env
DEBUG=true
LOG_LEVEL=debug
```

### Performance Monitoring

```bash
# View bot statistics (MTG Card Bot)
!stats
!cache

# Check system resources
top -p $(pgrep -f mtg-card-bot)
```

## Security Best Practices

1. **Never commit tokens** to version control
2. **Use environment variables** for sensitive configuration  
3. **Keep tokens secure** - treat them like passwords
4. **Regularly update dependencies**: `go mod tidy && mage vulnCheck`
5. **Monitor logs** for suspicious activity
6. **Use least-privilege permissions** when inviting bots
7. **Run vulnerability scans**: `mage vulnCheck`

## Production Considerations

### Resource Requirements

- **Memory**: 50-100MB per bot under normal load
- **CPU**: <5% per bot under normal load  
- **Storage**: Minimal (logs and SQLite databases for Music bot)
- **Network**: Depends on usage (Discord API calls)

### Scaling Considerations

- Each bot can handle 1000+ concurrent users
- Consider load balancing for very high traffic
- SQLite databases are sufficient for most use cases
- Monitor cache hit rates for MTG Card Bot performance

### Backup Strategy

```bash
# Backup configuration and databases
tar -czf bot-backup-$(date +%Y%m%d).tar.gz .env config.json music.db

# Automated daily backups
echo "0 2 * * * cd /path/to/go-discord-bots && tar -czf backup-\$(date +\%Y\%m\%d).tar.gz .env config.json *.db" | crontab -
```

## Need Help?

- Check the [main README](README.md) for architecture details
- Review logs for specific error messages  
- Use `mage help` to see all available commands
- Run `mage ci` to check code quality and tests
- Ensure all permissions are correctly configured
- Verify network connectivity and firewall settings

Your bots should now be running and responding to commands in your Discord server!