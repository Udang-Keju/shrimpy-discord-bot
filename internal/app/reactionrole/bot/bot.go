package bot

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
)

// BotHandler handles Discord bot reaction events and slash commands for reaction roles.
type BotHandler struct {
	reactionRoleSvc *service.ReactionRoleService
}

// NewBotHandler constructs a new BotHandler.
func NewBotHandler(reactionRoleSvc *service.ReactionRoleService) *BotHandler {
	return &BotHandler{
		reactionRoleSvc: reactionRoleSvc,
	}
}

// OnMessageReactionAdd handles roles being granted on reaction click.
func (h *BotHandler) OnMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
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
		err := h.reactionRoleSvc.HandleReactionAdd(context.Background(), s, discordMsgID, guildID, userID, emojiStr)
		if err != nil {
			fmt.Printf("Bot Error: failed to handle reaction role grant: %v\n", err)
		}
	}()
}

// OnMessageReactionRemove handles roles being revoked when reaction is removed.
func (h *BotHandler) OnMessageReactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
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
		err := h.reactionRoleSvc.HandleReactionRemove(context.Background(), s, discordMsgID, guildID, userID, emojiStr)
		if err != nil {
			fmt.Printf("Bot Error: failed to handle reaction role revoke: %v\n", err)
		}
	}()
}

// HandleReactionRoleCommand manages reaction role embeds and emoji mappings.
func (h *BotHandler) HandleReactionRoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, guildID int64, opt *discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	switch opt.Name {
	case "create":
		channel := opt.Options[0].ChannelValue(nil)
		title := opt.Options[1].StringValue()
		desc := opt.Options[2].StringValue()

		chID, _ := strconv.ParseInt(channel.ID, 10, 64)

		msg, err := h.reactionRoleSvc.Create(context.Background(), s, guildID, chID, title, desc, nil, nil)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Reaction role panel created in <#%d>. \nID: `%s` (Use this ID to add emoji role mappings).", chID, msg.ID), nil

	case "add":
		msgID := opt.Options[0].StringValue()
		emoji := opt.Options[1].StringValue()
		roleID, _ := discordutil.ParseID(opt.Options[2].StringValue())

		_, err := h.reactionRoleSvc.AddEmoji(context.Background(), s, msgID, emoji, roleID)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Added mapping: emoji %s will grant role <@&%d> on reaction role message.", emoji, roleID), nil

	case "remove":
		msgID := opt.Options[0].StringValue()
		emoji := opt.Options[1].StringValue()

		err := h.reactionRoleSvc.RemoveEmoji(context.Background(), s, msgID, emoji)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("✅ Removed mapping for emoji %s from reaction role message.", emoji), nil
	}

	return "Invalid subcommand", nil
}
