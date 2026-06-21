package handlers

import (
	"fmt"
	"strconv"

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
	case "setup":
		subCommand := data.Options[0]
		if subCommand.Name == "welcome" {
			response, err = ctx.WelcomeBot.HandleSetupWelcome(s, i, guildID, subCommand.Options)
		}
	case "staff":
		subCommand := data.Options[0]
		response, err = ctx.GuildBot.HandleStaffCommand(s, i, guildID, subCommand)
	case "ticket":
		subCommand := data.Options[0]
		response, err = ctx.TicketBot.HandleTicketCommand(s, i, guildID, subCommand)
	case "reactionrole":
		subCommand := data.Options[0]
		response, err = ctx.ReactionRoleBot.HandleReactionRoleCommand(s, i, guildID, subCommand)
	default:
		response = "Unknown command."
	}

	if err != nil {
		response = fmt.Sprintf("❌ Error: %v", err)
	}

	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &response,
	})
}
