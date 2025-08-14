// Package clippy provides the Clippy Discord bot implementation.
package clippy

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sawyer/discord-bot-framework/internal/config"
	"github.com/sawyer/discord-bot-framework/internal/errors"
	"github.com/sawyer/discord-bot-framework/internal/framework"
	"github.com/sawyer/discord-bot-framework/internal/logging"
)

// Bot represents the Clippy Discord bot.
type Bot struct {
	*framework.Bot
	logger         *slog.Logger
	quotes         []string
	wisdomQuotes   []string
	randomTicker   *time.Ticker
	stopRandomChan chan struct{}
}

// NewBot creates a new Clippy bot instance.
func NewBot(cfg *config.BotConfig) (*Bot, error) {
	frameworkBot, err := framework.NewBot(cfg)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		Bot:            frameworkBot,
		logger:         logging.WithComponent("clippy"),
		quotes:         getClippyQuotes(),
		wisdomQuotes:   getWisdomQuotes(),
		stopRandomChan: make(chan struct{}),
	}

	// Register commands
	bot.registerCommands()

	// Register message handler for random responses
	bot.RegisterMessageHandler(bot.handleRandomResponses)

	return bot, nil
}

// Start starts the Clippy bot.
func (b *Bot) Start() error {
	if err := b.Bot.Start(); err != nil {
		return err
	}

	// Start random responses if enabled
	if b.GetConfig().RandomResponses {
		b.startRandomResponses()
	}

	return nil
}

// Stop stops the Clippy bot.
func (b *Bot) Stop(ctx context.Context) error {
	// Stop random responses
	b.stopRandomResponses()

	return b.Bot.Stop(ctx)
}

// registerCommands registers all Clippy commands.
func (b *Bot) registerCommands() {
	// /clippy command
	clippyCommand := &discordgo.ApplicationCommand{
		Name:        "clippy",
		Description: "Get an unhinged Clippy response",
	}
	b.RegisterCommand(clippyCommand, b.handleClippyCommand)

	// /clippy_wisdom command
	wisdomCommand := &discordgo.ApplicationCommand{
		Name:        "clippy_wisdom",
		Description: "Receive Clippy's questionable wisdom",
	}
	b.RegisterCommand(wisdomCommand, b.handleWisdomCommand)

	// /clippy_help command
	helpCommand := &discordgo.ApplicationCommand{
		Name:        "clippy_help",
		Description: "Get help from Clippy (if you dare)",
	}
	b.RegisterCommand(helpCommand, b.handleHelpCommand)
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
				Value:  "`/clippy` - Get a classic unhinged Clippy response\n`/clippy_wisdom` - Receive questionable life advice\n`/clippy_help` - Get help (if you dare)",
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

// handleRandomResponses handles random responses to messages.
func (b *Bot) handleRandomResponses(s *discordgo.Session, m *discordgo.MessageCreate) error {
	// 2% chance to respond to any message
	if rand.Float64() > 0.02 {
		return nil
	}

	// Add a slight delay to make it feel more natural
	time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)

	quote := b.quotes[rand.Intn(len(b.quotes))]

	_, err := s.ChannelMessageSend(m.ChannelID, quote)
	if err != nil {
		return errors.NewDiscordError("failed to send random response", err)
	}

	b.logger.Info("Sent random response", "channel", m.ChannelID, "user", m.Author.Username)
	return nil
}

// startRandomResponses starts sending random responses at intervals.
func (b *Bot) startRandomResponses() {
	b.logger.Info("Starting random responses")

	// Send random messages every 30-90 minutes
	interval := time.Duration(rand.Intn(60)+30) * time.Minute
	b.randomTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-b.randomTicker.C:
				b.sendRandomMessage()
				// Reset ticker with new random interval
				b.randomTicker.Reset(time.Duration(rand.Intn(60)+30) * time.Minute)
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
	session := b.GetSession()

	if len(session.State.Guilds) == 0 {
		return
	}

	// Pick a random guild
	guild := session.State.Guilds[rand.Intn(len(session.State.Guilds))]

	// Find text channels
	var textChannels []*discordgo.Channel
	for _, channel := range guild.Channels {
		if channel.Type == discordgo.ChannelTypeGuildText {
			// Check permissions
			permissions, err := session.UserChannelPermissions(session.State.User.ID, channel.ID)
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

	_, err := session.ChannelMessageSend(channel.ID, quote)
	if err != nil {
		b.logger.Error("Failed to send random message", "error", err)
	} else {
		b.logger.Info("Sent random message", "guild", guild.Name, "channel", channel.Name)
	}
}

// getClippyQuotes returns all the unhinged Clippy quotes.
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

// getWisdomQuotes returns all the wisdom quotes.
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
