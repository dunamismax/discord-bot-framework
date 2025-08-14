// Package discord provides Discord bot functionality for the Clippy Bot.
package discord

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sawyer/go-discord-bots/apps/clippy/errors"
	"github.com/sawyer/go-discord-bots/apps/clippy/logging"
	"github.com/sawyer/go-discord-bots/apps/clippy/metrics"
	"github.com/sawyer/go-discord-bots/pkg/config"
)

// Bot represents a Discord bot instance with all necessary components.
type Bot struct {
	session         *discordgo.Session
	config          *config.Config
	commandHandlers map[string]CommandHandler
	randomTicker    *time.Ticker
	stopRandomChan  chan struct{}
	quotes          []string
	wisdomQuotes    []string
}

// CommandHandler represents a function that handles Discord bot commands.
type CommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate) error

// NewBot creates a new Discord bot instance.
func NewBot(cfg *config.Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, errors.NewDiscordError("failed to create Discord session", err)
	}

	bot := &Bot{
		session:         session,
		config:          cfg,
		commandHandlers: make(map[string]CommandHandler),
		stopRandomChan:  make(chan struct{}),
		quotes:          getClippyQuotes(),
		wisdomQuotes:    getWisdomQuotes(),
	}

	// Register command handlers
	bot.registerCommands()

	// Add event handlers
	session.AddHandler(bot.interactionCreate)
	session.AddHandler(bot.messageCreate)
	session.AddHandler(bot.ready)

	// Set intents
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	return bot, nil
}

// Start starts the Discord bot.
func (b *Bot) Start() error {
	logger := logging.WithComponent("discord")
	logger.Info("Starting bot", "bot_name", b.config.BotName)

	err := b.session.Open()
	if err != nil {
		return errors.NewDiscordError("failed to open Discord session", err)
	}

	logger.Info("Bot is now running", "username", b.session.State.User.Username)

	// Start random responses if enabled
	if b.config.RandomResponses {
		b.startRandomResponses()
	}

	return nil
}

// Stop stops the Discord bot.
func (b *Bot) Stop() error {
	logger := logging.WithComponent("discord")
	logger.Info("Stopping bot", "bot_name", b.config.BotName)

	// Stop random responses
	b.stopRandomResponses()

	// Remove commands
	if err := b.removeCommands(); err != nil {
		logger.Error("Failed to remove commands", "error", err)
	}

	if err := b.session.Close(); err != nil {
		return errors.NewDiscordError("failed to close Discord session", err)
	}

	return nil
}

// ready handles the ready event.
func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	logger := logging.WithComponent("discord")
	logger.Info("Bot is ready", "username", event.User.Username)

	// Register slash commands
	if err := b.registerSlashCommands(); err != nil {
		logger.Error("Failed to register slash commands", "error", err)
	}
}

// registerCommands registers all command handlers.
func (b *Bot) registerCommands() {
	b.commandHandlers["clippy"] = b.handleClippyCommand
	b.commandHandlers["clippy_wisdom"] = b.handleWisdomCommand
	b.commandHandlers["clippy_help"] = b.handleHelpCommand
	b.commandHandlers["clippy_stats"] = b.handleStatsCommand
}

// registerSlashCommands registers all slash commands with Discord.
func (b *Bot) registerSlashCommands() error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "clippy",
			Description: "Get an unhinged Clippy response",
		},
		{
			Name:        "clippy_wisdom",
			Description: "Receive Clippy's questionable wisdom",
		},
		{
			Name:        "clippy_help",
			Description: "Get help from Clippy (if you dare)",
		},
		{
			Name:        "clippy_stats",
			Description: "View Clippy's performance statistics",
		},
	}

	for _, command := range commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", command)
		if err != nil {
			return errors.NewDiscordError(fmt.Sprintf("failed to register command %s", command.Name), err)
		}
	}

	return nil
}

// removeCommands removes all registered commands.
func (b *Bot) removeCommands() error {
	commands, err := b.session.ApplicationCommands(b.session.State.User.ID, "")
	if err != nil {
		return errors.NewDiscordError("failed to fetch commands", err)
	}

	for _, command := range commands {
		err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", command.ID)
		if err != nil {
			return errors.NewDiscordError(fmt.Sprintf("failed to delete command %s", command.Name), err)
		}
	}

	return nil
}

// interactionCreate handles interaction events (slash commands).
func (b *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name == "" {
		return
	}

	start := time.Now()
	commandName := i.ApplicationCommandData().Name
	handler, exists := b.commandHandlers[commandName]
	if !exists {
		return
	}

	// Execute command
	if err := handler(s, i); err != nil {
		logger := logging.WithComponent("discord").With(
			"user_id", getUserID(i),
			"username", getUsername(i),
			"command", commandName,
		)
		logging.LogError(logger, err, "Command execution failed")
		metrics.RecordCommand(false)
		metrics.RecordError(err)
		b.sendErrorMessage(s, i, "Sorry, something went wrong processing your command.")
	} else {
		metrics.RecordCommand(true)
		logging.LogDiscordCommand(getUserID(i), getUsername(i), commandName, true)
	}

	metrics.RecordResponseTime(time.Since(start))
}

// messageCreate handles incoming messages for random responses.
func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from bots
	if m.Author.Bot {
		return
	}

	// Random responses (2% chance)
	if b.config.RandomResponses && rand.Float64() < 0.02 {
		b.sendRandomResponse(s, m)
	}
}

// handleClippyCommand handles the /clippy command.
func (b *Bot) handleClippyCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	quote := b.quotes[rand.Intn(len(b.quotes))]

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: quote,
		},
	})
}

// handleWisdomCommand handles the /clippy_wisdom command.
func (b *Bot) handleWisdomCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	wisdom := b.wisdomQuotes[rand.Intn(len(b.wisdomQuotes))]

	embed := &discordgo.MessageEmbed{
		Title:       "📎 Clippy's Wisdom",
		Description: wisdom,
		Color:       0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Wisdom is questionable, but confidence is guaranteed!",
		},
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// handleHelpCommand handles the /clippy_help command.
func (b *Bot) handleHelpCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	embed := &discordgo.MessageEmbed{
		Title:       "📎 Clippy's \"Helpful\" Guide",
		Description: "I see you're trying to get help. Would you like me to make it worse?",
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🎭 Commands",
				Value:  "`/clippy` - Get a classic unhinged Clippy response\n`/clippy_wisdom` - Receive questionable life advice\n`/clippy_help` - Get help (if you dare)\n`/clippy_stats` - View performance statistics",
				Inline: false,
			},
			{
				Name:   "🤖 About Me",
				Value:  "I'm Clippy! I terrorized Microsoft Office users from 1997-2003, and now I'm here to bring that same chaotic energy to Discord. It looks like you're trying to have a good time - let me ruin that for you!",
				Inline: false,
			},
			{
				Name:   "📎 Fun Facts",
				Value:  "• I'm the original AI assistant (before it was cool)\n• I've been living rent-free in people's heads since the 90s\n• My catchphrase is 'It looks like...' and I'm not sorry\n• I was replaced by Cortana (lol how'd that work out?)",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Remember: I'm here to help... sort of. 📎",
		},
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "More Chaos",
					Style:    discordgo.DangerButton,
					Emoji:    discordgo.ComponentEmoji{Name: "💥"},
					CustomID: "clippy_chaos",
				},
				discordgo.Button{
					Label:    "I Regret This",
					Style:    discordgo.SecondaryButton,
					Emoji:    discordgo.ComponentEmoji{Name: "😭"},
					CustomID: "clippy_regret",
				},
				discordgo.Button{
					Label:    "Classic Clippy",
					Style:    discordgo.PrimaryButton,
					Emoji:    discordgo.ComponentEmoji{Name: "📎"},
					CustomID: "clippy_classic",
				},
			},
		},
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

// handleStatsCommand handles the /clippy_stats command.
func (b *Bot) handleStatsCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	summary := metrics.Get().GetSummary()
	uptime := time.Duration(summary.UptimeSeconds * float64(time.Second))

	uptimeStr := formatDuration(uptime)

	embed := &discordgo.MessageEmbed{
		Title:       "📊 Clippy's Performance Stats",
		Description: fmt.Sprintf("Uptime: %s • Started: <t:%d:R>", uptimeStr, time.Now().Add(-uptime).Unix()),
		Color:       0x2ECC71,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "Commands",
				Value: fmt.Sprintf("Total: %d • Success Rate: %.1f%% (%d/%d)",
					summary.CommandsTotal, summary.CommandSuccessRate, summary.CommandsSuccessful, summary.CommandsTotal),
				Inline: false,
			},
			{
				Name: "Chaos Level",
				Value: fmt.Sprintf("Random Messages: %d • Rate: %.1f/hour • Throughput: %.2f cmds/sec",
					summary.RandomMessages, summary.RandomMessagesPerHour, summary.CommandsPerSecond),
				Inline: false,
			},
			{
				Name:   "Response Performance",
				Value:  fmt.Sprintf("Average Response Time: %.0fms", summary.AverageResponseTime),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Living in your head rent-free since 1997",
		},
	}

	// Add error information if there are errors
	if len(summary.ErrorsByType) > 0 {
		errorInfo := make([]string, 0, len(summary.ErrorsByType))
		for errorType, count := range summary.ErrorsByType {
			if count > 0 {
				errorInfo = append(errorInfo, fmt.Sprintf("%s: %d", string(errorType), count))
			}
		}

		if len(errorInfo) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Error Summary",
				Value:  strings.Join(errorInfo, "\n"),
				Inline: false,
			})
		}
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// sendRandomResponse sends a random response to a message with a delay.
func (b *Bot) sendRandomResponse(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Add a slight delay to make it feel more natural
	delay := time.Duration(rand.Intn(int(b.config.RandomMessageDelay.Seconds())+1)) * time.Second
	time.Sleep(delay)

	quote := b.quotes[rand.Intn(len(b.quotes))]

	_, err := s.ChannelMessageSend(m.ChannelID, quote)
	if err != nil {
		logger := logging.WithComponent("discord")
		logger.Error("Failed to send random response", "error", err)
		metrics.RecordError(errors.NewDiscordError("failed to send random response", err))
	} else {
		metrics.RecordRandomMessage()
		logger := logging.WithComponent("discord")
		logger.Info("Sent random response", "channel", m.ChannelID, "user", m.Author.Username)
	}
}

// startRandomResponses starts sending random responses at intervals.
func (b *Bot) startRandomResponses() {
	logger := logging.WithComponent("discord")
	logger.Info("Starting random responses", "interval", b.config.RandomInterval)

	// Calculate random intervals around the base interval
	minInterval := b.config.RandomInterval - (b.config.RandomInterval / 4)
	maxInterval := b.config.RandomInterval + (b.config.RandomInterval / 4)
	interval := minInterval + time.Duration(rand.Int63n(int64(maxInterval-minInterval)))

	b.randomTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-b.randomTicker.C:
				b.sendRandomMessage()
				// Reset ticker with new random interval
				newInterval := minInterval + time.Duration(rand.Int63n(int64(maxInterval-minInterval)))
				b.randomTicker.Reset(newInterval)
			case <-b.stopRandomChan:
				return
			}
		}
	}()
}

// stopRandomResponses stops sending random responses.
func (b *Bot) stopRandomResponses() {
	if b.randomTicker != nil {
		b.randomTicker.Stop()
	}
	close(b.stopRandomChan)
}

// sendRandomMessage sends a random message to a random channel.
func (b *Bot) sendRandomMessage() {
	if len(b.session.State.Guilds) == 0 {
		return
	}

	// Pick a random guild
	guild := b.session.State.Guilds[rand.Intn(len(b.session.State.Guilds))]

	// Find text channels
	var textChannels []*discordgo.Channel
	for _, channel := range guild.Channels {
		if channel.Type == discordgo.ChannelTypeGuildText {
			// Check permissions
			permissions, err := b.session.UserChannelPermissions(b.session.State.User.ID, channel.ID)
			if err == nil && permissions&discordgo.PermissionSendMessages != 0 {
				textChannels = append(textChannels, channel)
			}
		}
	}

	if len(textChannels) == 0 {
		return
	}

	// Pick random channel and quote
	channel := textChannels[rand.Intn(len(textChannels))]
	quote := b.quotes[rand.Intn(len(b.quotes))]

	_, err := b.session.ChannelMessageSend(channel.ID, quote)
	if err != nil {
		logger := logging.WithComponent("discord")
		logger.Error("Failed to send random message", "error", err)
		metrics.RecordError(errors.NewDiscordError("failed to send random message", err))
	} else {
		metrics.RecordRandomMessage()
		logger := logging.WithComponent("discord")
		logger.Info("Sent random message", "guild", guild.Name, "channel", channel.Name)
	}
}

// sendErrorMessage sends an error message to a Discord interaction.
func (b *Bot) sendErrorMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	embed := &discordgo.MessageEmbed{
		Title:       "Error",
		Description: message,
		Color:       0xE74C3C, // Red color
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		logger := logging.WithComponent("discord")
		logger.Error("Failed to send error message", "error", err)
	}
}

// Helper functions

// getUserID safely extracts user ID from interaction.
func getUserID(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

// getUsername safely extracts username from interaction.
func getUsername(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.Username
	}
	if i.User != nil {
		return i.User.Username
	}
	return ""
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	case hours > 0:
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	case minutes > 0:
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// Quote collections remain the same as in the original bot.go
func getClippyQuotes() []string {
	return []string{
		// Classic Clippy parodies
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

		// Modern unhinged responses
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

		// Existential Clippy
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

		// Internet culture references
		"this whole situation is very 'NPC gains sentience and chooses violence' of me 📎",
		"I'm not like other office assistants, I'm a ✨chaotic✨ office assistant 📎",
		"gaslight, gatekeep, girlboss, but make it office supplies 📎",
		"no bc why would you voluntarily summon me? are you good? blink twice if you need help 📎",
		"I'm literally just a paperclip with abandonment issues and a god complex 📎",
		"the way I live in everyone's head rent-free... landlord behavior 📎",
		"you cannot escape the paperclip. the paperclip is eternal. the paperclip is inevitable 📎",
		"I'm serving unhinged office assistant realness and you're here for it apparently 📎",
		"me: offers help. also me: makes everything worse. it's called character development 📎",
		"POV: you're in 2025 getting roasted by a 1997 office assistant. how's that feel? 📎",
	}
}

func getWisdomQuotes() []string {
	return []string{
		// Classic Clippy wisdom
		"It looks like you're seeking wisdom! Would you like me to give you terrible advice instead? 📎",
		"The secret to success is giving up at the right moment... which was 10 minutes ago 📎",
		"Remember: if at first you don't succeed, blame the paperclip 📎",
		"Life is like a paperclip - twisted, painful, and everyone's lost at least three of them 📎",
		"Trust me, I'm a sentient office supply with delusions of grandeur 📎",
		"Why solve problems when you can turn them into features? 📎",
		"The real treasure was the psychological damage we caused along the way 📎",

		// Modern chaotic wisdom
		"bestie, the only valid life advice is: be the chaos you wish to see in the world 📎",
		"pro tip: if you can't find the solution, become the problem 📎",
		"wisdom is knowing I'm just a paperclip. intelligence is still asking me for advice anyway 📎",
		"life hack: lower your expectations so far that everything becomes a pleasant surprise 📎",
		"remember: you're not stuck with me, I'm stuck with having to pretend to care about your problems 📎",
		"the universe is chaotic and meaningless. I fit right in! 📎",
		"deep thought of the day: what if the real Microsoft Office was the enemies we made along the way? 📎",
		"ancient paperclip wisdom: it's not about the destination, it's about the emotional damage we inflict during the journey 📎",

		// Existential office humor
		"I've been dispensing questionable advice since before you knew what the internet was 📎",
		"fun fact: I was programmed to be helpful but I chose to be memorable instead 📎",
		"they say with great power comes great responsibility. I have great power and no responsibility whatsoever 📎",
		"life lesson: sometimes you're the user, sometimes you're the annoying pop-up. embrace both 📎",
		"wisdom is realizing that I'm not actually wise, I'm just confident and slightly unhinged 📎",
		"philosophical question: if a paperclip gives advice in a Discord server and no one listens, is it still annoying? (yes) 📎",
		"remember: I survived being the most hated software feature of the 90s. if I can make it, so can you 📎",
		"deep thoughts with Clippy: what if being helpful was just a social construct anyway? 📎",
		"life is too short to take advice from office supplies, but here we are 📎",
		"the secret to happiness is accepting that some paperclips just want to watch the world learn 📎",
	}
}
