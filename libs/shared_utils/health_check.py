"""Health check system for Discord bots."""

import logging
import time
from collections import defaultdict
from datetime import datetime

import psutil
from aiohttp import web


class HealthCheckServer:
    """HTTP server for health checks and monitoring."""

    def __init__(self, bot, port: int = 8080):
        self.bot = bot
        self.port = port
        self.app = web.Application()
        self.runner = None
        self.site = None
        self.logger = logging.getLogger(__name__)

        # Rate limiting
        self.rate_limits = defaultdict(list)
        self.rate_limit_window = 60  # 60 seconds
        self.rate_limit_max_requests = 30  # 30 requests per minute per IP

        # Set up routes
        self.app.router.add_get('/health', self.rate_limited(self.health_check))
        self.app.router.add_get('/metrics', self.rate_limited(self.metrics))
        self.app.router.add_get('/status', self.rate_limited(self.status))

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

    def rate_limited(self, handler):
        """Rate limiting decorator for endpoints."""
        async def wrapper(request):
            # Get client IP
            client_ip = request.remote
            if 'X-Forwarded-For' in request.headers:
                client_ip = request.headers['X-Forwarded-For'].split(',')[0].strip()
            elif 'X-Real-IP' in request.headers:
                client_ip = request.headers['X-Real-IP']

            current_time = time.time()

            # Clean old requests outside the window
            self.rate_limits[client_ip] = [
                req_time for req_time in self.rate_limits[client_ip]
                if current_time - req_time < self.rate_limit_window
            ]

            # Check if rate limit exceeded
            if len(self.rate_limits[client_ip]) >= self.rate_limit_max_requests:
                self.logger.warning(f"Rate limit exceeded for IP: {client_ip}")
                return web.json_response(
                    {
                        'error': 'Rate limit exceeded',
                        'limit': self.rate_limit_max_requests,
                        'window_seconds': self.rate_limit_window,
                        'retry_after': self.rate_limit_window
                    },
                    status=429,
                    headers={'Retry-After': str(self.rate_limit_window)}
                )

            # Add current request to rate limit tracking
            self.rate_limits[client_ip].append(current_time)

            # Call the actual handler
            return await handler(request)

        return wrapper

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

            # System performance metrics
            process = psutil.Process()
            memory_info = process.memory_info()
            cpu_percent = process.cpu_percent()

            # System-wide metrics
            system_memory = psutil.virtual_memory()
            system_disk = psutil.disk_usage('/')

            # Calculate uptime
            uptime_seconds = None
            if hasattr(self.bot, 'start_time'):
                uptime_seconds = (datetime.utcnow() - self.bot.start_time).total_seconds()

            metrics_data = {
                'bot_info': {
                    'name': self.bot.__class__.__name__,
                    'user_id': self.bot.user.id if self.bot.user else None,
                    'username': str(self.bot.user) if self.bot.user else None,
                    'start_time': self.bot.start_time.isoformat() if hasattr(self.bot, 'start_time') else None,
                    'uptime_seconds': uptime_seconds
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
                'performance': {
                    'process': {
                        'cpu_percent': cpu_percent,
                        'memory_mb': round(memory_info.rss / 1024 / 1024, 2),
                        'memory_percent': round(memory_info.rss / system_memory.total * 100, 2),
                        'threads': process.num_threads(),
                        'open_files': len(process.open_files()) if hasattr(process, 'open_files') else 0
                    },
                    'system': {
                        'cpu_percent': psutil.cpu_percent(interval=1),
                        'memory_percent': system_memory.percent,
                        'memory_available_mb': round(system_memory.available / 1024 / 1024, 2),
                        'disk_percent': system_disk.percent,
                        'disk_free_gb': round(system_disk.free / 1024 / 1024 / 1024, 2),
                        'load_average': psutil.getloadavg() if hasattr(psutil, 'getloadavg') else None
                    }
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
