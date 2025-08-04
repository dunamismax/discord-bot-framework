"""Base bot class with common functionality."""

import logging
import traceback
import sys
from datetime import datetime
import discord
from discord.ext import commands
from typing import Optional, List, Dict, Any

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
        self.start_time = datetime.utcnow()
        self.setup_logging()
    
    def setup_logging(self):
        """Set up comprehensive logging for the bot."""
        level = logging.DEBUG if self.config.debug else logging.INFO
        
        # Create formatters
        detailed_formatter = logging.Formatter(
            "%(asctime)s - %(name)s - %(levelname)s - %(funcName)s:%(lineno)d - %(message)s"
        )
        simple_formatter = logging.Formatter(
            "%(asctime)s - %(levelname)s - %(message)s"
        )
        
        # Set up root logger
        root_logger = logging.getLogger()
        root_logger.setLevel(level)
        
        # Clear existing handlers
        for handler in root_logger.handlers[:]:
            root_logger.removeHandler(handler)
        
        # Console handler
        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setLevel(level)
        console_handler.setFormatter(simple_formatter if not self.config.debug else detailed_formatter)
        root_logger.addHandler(console_handler)
        
        # Set up bot-specific logger
        self.logger = logging.getLogger(self.__class__.__name__)
        
        # Suppress Discord.py debug logs unless in debug mode
        if not self.config.debug:
            logging.getLogger('discord').setLevel(logging.WARNING)
            logging.getLogger('discord.http').setLevel(logging.WARNING)
    
    async def on_ready(self):
        """Called when the bot is ready."""
        self.logger.info(f"{self.user} is ready and online!")
        self.logger.info(f"Bot is in {len(self.guilds)} guilds")
    
    async def on_command_error(self, ctx, error):
        """Handle command errors with comprehensive logging."""
        if isinstance(error, commands.CommandNotFound):
            return
        
        # Log error with full traceback
        error_msg = f"Command error in {ctx.command}: {error}"
        self.logger.error(error_msg)
        
        if self.config.debug:
            self.logger.error(f"Full traceback:\n{''.join(traceback.format_exception(type(error), error, error.__traceback__))}")
        
        # User-friendly error responses
        if isinstance(error, commands.MissingRequiredArgument):
            await ctx.send(f"❌ Missing required argument: `{error.param}`")
        elif isinstance(error, commands.BadArgument):
            await ctx.send(f"❌ Invalid argument: {error}")
        elif isinstance(error, commands.CheckFailure):
            await ctx.send("❌ You don't have permission to use this command.")
        elif isinstance(error, commands.CommandOnCooldown):
            await ctx.send(f"❌ Command is on cooldown. Try again in {error.retry_after:.2f} seconds.")
        elif isinstance(error, discord.Forbidden):
            await ctx.send("❌ I don't have permission to perform this action.")
        else:
            await ctx.send("❌ An unexpected error occurred while processing the command.")
    
    async def on_application_command_error(self, ctx, error):
        """Handle application command (slash command) errors."""
        error_msg = f"Slash command error in {ctx.command}: {error}"
        self.logger.error(error_msg)
        
        if self.config.debug:
            self.logger.error(f"Full traceback:\n{''.join(traceback.format_exception(type(error), error, error.__traceback__))}")
        
        # User-friendly error responses for slash commands
        try:
            if isinstance(error, commands.MissingRequiredArgument):
                await ctx.respond(f"❌ Missing required argument: `{error.param}`", ephemeral=True)
            elif isinstance(error, commands.BadArgument):
                await ctx.respond(f"❌ Invalid argument: {error}", ephemeral=True)
            elif isinstance(error, commands.CheckFailure):
                await ctx.respond("❌ You don't have permission to use this command.", ephemeral=True)
            elif isinstance(error, commands.CommandOnCooldown):
                await ctx.respond(f"❌ Command is on cooldown. Try again in {error.retry_after:.2f} seconds.", ephemeral=True)
            elif isinstance(error, discord.Forbidden):
                await ctx.respond("❌ I don't have permission to perform this action.", ephemeral=True)
            else:
                await ctx.respond("❌ An unexpected error occurred while processing the command.", ephemeral=True)
        except discord.NotFound:
            # Interaction token is invalid/expired
            self.logger.warning("Could not respond to interaction - token may be expired")
    
    async def on_error(self, event_method, *args, **kwargs):
        """Handle general bot errors."""
        exc_type, exc_value, exc_traceback = sys.exc_info()
        error_msg = f"Error in {event_method}: {exc_value}"
        self.logger.error(error_msg)
        
        if self.config.debug:
            self.logger.error(f"Full traceback:\n{''.join(traceback.format_exception(exc_type, exc_value, exc_traceback))}")
    
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