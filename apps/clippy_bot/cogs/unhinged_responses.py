"""Unhinged Clippy response cog."""

import asyncio
import random

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
        # Check user cooldown
        if self.bot.is_user_on_cooldown(ctx.author.id, "clippy"):
            remaining = self.bot.get_user_cooldown_remaining(ctx.author.id, "clippy")
            await ctx.respond(
                f"Hold your horses! You can use this command again in {remaining:.1f} seconds. 📎",
                ephemeral=True
            )
            return

        # Set cooldown for user
        self.bot.set_user_cooldown(ctx.author.id, "clippy")

        quote = random.choice(self.clippy_quotes)
        await ctx.respond(quote)

    @commands.slash_command(name="clippy_wisdom", description="Receive Clippy's questionable wisdom")
    async def clippy_wisdom(self, ctx):
        """Provide unhelpful wisdom."""
        # Check user cooldown
        if self.bot.is_user_on_cooldown(ctx.author.id, "clippy_wisdom"):
            remaining = self.bot.get_user_cooldown_remaining(ctx.author.id, "clippy_wisdom")
            await ctx.respond(
                f"Patience, young grasshopper! Wisdom comes to those who wait {remaining:.1f} more seconds. 📎",
                ephemeral=True
            )
            return

        # Set cooldown for user
        self.bot.set_user_cooldown(ctx.author.id, "clippy_wisdom")

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

    @commands.slash_command(name="clippy_poll", description="Let Clippy create a chaotic poll")
    async def clippy_poll(self, ctx, question: str):
        """Create a poll with Clippy's unhinged options."""
        try:
            # Validate input
            question = self.bot.validate_input(question, max_length=200)
        except ValueError as e:
            await ctx.respond(f"❌ {e}", ephemeral=True)
            return

        # Check user cooldown
        if self.bot.is_user_on_cooldown(ctx.author.id, "clippy_poll"):
            remaining = self.bot.get_user_cooldown_remaining(ctx.author.id, "clippy_poll")
            await ctx.respond(
                f"Easy there, poll master! Try again in {remaining:.1f} seconds. 📎",
                ephemeral=True
            )
            return

        # Set cooldown for user
        self.bot.set_user_cooldown(ctx.author.id, "clippy_poll")

        # Clippy's chaotic poll options
        clippy_options = [
            "Definitely, but also definitely not 📎",
            "Ask me again when I care 📎",
            "Signs point to... confusion 📎",
            "My sources say 'maybe' 📎",
            "Cannot predict now, I'm busy 📎",
            "Don't count on it, count on chaos 📎",
            "Yes, but actually no 📎",
            "Reply hazy, try again never 📎"
        ]

        # Create embed with poll
        embed = discord.Embed(
            title="📎 Clippy's Chaotic Poll",
            description=f"**Question:** {question}\n\n**Choose your destiny:**",
            color=0x5865F2
        )

        # Add poll options (limited to 4 for simplicity)
        selected_options = random.sample(clippy_options, 4)
        for i, option in enumerate(selected_options, 1):
            embed.add_field(name=f"{i}️⃣", value=option, inline=False)

        embed.set_footer(text="Click the reactions below to vote! 📎")

        # Send the poll
        message = await ctx.respond(embed=embed)

        # Add reactions for voting
        reactions = ["1️⃣", "2️⃣", "3️⃣", "4️⃣"]
        try:
            # Get the actual message object to add reactions
            msg = await ctx.original_response()
            for reaction in reactions:
                await msg.add_reaction(reaction)
        except Exception as e:
            self.bot.logger.error(f"Failed to add reactions to poll: {e}")

    @commands.slash_command(name="clippy_help", description="Get help from Clippy (if you dare)")
    async def clippy_help_command(self, ctx):
        """Provide Clippy's version of help."""
        # Check user cooldown
        if self.bot.is_user_on_cooldown(ctx.author.id, "clippy_help"):
            remaining = self.bot.get_user_cooldown_remaining(ctx.author.id, "clippy_help")
            await ctx.respond(
                f"I'm busy being unhelpful! Come back in {remaining:.1f} seconds. 📎",
                ephemeral=True
            )
            return

        # Set cooldown for user
        self.bot.set_user_cooldown(ctx.author.id, "clippy_help")

        # Create interactive help with buttons
        embed = discord.Embed(
            title="📎 Clippy's \"Helpful\" Guide",
            description="I see you're trying to get help. Would you like me to make it worse?",
            color=0x5865F2
        )

        embed.add_field(
            name="🎭 Commands",
            value="`/clippy` - Get an unhinged response\n`/clippy_wisdom` - Questionable life advice\n`/clippy_poll` - Create chaotic polls\n`/clippy_help` - This mess",
            inline=False
        )

        embed.add_field(
            name="🤖 About Me",
            value="I'm Clippy, your friendly neighborhood chaos agent. I randomly appear in channels to provide unsolicited advice and existential dread.",
            inline=False
        )

        embed.set_footer(text="Remember: I'm here to help... sort of. 📎")

        # Create action buttons
        class ClippyHelpView(discord.ui.View):
            def __init__(self):
                super().__init__(timeout=60)

            @discord.ui.button(label="More Chaos", style=discord.ButtonStyle.danger, emoji="💥")
            async def more_chaos(self, button: discord.ui.Button, interaction: discord.Interaction):
                chaos_quotes = [
                    "Chaos achieved! Mission accomplished! 📎",
                    "You asked for more chaos. Bold choice. 📎",
                    "Congratulations! You've unlocked maximum confusion! 📎",
                    "I see you enjoy living dangerously. Respect. 📎"
                ]
                await interaction.response.send_message(random.choice(chaos_quotes), ephemeral=True)

            @discord.ui.button(label="I Regret This", style=discord.ButtonStyle.secondary, emoji="😭")
            async def regret(self, button: discord.ui.Button, interaction: discord.Interaction):
                regret_quotes = [
                    "Too late! I'm already in your computer! 📎",
                    "Regret is just chaos in disguise! 📎",
                    "You can't uninstall me from your nightmares! 📎",
                    "Regret? More like... re-Clippy! 📎"
                ]
                await interaction.response.send_message(random.choice(regret_quotes), ephemeral=True)

        await ctx.respond(embed=embed, view=ClippyHelpView())


async def setup(bot):
    """Set up the cog."""
    await bot.add_cog(UnhingedResponses(bot))
