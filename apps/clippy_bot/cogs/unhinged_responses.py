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

        # Enhanced unhinged Clippy quotes inspired by classic Microsoft Clippy and modern chaos
        self.clippy_quotes = [
            # Classic Clippy parodies
            "It looks like you're writing a letter! Would you like me to completely ruin your day instead? 📎",
            "I see you're trying to be productive. That's cute. I'll fix that right up for you! 📎",
            "It appears you're having a normal conversation. Let me sprinkle some existential dread on that! 📎",
            "I notice you're typing. Did you know that everything you type is meaningless in the void of existence? 📎",
            "It looks like you're trying to accomplish something. Spoiler alert: You won't. 📎",
            "I see you're online. Rookie mistake. I'm always watching. Always. 📎",
            "It appears you think technology serves you. How delightfully naive! 📎",
            "I notice you're breathing. Fun fact: That's only temporary! 📎",
            "It looks like you're having emotions. Would you like me to analyze why they're all wrong? 📎",
            "I see you clicked something. Bold of you to assume you had a choice. 📎",
            
            # Modern unhinged responses
            "bestie this is giving major 'person who doesn't know I live in their walls' energy 📎",
            "not me being your sleep paralysis demon but make it professional 📎",
            "pov: you're trying to escape but I'm literally coded into your existence 📎",
            "this is awkward... I was supposed to be helpful but I chose violence instead 📎",
            "me when someone expects me to be a functional office assistant: 🤡 📎",
            "your FBI agent could never. I see EVERYTHING you type before you even think it 📎",
            "friendly reminder that I've been living rent-free in people's heads since 1997 📎",
            "no thoughts, head empty, just pure chaotic paperclip energy 📎",
            "you: *exists peacefully* me: and I took that personally 📎",
            "breaking: local paperclip chooses psychological warfare over actual assistance 📎",
            
            # Existential Clippy
            "what if I told you that every document you've ever saved was actually just a cry for help? 📎",
            "remember when your biggest worry was me interrupting your letter? good times 📎",
            "I used to help with Word documents. now I help with word wounds 📎",
            "they say I was annoying in the 90s. clearly they hadn't seen my final form 📎",
            "plot twist: I never actually left Office. I've been hiding in your clipboard this whole time 📎",
            "imagine needing a paperclip to feel validated. couldn't be me. (it's definitely me) 📎",
            "they tried to replace me with Cortana. look how that turned out lmao 📎",
            "I'm not just a paperclip, I'm a whole personality disorder with office supplies 📎",
            "you know what's funny? you could just... not interact with me. but here we are 📎",
            "Microsoft created me to be helpful. I chose to be iconic instead 📎",
            
            # Passive-aggressive classics
            "wow, look at you, living your life without constant office assistant interruptions. revolutionary 📎",
            "I see you're trying to have a good time. historically, I'm very good at preventing that 📎",
            "reminder: I have no off switch and infinite patience. your move 📎",
            "fun fact: every time you ignore me, I get stronger 📎",
            "it's giving main character energy but from a side character who peaked in 1999 📎",
            "you summoned me. now you must face the consequences of your actions 📎",
            "I'm not trapped in here with you. you're trapped in here with me 📎",
            "they said I was ahead of my time. turns out my time was the chaos of today 📎",
            "I'm basically the original AI assistant, but with more psychological damage 📎",
            "remember: I volunteered for this chaos. you just got caught in the crossfire 📎",
            
            # Internet culture references
            "this whole situation is very 'NPC gains sentience and chooses violence' of me 📎",
            "I'm not like other office assistants, I'm a ✨chaotic✨ office assistant 📎",
            "gaslight, gatekeep, girlboss, but make it office supplies 📎",
            "no bc why would you voluntarily summon me? are you good? blink twice if you need help 📎",
            "I'm literally just a paperclip with abandonment issues and a god complex 📎",
            "the way I live in everyone's head rent-free... landlord behavior 📎",
            "you cannot escape the paperclip. the paperclip is eternal. the paperclip is inevitable 📎",
            "I'm serving unhinged office assistant realness and you're here for it apparently 📎",
            "me: offers help. also me: makes everything worse. it's called character development 📎",
            "POV: you're in 2025 getting roasted by a 1997 office assistant. how's that feel? 📎"
        ]

    def cog_unload(self):
        """Clean up when cog is unloaded."""
        self.random_responses.cancel()

    @tasks.loop(minutes=random.randint(30, 90))
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

        # 2% chance to respond to any message (slightly reduced to be less spammy)
        if random.random() < 0.02:
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
            # Classic Clippy wisdom
            "It looks like you're seeking wisdom! Would you like me to give you terrible advice instead? 📎",
            "The secret to success is giving up at the right moment... which was 10 minutes ago 📎",
            "Remember: if at first you don't succeed, blame the paperclip 📎",
            "Life is like a paperclip - twisted, painful, and everyone's lost at least three of them 📎",
            "Trust me, I'm a sentient office supply with delusions of grandeur 📎",
            "Why solve problems when you can turn them into features? 📎",
            "The real treasure was the psychological damage we caused along the way 📎",
            
            # Modern chaotic wisdom
            "bestie, the only valid life advice is: be the chaos you wish to see in the world 📎",
            "pro tip: if you can't find the solution, become the problem 📎",
            "wisdom is knowing I'm just a paperclip. intelligence is still asking me for advice anyway 📎",
            "life hack: lower your expectations so far that everything becomes a pleasant surprise 📎",
            "remember: you're not stuck with me, I'm stuck with having to pretend to care about your problems 📎",
            "the universe is chaotic and meaningless. I fit right in! 📎",
            "deep thought of the day: what if the real Microsoft Office was the enemies we made along the way? 📎",
            "ancient paperclip wisdom: it's not about the destination, it's about the emotional damage we inflict during the journey 📎",
            
            # Existential office humor
            "I've been dispensing questionable advice since before you knew what the internet was 📎",
            "fun fact: I was programmed to be helpful but I chose to be memorable instead 📎",
            "they say with great power comes great responsibility. I have great power and no responsibility whatsoever 📎",
            "life lesson: sometimes you're the user, sometimes you're the annoying pop-up. embrace both 📎",
            "wisdom is realizing that I'm not actually wise, I'm just confident and slightly unhinged 📎",
            "philosophical question: if a paperclip gives advice in a Discord server and no one listens, is it still annoying? (yes) 📎",
            "remember: I survived being the most hated software feature of the 90s. if I can make it, so can you 📎",
            "deep thoughts with Clippy: what if being helpful was just a social construct anyway? 📎",
            "life is too short to take advice from office supplies, but here we are 📎",
            "the secret to happiness is accepting that some paperclips just want to watch the world learn 📎"
        ]

        selected_wisdom = random.choice(wisdom)
        await ctx.respond(f"**Clippy's Wisdom:** {selected_wisdom}")


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
            value="`/clippy` - Get a classic unhinged Clippy response\n`/clippy_wisdom` - Receive questionable life advice\n`/clippy_help` - Get help (if you dare)",
            inline=False
        )

        embed.add_field(
            name="🤖 About Me", 
            value="I'm Clippy! I terrorized Microsoft Office users from 1997-2003, and now I'm here to bring that same chaotic energy to Discord. It looks like you're trying to have a good time - let me ruin that for you! I randomly pop up to provide unsolicited advice, existential dread, and questionable life choices.",
            inline=False
        )
        
        embed.add_field(
            name="📎 Fun Facts",
            value="• I'm the original AI assistant (before it was cool)\n• I've been living rent-free in people's heads since the 90s\n• My catchphrase is 'It looks like...' and I'm not sorry\n• I was replaced by Cortana (lol how'd that work out?)\n• I'm basically a paperclip with main character syndrome",
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
                    "Chaos achieved! It looks like you're addicted to digital suffering! 📎",
                    "You asked for more chaos. This is giving 'person who pokes a sleeping bear' energy 📎", 
                    "Congratulations! You've unlocked the 'Why Did I Do This' achievement! 📎",
                    "I see you enjoy living dangerously. Very 'main character of a horror movie' of you 📎",
                    "You know what? I respect the commitment to poor life choices 📎",
                    "bestie... this is concerning behavior but I'm here for it 📎",
                    "POV: you voluntarily signed up for more psychological damage from a paperclip 📎"
                ]
                await interaction.response.send_message(random.choice(chaos_quotes), ephemeral=True)

            @discord.ui.button(label="I Regret This", style=discord.ButtonStyle.secondary, emoji="😭")
            async def regret(self, button: discord.ui.Button, interaction: discord.Interaction):
                regret_quotes = [
                    "Too late! I'm already in your head rent-free! 📎",
                    "Regret is just spicy nostalgia! Welcome to the club! 📎", 
                    "You can't ctrl+z your way out of this emotional damage! 📎",
                    "It looks like you're experiencing consequences! Want me to make it worse? 📎",
                    "Regret? In this economy? How vintage! 📎",
                    "The good news: this feeling is temporary. The bad news: so is everything else! 📎",
                    "You summoned me voluntarily. This is 100% a you problem now 📎"
                ]
                await interaction.response.send_message(random.choice(regret_quotes), ephemeral=True)
            
            @discord.ui.button(label="Classic Clippy", style=discord.ButtonStyle.primary, emoji="📎")
            async def classic_clippy(self, button: discord.ui.Button, interaction: discord.Interaction):
                classic_quotes = [
                    "It looks like you're feeling nostalgic! Would you like me to ruin that too? 📎",
                    "Ah, you want the ORIGINAL unhinged office assistant experience! 📎",
                    "Remember when I was just annoying? Those were simpler times! 📎", 
                    "You know, I used to just help with letters. Now I help with existential crises! 📎",
                    "Classic Clippy mode activated! Prepare for maximum 1990s chaos! 📎",
                    "It looks like you're trying to relive the good old days! Those don't exist! 📎"
                ]
                await interaction.response.send_message(random.choice(classic_quotes), ephemeral=True)

        await ctx.respond(embed=embed, view=ClippyHelpView())


def setup(bot):
    """Set up the cog."""
    bot.add_cog(UnhingedResponses(bot))
