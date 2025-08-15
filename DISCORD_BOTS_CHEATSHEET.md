# Discord Bots Cheat Sheet

This repository contains three independent Discord bots. Each bot is its own application in the Discord Developer Portal, has its own bot token, and must be invited to your server separately. There are no shared commands between bots.

## Clippy Bot (Slash Commands)

- /clippy: Unhinged Clippy response.
- /clippy_wisdom: An embedded dose of “wisdom.”
- /clippy_help: Help embed with interactive buttons.
- Random behavior: 2% chance to reply to any message; periodic random posts every 30–90 minutes (if enabled).
- Cooldowns: Per-user command cooldowns enforced by the framework.

Required permissions when inviting: Send Messages, Use Application Commands, Add Reactions, Read Message History.

## Music Bot (Slash Commands)

Playback

- /play query:<url or search>: Join your voice channel and queue a track (must be in a voice channel).
- /pause: Pause current song.
- /resume: Resume playback.
- /skip: Skip current song.
- /stop: Stop playback, clear queue, and disconnect.
- /queue: Show now playing and upcoming items.

Playlists (top-level commands; requires database configured)

- /playlist_create name:<text>: Create a playlist.
- /playlist_list: List your playlists.
- /playlist_show playlist_id:<number>: Show songs in a playlist.
- /playlist_play playlist_id:<number>: Queue all songs in a playlist. (Not yet implemented in code)
- /playlist_add playlist_id:<number>: Add current song to a playlist. (Not yet implemented in code)
- /playlist_remove playlist_id:<number> song_number:<number>: Remove a song. (Not yet implemented in code)
- /playlist_delete playlist_id:<number>: Delete a playlist. (Not yet implemented in code)

Notes

- Host requirements: yt-dlp and FFmpeg installed and on PATH (yt-dlp used for extraction; FFmpeg required for real audio playback; current code simulates playback timing).
- Inactivity: Auto-disconnect after configured idle time.
- Cooldowns: Per-user command cooldowns enforced by the framework.

Required permissions when inviting: Send Messages, Use Application Commands, Connect, Speak, Read Message History.

## MTG Card Bot (Prefix Commands)

Prefix: Default is ! (configurable via env/Config).

Core

- !<card name>: Fetch card info; supports partial/fuzzy names.
- !random: Random card.
- !help: Help and examples.
- !stats: Live bot metrics summary.
- !cache: Cache performance stats.

Multi-card grid

- Use semicolons in a single message to fetch multiple cards and send a grid of images:
  - Example: "!city of brass e:arn; library of alexandria e:arn; juzam djinn e:arn; serendib efreet e:arn"

Filters (examples)

- e:<set>, set:<set>, frame:<year>, border:<color>, is:foil, is:nonfoil, is:fullart, is:textless, is:borderless, rarity:<r>

Required permissions when inviting: Send Messages, Embed Links, Attach Files, Read Message History.
Also enable the “Message Content Intent” for this bot in the Developer Portal (it reads message content to parse prefix commands).

## Invite and Permissions (Must Do Per Bot)

- Scopes: Include both bot and applications.commands when generating invite URLs for Clippy and Music (slash-command bots). MTG Card Bot only needs bot.
- Permissions: Choose the permissions listed above per bot when generating the invite URL.

## Slash Commands Not Appearing? Quick Fix Checklist

1) Ensure the correct OAuth scopes

- Re-invite each slash-command bot with scopes: bot and applications.commands.

2) Use Guild-scoped registration for fast propagation

- Set CLIPPY_GUILD_ID and/or MUSIC_GUILD_ID (or set guild_id in config.json). Guild commands appear almost instantly. Without a guild_id, commands are registered globally and may take up to 1 hour to appear.

3) Verify logs on startup

- You should see logs like "Registered command" for each slash command. Enable debug logs via CLIPPY_DEBUG=true or MUSIC_DEBUG=true (or log_level DEBUG in config).

4) Confirm the token and app match

- Each bot must use its own application’s bot token (from the app you invited). Mismatched tokens cause commands to register under a different app.

5) Permissions in Discord

- Ensure the bot role isn’t restricted from Use Application Commands in your server/channels.

6) Last resort: wait or re-register

- Global commands can take time. With a guild_id configured, restart the bot to remove and re-register commands for the target guild.

That’s it—use this as the single reference for all three bots in this repo.
