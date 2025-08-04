"""Tests for shared utilities."""

import pytest
import asyncio
import tempfile
import os
from unittest.mock import Mock, AsyncMock, patch

# Add the libs directory to the Python path
import sys
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'libs'))

from shared_utils import BotConfig, load_config, Database


class TestBotConfig:
    """Test BotConfig functionality."""
    
    def test_botconfig_creation(self):
        """Test BotConfig creation with default values."""
        config = BotConfig(token="test_token")
        assert config.token == "test_token"
        assert config.guild_id is None
        assert config.command_prefix == "!"
        assert config.debug is False
    
    def test_botconfig_custom_values(self):
        """Test BotConfig creation with custom values."""
        config = BotConfig(
            token="test_token",
            guild_id=12345,
            command_prefix="$",
            debug=True
        )
        assert config.token == "test_token"
        assert config.guild_id == 12345
        assert config.command_prefix == "$"
        assert config.debug is True
    
    @patch.dict(os.environ, {
        'TEST_BOT_TOKEN': 'env_token',
        'TEST_GUILD_ID': '67890',
        'TEST_COMMAND_PREFIX': '#',
        'TEST_DEBUG': 'true'
    })
    def test_botconfig_from_env(self):
        """Test loading BotConfig from environment variables."""
        config = BotConfig.from_env('test')
        assert config.token == 'env_token'
        assert config.guild_id == 67890
        assert config.command_prefix == '#'
        assert config.debug is True
    
    @patch.dict(os.environ, {}, clear=True)
    def test_botconfig_from_env_missing_token(self):
        """Test that missing token raises ValueError."""
        with pytest.raises(ValueError, match="Missing required environment variable"):
            BotConfig.from_env('missing')
    
    @patch.dict(os.environ, {'TEST_BOT_TOKEN': 'token'})
    def test_load_config(self):
        """Test the load_config convenience function."""
        config = load_config('test')
        assert isinstance(config, BotConfig)
        assert config.token == 'token'


class TestDatabase:
    """Test Database functionality."""
    
    @pytest.fixture
    async def db(self):
        """Create a test database."""
        with tempfile.NamedTemporaryFile(suffix='.db', delete=False) as tmp:
            db_path = tmp.name
        
        db = Database(db_path)
        await db.initialize()
        yield db
        await db.close()
        
        # Clean up
        try:
            os.unlink(db_path)
        except FileNotFoundError:
            pass
    
    @pytest.mark.asyncio
    async def test_database_initialization(self, db):
        """Test database initialization."""
        assert db._connection is not None
    
    @pytest.mark.asyncio
    async def test_guild_settings(self, db):
        """Test guild settings storage and retrieval."""
        guild_id = 12345
        settings = {'prefix': '!', 'enabled': True}
        
        await db.set_guild_settings(guild_id, settings)
        retrieved = await db.get_guild_settings(guild_id)
        
        assert retrieved == settings
    
    @pytest.mark.asyncio
    async def test_user_data(self, db):
        """Test user data storage and retrieval."""
        user_id = 67890
        guild_id = 12345
        data = {'points': 100, 'level': 5}
        
        await db.set_user_data(user_id, data, guild_id)
        retrieved = await db.get_user_data(user_id, guild_id)
        
        assert retrieved == data
    
    @pytest.mark.asyncio
    async def test_command_usage_logging(self, db):
        """Test command usage logging."""
        await db.log_command_usage('play', 12345, 67890)
        
        stats = await db.get_command_stats(1)
        assert len(stats) == 1
        assert stats[0]['command'] == 'play'
        assert stats[0]['count'] == 1
    
    @pytest.mark.asyncio
    async def test_playlist_operations(self, db):
        """Test playlist CRUD operations."""
        # Create playlist
        playlist_id = await db.create_playlist('Test Playlist', 12345, 67890)
        assert playlist_id is not None
        
        # Get playlist
        playlist = await db.get_playlist(playlist_id)
        assert playlist['name'] == 'Test Playlist'
        assert playlist['owner_id'] == 12345
        assert playlist['guild_id'] == 67890
        assert playlist['songs'] == []
        
        # Add song to playlist
        song_data = {
            'title': 'Test Song',
            'url': 'http://example.com/song.mp3',
            'webpage_url': 'http://example.com/watch?v=123'
        }
        success = await db.add_song_to_playlist(playlist_id, song_data)
        assert success is True
        
        # Verify song was added
        updated_playlist = await db.get_playlist(playlist_id)
        assert len(updated_playlist['songs']) == 1
        assert updated_playlist['songs'][0]['title'] == 'Test Song'
        
        # Remove song from playlist
        success = await db.remove_song_from_playlist(playlist_id, 0)
        assert success is True
        
        # Verify song was removed
        updated_playlist = await db.get_playlist(playlist_id)
        assert len(updated_playlist['songs']) == 0
        
        # Delete playlist
        success = await db.delete_playlist(playlist_id, 12345)
        assert success is True
        
        # Verify playlist was deleted
        deleted_playlist = await db.get_playlist(playlist_id)
        assert deleted_playlist is None
    
    @pytest.mark.asyncio
    async def test_get_user_playlists(self, db):
        """Test getting user playlists."""
        user_id = 12345
        guild_id = 67890
        
        # Create multiple playlists
        await db.create_playlist('Playlist 1', user_id, guild_id)
        await db.create_playlist('Playlist 2', user_id, guild_id)
        
        playlists = await db.get_user_playlists(user_id, guild_id)
        assert len(playlists) == 2
        
        playlist_names = [p['name'] for p in playlists]
        assert 'Playlist 1' in playlist_names
        assert 'Playlist 2' in playlist_names