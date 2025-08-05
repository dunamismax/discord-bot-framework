"""Help system for Discord bots."""


import discord
from discord.ext import commands


class HelpSystem(commands.Cog):
    """Universal help system for Discord bots."""

    def __init__(self, bot):
        self.bot = bot
        self.bot_name = bot.__class__.__name__
        self.color = 0x00ff00

    def get_command_info(self) -> dict[str, list[dict[str, str]]]:
        """Get organized command information."""
        categories = {}

        for command in self.bot.walk_application_commands():
            if isinstance(command, discord.SlashCommandGroup):
                # Handle command groups
                category = command.name.title()
                if category not in categories:
                    categories[category] = []

                for subcommand in command.subcommands:
                    categories[category].append({
                        'name': f"{command.name} {subcommand.name}",
                        'description': subcommand.description or "No description available"
                    })
            else:
                # Handle regular slash commands
                category = getattr(command.cog, 'qualified_name', 'General') if command.cog else 'General'
                if category not in categories:
                    categories[category] = []

                categories[category].append({
                    'name': command.name,
                    'description': command.description or "No description available"
                })

        return categories

    @commands.slash_command(name="help", description="Show bot commands and information")
    async def help_command(self, ctx, category: str = None):
        """Display help information."""
        command_info = self.get_command_info()

        if category:
            # Show specific category
            category_title = category.title()
            if category_title not in command_info:
                available_categories = ", ".join(command_info.keys())
                await ctx.respond(
                    f"❌ Category '{category}' not found. Available categories: {available_categories}",
                    ephemeral=True
                )
                return

            embed = discord.Embed(
                title=f"📚 {self.bot_name} - {category_title} Commands",
                color=self.color
            )

            for command in command_info[category_title]:
                embed.add_field(
                    name=f"/{command['name']}",
                    value=command['description'],
                    inline=False
                )
        else:
            # Show overview of all categories
            embed = discord.Embed(
                title=f"📚 {self.bot_name} Help",
                description="Here are all available command categories:",
                color=self.color
            )

            for cat_name, commands in command_info.items():
                command_list = [f"/{cmd['name']}" for cmd in commands[:5]]  # Show first 5 commands
                if len(commands) > 5:
                    command_list.append(f"... and {len(commands) - 5} more")

                embed.add_field(
                    name=f"{cat_name} ({len(commands)} commands)",
                    value="\n".join(command_list),
                    inline=True
                )

            embed.add_field(
                name="📖 Get Detailed Help",
                value="Use `/help <category>` to see detailed information about a specific category.",
                inline=False
            )

        # Add footer with bot info
        embed.set_footer(
            text=f"{self.bot_name} | {len(self.bot.guilds)} servers | Ping: {round(self.bot.latency * 1000)}ms"
        )

        await ctx.respond(embed=embed)

    @commands.slash_command(name="info", description="Show bot information and statistics")
    async def info_command(self, ctx):
        """Display bot information and statistics."""
        embed = discord.Embed(
            title=f"ℹ️ {self.bot_name} Information",
            color=self.color
        )

        # Basic bot info
        embed.add_field(
            name="🤖 Bot Info",
            value=f"**User:** {self.bot.user.mention}\n"
                  f"**ID:** {self.bot.user.id}\n"
                  f"**Created:** <t:{int(self.bot.user.created_at.timestamp())}:D>",
            inline=True
        )

        # Statistics
        total_members = sum(guild.member_count for guild in self.bot.guilds)
        embed.add_field(
            name="📊 Statistics",
            value=f"**Guilds:** {len(self.bot.guilds)}\n"
                  f"**Members:** {total_members:,}\n"
                  f"**Ping:** {round(self.bot.latency * 1000)}ms",
            inline=True
        )

        # Technical info
        embed.add_field(
            name="⚙️ Technical",
            value=f"**Library:** discord.py\n"
                  f"**Python:** 3.8+\n"
                  f"**Uptime:** <t:{int(self.bot.start_time.timestamp()) if hasattr(self.bot, 'start_time') else 0}:R>",
            inline=True
        )

        # Add command stats if database is available
        if hasattr(self.bot, 'db') and self.bot.db:
            try:
                stats = await self.bot.db.get_command_stats(7)  # Last 7 days
                if stats:
                    top_commands = stats[:5]  # Top 5 commands
                    command_text = "\n".join([f"/{cmd['command']}: {cmd['count']}" for cmd in top_commands])
                    embed.add_field(
                        name="📈 Top Commands (7 days)",
                        value=command_text,
                        inline=False
                    )
            except Exception as e:
                self.bot.logger.error(f"Error getting command stats: {e}")

        embed.set_thumbnail(url=self.bot.user.display_avatar.url)
        embed.set_footer(text="Thank you for using our bot! 💙")

        await ctx.respond(embed=embed)


def setup(bot):
    """Set up the help system cog."""
    bot.add_cog(HelpSystem(bot))
