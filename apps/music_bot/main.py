"""YouTube Music Discord Bot."""

import asyncio
import os
import sys

# Add the libs and project root directories to the Python path
project_root = os.path.join(os.path.dirname(__file__), '..', '..')
sys.path.insert(0, os.path.join(project_root, 'libs'))
sys.path.insert(0, project_root)

import discord

from shared_utils import BaseBot, DatabaseMixin, load_config


class MusicBot(DatabaseMixin, BaseBot):
    """YouTube Music Discord Bot."""

    def __init__(self):
        config = load_config("music")

        # Music bot needs voice state intents
        intents = discord.Intents.default()
        intents.message_content = True
        intents.voice_states = True
        intents.guilds = True

        super().__init__(config, intents)

    async def setup_hook(self):
        """Set up the bot when it starts."""
        if hasattr(self, '_setup_hook_called'):
            self.logger.info("setup_hook already called, skipping...")
            return
            
        self._setup_hook_called = True
        self.logger.info("Starting setup_hook...")
        
        # Initialize database
        try:
            await self.setup_database()
            self.logger.info("Database setup completed")
        except Exception as e:
            self.logger.error(f"Database setup failed: {e}")

        try:
            await self.load_cogs([
                "apps.music_bot.cogs.music_player",
                "shared_utils.help_system"
            ])
        except Exception as e:
            self.logger.error(f"Failed to load cogs: {e}")
            import traceback
            self.logger.error(f"Traceback: {traceback.format_exc()}")

        # Sync slash commands
        self.logger.info("Starting command sync...")
        if self.config.guild_id:
            guild = self.get_guild(self.config.guild_id)
            if guild:
                await self.sync_commands(guild_ids=[guild.id])
                self.logger.info(f"Synced commands to guild: {guild.name}")
            else:
                self.logger.error(f"Could not find guild with ID: {self.config.guild_id}")
        else:
            await self.sync_commands()
            self.logger.info("Synced commands globally")
        
        self.logger.info("setup_hook completed!")

    async def on_ready(self):
        """Called when the bot is ready."""
        await super().on_ready()
        # Manually call setup_hook if it wasn't called automatically
        if not hasattr(self, '_setup_hook_called'):
            await self.setup_hook()

    async def on_voice_state_update(self, member, before, after):
        """Handle voice state updates."""
        # If the bot is left alone in a voice channel, disconnect
        if member == self.user:
            return

        # Check if bot is in a voice channel and is now alone
        for guild in self.guilds:
            voice_client = guild.voice_client
            if voice_client and voice_client.channel:
                # Count non-bot members in the voice channel
                non_bot_members = [m for m in voice_client.channel.members if not m.bot]
                if len(non_bot_members) == 0:
                    # Bot is alone, disconnect after a short delay
                    await asyncio.sleep(5)
                    # Check again after delay
                    non_bot_members = [m for m in voice_client.channel.members if not m.bot]
                    if len(non_bot_members) == 0:
                        await voice_client.disconnect()
                        self.logger.info(f"Disconnected from {guild.name} - no users in voice channel")


async def main():
    """Main function to run the bot."""
    bot = MusicBot()

    try:
        await bot.start(bot.config.token)
    except KeyboardInterrupt:
        bot.logger.info("Bot shutdown requested")
    finally:
        if not bot.is_closed():
            await bot.close_database()
            await bot.close()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nBot shutdown complete")
