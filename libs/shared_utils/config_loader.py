"""Configuration loading utilities for Discord bots."""

import os
from dataclasses import dataclass
from typing import Optional
from dotenv import load_dotenv


@dataclass
class BotConfig:
    """Configuration class for Discord bots."""
    
    token: str
    guild_id: Optional[int] = None
    command_prefix: str = "!"
    debug: bool = False
    
    @classmethod
    def from_env(cls, bot_name: str) -> "BotConfig":
        """Load configuration from environment variables."""
        load_dotenv()
        
        token_key = f"{bot_name.upper()}_BOT_TOKEN"
        guild_key = f"{bot_name.upper()}_GUILD_ID"
        prefix_key = f"{bot_name.upper()}_COMMAND_PREFIX"
        debug_key = f"{bot_name.upper()}_DEBUG"
        
        token = os.getenv(token_key)
        if not token:
            raise ValueError(f"Missing required environment variable: {token_key}")
        
        guild_id = os.getenv(guild_key)
        guild_id = int(guild_id) if guild_id else None
        
        return cls(
            token=token,
            guild_id=guild_id,
            command_prefix=os.getenv(prefix_key, "!"),
            debug=os.getenv(debug_key, "false").lower() == "true"
        )


def load_config(bot_name: str) -> BotConfig:
    """Load configuration for a specific bot."""
    return BotConfig.from_env(bot_name)