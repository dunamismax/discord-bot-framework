"""Unhinged Microsoft Clippy Discord Bot."""

import asyncio
import os
import sys

# Add the libs directory to the Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..', 'libs'))

from shared_utils import BaseBot, load_config


class ClippyBot(BaseBot):
    """The Unhinged Microsoft Clippy Bot."""

    def __init__(self):
        config = load_config("clippy")
        super().__init__(config)

    async def setup_hook(self):
        """Set up the bot when it starts."""
        await self.load_cogs([
            "cogs.unhinged_responses",
            "shared_utils.help_system"
        ])

        # Sync slash commands
        if self.config.guild_id:
            guild = self.get_guild(self.config.guild_id)
            if guild:
                await self.tree.sync(guild=guild)
                self.logger.info(f"Synced commands to guild: {guild.name}")
        else:
            await self.tree.sync()
            self.logger.info("Synced commands globally")


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
