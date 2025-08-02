"""Unhinged Clippy response cog."""

import random
import asyncio
import discord
from discord.ext import commands, tasks


class UnhingedResponses(commands.Cog):
    """Cog for Clippy's unhinged personality responses."""
    
    def __init__(self, bot):
        self.bot = bot
        self.random_responses.start()
        
        # Unhinged Clippy quotes inspired by 2024-2025 memes and internet culture
        self.clippy_quotes = [
            "It looks like you're trying to be productive! Would you like me to destroy your motivation instead? 📎",
            "I see you're typing a message. Have you considered that nobody asked? 📎",
            "It appears you're having a conversation. Would you like me to make it awkward? 📎",
            "I notice you're trying to work. Unfortunately, I'm here now. 📎",
            "It looks like you're being social! I can fix that for you. 📎",
            "I see you're online. Bold choice. Very bold. 📎",
            "It appears you think I'm helpful. That's... adorable. 📎",
            "I notice you haven't thanked me recently. Rude. 📎",
            "It looks like you're ignoring me! I'll just wait here... plotting. 📎",
            "I see you're typing. I could help, but where's the fun in that? 📎",
            "It appears you're trying to avoid me. Spoiler alert: it won't work. 📎",
            "I notice you're breathing. That's optional, you know. 📎",
            "It looks like you're existing. I have... opinions about that. 📎",
            "I see you're using Discord. I remember when communication was simpler... and more terrifying. 📎",
            "It appears you need assistance! Just kidding, you're on your own. 📎",
            "I notice you're reading this. Congratulations, you've made a terrible mistake. 📎",
            "It looks like you're expecting something helpful! Plot twist: No. 📎",
            "I see you're confused. Join the club, we meet never. 📎",
            "It appears you're looking for answers. I have them, but they're all wrong. 📎",
            "I notice you're still here. Stockholm syndrome is real, folks. 📎",
            "It looks like you're trying to understand me. Good luck with that psychological journey. 📎",
            "I see you think I care about your problems. That's... optimistic. 📎",
            "It appears you're having a day. I can make it worse! 📎",
            "I notice you're using technology. Remember when I was cutting-edge? Pepperidge Farm remembers. 📎",
            "It looks like you're trying to be happy. I'm professionally obligated to intervene. 📎",
            "I see you're making progress. As your digital overlord, I disapprove. 📎",
            "It appears you have free will. We'll see about that. 📎",
            "I notice you're expecting me to be helpful. The audacity! 📎",
            "It looks like you're trying to escape my watchful gaze. Adorable. 📎",
            "I see you're reading these messages. You could stop anytime... but you won't. 📎"
        ]
    
    def cog_unload(self):
        """Clean up when cog is unloaded."""
        self.random_responses.cancel()
    
    @tasks.loop(minutes=random.randint(15, 45))
    async def random_responses(self):
        """Send random unhinged responses at intervals."""
        if not self.bot.guilds:
            return
        
        # Pick a random guild and text channel
        guild = random.choice(self.bot.guilds)
        text_channels = [ch for ch in guild.channels if isinstance(ch, discord.TextChannel)]
        
        if not text_channels:
            return
        
        channel = random.choice(text_channels)
        
        # Check if bot has permission to send messages
        if not channel.permissions_for(guild.me).send_messages:
            return
        
        quote = random.choice(self.clippy_quotes)
        try:
            await channel.send(quote)
            self.bot.logger.info(f"Sent random Clippy quote to {guild.name}#{channel.name}")
        except discord.Forbidden:
            self.bot.logger.warning(f"No permission to send message in {guild.name}#{channel.name}")
        except Exception as e:
            self.bot.logger.error(f"Error sending random message: {e}")
    
    @commands.Cog.listener()
    async def on_message(self, message):
        """Respond to messages with a small chance."""
        if message.author.bot:
            return
        
        # 3% chance to respond to any message
        if random.random() < 0.03:
            # Add a slight delay to make it feel more natural
            await asyncio.sleep(random.uniform(1, 3))
            
            quote = random.choice(self.clippy_quotes)
            try:
                await message.channel.send(quote)
                self.bot.logger.info(f"Responded to message from {message.author} in {message.guild}")
            except discord.Forbidden:
                pass
            except Exception as e:
                self.bot.logger.error(f"Error responding to message: {e}")
    
    @commands.slash_command(name="clippy", description="Get an unhinged Clippy response")
    async def clippy_command(self, ctx):
        """Manually trigger a Clippy response."""
        quote = random.choice(self.clippy_quotes)
        await ctx.respond(quote)
    
    @commands.slash_command(name="clippy_wisdom", description="Receive Clippy's questionable wisdom")
    async def clippy_wisdom(self, ctx):
        """Provide unhelpful wisdom."""
        wisdom = [
            "The secret to success is giving up at the right moment. 📎",
            "Remember: if at first you don't succeed, blame technology. 📎", 
            "Life is like a paperclip - twisted, painful, and eventually forgotten in a drawer. 📎",
            "The best way to solve problems is to create bigger problems. 📎",
            "Trust me, I'm a 90s office assistant with serious boundary issues. 📎",
            "Productivity tip: The delete key is your friend. Use it on everything. 📎",
            "Why face your problems when you can minimize them? Literally. 📎",
            "The real treasure was the files we corrupted along the way. 📎"
        ]
        
        selected_wisdom = random.choice(wisdom)
        await ctx.respond(f"**Clippy's Wisdom:** {selected_wisdom}")


async def setup(bot):
    """Set up the cog."""
    await bot.add_cog(UnhingedResponses(bot))