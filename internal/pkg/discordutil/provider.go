package discordutil

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// DiscordSessionProvider interface allows retrieving guild-specific
// and app-specific discordgo sessions dynamically in a multi-bot environment.
type DiscordSessionProvider interface {
	GetSessionForGuild(ctx context.Context, guildID int64) (*discordgo.Session, error)
	IsBotInGuild(guildID string) bool
}
