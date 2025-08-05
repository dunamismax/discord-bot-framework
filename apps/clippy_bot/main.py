"""Unhinged Microsoft Clippy Discord Bot."""

import asyncio
import os
import sys

# Add the libs and project root directories to the Python path
project_root = os.path.join(os.path.dirname(__file__), '..', '..')
sys.path.insert(0, os.path.join(project_root, 'libs'))
sys.path.insert(0, project_root)

from shared_utils import BaseBot, load_config


class ClippyBot(BaseBot):
    """The Unhinged Microsoft Clippy Bot."""

    def __init__(self):
        config = load_config("clippy")
        super().__init__(config)

    async def setup_hook(self):
        """Set up the bot when it starts."""
        if hasattr(self, '_setup_hook_called'):
            self.logger.info("setup_hook already called, skipping...")
            return
            
        self._setup_hook_called = True
        self.logger.info("Starting setup_hook...")
        
        try:
            await self.load_cogs([
                "apps.clippy_bot.cogs.unhinged_responses",
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


async def main():
    """Main function to run the bot."""
    bot = ClippyBot()

    try:
        await bot.start(bot.config.token)
    except KeyboardInterrupt:
        bot.logger.info("Bot shutdown requested")
    finally:
        if not bot.is_closed():
            await bot.close()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nBot shutdown complete")
