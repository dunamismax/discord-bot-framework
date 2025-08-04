"""Database utilities for Discord bots."""

import asyncio
import aiosqlite
import logging
from pathlib import Path
from typing import Any, Dict, List, Optional, Union
from datetime import datetime
import json


class Database:
    """Async SQLite database wrapper for Discord bots."""
    
    def __init__(self, db_path: Union[str, Path]):
        self.db_path = Path(db_path)
        self.logger = logging.getLogger(__name__)
        self._connection = None
        
    async def initialize(self):
        """Initialize the database and create tables."""
        # Ensure directory exists
        self.db_path.parent.mkdir(parents=True, exist_ok=True)
        
        self._connection = await aiosqlite.connect(self.db_path)
        self._connection.row_factory = aiosqlite.Row
        
        await self._create_tables()
        self.logger.info(f"Database initialized at {self.db_path}")
    
    async def _create_tables(self):
        """Create base tables for bot functionality."""
        # Guild settings table
        await self._connection.execute("""
            CREATE TABLE IF NOT EXISTS guild_settings (
                guild_id INTEGER PRIMARY KEY,
                settings TEXT NOT NULL DEFAULT '{}',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)
        
        # User data table
        await self._connection.execute("""
            CREATE TABLE IF NOT EXISTS user_data (
                user_id INTEGER PRIMARY KEY,
                guild_id INTEGER,
                data TEXT NOT NULL DEFAULT '{}',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (guild_id) REFERENCES guild_settings (guild_id)
            )
        """)
        
        # Command usage statistics
        await self._connection.execute("""
            CREATE TABLE IF NOT EXISTS command_usage (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                command_name TEXT NOT NULL,
                user_id INTEGER NOT NULL,
                guild_id INTEGER,
                used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)
        
        # Music bot specific tables
        await self._connection.execute("""
            CREATE TABLE IF NOT EXISTS music_playlists (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                owner_id INTEGER NOT NULL,
                guild_id INTEGER NOT NULL,
                songs TEXT NOT NULL DEFAULT '[]',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)
        
        await self._connection.commit()
    
    async def close(self):
        """Close the database connection."""
        if self._connection:
            await self._connection.close()
            self.logger.info("Database connection closed")
    
    async def get_guild_settings(self, guild_id: int) -> Dict[str, Any]:
        """Get settings for a guild."""
        cursor = await self._connection.execute(
            "SELECT settings FROM guild_settings WHERE guild_id = ?",
            (guild_id,)
        )
        row = await cursor.fetchone()
        
        if row:
            return json.loads(row['settings'])
        return {}
    
    async def set_guild_settings(self, guild_id: int, settings: Dict[str, Any]):
        """Set settings for a guild."""
        settings_json = json.dumps(settings)
        
        await self._connection.execute("""
            INSERT OR REPLACE INTO guild_settings (guild_id, settings, updated_at)
            VALUES (?, ?, CURRENT_TIMESTAMP)
        """, (guild_id, settings_json))
        
        await self._connection.commit()
    
    async def get_user_data(self, user_id: int, guild_id: Optional[int] = None) -> Dict[str, Any]:
        """Get data for a user."""
        if guild_id:
            cursor = await self._connection.execute(
                "SELECT data FROM user_data WHERE user_id = ? AND guild_id = ?",
                (user_id, guild_id)
            )
        else:
            cursor = await self._connection.execute(
                "SELECT data FROM user_data WHERE user_id = ? AND guild_id IS NULL",
                (user_id,)
            )
        
        row = await cursor.fetchone()
        
        if row:
            return json.loads(row['data'])
        return {}
    
    async def set_user_data(self, user_id: int, data: Dict[str, Any], guild_id: Optional[int] = None):
        """Set data for a user."""
        data_json = json.dumps(data)
        
        await self._connection.execute("""
            INSERT OR REPLACE INTO user_data (user_id, guild_id, data, updated_at)
            VALUES (?, ?, ?, CURRENT_TIMESTAMP)
        """, (user_id, guild_id, data_json))
        
        await self._connection.commit()
    
    async def log_command_usage(self, command_name: str, user_id: int, guild_id: Optional[int] = None):
        """Log command usage for analytics."""
        await self._connection.execute("""
            INSERT INTO command_usage (command_name, user_id, guild_id)
            VALUES (?, ?, ?)
        """, (command_name, user_id, guild_id))
        
        await self._connection.commit()
    
    async def get_command_stats(self, days: int = 30) -> List[Dict[str, Any]]:
        """Get command usage statistics."""
        cursor = await self._connection.execute("""
            SELECT command_name, COUNT(*) as usage_count
            FROM command_usage
            WHERE used_at >= datetime('now', '-{} days')
            GROUP BY command_name
            ORDER BY usage_count DESC
        """.format(days))
        
        rows = await cursor.fetchall()
        return [{"command": row['command_name'], "count": row['usage_count']} for row in rows]
    
    async def create_playlist(self, name: str, owner_id: int, guild_id: int) -> int:
        """Create a new music playlist."""
        cursor = await self._connection.execute("""
            INSERT INTO music_playlists (name, owner_id, guild_id)
            VALUES (?, ?, ?)
        """, (name, owner_id, guild_id))
        
        await self._connection.commit()
        return cursor.lastrowid
    
    async def get_playlist(self, playlist_id: int) -> Optional[Dict[str, Any]]:
        """Get a music playlist by ID."""
        cursor = await self._connection.execute("""
            SELECT * FROM music_playlists WHERE id = ?
        """, (playlist_id,))
        
        row = await cursor.fetchone()
        if row:
            return {
                "id": row['id'],
                "name": row['name'],
                "owner_id": row['owner_id'],
                "guild_id": row['guild_id'],
                "songs": json.loads(row['songs']),
                "created_at": row['created_at'],
                "updated_at": row['updated_at']
            }
        return None
    
    async def get_user_playlists(self, user_id: int, guild_id: int) -> List[Dict[str, Any]]:
        """Get all playlists for a user in a guild."""
        cursor = await self._connection.execute("""
            SELECT * FROM music_playlists 
            WHERE owner_id = ? AND guild_id = ?
            ORDER BY updated_at DESC
        """, (user_id, guild_id))
        
        rows = await cursor.fetchall()
        return [
            {
                "id": row['id'],
                "name": row['name'],
                "owner_id": row['owner_id'],
                "guild_id": row['guild_id'],
                "songs": json.loads(row['songs']),
                "created_at": row['created_at'],
                "updated_at": row['updated_at']
            }
            for row in rows
        ]
    
    async def add_song_to_playlist(self, playlist_id: int, song_data: Dict[str, Any]):
        """Add a song to a playlist."""
        # Get current playlist
        playlist = await self.get_playlist(playlist_id)
        if not playlist:
            return False
        
        # Add song to the list
        songs = playlist['songs']
        songs.append(song_data)
        
        # Update playlist
        await self._connection.execute("""
            UPDATE music_playlists 
            SET songs = ?, updated_at = CURRENT_TIMESTAMP
            WHERE id = ?
        """, (json.dumps(songs), playlist_id))
        
        await self._connection.commit()
        return True
    
    async def remove_song_from_playlist(self, playlist_id: int, song_index: int):
        """Remove a song from a playlist by index."""
        playlist = await self.get_playlist(playlist_id)
        if not playlist:
            return False
        
        songs = playlist['songs']
        if 0 <= song_index < len(songs):
            songs.pop(song_index)
            
            await self._connection.execute("""
                UPDATE music_playlists 
                SET songs = ?, updated_at = CURRENT_TIMESTAMP
                WHERE id = ?
            """, (json.dumps(songs), playlist_id))
            
            await self._connection.commit()
            return True
        
        return False
    
    async def delete_playlist(self, playlist_id: int, owner_id: int) -> bool:
        """Delete a playlist (only by owner)."""
        cursor = await self._connection.execute("""
            DELETE FROM music_playlists 
            WHERE id = ? AND owner_id = ?
        """, (playlist_id, owner_id))
        
        await self._connection.commit()
        return cursor.rowcount > 0


class DatabaseMixin:
    """Mixin to add database functionality to bot classes."""
    
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.db: Optional[Database] = None
    
    async def setup_database(self, db_path: Union[str, Path] = None):
        """Set up the database connection."""
        if db_path is None:
            db_path = f"data/{self.__class__.__name__.lower()}.db"
        
        self.db = Database(db_path)
        await self.db.initialize()
        
        # Log command usage automatically
        self.add_listener(self._log_command_usage, 'on_application_command_completion')
    
    async def _log_command_usage(self, ctx):
        """Automatically log command usage."""
        if self.db and hasattr(ctx, 'command') and ctx.command:
            await self.db.log_command_usage(
                ctx.command.name,
                ctx.author.id,
                ctx.guild.id if ctx.guild else None
            )
    
    async def close_database(self):
        """Close database connection."""
        if self.db:
            await self.db.close()