package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/bwmarrin/discordgo"
)

// GetSlashCommands returns the list of ApplicationCommand definitions to register with Discord.
func GetSlashCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "botinfo",
			Description: "Displays general info and statistics about Shrimpy bot",
		},
		{
			Name:        "setup",
			Description: "Configure Shrimpy features",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "welcome",
					Description: "Configure onboarding and welcome greetings",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "channel",
							Description: "The channel to post welcome messages to",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "message",
							Description: "The greeting template. Supports {user}, {server}, {membercount}",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "dm-message",
							Description: "Optional DM greeting template sent directly to users",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "enabled",
							Description: "Turn the welcome feature on or off",
							Required:    false,
						},
					},
				},
			},
		},
		{
			Name:        "staff",
			Description: "Manage support staff roles",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "add",
					Description: "Add a Discord role to the support staff list",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionRole,
							Name:        "role",
							Description: "The role to add",
							Required:    true,
						},
					},
				},
				{
					Name:        "remove",
					Description: "Remove a Discord role from the support staff list",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionRole,
							Name:        "role",
							Description: "The role to remove",
							Required:    true,
						},
					},
				},
				{
					Name:        "list",
					Description: "List all configured support staff roles",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
		{
			Name:        "ticket",
			Description: "Manage support tickets within ticket channels",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "claim",
					Description: "Claim the current ticket",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "unclaim",
					Description: "Release claim on the current ticket",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "close",
					Description: "Close the current ticket",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "reason",
							Description: "Why the ticket is being closed",
							Required:    false,
						},
					},
				},
				{
					Name:        "priority",
					Description: "Update ticket priority",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "level",
							Description: "Priority level (low, medium, high, urgent)",
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "Low", Value: "low"},
								{Name: "Medium", Value: "medium"},
								{Name: "High", Value: "high"},
								{Name: "Urgent", Value: "urgent"},
							},
						},
					},
				},
				{
					Name:        "add",
					Description: "Add a member to the current ticket",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to add",
							Required:    true,
						},
					},
				},
				{
					Name:        "remove",
					Description: "Remove a member from the current ticket",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to remove",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "reactionrole",
			Description: "Manage reaction roles",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "create",
					Description: "Post a new reaction role embed",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "channel",
							Description: "The channel to post the embed to",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "title",
							Description: "The embed title",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "description",
							Description: "The embed description",
							Required:    true,
						},
					},
				},
				{
					Name:        "add",
					Description: "Map an emoji to a role on an existing reaction role message",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "message-id",
							Description: "The UUID of the reaction role message (from database)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "emoji",
							Description: "The emoji to react with",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionRole,
							Name:        "role",
							Description: "The role to grant on click",
							Required:    true,
						},
					},
				},
				{
					Name:        "remove",
					Description: "Remove an emoji mapping from a reaction role message",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "message-id",
							Description: "The UUID of the reaction role message",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "emoji",
							Description: "The emoji to remove",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// OnInteractionCreate routes slash commands and button components.
func (ctx *HandlerContext) OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		ctx.handleSlashCommand(s, i)
	case discordgo.InteractionMessageComponent:
		ctx.HandleComponentInteraction(s, i)
	}
}

func (ctx *HandlerContext) handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	// Initial ephemeral defer response for slash commands to prevent timeouts (3s limit)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	ctx.runCommandLogic(s, i, &data)
}

func (ctx *HandlerContext) runCommandLogic(s *discordgo.Session, i *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	var response string
	var err error

	guildID, _ := strconv.ParseInt(i.GuildID, 10, 64)

	switch data.Name {
	case "botinfo":
		response = "🦐 **Shrimpy v1.0.0**\nRobust Ticket & Onboarding Assistant for Discord.\n*Built with GORM, pgxpool, and Go standard library.*"
	case "botnickname":
		// Handle nicknames
	case "setup":
		subCommand := data.Options[0]
		if subCommand.Name == "welcome" {
			response, err = ctx.handleSetupWelcome(i, guildID, subCommand.Options)
		}
	case "staff":
		subCommand := data.Options[0]
		response, err = ctx.handleStaffCommand(i, guildID, subCommand)
	case "ticket":
		subCommand := data.Options[0]
		response, err = ctx.handleTicketCommand(s, i, guildID, subCommand)
	case "reactionrole":
		subCommand := data.Options[0]
		response, err = ctx.handleReactionRoleCommand(s, i, guildID, subCommand)
	default:
		response = "Unknown command."
	}

	if err != nil {
		response = fmt.Sprintf("❌ Error: %v", err)
	}

	// Update the deferred response
	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &response,
	})
}

// ─── Setup Subcommand ────────────────────────────────────────────────────────

func (ctx *HandlerContext) handleSetupWelcome(i *discordgo.InteractionCreate, guildID int64, options []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	var channel *discordgo.Channel
	var message string
	var dmMessage *string
	enabled := true

	for _, opt := range options {
		switch opt.Name {
		case "channel":
			channel = opt.ChannelValue(nil)
		case "message":
			message = opt.StringValue()
		case "dm-message":
			dm := opt.StringValue()
			dmMessage = &dm
		case "enabled":
			enabled = opt.BoolValue()
		}
	}

	if channel == nil {
		return "", fmt.Errorf("invalid channel provided")
	}

	chID, err := repository.ParseID(channel.ID)
	if err != nil {
		return "", err
	}

	// Create/Update WelcomeConfig
	cfg := &repository.WelcomeConfig{
		GuildID:        guildID,
		Enabled:        enabled,
		ChannelID:      &chID,
		ChannelMessage: &message,
		DMMessage:      dmMessage,
	}

	_, err = ctx.WelcomeSvc.Save(context.Background(), cfg)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("✅ Welcome system configured successfully in <#%d> (Enabled: %v)", chID, enabled), nil
}

// ─── Staff Subcommands ────────────────────────────────────────────────────────

func (ctx *HandlerContext) handleStaffCommand(i *discordgo.InteractionCreate, guildID int64, opt *discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	switch opt.Name {
	case "add":
		roleID, _ := repository.ParseID(opt.Options[0].StringValue())
		_, err := ctx.GuildSvc.AddStaffRole(context.Background(), guildID, roleID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ Added <@&%d> to the support staff role list.", roleID), nil

	case "remove":
		roleID, _ := repository.ParseID(opt.Options[0].StringValue())
		err := ctx.GuildSvc.RemoveStaffRole(context.Background(), guildID, roleID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ Removed <@&%d> from the support staff role list.", roleID), nil

	case "list":
		roles, err := ctx.GuildSvc.ListStaffRoles(context.Background(), guildID)
		if err != nil {
			return "", err
		}

		if len(roles) == 0 {
			return "ℹ️ No support staff roles configured for this server.", nil
		}

		var sb strings.Builder
		sb.WriteString("📋 **Support Staff Roles:**\n")
		for _, r := range roles {
			sb.WriteString(fmt.Sprintf("- <@&%d>\n", r.RoleID))
		}
		return sb.String(), nil
	}

	return "Invalid subcommand", nil
}

// ─── Ticket Subcommands ───────────────────────────────────────────────────────

func (ctx *HandlerContext) handleTicketCommand(s *discordgo.Session, i *discordgo.InteractionCreate, guildID int64, opt *discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	channelID, _ := repository.ParseID(i.ChannelID)
	userID, _ := repository.ParseID(i.Member.User.ID)

	// Fetch ticket from active channel
	ticket, err := ctx.TicketSvc.GetByChannelID(context.Background(), channelID)
	if err != nil {
		return "", fmt.Errorf("this command can only be used inside active ticket channels")
	}

	switch opt.Name {
	case "claim":
		_, err = ctx.TicketSvc.Claim(context.Background(), s, ticket.ID, userID)
		if err != nil {
			return "", err
		}
		return "Claim request submitted.", nil

	case "unclaim":
		_, err = ctx.TicketSvc.Unclaim(context.Background(), s, ticket.ID)
		if err != nil {
			return "", err
		}
		return "Unclaim request submitted.", nil

	case "close":
		var reason *string
		if len(opt.Options) > 0 {
			r := opt.Options[0].StringValue()
			reason = &r
		}
		_, err = ctx.TicketSvc.Close(context.Background(), s, ticket.ID, reason, userID)
		if err != nil {
			return "", err
		}
		return "Ticket closing initiated.", nil

	case "priority":
		level := opt.Options[0].StringValue()
		prio := repository.TicketPriority(level)
		_, err = ctx.TicketSvc.UpdatePriority(context.Background(), ticket.ID, prio)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ Ticket priority updated to **%s**", strings.ToUpper(level)), nil

	case "add":
		targetUser := opt.Options[0].UserValue(nil)
		chIDStr := fmt.Sprintf("%d", channelID)
		err = s.ChannelPermissionSet(chIDStr, targetUser.ID, discordgo.PermissionOverwriteTypeMember,
			discordgo.PermissionReadMessages|discordgo.PermissionSendMessages|discordgo.PermissionEmbedLinks|discordgo.PermissionAttachFiles, 0)
		if err != nil {
			return "", fmt.Errorf("failed to add member permissions on channel: %w", err)
		}
		return fmt.Sprintf("✅ Added <@%s> to this ticket.", targetUser.ID), nil

	case "remove":
		targetUser := opt.Options[0].UserValue(nil)
		chIDStr := fmt.Sprintf("%d", channelID)
		err = s.ChannelPermissionDelete(chIDStr, targetUser.ID)
		if err != nil {
			return "", fmt.Errorf("failed to remove member permissions: %w", err)
		}
		return fmt.Sprintf("✅ Removed <@%s> from this ticket.", targetUser.ID), nil
	}

	return "Invalid subcommand", nil
}

// ─── ReactionRole Subcommands ─────────────────────────────────────────────────

func (ctx *HandlerContext) handleReactionRoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, guildID int64, opt *discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	switch opt.Name {
	case "create":
		channel := opt.Options[0].ChannelValue(nil)
		title := opt.Options[1].StringValue()
		desc := opt.Options[2].StringValue()

		chID, _ := repository.ParseID(channel.ID)

		msg, err := ctx.ReactionRoleSvc.Create(context.Background(), s, guildID, chID, title, desc, nil, nil)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Reaction role panel created in <#%d>. \nID: `%s` (Use this ID to add emoji role mappings).", chID, msg.ID), nil

	case "add":
		msgID := opt.Options[0].StringValue()
		emoji := opt.Options[1].StringValue()
		roleID, _ := repository.ParseID(opt.Options[2].StringValue())

		_, err := ctx.ReactionRoleSvc.AddEmoji(context.Background(), s, msgID, emoji, roleID)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Added mapping: emoji %s will grant role <@&%d> on reaction role message.", emoji, roleID), nil

	case "remove":
		msgID := opt.Options[0].StringValue()
		emoji := opt.Options[1].StringValue()

		err := ctx.ReactionRoleSvc.RemoveEmoji(context.Background(), s, msgID, emoji)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Removed mapping for emoji %s from reaction role message.", emoji), nil
	}

	return "Invalid subcommand", nil
}
