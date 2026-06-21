package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/service"
	"github.com/bwmarrin/discordgo"
)

// BotHandler handles Discord bot interactions, commands, and events for the Guild/Staff feature.
type BotHandler struct {
	guildSvc *service.GuildService
}

// NewBotHandler constructs a new BotHandler.
func NewBotHandler(guildSvc *service.GuildService) *BotHandler {
	return &BotHandler{guildSvc: guildSvc}
}

// OnGuildCreate registers the guild in the DB on invite or reconnect.
func (h *BotHandler) OnGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	guildID, err := strconv.ParseInt(g.ID, 10, 64)
	if err != nil {
		return
	}

	// Auto-register guild config in DB (upsert) to ensure defaults are populated
	_, err = h.guildSvc.GetConfig(context.Background(), guildID)
	if err != nil {
		fmt.Printf("Bot Error: failed to register guild %d: %v\n", guildID, err)
	} else {
		fmt.Printf("Bot: Guild registered/loaded: %s (%d)\n", g.Name, guildID)
	}
}

// OnGuildDelete handles when the bot is kicked from a server.
func (h *BotHandler) OnGuildDelete(s *discordgo.Session, g *discordgo.GuildDelete) {
	guildID, err := strconv.ParseInt(g.ID, 10, 64)
	if err != nil {
		return
	}

	err = h.guildSvc.Deactivate(context.Background(), guildID)
	if err != nil {
		fmt.Printf("Bot Error: failed to deactivate guild %d: %v\n", guildID, err)
	} else {
		fmt.Printf("Bot: Deactivated guild config: %d\n", guildID)
	}
}

// HandleStaffCommand manages adding, removing, and listing support staff roles.
func (h *BotHandler) HandleStaffCommand(s *discordgo.Session, i *discordgo.InteractionCreate, guildID int64, opt *discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	switch opt.Name {
	case "add":
		roleID, _ := strconv.ParseInt(opt.Options[0].StringValue(), 10, 64)
		_, err := h.guildSvc.AddStaffRole(context.Background(), guildID, roleID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ Added <@&%d> to the support staff role list.", roleID), nil

	case "remove":
		roleID, _ := strconv.ParseInt(opt.Options[0].StringValue(), 10, 64)
		err := h.guildSvc.RemoveStaffRole(context.Background(), guildID, roleID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ Removed <@&%d> from the support staff role list.", roleID), nil

	case "list":
		roles, err := h.guildSvc.ListStaffRoles(context.Background(), guildID)
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

// HandleSetPrefix handles the legacy text command to set the prefix.
func (h *BotHandler) HandleSetPrefix(s *discordgo.Session, m *discordgo.MessageCreate, newPrefix string) {
	perms, err := s.State.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		member, err := s.State.Member(m.GuildID, m.Author.ID)
		if err == nil {
			for _, roleID := range member.Roles {
				role, err := s.State.Role(m.GuildID, roleID)
				if err == nil && (role.Permissions&discordgo.PermissionAdministrator != 0) {
					perms = discordgo.PermissionAdministrator
					break
				}
			}
		}
	}

	if perms&discordgo.PermissionAdministrator == 0 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "❌ Only server administrators can update the command prefix.")
		return
	}

	if len(newPrefix) > 10 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "❌ Prefix cannot exceed 10 characters.")
		return
	}

	guildID, _ := strconv.ParseInt(m.GuildID, 10, 64)
	_, err = h.guildSvc.UpdateConfig(context.Background(), guildID, map[string]interface{}{
		"prefix": newPrefix,
	})
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Failed to update prefix: %v", err))
		return
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Command prefix updated to: `%s`", newPrefix))
}
