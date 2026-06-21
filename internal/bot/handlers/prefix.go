package handlers

import (
	"fmt"
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
			ctx.GuildBot.HandleSetPrefix(s, m, newValue)
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
