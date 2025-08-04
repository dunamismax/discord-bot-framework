"""Health check system for Discord bots."""

import asyncio
import aiohttp
from aiohttp import web
import json
import logging
from datetime import datetime
from typing import Dict, Any


class HealthCheckServer:
    """HTTP server for health checks and monitoring."""
    
    def __init__(self, bot, port: int = 8080):
        self.bot = bot
        self.port = port
        self.app = web.Application()
        self.runner = None
        self.site = None
        self.logger = logging.getLogger(__name__)
        
        # Set up routes
        self.app.router.add_get('/health', self.health_check)
        self.app.router.add_get('/metrics', self.metrics)
        self.app.router.add_get('/status', self.status)
    
    async def start(self):
        """Start the health check server."""
        try:
            self.runner = web.AppRunner(self.app)
            await self.runner.setup()
            
            self.site = web.TCPSite(self.runner, 'localhost', self.port)
            await self.site.start()
            
            self.logger.info(f"Health check server started on http://localhost:{self.port}")
        except Exception as e:
            self.logger.error(f"Failed to start health check server: {e}")
    
    async def stop(self):
        """Stop the health check server."""
        if self.site:
            await self.site.stop()
        if self.runner:
            await self.runner.cleanup()
        self.logger.info("Health check server stopped")
    
    async def health_check(self, request):
        """Basic health check endpoint."""
        is_healthy = (
            self.bot.is_ready() and
            not self.bot.is_closed() and
            self.bot.latency > 0
        )
        
        status_code = 200 if is_healthy else 503
        response_data = {
            'status': 'healthy' if is_healthy else 'unhealthy',
            'timestamp': datetime.utcnow().isoformat(),
            'bot_ready': self.bot.is_ready(),
            'bot_closed': self.bot.is_closed(),
            'latency_ms': round(self.bot.latency * 1000, 2) if self.bot.latency > 0 else None
        }
        
        return web.json_response(response_data, status=status_code)
    
    async def metrics(self, request):
        """Detailed metrics endpoint."""
        try:
            total_members = sum(guild.member_count for guild in self.bot.guilds)
            
            # Command stats if available
            command_stats = []
            if hasattr(self.bot, 'db') and self.bot.db:
                try:
                    stats = await self.bot.db.get_command_stats(1)  # Last 24 hours
                    command_stats = stats[:10]  # Top 10 commands
                except Exception as e:
                    self.logger.error(f"Error getting command stats: {e}")
            
            metrics_data = {
                'bot_info': {
                    'name': self.bot.__class__.__name__,
                    'user_id': self.bot.user.id if self.bot.user else None,
                    'username': str(self.bot.user) if self.bot.user else None,
                    'start_time': self.bot.start_time.isoformat() if hasattr(self.bot, 'start_time') else None
                },
                'status': {
                    'ready': self.bot.is_ready(),
                    'closed': self.bot.is_closed(),
                    'latency_ms': round(self.bot.latency * 1000, 2) if self.bot.latency > 0 else None
                },
                'statistics': {
                    'guilds': len(self.bot.guilds),
                    'total_members': total_members,
                    'cogs_loaded': len(self.bot.cogs)
                },
                'commands': {
                    'total_application_commands': len(list(self.bot.walk_application_commands())),
                    'top_commands_24h': command_stats
                },
                'timestamp': datetime.utcnow().isoformat()
            }
            
            return web.json_response(metrics_data)
            
        except Exception as e:
            self.logger.error(f"Error generating metrics: {e}")
            return web.json_response(
                {'error': 'Failed to generate metrics', 'timestamp': datetime.utcnow().isoformat()},
                status=500
            )
    
    async def status(self, request):
        """Simple status endpoint for load balancers."""
        if self.bot.is_ready() and not self.bot.is_closed():
            return web.Response(text='OK', status=200)
        else:
            return web.Response(text='NOT OK', status=503)


class HealthCheckMixin:
    """Mixin to add health check functionality to bot classes."""
    
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.health_server: HealthCheckServer = None
    
    async def start_health_server(self, port: int = 8080):
        """Start the health check server."""
        self.health_server = HealthCheckServer(self, port)
        await self.health_server.start()
    
    async def stop_health_server(self):
        """Stop the health check server."""
        if self.health_server:
            await self.health_server.stop()