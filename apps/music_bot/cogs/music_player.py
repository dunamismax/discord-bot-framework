"""YouTube music player cog."""

import asyncio
import discord
import yt_dlp
from discord.ext import commands, tasks
from typing import Dict, List, Optional
from dataclasses import dataclass
from collections import deque


@dataclass
class Song:
    """Represents a song in the queue."""
    title: str
    url: str
    webpage_url: str
    duration: Optional[int] = None
    requester: Optional[discord.Member] = None


class MusicQueue:
    """Manages the music queue for a guild."""
    
    def __init__(self):
        self.queue: deque[Song] = deque()
        self.current: Optional[Song] = None
        self.is_playing = False
        self.is_paused = False
    
    def add(self, song: Song):
        """Add a song to the queue."""
        self.queue.append(song)
    
    def next(self) -> Optional[Song]:
        """Get the next song from the queue."""
        if self.queue:
            return self.queue.popleft()
        return None
    
    def skip(self) -> Optional[Song]:
        """Skip current song and get next."""
        if self.queue:
            return self.queue.popleft()
        return None
    
    def clear(self):
        """Clear the entire queue."""
        self.queue.clear()
        self.current = None
        self.is_playing = False
        self.is_paused = False
    
    def get_queue_list(self) -> List[Song]:
        """Get a list of songs in the queue."""
        return list(self.queue)


class MusicPlayer(commands.Cog):
    """YouTube music player cog."""
    
    def __init__(self, bot):
        self.bot = bot
        self.queues: Dict[int, MusicQueue] = {}
        self.voice_clients: Dict[int, discord.VoiceClient] = {}
        self.inactivity_timers: Dict[int, asyncio.Task] = {}
        
        # yt-dlp options for best audio quality streaming
        self.ydl_opts = {
            'format': 'bestaudio/best',
            'noplaylist': True,
            'quiet': True,
            'no_warnings': True,
            'extractaudio': True,
            'audioformat': 'mp3',
            'outtmpl': '%(extractor)s-%(id)s-%(title)s.%(ext)s',
            'restrictfilenames': True,
            'logtostderr': False,
        }
        
        # FFmpeg options for Discord voice
        self.ffmpeg_options = {
            'before_options': '-nostdin',
            'options': '-vn -b:a 128k'
        }
    
    def get_queue(self, guild_id: int) -> MusicQueue:
        """Get or create a music queue for a guild."""
        if guild_id not in self.queues:
            self.queues[guild_id] = MusicQueue()
        return self.queues[guild_id]
    
    async def extract_song_info(self, query: str) -> Optional[Song]:
        """Extract song information from YouTube URL or search query."""
        try:
            with yt_dlp.YoutubeDL(self.ydl_opts) as ydl:
                # If it's not a URL, search for it
                if not query.startswith(('http://', 'https://')):
                    query = f"ytsearch:{query}"
                
                info = ydl.extract_info(query, download=False)
                
                if 'entries' in info:
                    # Take the first result from search
                    info = info['entries'][0]
                
                return Song(
                    title=info.get('title', 'Unknown'),
                    url=info.get('url'),
                    webpage_url=info.get('webpage_url'),
                    duration=info.get('duration')
                )
        except Exception as e:
            self.bot.logger.error(f"Error extracting song info: {e}")
            return None
    
    async def play_next(self, guild_id: int):
        """Play the next song in the queue."""
        queue = self.get_queue(guild_id)
        voice_client = self.voice_clients.get(guild_id)
        
        if not voice_client or not voice_client.is_connected():
            return
        
        next_song = queue.next()
        if next_song:
            queue.current = next_song
            queue.is_playing = True
            queue.is_paused = False
            
            try:
                source = discord.FFmpegPCMAudio(
                    next_song.url,
                    **self.ffmpeg_options
                )
                
                def after_playing(error):
                    if error:
                        self.bot.logger.error(f"Player error: {error}")
                    
                    # Schedule next song
                    asyncio.run_coroutine_threadsafe(
                        self.play_next(guild_id),
                        self.bot.loop
                    )
                
                voice_client.play(source, after=after_playing)
                self.bot.logger.info(f"Now playing: {next_song.title}")
                
                # Reset inactivity timer
                await self.reset_inactivity_timer(guild_id)
                
            except Exception as e:
                self.bot.logger.error(f"Error playing song: {e}")
                queue.is_playing = False
        else:
            queue.is_playing = False
            queue.current = None
            # Start inactivity timer when queue is empty
            await self.start_inactivity_timer(guild_id)
    
    async def start_inactivity_timer(self, guild_id: int):
        """Start the 5-minute inactivity timer."""
        await self.cancel_inactivity_timer(guild_id)
        
        async def inactivity_timeout():
            await asyncio.sleep(300)  # 5 minutes
            voice_client = self.voice_clients.get(guild_id)
            if voice_client and voice_client.is_connected():
                queue = self.get_queue(guild_id)
                if not queue.is_playing and len(queue.queue) == 0:
                    await voice_client.disconnect()
                    self.voice_clients.pop(guild_id, None)
                    self.bot.logger.info(f"Disconnected from guild {guild_id} due to inactivity")
        
        self.inactivity_timers[guild_id] = asyncio.create_task(inactivity_timeout())
    
    async def reset_inactivity_timer(self, guild_id: int):
        """Reset the inactivity timer."""
        await self.cancel_inactivity_timer(guild_id)
    
    async def cancel_inactivity_timer(self, guild_id: int):
        """Cancel the inactivity timer."""
        if guild_id in self.inactivity_timers:
            self.inactivity_timers[guild_id].cancel()
            del self.inactivity_timers[guild_id]
    
    @commands.slash_command(name="play", description="Play music from YouTube")
    async def play(self, ctx, *, query: str):
        """Play a song from YouTube."""
        await ctx.defer()
        
        # Check if user is in a voice channel
        if not ctx.author.voice:
            await ctx.followup.send("You need to be in a voice channel to use this command!")
            return
        
        # Get or connect to voice channel
        voice_channel = ctx.author.voice.channel
        voice_client = self.voice_clients.get(ctx.guild.id)
        
        if not voice_client:
            try:
                voice_client = await voice_channel.connect()
                self.voice_clients[ctx.guild.id] = voice_client
            except Exception as e:
                await ctx.followup.send(f"Failed to connect to voice channel: {e}")
                return
        elif voice_client.channel != voice_channel:
            await voice_client.move_to(voice_channel)
        
        # Extract song information
        song = await self.extract_song_info(query)
        if not song:
            await ctx.followup.send("Could not find or load the requested song.")
            return
        
        song.requester = ctx.author
        queue = self.get_queue(ctx.guild.id)
        
        # Add to queue
        queue.add(song)
        
        if not queue.is_playing:
            await self.play_next(ctx.guild.id)
            await ctx.followup.send(f"🎵 Now playing: **{song.title}**")
        else:
            await ctx.followup.send(f"🎵 Added to queue: **{song.title}**\nPosition in queue: {len(queue.queue)}")
    
    @commands.slash_command(name="pause", description="Pause the current song")
    async def pause(self, ctx):
        """Pause the current song."""
        voice_client = self.voice_clients.get(ctx.guild.id)
        queue = self.get_queue(ctx.guild.id)
        
        if voice_client and voice_client.is_playing():
            voice_client.pause()
            queue.is_paused = True
            await ctx.respond("⏸️ Paused the current song.")
        else:
            await ctx.respond("Nothing is currently playing.")
    
    @commands.slash_command(name="resume", description="Resume the current song")
    async def resume(self, ctx):
        """Resume the current song."""
        voice_client = self.voice_clients.get(ctx.guild.id)
        queue = self.get_queue(ctx.guild.id)
        
        if voice_client and voice_client.is_paused():
            voice_client.resume()
            queue.is_paused = False
            await ctx.respond("▶️ Resumed the current song.")
        else:
            await ctx.respond("Nothing is currently paused.")
    
    @commands.slash_command(name="skip", description="Skip the current song")
    async def skip(self, ctx):
        """Skip the current song."""
        voice_client = self.voice_clients.get(ctx.guild.id)
        queue = self.get_queue(ctx.guild.id)
        
        if voice_client and (voice_client.is_playing() or voice_client.is_paused()):
            voice_client.stop()
            await ctx.respond("⏭️ Skipped the current song.")
        else:
            await ctx.respond("Nothing is currently playing.")
    
    @commands.slash_command(name="stop", description="Stop music and clear the queue")
    async def stop(self, ctx):
        """Stop music and clear the queue."""
        voice_client = self.voice_clients.get(ctx.guild.id)
        queue = self.get_queue(ctx.guild.id)
        
        if voice_client:
            if voice_client.is_playing() or voice_client.is_paused():
                voice_client.stop()
            
            await voice_client.disconnect()
            self.voice_clients.pop(ctx.guild.id, None)
            await self.cancel_inactivity_timer(ctx.guild.id)
        
        queue.clear()
        await ctx.respond("⏹️ Stopped music and cleared the queue. Disconnected from voice channel.")
    
    @commands.slash_command(name="queue", description="Show the current music queue")
    async def show_queue(self, ctx):
        """Show the current music queue."""
        queue = self.get_queue(ctx.guild.id)
        
        if not queue.current and len(queue.queue) == 0:
            await ctx.respond("The queue is empty.")
            return
        
        embed = discord.Embed(title="🎵 Music Queue", color=0x00ff00)
        
        if queue.current:
            status = "⏸️ Paused" if queue.is_paused else "▶️ Playing"
            embed.add_field(
                name=f"{status} Now",
                value=f"**{queue.current.title}**\nRequested by: {queue.current.requester.mention if queue.current.requester else 'Unknown'}",
                inline=False
            )
        
        if queue.queue:
            queue_list = []
            for i, song in enumerate(list(queue.queue)[:10], 1):  # Show first 10 songs
                queue_list.append(f"{i}. **{song.title}** - {song.requester.mention if song.requester else 'Unknown'}")
            
            embed.add_field(
                name="Up Next",
                value="\n".join(queue_list),
                inline=False
            )
            
            if len(queue.queue) > 10:
                embed.add_field(
                    name="",
                    value=f"... and {len(queue.queue) - 10} more songs",
                    inline=False
                )
        
        await ctx.respond(embed=embed)


async def setup(bot):
    """Set up the cog."""
    await bot.add_cog(MusicPlayer(bot))