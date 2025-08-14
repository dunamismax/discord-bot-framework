# Slash Commands Reference

This document lists available slash commands for each independent bot app in this repository. There is no shared command surface between bots.

## Clippy Bot (Slash Commands)

- /clippy: Get an unhinged Clippy response.
- /clippy_wisdom: Receive Clippy’s questionable wisdom.
- /clippy_help: Get help (with interactive buttons).

Automatic features
- Random messages every 30–90 minutes (if enabled in config).
- ~2% chance to reply to any message with a quote.

Required invite scopes/permissions
- Scopes: bot, applications.commands
- Permissions: Send Messages, Use Application Commands, Add Reactions, Read Message History

## Music Bot (Slash Commands)

Playback
- /play query:<url or search>: Play from YouTube (must be in a voice channel).
- /pause: Pause current song.
- /resume: Resume playback.
- /skip: Skip current song.
- /stop: Stop, clear queue, and disconnect.
- /queue: Show the queue.

Playlists (top-level commands; requires DB)
- /playlist_create name:<text>: Create a playlist.
- /playlist_list: List your playlists.
- /playlist_show playlist_id:<number>: Show songs in a playlist.
- /playlist_play playlist_id:<number>: Queue a playlist. (Not yet implemented)
- /playlist_add playlist_id:<number>: Add current song. (Not yet implemented)
- /playlist_remove playlist_id:<number> song_number:<number>: Remove a song. (Not yet implemented)
- /playlist_delete playlist_id:<number>: Delete a playlist. (Not yet implemented)

Notes
- Requires yt-dlp and FFmpeg on host (FFmpeg playback not fully implemented; current code simulates timing).
- Inactivity timeout disconnects the bot when idle.

Required invite scopes/permissions
- Scopes: bot, applications.commands
- Permissions: Send Messages, Use Application Commands, Connect, Speak, Read Message History

## Note on the MTG Card Bot

The MTG Card Bot uses prefix commands (default "!") and does not register slash commands. See DISCORD_BOTS_CHEATSHEET.md for its command list and usage.

## Troubleshooting Slash Commands

- Invite link must include applications.commands scope for Clippy and Music.
- Set a guild_id in config (or CLIPPY_GUILD_ID/MUSIC_GUILD_ID) for fast, guild-scoped registration; global commands can take up to 1 hour to appear.
- Check startup logs for "Registered command" messages; enable DEBUG logging if needed.
