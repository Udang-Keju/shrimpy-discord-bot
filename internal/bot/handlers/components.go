package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/bwmarrin/discordgo"
)

// HandleComponentInteraction routes button clicks and dropdown select menu choices.
func (ctx *HandlerContext) HandleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()
	parts := strings.Split(data.CustomID, ":")

	if len(parts) < 3 || parts[0] != "ticket" {
		return
	}

	action := parts[1]
	targetID := parts[2]

	guildID, _ := strconv.ParseInt(i.GuildID, 10, 64)
	userID, _ := strconv.ParseInt(i.Member.User.ID, 10, 64)

	// Defer response to prevent Gateway timeouts
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
		// targetID = Category UUID
		var ticket *repository.Ticket
		ticket, err = ctx.TicketSvc.Open(context.Background(), s, guildID, targetID, userID)
		if err == nil {
			if ticket.ChannelID != nil {
				response = fmt.Sprintf("✅ Support ticket created successfully in <#%d>!", *ticket.ChannelID)
			} else {
				response = "✅ Support ticket created successfully!"
			}
		} else if err == repository.ErrLimitReached {
			response = "❌ You have reached the limit of open tickets in this category."
		} else {
			response = fmt.Sprintf("❌ Failed to create ticket: %v", err)
		}

	case "claim":
		// targetID = Ticket UUID
		_, err = ctx.TicketSvc.Claim(context.Background(), s, targetID, userID)
		if err == nil {
			response = "✅ You have successfully claimed this ticket."
		} else {
			response = fmt.Sprintf("❌ Failed to claim ticket: %v", err)
		}

	case "close":
		// targetID = Ticket UUID
		_, err = ctx.TicketSvc.Close(context.Background(), s, targetID, nil, userID)
		if err == nil {
			response = "✅ Ticket has been successfully closed."
		} else {
			response = fmt.Sprintf("❌ Failed to close ticket: %v", err)
		}

	case "reopen":
		// targetID = Ticket UUID
		_, err = ctx.TicketSvc.Reopen(context.Background(), s, targetID)
		if err == nil {
			response = "✅ Ticket has been successfully reopened."
		} else {
			response = fmt.Sprintf("❌ Failed to reopen ticket: %v", err)
		}

	case "archive":
		// targetID = Ticket UUID
		err = ctx.TicketSvc.Archive(context.Background(), s, targetID)
		if err == nil {
			response = "✅ Ticket has been successfully archived."
		} else {
			response = fmt.Sprintf("❌ Failed to archive ticket: %v", err)
		}

	default:
		response = "Unknown component action."
	}

	// Edit deferred response
	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &response,
	})
}
