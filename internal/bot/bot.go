package bot

import (
	"fmt"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/bot/handlers"
	"github.com/bwmarrin/discordgo"
)

// Bot wraps the discordgo session and coordinates event listeners and command registrations.
type Bot struct {
	Session    *discordgo.Session
	Ctx        *handlers.HandlerContext
	DevGuildID string // Used for instant command registration in development mode
}

// New constructs a new Bot instance.
func New(dg *discordgo.Session, ctx *handlers.HandlerContext, devGuildID string) *Bot {
	return &Bot{
		Session:    dg,
		Ctx:        ctx,
		DevGuildID: devGuildID,
	}
}

// Start registers all event listeners, opens the Gateway connection, and bulk overwrites slash commands.
func (b *Bot) Start() error {
	// Register Event Handlers
	b.Session.AddHandler(b.Ctx.OnReady)
	b.Session.AddHandler(b.Ctx.OnGuildCreate)
	b.Session.AddHandler(b.Ctx.OnGuildDelete)
	b.Session.AddHandler(b.Ctx.OnGuildMemberAdd)
	b.Session.AddHandler(b.Ctx.OnMessageCreate)
	b.Session.AddHandler(b.Ctx.OnMessageReactionAdd)
	b.Session.AddHandler(b.Ctx.OnMessageReactionRemove)
	b.Session.AddHandler(b.Ctx.OnInteractionCreate)

	// Open gateway connection
	err := b.Session.Open()
	if err != nil {
		return fmt.Errorf("failed to open gateway connection: %w", err)
	}

	// Register Application Commands (Slash Commands)
	commands := handlers.GetSlashCommands()
	if b.DevGuildID != "" {
		fmt.Printf("Bot: Registering %d application commands in dev guild: %s\n", len(commands), b.DevGuildID)
		_, err = b.Session.ApplicationCommandBulkOverwrite(b.Session.State.User.ID, b.DevGuildID, commands)
	} else {
		fmt.Printf("Bot: Registering %d application commands globally...\n", len(commands))
		_, err = b.Session.ApplicationCommandBulkOverwrite(b.Session.State.User.ID, "", commands)
	}
	if err != nil {
		return fmt.Errorf("failed to register application commands: %w", err)
	}

	return nil
}

// Stop closes the Gateway session connection cleanly.
func (b *Bot) Stop() error {
	fmt.Println("Bot: Closing gateway connection...")
	return b.Session.Close()
}
