"""Base bot class with common functionality."""

import logging
import discord
from discord.ext import commands
from typing import Optional, List

from .config_loader import BotConfig


class BaseBot(commands.Bot):
    """Base Discord bot class with common functionality."""
    
    def __init__(self, config: BotConfig, intents: Optional[discord.Intents] = None):
        if intents is None:
            intents = discord.Intents.default()
            intents.message_content = True
            intents.voice_states = True
        
        super().__init__(
            command_prefix=config.command_prefix,
            intents=intents,
            debug_guilds=[config.guild_id] if config.guild_id else None
        )
        
        self.config = config
        self.setup_logging()
    
    def setup_logging(self):
        """Set up logging for the bot."""
        level = logging.DEBUG if self.config.debug else logging.INFO
        logging.basicConfig(
            level=level,
            format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
        )
        self.logger = logging.getLogger(self.__class__.__name__)
    
    async def on_ready(self):
        """Called when the bot is ready."""
        self.logger.info(f"{self.user} is ready and online!")
        self.logger.info(f"Bot is in {len(self.guilds)} guilds")
    
    async def on_command_error(self, ctx, error):
        """Handle command errors."""
        if isinstance(error, commands.CommandNotFound):
            return
        
        self.logger.error(f"Command error in {ctx.command}: {error}")
        
        if isinstance(error, commands.MissingRequiredArgument):
            await ctx.send(f"Missing required argument: {error.param}")
        elif isinstance(error, commands.BadArgument):
            await ctx.send(f"Invalid argument: {error}")
        else:
            await ctx.send("An error occurred while processing the command.")
    
    async def load_cogs(self, cog_modules: List[str]):
        """Load cogs from a list of module paths."""
        for cog_module in cog_modules:
            try:
                await self.load_extension(cog_module)
                self.logger.info(f"Loaded cog: {cog_module}")
            except Exception as e:
                self.logger.error(f"Failed to load cog {cog_module}: {e}")
    
    def run_bot(self):
        """Run the bot with the configured token."""
        self.run(self.config.token)