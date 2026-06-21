package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app"
	guild_bot "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/bot"
	guild_svc "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/service"
	rr_bot "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/bot"
	rr_svc "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/service"
	ticket_bot "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/bot"
	ticket_svc "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/service"
	welcome_bot "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/bot"
	welcome_svc "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/service"
	"github.com/bwmarrin/discordgo"
)

// HandlerContext houses references to the new feature handlers and services.
type HandlerContext struct {
	GuildSvc        *guild_svc.GuildService
	AutoRoleSvc     *guild_svc.AutoRoleService
	WelcomeSvc      *welcome_svc.WelcomeService
	ReactionRoleSvc *rr_svc.ReactionRoleService
	TicketSvc       *ticket_svc.TicketService

	GuildBot        *guild_bot.BotHandler
	WelcomeBot      *welcome_bot.BotHandler
	ReactionRoleBot *rr_bot.BotHandler
	TicketBot       *ticket_bot.BotHandler
}

// NewHandlerContext creates a new HandlerContext.
func NewHandlerContext(modules *app.Modules) *HandlerContext {
	return &HandlerContext{
		GuildSvc:        modules.Guild.Service,
		AutoRoleSvc:     modules.Guild.AutoRoleSvc,
		WelcomeSvc:      modules.Welcome.Service,
		ReactionRoleSvc: modules.ReactionRole.Service,
		TicketSvc:       modules.Ticket.TicketSvc,
		GuildBot:        modules.Guild.Bot,
		WelcomeBot:      modules.Welcome.Bot,
		ReactionRoleBot: modules.ReactionRole.Bot,
		TicketBot:       modules.Ticket.Bot,
	}
}



// OnReady logs when the bot is successfully connected to the Gateway.
func (ctx *HandlerContext) OnReady(s *discordgo.Session, r *discordgo.Ready) {
	fmt.Printf("Bot: Logged in as %s#%s (%s)\n", r.User.Username, r.User.Discriminator, r.User.ID)
}

// OnGuildCreate registers the guild in the DB on invite or reconnect.
func (ctx *HandlerContext) OnGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	ctx.GuildBot.OnGuildCreate(s, g)
}

// OnGuildDelete handles when the bot is kicked from a server.
func (ctx *HandlerContext) OnGuildDelete(s *discordgo.Session, g *discordgo.GuildDelete) {
	ctx.GuildBot.OnGuildDelete(s, g)
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
	if m.Author.ID == s.State.User.ID {
		return
	}

	guildID, err := strconv.ParseInt(m.GuildID, 10, 64)
	if err != nil {
		return
	}

	// 1. Log messages for tickets
	ctx.TicketBot.OnMessageCreate(s, m)

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
	ctx.ReactionRoleBot.OnMessageReactionAdd(s, r)
}

// OnMessageReactionRemove handles roles being revoked when reaction is removed.
func (ctx *HandlerContext) OnMessageReactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	ctx.ReactionRoleBot.OnMessageReactionRemove(s, r)
}
