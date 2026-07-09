package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
)

// BotHandler handles Discord bot events, interactions, and commands for the Ticket feature.
type BotHandler struct {
	ticketSvc *service.TicketService
}

// actorFromInteraction builds a service.Actor from the interaction's member,
// carrying the member's roles and whether they hold Administrator / Manage
// Server (which bypasses the ticket-role checks).
func actorFromInteraction(i *discordgo.InteractionCreate) service.Actor {
	if i.Member == nil || i.Member.User == nil {
		return service.Actor{}
	}
	userID, _ := strconv.ParseInt(i.Member.User.ID, 10, 64)
	roleIDs := make([]int64, 0, len(i.Member.Roles))
	for _, r := range i.Member.Roles {
		if id, err := strconv.ParseInt(r, 10, 64); err == nil {
			roleIDs = append(roleIDs, id)
		}
	}
	const adminPerms = discordgo.PermissionAdministrator | discordgo.PermissionManageServer
	return service.Actor{
		UserID:     userID,
		RoleIDs:    roleIDs,
		Privileged: i.Member.Permissions&adminPerms != 0,
	}
}

// NewBotHandler constructs a new BotHandler.
func NewBotHandler(ticketSvc *service.TicketService) *BotHandler {
	return &BotHandler{
		ticketSvc: ticketSvc,
	}
}

// OnMessageCreate logs messages in active ticket channels to build transcripts.
func (h *BotHandler) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	channelID, err := strconv.ParseInt(m.ChannelID, 10, 64)
	if err != nil {
		return
	}

	go func() {
		ticket, err := h.ticketSvc.GetByChannelID(context.Background(), channelID)
		if err == nil && ticket != nil && ticket.Status != model.TicketStatusArchived {
			authorID, _ := strconv.ParseInt(m.Author.ID, 10, 64)
			content := m.Content

			var attachments []discordutil.Attachment
			for _, att := range m.Attachments {
				attachments = append(attachments, discordutil.Attachment{
					Filename: att.Filename,
					URL:      att.URL,
					Size:     att.Size,
				})
			}

			isStaffNote := false
			if strings.HasPrefix(m.Content, "*") && strings.HasSuffix(m.Content, "*") {
				isStaffNote = true
			}

			_ = h.ticketSvc.LogMessage(context.Background(), ticket.ID, authorID, m.Author.Username, &content, isStaffNote, attachments)
		}
	}()
}

// HandleComponentInteraction routes button clicks for opening/claiming/closing/re-opening/archiving tickets.
func (h *BotHandler) HandleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	parts := strings.Split(data.CustomID, ":")

	if len(parts) < 2 || parts[0] != "ticket" {
		return
	}

	action := parts[1]
	var targetID string
	if action == "open_select" {
		if len(data.Values) == 0 {
			return
		}
		action = "open"
		targetID = data.Values[0]
	} else {
		if len(parts) < 3 {
			return
		}
		targetID = parts[2]
	}

	guildID, _ := strconv.ParseInt(i.GuildID, 10, 64)
	userID, _ := strconv.ParseInt(i.Member.User.ID, 10, 64)
	actor := actorFromInteraction(i)

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	var response string
	var err error

	switch action {
	case "open":
		var ticket *model.Ticket
		ticket, err = h.ticketSvc.Open(context.Background(), s, guildID, targetID, userID)
		if err == nil {
			if ticket.ChannelID != nil {
				response = fmt.Sprintf("✅ Support ticket created successfully in <#%d>!", *ticket.ChannelID)
			} else {
				response = "✅ Support ticket created successfully!"
			}
		} else if err == model.ErrLimitReached {
			response = "❌ You have reached the limit of open tickets in this category."
		} else {
			response = fmt.Sprintf("❌ Failed to create ticket: %v", err)
		}

	case "claim":
		_, err = h.ticketSvc.Claim(context.Background(), s, targetID, actor)
		if err == nil {
			response = "✅ You have successfully claimed this ticket."
		} else {
			response = fmt.Sprintf("❌ Failed to claim ticket: %v", err)
		}

	case "resolve":
		_, err = h.ticketSvc.Resolve(context.Background(), s, targetID, actor)
		if err == nil {
			response = "✅ Ticket has been marked as resolved."
		} else {
			response = fmt.Sprintf("❌ Failed to resolve ticket: %v", err)
		}

	case "unresolve":
		_, err = h.ticketSvc.Unresolve(context.Background(), s, targetID, actor)
		if err == nil {
			response = "✅ Ticket has been un-resolved and is active again."
		} else {
			response = fmt.Sprintf("❌ Failed to un-resolve ticket: %v", err)
		}

	case "close":
		_, err = h.ticketSvc.Close(context.Background(), s, targetID, nil, actor)
		if err == nil {
			response = "✅ Ticket has been successfully closed."
		} else {
			response = fmt.Sprintf("❌ Failed to close ticket: %v", err)
		}

	case "reopen":
		_, err = h.ticketSvc.Reopen(context.Background(), s, targetID, actor)
		if err == nil {
			response = "✅ Ticket has been successfully reopened."
		} else {
			response = fmt.Sprintf("❌ Failed to reopen ticket: %v", err)
		}

	case "archive":
		err = h.ticketSvc.Archive(context.Background(), s, targetID, actor)
		if err == nil {
			response = "✅ Ticket has been successfully archived."
		} else {
			response = fmt.Sprintf("❌ Failed to archive ticket: %v", err)
		}

	default:
		response = "Unknown component action."
	}

	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &response,
	})
}

// HandleTicketCommand routes slash commands for ticket actions inside ticket channels.
func (h *BotHandler) HandleTicketCommand(s *discordgo.Session, i *discordgo.InteractionCreate, guildID int64, opt *discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	channelID, _ := strconv.ParseInt(i.ChannelID, 10, 64)
	actor := actorFromInteraction(i)

	ticket, err := h.ticketSvc.GetByChannelID(context.Background(), channelID)
	if err != nil {
		return "", fmt.Errorf("this command can only be used inside active ticket channels")
	}

	switch opt.Name {
	case "claim":
		_, err = h.ticketSvc.Claim(context.Background(), s, ticket.ID, actor)
		if err != nil {
			return "", err
		}
		return "Claim request submitted.", nil

	case "unclaim":
		_, err = h.ticketSvc.Unclaim(context.Background(), s, ticket.ID, actor)
		if err != nil {
			return "", err
		}
		return "Unclaim request submitted.", nil

	case "resolve":
		_, err = h.ticketSvc.Resolve(context.Background(), s, ticket.ID, actor)
		if err != nil {
			return "", err
		}
		return "Ticket marked as resolved.", nil

	case "close":
		var reason *string
		if len(opt.Options) > 0 {
			r := opt.Options[0].StringValue()
			reason = &r
		}
		_, err = h.ticketSvc.Close(context.Background(), s, ticket.ID, reason, actor)
		if err != nil {
			return "", err
		}
		return "Ticket closing initiated.", nil

	case "priority":
		level := opt.Options[0].StringValue()
		prio := model.TicketPriority(level)
		_, err = h.ticketSvc.UpdatePriority(context.Background(), ticket.ID, prio)
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
