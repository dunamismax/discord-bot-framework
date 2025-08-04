"""Tests for music bot functionality."""

import pytest
import asyncio
import os
import sys
from unittest.mock import Mock, AsyncMock, patch, MagicMock

# Add the apps directory to the Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'apps', 'music_bot'))
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'libs'))

from cogs.music_player import Song, MusicQueue, MusicPlayer


class TestSong:
    """Test Song dataclass."""
    
    def test_song_creation(self):
        """Test Song creation with basic data."""
        song = Song(
            title="Test Song",
            url="http://example.com/song.mp3",
            webpage_url="http://example.com/watch?v=123"
        )
        assert song.title == "Test Song"
        assert song.url == "http://example.com/song.mp3"
        assert song.webpage_url == "http://example.com/watch?v=123"
        assert song.duration is None
        assert song.requester is None
    
    def test_song_with_optional_fields(self):
        """Test Song creation with optional fields."""
        mock_user = Mock()
        song = Song(
            title="Test Song",
            url="http://example.com/song.mp3",
            webpage_url="http://example.com/watch?v=123",
            duration=180,
            requester=mock_user
        )
        assert song.duration == 180
        assert song.requester == mock_user


class TestMusicQueue:
    """Test MusicQueue functionality."""
    
    @pytest.fixture
    def queue(self):
        """Create a test music queue."""
        return MusicQueue()
    
    @pytest.fixture
    def sample_song(self):
        """Create a sample song."""
        return Song(
            title="Test Song",
            url="http://example.com/song.mp3",
            webpage_url="http://example.com/watch?v=123"
        )
    
    def test_queue_initialization(self, queue):
        """Test queue initialization."""
        assert len(queue.queue) == 0
        assert queue.current is None
        assert queue.is_playing is False
        assert queue.is_paused is False
    
    def test_add_song(self, queue, sample_song):
        """Test adding a song to the queue."""
        queue.add(sample_song)
        assert len(queue.queue) == 1
        assert queue.queue[0] == sample_song
    
    def test_next_song(self, queue, sample_song):
        """Test getting the next song from the queue."""
        queue.add(sample_song)
        next_song = queue.next()
        assert next_song == sample_song
        assert len(queue.queue) == 0
    
    def test_next_empty_queue(self, queue):
        """Test getting next song from empty queue."""
        next_song = queue.next()
        assert next_song is None
    
    def test_skip_song(self, queue, sample_song):
        """Test skipping a song."""
        queue.add(sample_song)
        skipped = queue.skip()
        assert skipped == sample_song
        assert len(queue.queue) == 0
    
    def test_clear_queue(self, queue, sample_song):
        """Test clearing the queue."""
        queue.add(sample_song)
        queue.current = sample_song
        queue.is_playing = True
        queue.is_paused = True
        
        queue.clear()
        
        assert len(queue.queue) == 0
        assert queue.current is None
        assert queue.is_playing is False
        assert queue.is_paused is False
    
    def test_get_queue_list(self, queue):
        """Test getting the queue as a list."""
        song1 = Song("Song 1", "url1", "web1")
        song2 = Song("Song 2", "url2", "web2")
        
        queue.add(song1)
        queue.add(song2)
        
        queue_list = queue.get_queue_list()
        assert len(queue_list) == 2
        assert queue_list[0] == song1
        assert queue_list[1] == song2


class TestMusicPlayer:
    """Test MusicPlayer cog functionality."""
    
    @pytest.fixture
    def mock_bot(self):
        """Create a mock bot."""
        bot = Mock()
        bot.logger = Mock()
        bot.loop = asyncio.get_event_loop()
        return bot
    
    @pytest.fixture
    def music_player(self, mock_bot):
        """Create a MusicPlayer instance."""
        return MusicPlayer(mock_bot)
    
    def test_music_player_initialization(self, music_player):
        """Test MusicPlayer initialization."""
        assert isinstance(music_player.queues, dict)
        assert isinstance(music_player.voice_clients, dict)
        assert isinstance(music_player.inactivity_timers, dict)
        assert 'format' in music_player.ydl_opts
        assert 'before_options' in music_player.ffmpeg_options
    
    def test_get_queue(self, music_player):
        """Test getting or creating a queue for a guild."""
        guild_id = 12345
        
        # First call should create a new queue
        queue1 = music_player.get_queue(guild_id)
        assert isinstance(queue1, MusicQueue)
        assert guild_id in music_player.queues
        
        # Second call should return the same queue
        queue2 = music_player.get_queue(guild_id)
        assert queue1 is queue2
    
    @pytest.mark.asyncio
    async def test_extract_song_info_url(self, music_player):
        """Test extracting song info from URL."""
        mock_info = {
            'title': 'Test Song',
            'url': 'http://example.com/song.mp3',
            'webpage_url': 'http://example.com/watch?v=123',
            'duration': 180
        }
        
        with patch('yt_dlp.YoutubeDL') as mock_ydl_class:
            mock_ydl = Mock()
            mock_ydl.extract_info.return_value = mock_info
            mock_ydl_class.return_value.__enter__.return_value = mock_ydl
            
            song = await music_player.extract_song_info('http://example.com/watch?v=123')
            
            assert song is not None
            assert song.title == 'Test Song'
            assert song.url == 'http://example.com/song.mp3'
            assert song.webpage_url == 'http://example.com/watch?v=123'
            assert song.duration == 180
    
    @pytest.mark.asyncio
    async def test_extract_song_info_search(self, music_player):
        """Test extracting song info from search query."""
        mock_info = {
            'entries': [{
                'title': 'Search Result',
                'url': 'http://example.com/result.mp3',
                'webpage_url': 'http://example.com/watch?v=456',
                'duration': 240
            }]
        }
        
        with patch('yt_dlp.YoutubeDL') as mock_ydl_class:
            mock_ydl = Mock()
            mock_ydl.extract_info.return_value = mock_info
            mock_ydl_class.return_value.__enter__.return_value = mock_ydl
            
            song = await music_player.extract_song_info('test search query')
            
            assert song is not None
            assert song.title == 'Search Result'
            mock_ydl.extract_info.assert_called_once_with('ytsearch:test search query', download=False)
    
    @pytest.mark.asyncio
    async def test_extract_song_info_error(self, music_player):
        """Test extracting song info with error."""
        with patch('yt_dlp.YoutubeDL') as mock_ydl_class:
            mock_ydl = Mock()
            mock_ydl.extract_info.side_effect = Exception("Test error")
            mock_ydl_class.return_value.__enter__.return_value = mock_ydl
            
            song = await music_player.extract_song_info('error query')
            
            assert song is None
            music_player.bot.logger.error.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_cancel_inactivity_timer(self, music_player):
        """Test canceling inactivity timer."""
        guild_id = 12345
        
        # Create a mock task
        mock_task = Mock()
        music_player.inactivity_timers[guild_id] = mock_task
        
        await music_player.cancel_inactivity_timer(guild_id)
        
        mock_task.cancel.assert_called_once()
        assert guild_id not in music_player.inactivity_timers
    
    @pytest.mark.asyncio
    async def test_cancel_inactivity_timer_no_timer(self, music_player):
        """Test canceling inactivity timer when none exists."""
        guild_id = 12345
        
        # Should not raise an error
        await music_player.cancel_inactivity_timer(guild_id)
        
        assert guild_id not in music_player.inactivity_timers