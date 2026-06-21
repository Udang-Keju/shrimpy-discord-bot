package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// HandlePrefixCommand parses and routes legacy text-based commands (e.g. !help, !botinfo, !set prefix).
func (ctx *HandlerContext) HandlePrefixCommand(s *discordgo.Session, m *discordgo.MessageCreate, prefix string) {
	content := m.Content[len(prefix):]
	args := strings.Fields(content)

	if len(args) == 0 {
		return
	}

	commandName := strings.ToLower(args[0])

	switch commandName {
	case "help":
		ctx.handlePrefixHelp(s, m, prefix)
	case "botinfo":
		_, _ = s.ChannelMessageSend(m.ChannelID, "🦐 **Shrimpy v1.0.0**\nRobust Ticket & Onboarding Assistant for Discord.\n*Use slash commands (e.g. `/botinfo`) for full interactivity.*")
	case "set":
		if len(args) < 3 {
			_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("ℹ️ Usage: `%sset prefix [new_prefix]`", prefix))
			return
		}

		subSetting := strings.ToLower(args[1])
		newValue := args[2]

		if subSetting == "prefix" {
			ctx.handleSetPrefix(s, m, newValue)
		}
	}
}

func (ctx *HandlerContext) handlePrefixHelp(s *discordgo.Session, m *discordgo.MessageCreate, prefix string) {
	helpText := fmt.Sprintf(`📋 **Shrimpy Text Commands Help:**
- `+"`%shelp`"+` — Displays this list.
- `+"`%sbotinfo`"+` — Displays bot runtime information.
- `+"`%sset prefix [prefix]`"+` — Update command prefix (Admin only).

*Note: Shrimpy is fully optimized for Slash Commands! Type `+"`/`"+` to view all interactive commands (e.g., ticket claim/close, panel setup).*`, prefix, prefix, prefix)

	_, _ = s.ChannelMessageSend(m.ChannelID, helpText)
}

func (ctx *HandlerContext) handleSetPrefix(s *discordgo.Session, m *discordgo.MessageCreate, newPrefix string) {
	// Administrator permission check
	perms, err := s.State.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		// Fallback check
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
	_, err = ctx.GuildSvc.UpdateConfig(context.Background(), guildID, map[string]interface{}{
		"prefix": newPrefix,
	})
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Failed to update prefix: %v", err))
		return
	}

	_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Command prefix updated to: `%s`", newPrefix))
}
