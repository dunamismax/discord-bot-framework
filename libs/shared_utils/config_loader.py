"""Configuration loading utilities for Discord bots."""

import os
import re
from dataclasses import dataclass, field
from typing import Optional, Dict, Any, List
from pathlib import Path
from dotenv import load_dotenv
import json
import logging


@dataclass
class BotConfig:
    """Configuration class for Discord bots."""
    
    token: str
    guild_id: Optional[int] = None
    command_prefix: str = "!"
    debug: bool = False
    database_url: Optional[str] = None
    health_check_port: int = 8080
    max_playlist_size: int = 100
    max_queue_size: int = 50
    auto_disconnect_timeout: int = 300  # 5 minutes
    command_cooldown: float = 1.0  # seconds
    allowed_file_formats: List[str] = field(default_factory=lambda: ['mp3', 'mp4', 'webm', 'ogg'])
    max_song_duration: int = 3600  # 1 hour in seconds
    ytdl_options: Dict[str, Any] = field(default_factory=dict)
    ffmpeg_options: Dict[str, Any] = field(default_factory=dict)
    
    @classmethod
    def from_env(cls, bot_name: str) -> "BotConfig":
        """Load configuration from environment variables."""
        load_dotenv()
        
        prefix = bot_name.upper()
        
        # Required settings
        token = os.getenv(f"{prefix}_BOT_TOKEN")
        if not token:
            raise ValueError(f"Missing required environment variable: {prefix}_BOT_TOKEN")
        
        # Optional settings with defaults
        guild_id = os.getenv(f"{prefix}_GUILD_ID")
        guild_id = int(guild_id) if guild_id else None
        
        return cls(
            token=token,
            guild_id=guild_id,
            command_prefix=os.getenv(f"{prefix}_COMMAND_PREFIX", "!"),
            debug=os.getenv(f"{prefix}_DEBUG", "false").lower() == "true",
            database_url=os.getenv(f"{prefix}_DATABASE_URL"),
            health_check_port=int(os.getenv(f"{prefix}_HEALTH_CHECK_PORT", "8080")),
            max_playlist_size=int(os.getenv(f"{prefix}_MAX_PLAYLIST_SIZE", "100")),
            max_queue_size=int(os.getenv(f"{prefix}_MAX_QUEUE_SIZE", "50")),
            auto_disconnect_timeout=int(os.getenv(f"{prefix}_AUTO_DISCONNECT_TIMEOUT", "300")),
            command_cooldown=float(os.getenv(f"{prefix}_COMMAND_COOLDOWN", "1.0")),
            max_song_duration=int(os.getenv(f"{prefix}_MAX_SONG_DURATION", "3600")),
        )
    
    @classmethod
    def from_file(cls, config_path: Path, bot_name: str) -> "BotConfig":
        """Load configuration from a JSON file."""
        if not config_path.exists():
            raise FileNotFoundError(f"Configuration file not found: {config_path}")
        
        with open(config_path, 'r') as f:
            data = json.load(f)
        
        bot_config = data.get(bot_name, {})
        
        # Token is still required from environment for security
        load_dotenv()
        token = os.getenv(f"{bot_name.upper()}_BOT_TOKEN")
        if not token:
            raise ValueError(f"Missing required environment variable: {bot_name.upper()}_BOT_TOKEN")
        
        return cls(
            token=token,
            guild_id=bot_config.get('guild_id'),
            command_prefix=bot_config.get('command_prefix', '!'),
            debug=bot_config.get('debug', False),
            database_url=bot_config.get('database_url'),
            health_check_port=bot_config.get('health_check_port', 8080),
            max_playlist_size=bot_config.get('max_playlist_size', 100),
            max_queue_size=bot_config.get('max_queue_size', 50),
            auto_disconnect_timeout=bot_config.get('auto_disconnect_timeout', 300),
            command_cooldown=bot_config.get('command_cooldown', 1.0),
            allowed_file_formats=bot_config.get('allowed_file_formats', ['mp3', 'mp4', 'webm', 'ogg']),
            max_song_duration=bot_config.get('max_song_duration', 3600),
            ytdl_options=bot_config.get('ytdl_options', {}),
            ffmpeg_options=bot_config.get('ffmpeg_options', {})
        )
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert config to dictionary (excluding sensitive data)."""
        config_dict = {
            'guild_id': self.guild_id,
            'command_prefix': self.command_prefix,
            'debug': self.debug,
            'database_url': '***' if self.database_url else None,
            'health_check_port': self.health_check_port,
            'max_playlist_size': self.max_playlist_size,
            'max_queue_size': self.max_queue_size,
            'auto_disconnect_timeout': self.auto_disconnect_timeout,
            'command_cooldown': self.command_cooldown,
            'allowed_file_formats': self.allowed_file_formats,
            'max_song_duration': self.max_song_duration,
            'ytdl_options': self.ytdl_options,
            'ffmpeg_options': self.ffmpeg_options
        }
        return config_dict
    
    def validate(self) -> List[str]:
        """Validate configuration and return any errors."""
        errors = []
        
        # Validate bot token
        if not self.token:
            errors.append("Bot token is required")
        elif len(self.token) < 50:
            errors.append("Bot token appears to be invalid (too short)")
        elif not re.match(r'^[A-Za-z0-9._-]+$', self.token):
            errors.append("Bot token contains invalid characters")
        
        # Validate guild ID if provided
        if self.guild_id is not None:
            if not isinstance(self.guild_id, int) or self.guild_id <= 0:
                errors.append("Guild ID must be a positive integer")
            elif len(str(self.guild_id)) < 15:  # Discord snowflakes are typically 17-19 digits
                errors.append("Guild ID appears to be invalid (too short)")
        
        # Validate ports
        if not (1 <= self.health_check_port <= 65535):
            errors.append("Health check port must be between 1 and 65535")
        
        # Validate limits
        if self.max_playlist_size <= 0:
            errors.append("Max playlist size must be positive")
        elif self.max_playlist_size > 1000:
            errors.append("Max playlist size is too large (max: 1000)")
        
        if self.max_queue_size <= 0:
            errors.append("Max queue size must be positive")
        elif self.max_queue_size > 500:
            errors.append("Max queue size is too large (max: 500)")
        
        if self.auto_disconnect_timeout < 30:
            errors.append("Auto disconnect timeout must be at least 30 seconds")
        elif self.auto_disconnect_timeout > 3600:
            errors.append("Auto disconnect timeout is too large (max: 1 hour)")
        
        if self.command_cooldown < 0:
            errors.append("Command cooldown cannot be negative")
        elif self.command_cooldown > 60:
            errors.append("Command cooldown is too large (max: 60 seconds)")
        
        if self.max_song_duration < 30:
            errors.append("Max song duration must be at least 30 seconds")
        elif self.max_song_duration > 7200:  # 2 hours
            errors.append("Max song duration is too large (max: 2 hours)")
        
        # Validate file formats
        valid_formats = {'mp3', 'mp4', 'webm', 'ogg', 'wav', 'm4a', 'flac'}
        invalid_formats = set(self.allowed_file_formats) - valid_formats
        if invalid_formats:
            errors.append(f"Invalid file formats: {', '.join(invalid_formats)}")
        
        # Validate command prefix
        if not self.command_prefix:
            errors.append("Command prefix cannot be empty")
        elif len(self.command_prefix) > 5:
            errors.append("Command prefix is too long (max: 5 characters)")
        
        return errors


def load_config(bot_name: str, config_file: Optional[Path] = None, validate_config: bool = True) -> BotConfig:
    """Load configuration for a specific bot.
    
    Args:
        bot_name: Name of the bot to load config for
        config_file: Optional path to JSON config file. If provided, loads from file.
                    Otherwise loads from environment variables.
        validate_config: Whether to validate the loaded configuration
    
    Returns:
        BotConfig instance
    
    Raises:
        ValueError: If configuration validation fails
    """
    if config_file and config_file.exists():
        config = BotConfig.from_file(config_file, bot_name)
    else:
        config = BotConfig.from_env(bot_name)
    
    if validate_config:
        errors = config.validate()
        if errors:
            error_msg = f"Configuration validation failed for {bot_name}:\n" + "\n".join(f"  - {error}" for error in errors)
            raise ValueError(error_msg)
    
    return config