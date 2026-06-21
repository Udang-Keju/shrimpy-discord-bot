package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/service"
	"github.com/bwmarrin/discordgo"
)

// HandlerContext houses references to all internal services. All event, slash command,
// and component handlers are methods on this context to enable dependency injection.
type HandlerContext struct {
	GuildSvc        *service.GuildService
	WelcomeSvc      *service.WelcomeService
	AutoRoleSvc     *service.AutoRoleService
	ReactionRoleSvc *service.ReactionRoleService
	TicketSvc       *service.TicketService
}

// NewHandlerContext creates a new HandlerContext.
func NewHandlerContext(
	guildSvc *service.GuildService,
	welcomeSvc *service.WelcomeService,
	autoRoleSvc *service.AutoRoleService,
	reactionRoleSvc *service.ReactionRoleService,
	ticketSvc *service.TicketService,
) *HandlerContext {
	return &HandlerContext{
		GuildSvc:        guildSvc,
		WelcomeSvc:      welcomeSvc,
		AutoRoleSvc:     autoRoleSvc,
		ReactionRoleSvc: reactionRoleSvc,
		TicketSvc:       ticketSvc,
	}
}

// OnReady logs when the bot is successfully connected to the Gateway.
func (ctx *HandlerContext) OnReady(s *discordgo.Session, r *discordgo.Ready) {
	fmt.Printf("Bot: Logged in as %s#%s (%s)\n", r.User.Username, r.User.Discriminator, r.User.ID)
}

// OnGuildCreate registers the guild in the DB on invite or reconnect.
func (ctx *HandlerContext) OnGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	guildID, err := strconv.ParseInt(g.ID, 10, 64)
	if err != nil {
		return
	}

	// Auto-register guild config in DB (upsert) to ensure defaults are populated
	_, err = ctx.GuildSvc.GetConfig(context.Background(), guildID)
	if err != nil {
		fmt.Printf("Bot Error: failed to register guild %d: %v\n", guildID, err)
	} else {
		fmt.Printf("Bot: Guild registered/loaded: %s (%d)\n", g.Name, guildID)
	}
}

// OnGuildDelete handles when the bot is kicked from a server.
func (ctx *HandlerContext) OnGuildDelete(s *discordgo.Session, g *discordgo.GuildDelete) {
	guildID, err := strconv.ParseInt(g.ID, 10, 64)
	if err != nil {
		return
	}

	err = ctx.GuildSvc.Deactivate(context.Background(), guildID)
	if err != nil {
		fmt.Printf("Bot Error: failed to deactivate guild %d: %v\n", guildID, err)
	} else {
		fmt.Printf("Bot: Deactivated guild config: %d\n", guildID)
	}
}

// OnGuildMemberAdd handles when a new user joins the server (welcome message + auto role).
func (ctx *HandlerContext) OnGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	guildID, err := strconv.ParseInt(m.GuildID, 10, 64)
	if err != nil {
		return
	}

	userID, err := strconv.ParseInt(m.User.ID, 10, 64)
	if err != nil {
		return
	}

	go func() {
		// 1. Assign auto-roles
		err = ctx.AutoRoleSvc.AssignRoles(context.Background(), s, guildID, userID)
		if err != nil {
			fmt.Printf("Bot Error: failed to assign auto roles to %s: %v\n", m.User.Username, err)
		}

		// 2. Trigger welcome flow
		err = ctx.WelcomeSvc.SendWelcome(context.Background(), s, guildID, m.Member)
		if err != nil {
			fmt.Printf("Bot Error: failed to run welcome flow for %s: %v\n", m.User.Username, err)
		}
	}()
}

// OnMessageCreate logs messages in active ticket channels to build transcripts, and routes prefix commands.
func (ctx *HandlerContext) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore self messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	guildID, err := strconv.ParseInt(m.GuildID, 10, 64)
	if err != nil {
		return
	}

	// 1. Check if the message is in a ticket channel to log it for transcripts
	channelID, err := strconv.ParseInt(m.ChannelID, 10, 64)
	if err == nil {
		// Check if we can find this ticket in DB by channel ID
		// In Go, since repo is in Service Layer, we check if TicketSvc can find it
		// We'll add a helper or directly access repository if needed.
		// For simplicity, we can try to look it up using the TicketSvc or a lightweight checker.
		// Since we want to log messages in ticket channels, let's fetch the ticket.
		// Note: to avoid DB lookup on every single chat message, in production we'd cache this,
		// but since it's a simple bot, direct lookup or light caching works.
		// We can add a lightweight endpoint on service to find active tickets by channel.
		// Wait, did we define GetByChannelID on TicketRepo? Yes!
		// Let's implement message logging in a goroutine.
		go func() {
			ticket, err := ctx.TicketSvc.GetByChannelID(context.Background(), channelID)
			if err == nil && ticket != nil && ticket.Status != repository.TicketStatusArchived {
				authorID, _ := strconv.ParseInt(m.Author.ID, 10, 64)
				content := m.Content

				// Parse attachments
				var attachments []repository.Attachment
				for _, att := range m.Attachments {
					attachments = append(attachments, repository.Attachment{
						Filename: att.Filename,
						URL:      att.URL,
						Size:     att.Size,
					})
				}

				// Check if the message is a staff note (e.g. starts with custom prefix or in staff command)
				isStaffNote := false
				if strings.HasPrefix(m.Content, "*") && strings.HasSuffix(m.Content, "*") {
					// Markdown italics note style, or prefix check
					isStaffNote = true
				}

				_ = ctx.TicketSvc.LogMessage(context.Background(), ticket.ID, authorID, m.Author.Username, &content, isStaffNote, attachments)
			}
		}()
	}

	// 2. Handle prefix commands if message starts with guild prefix
	go func() {
		cfg, err := ctx.GuildSvc.GetConfig(context.Background(), guildID)
		if err != nil {
			return
		}

		if strings.HasPrefix(m.Content, cfg.Prefix) {
			ctx.HandlePrefixCommand(s, m, cfg.Prefix)
		}
	}()
}

// OnMessageReactionAdd handles roles being granted on reaction click.
func (ctx *HandlerContext) OnMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Ignore self reactions
	if r.UserID == s.State.User.ID {
		return
	}

	discordMsgID, err := strconv.ParseInt(r.MessageID, 10, 64)
	if err != nil {
		return
	}

	guildID, err := strconv.ParseInt(r.GuildID, 10, 64)
	if err != nil {
		return
	}

	userID, err := strconv.ParseInt(r.UserID, 10, 64)
	if err != nil {
		return
	}

	// Emoji identifier
	emojiStr := r.Emoji.Name
	if r.Emoji.ID != "" {
		// Custom emoji
		emojiStr = r.Emoji.APIName()
	}

	go func() {
		err := ctx.ReactionRoleSvc.HandleReactionAdd(context.Background(), s, discordMsgID, guildID, userID, emojiStr)
		if err != nil {
			fmt.Printf("Bot Error: failed to handle reaction role grant: %v\n", err)
		}
	}()
}

// OnMessageReactionRemove handles roles being revoked when reaction is removed.
func (ctx *HandlerContext) OnMessageReactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	// Ignore self reactions
	if r.UserID == s.State.User.ID {
		return
	}

	discordMsgID, err := strconv.ParseInt(r.MessageID, 10, 64)
	if err != nil {
		return
	}

	guildID, err := strconv.ParseInt(r.GuildID, 10, 64)
	if err != nil {
		return
	}

	userID, err := strconv.ParseInt(r.UserID, 10, 64)
	if err != nil {
		return
	}

	emojiStr := r.Emoji.Name
	if r.Emoji.ID != "" {
		emojiStr = r.Emoji.APIName()
	}

	go func() {
		err := ctx.ReactionRoleSvc.HandleReactionRemove(context.Background(), s, discordMsgID, guildID, userID, emojiStr)
		if err != nil {
			fmt.Printf("Bot Error: failed to handle reaction role revoke: %v\n", err)
		}
	}()
}
