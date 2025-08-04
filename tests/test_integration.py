"""Integration tests for the Discord bot framework."""

import pytest
import asyncio
import tempfile
import os
from unittest.mock import Mock, AsyncMock, patch
import discord

# Add the libs directory to the Python path
import sys
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'libs'))

from shared_utils import BotConfig, load_config, BaseBot, Database, DatabaseMixin


class TestBotIntegration:
    """Integration tests for the complete bot lifecycle."""
    
    @pytest.fixture
    async def mock_bot(self):
        """Create a mock bot for testing."""
        config = BotConfig(
            token="test_token_12345678901234567890123456789012345678901234567890",
            guild_id=123456789012345678,
            debug=True
        )
        
        # Mock Discord client
        with patch('discord.Client.__init__', return_value=None):
            bot = BaseBot(config)
            bot.user = Mock()
            bot.user.id = 123456789
            bot.guilds = []
            bot.is_ready = Mock(return_value=True)
            bot.is_closed = Mock(return_value=False)
            bot.latency = 0.1
            yield bot
    
    @pytest.mark.asyncio
    async def test_bot_initialization(self, mock_bot):
        """Test bot initialization with configuration."""
        assert mock_bot.config.token == "test_token_12345678901234567890123456789012345678901234567890"
        assert mock_bot.config.guild_id == 123456789012345678
        assert mock_bot.config.debug is True
        assert hasattr(mock_bot, 'user_cooldowns')
        assert isinstance(mock_bot.user_cooldowns, dict)
    
    @pytest.mark.asyncio
    async def test_configuration_validation(self):
        """Test configuration validation system."""
        # Test valid configuration
        valid_config = BotConfig(
            token="test_token_12345678901234567890123456789012345678901234567890",
            guild_id=123456789012345678,
            health_check_port=8080
        )
        errors = valid_config.validate()
        assert len(errors) == 0
        
        # Test invalid configuration
        invalid_config = BotConfig(
            token="short",  # Too short
            guild_id=-1,  # Invalid
            health_check_port=99999,  # Too high
            max_playlist_size=-1  # Negative
        )
        errors = invalid_config.validate()
        assert len(errors) > 0
        assert any("token appears to be invalid" in error for error in errors)
        assert any("Guild ID must be a positive integer" in error for error in errors)
        assert any("Health check port must be between" in error for error in errors)
        assert any("Max playlist size must be positive" in error for error in errors)
    
    @pytest.mark.asyncio
    async def test_user_cooldown_system(self, mock_bot):
        """Test user cooldown functionality."""
        user_id = 123456789
        command_name = "test_command"
        
        # Initially no cooldown
        assert not mock_bot.is_user_on_cooldown(user_id, command_name)
        assert mock_bot.get_user_cooldown_remaining(user_id, command_name) == 0.0
        
        # Set cooldown
        mock_bot.set_user_cooldown(user_id, command_name)
        
        # Should be on cooldown now
        assert mock_bot.is_user_on_cooldown(user_id, command_name)
        remaining = mock_bot.get_user_cooldown_remaining(user_id, command_name)
        assert remaining > 0
        assert remaining <= mock_bot.config.command_cooldown
    
    @pytest.mark.asyncio
    async def test_input_validation(self, mock_bot):
        """Test input validation system."""
        # Valid input
        valid_input = "Hello, world!"
        result = mock_bot.validate_input(valid_input)
        assert result == "Hello, world!"
        
        # Empty input
        with pytest.raises(ValueError, match="Input cannot be empty"):
            mock_bot.validate_input("", allow_empty=False)
        
        # Too long input
        long_input = "a" * 2001
        with pytest.raises(ValueError, match="Input too long"):
            mock_bot.validate_input(long_input, max_length=2000)
        
        # Dangerous input
        dangerous_input = "<script>alert('xss')</script>"
        with pytest.raises(ValueError, match="potentially dangerous content"):
            mock_bot.validate_input(dangerous_input)
    
    @pytest.mark.asyncio
    async def test_database_integration(self):
        """Test database integration with bot."""
        # Create temporary database
        with tempfile.NamedTemporaryFile(suffix='.db', delete=False) as tmp:
            db_path = tmp.name
        
        try:
            # Create a mock database mixin bot
            class TestBot(DatabaseMixin, BaseBot):
                def __init__(self):
                    config = BotConfig(
                        token="test_token_12345678901234567890123456789012345678901234567890"
                    )
                    with patch('discord.Client.__init__', return_value=None):
                        super().__init__(config)
                    self.user = Mock()
                    self.user.id = 123456789
            
            bot = TestBot()
            
            # Setup database
            await bot.setup_database(db_path)
            assert bot.db is not None
            
            # Test database operations
            guild_id = 123456789012345678
            settings = {'test_key': 'test_value'}
            
            await bot.db.set_guild_settings(guild_id, settings)
            retrieved_settings = await bot.db.get_guild_settings(guild_id)
            assert retrieved_settings == settings
            
            # Test command logging
            await bot.db.log_command_usage('test_command', 987654321, guild_id)
            stats = await bot.db.get_command_stats(1)
            assert len(stats) == 1
            assert stats[0]['command'] == 'test_command'
            
            # Cleanup
            await bot.close_database()
            
        finally:
            # Clean up temporary file
            try:
                os.unlink(db_path)
            except FileNotFoundError:
                pass
    
    @pytest.mark.asyncio
    async def test_retry_mechanism(self, mock_bot):
        """Test retry mechanism for API calls."""
        # Mock function that fails twice then succeeds
        call_count = 0
        
        async def failing_function():
            nonlocal call_count
            call_count += 1
            if call_count < 3:
                raise discord.HTTPException(Mock(), "Test error")
            return "success"
        
        # Should succeed after retries
        result = await mock_bot.safe_api_call(failing_function)
        assert result == "success"
        assert call_count == 3
    
    @pytest.mark.asyncio
    async def test_error_handling_integration(self, mock_bot):
        """Test comprehensive error handling."""
        # Test different types of Discord errors
        test_cases = [
            (discord.Forbidden(Mock(), "Forbidden"), "Forbidden action"),
            (discord.NotFound(Mock(), "Not found"), "Resource not found"),
        ]
        
        for exception, expected_log in test_cases:
            async def failing_function():
                raise exception
            
            with pytest.raises(type(exception)):
                await mock_bot.safe_api_call(failing_function)


class TestConfigurationIntegration:
    """Test configuration loading and validation integration."""
    
    @pytest.mark.asyncio
    async def test_load_config_with_validation(self):
        """Test loading configuration with validation enabled."""
        # Test with invalid token (should raise ValueError)
        with patch.dict(os.environ, {'TEST_BOT_TOKEN': 'invalid_token'}):
            with pytest.raises(ValueError, match="Configuration validation failed"):
                load_config('test', validate_config=True)
    
    @pytest.mark.asyncio
    async def test_load_config_without_validation(self):
        """Test loading configuration with validation disabled."""
        # Should succeed even with invalid token
        with patch.dict(os.environ, {'TEST_BOT_TOKEN': 'invalid_token'}):
            config = load_config('test', validate_config=False)
            assert config.token == 'invalid_token'


class TestHealthCheckIntegration:
    """Test health check system integration."""
    
    @pytest.fixture
    async def mock_health_server(self, mock_bot):
        """Create a mock health check server."""
        from shared_utils.health_check import HealthCheckServer
        
        server = HealthCheckServer(mock_bot, port=8888)
        yield server
    
    @pytest.mark.asyncio
    async def test_health_check_server_lifecycle(self, mock_health_server):
        """Test health check server start/stop lifecycle."""
        # Start server
        await mock_health_server.start()
        assert mock_health_server.runner is not None
        assert mock_health_server.site is not None
        
        # Stop server
        await mock_health_server.stop()
    
    @pytest.mark.asyncio
    async def test_rate_limiting_functionality(self, mock_health_server):
        """Test rate limiting in health check endpoints."""
        # Mock request object
        mock_request = Mock()
        mock_request.remote = "127.0.0.1"
        mock_request.headers = {}
        
        # Create rate limited handler
        async def dummy_handler(request):
            return {"status": "ok"}
        
        rate_limited_handler = mock_health_server.rate_limited(dummy_handler)
        
        # First request should succeed
        result = await rate_limited_handler(mock_request)
        assert result is not None
        
        # Subsequent requests within rate limit should also succeed
        # (Note: This is a simplified test - in real usage, you'd test the actual rate limiting)