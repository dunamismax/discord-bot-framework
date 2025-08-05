# Slash Commands Reference

This document lists all available slash commands for both Discord bots in this framework.

## 🤖 Clippy Bot Commands

The Unhinged Microsoft Clippy Bot provides chaotic entertainment and questionable assistance.

### Core Commands

#### `/clippy`

- **Description**: Get an unhinged Clippy response
- **Usage**: `/clippy`
- **Cooldown**: User-based cooldown applies
- **Example**: `/clippy` → "It looks like you're trying to be productive! Would you like me to destroy your motivation instead? 📎"

#### `/clippy_wisdom`

- **Description**: Receive Clippy's questionable wisdom
- **Usage**: `/clippy_wisdom`
- **Cooldown**: User-based cooldown applies
- **Example**: `/clippy_wisdom` → "**Clippy's Wisdom:** The secret to success is giving up at the right moment. 📎"

#### `/clippy_poll <question>`

- **Description**: Let Clippy create a chaotic poll
- **Usage**: `/clippy_poll question:`
- **Parameters**:
  - `question` (required): The poll question (max 200 characters)
- **Cooldown**: User-based cooldown applies
- **Example**: `/clippy_poll question: Should we order pizza?` → Creates a poll with 4 random chaotic options

#### `/clippy_help`

- **Description**: Get help from Clippy (if you dare)
- **Usage**: `/clippy_help`
- **Cooldown**: User-based cooldown applies
- **Features**: Interactive help with buttons for "More Chaos" and "I Regret This"

### Automatic Features

- **Random Messages**: Clippy randomly sends unhinged messages to text channels (15-45 minute intervals)
- **Message Responses**: 3% chance to respond to any user message with a random quote

---

## 🎵 Music Bot Commands

The YouTube Music Bot provides full-featured music playback and playlist management.

### Playback Commands

#### `/play <query>`

- **Description**: Play music from YouTube
- **Usage**: `/play query:`
- **Parameters**:
  - `query` (required): YouTube URL or search terms
- **Requirements**: Must be in a voice channel
- **Example**: `/play query: Never Gonna Give You Up`

#### `/pause`

- **Description**: Pause the current song
- **Usage**: `/pause`
- **Requirements**: Music must be playing

#### `/resume`

- **Description**: Resume the current song
- **Usage**: `/resume`
- **Requirements**: Music must be paused

#### `/skip`

- **Description**: Skip the current song
- **Usage**: `/skip`
- **Requirements**: Music must be playing

#### `/stop`

- **Description**: Stop music and clear the queue
- **Usage**: `/stop`
- **Effect**: Stops playback, clears queue, and disconnects from voice channel

#### `/queue`

- **Description**: Show the current music queue
- **Usage**: `/queue`
- **Display**: Shows currently playing song and upcoming songs in queue

### Playlist Management Commands

Playlist commands are organized under the `/playlist` command group:

#### `/playlist create <name>`

- **Description**: Create a new playlist
- **Usage**: `/playlist create name:`
- **Parameters**:
  - `name` (required): Playlist name (max 100 characters)
- **Limit**: Maximum 10 playlists per user per server

#### `/playlist list`

- **Description**: List your playlists
- **Usage**: `/playlist list`
- **Display**: Shows all your playlists in the current server with IDs

#### `/playlist show <playlist_id>`

- **Description**: Show songs in a playlist
- **Usage**: `/playlist show playlist_id:`
- **Parameters**:
  - `playlist_id` (required): ID of the playlist to display

#### `/playlist play <playlist_id>`

- **Description**: Play a playlist
- **Usage**: `/playlist play playlist_id:`
- **Parameters**:
  - `playlist_id` (required): ID of the playlist to play
- **Requirements**: Must be in a voice channel
- **Effect**: Adds all songs from playlist to the queue

#### `/playlist add <playlist_id>`

- **Description**: Add current song to a playlist
- **Usage**: `/playlist add playlist_id:`
- **Parameters**:
  - `playlist_id` (required): ID of your playlist
- **Requirements**: A song must be currently playing
- **Permissions**: Can only add to your own playlists

#### `/playlist remove <playlist_id> <song_number>`

- **Description**: Remove a song from a playlist
- **Usage**: `/playlist remove playlist_id: song_number:`
- **Parameters**:
  - `playlist_id` (required): ID of your playlist
  - `song_number` (required): Position of song in playlist (1-based)
- **Permissions**: Can only modify your own playlists

#### `/playlist delete <playlist_id>`

- **Description**: Delete a playlist
- **Usage**: `/playlist delete playlist_id:`
- **Parameters**:
  - `playlist_id` (required): ID of your playlist to delete
- **Permissions**: Can only delete your own playlists
- **Warning**: This action is permanent

---

## 📚 Shared Commands

These commands are available for both bots through the shared help system:

#### `/help [category]`

- **Description**: Show bot commands and information
- **Usage**: `/help` or `/help category:`
- **Parameters**:
  - `category` (optional): Specific command category to view
- **Display**: Shows all available commands organized by category

#### `/info`

- **Description**: Show bot information and statistics
- **Usage**: `/info`
- **Display**: Shows bot details, server count, member count, uptime, and technical information

---

## 🔧 Technical Notes

### Bot Architecture

- **Clippy Bot**: Based on `discord.Bot` for slash command support
- **Music Bot**: Based on `discord.Bot` with database integration for playlist persistence
- **Shared Utilities**: Common functionality including help system, error handling, and logging

### Permissions Required

- **Clippy Bot**: Send Messages, Use Slash Commands, Add Reactions
- **Music Bot**: Send Messages, Use Slash Commands, Connect to Voice, Speak in Voice

### Configuration

- Both bots use environment variables for configuration (`.env` file)
- Guild-specific command registration for faster updates
- Debug logging available when `DEBUG=true` in environment

### Cooldowns

- User-based cooldowns prevent spam (configurable via bot config)
- Cooldown messages provide helpful feedback with remaining time

### Error Handling

- Comprehensive error logging for debugging
- User-friendly error messages for common issues
- Automatic retry logic for Discord API calls
