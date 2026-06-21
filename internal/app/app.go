package app

import (
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings"
	settings_svc "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"gorm.io/gorm"
)

// Modules holds all compiled vertical feature modules for the application.
type Modules struct {
	Settings     *settings.Module
	Auth         *auth.Module
	Guild        *guild.Module
	Welcome      *welcome.Module
	ReactionRole *reactionrole.Module
	Ticket       *ticket.Module
}

// Build compiles and connects all modules with their respective layers.
func Build(
	db *gorm.DB,
	provider discordutil.DiscordSessionProvider,
	controller settings_svc.BotSessionController,
	jwtSecret []byte,
	tokenEncKey []byte,
	guildCacheTTL time.Duration,
) *Modules {
	settingsMod := settings.Build(db, tokenEncKey, controller)
	authMod := auth.Build(db, jwtSecret, tokenEncKey, settingsMod.Service)
	guildMod := guild.Build(db, guildCacheTTL, provider)
	welcomeMod := welcome.Build(db)
	reactionRoleMod := reactionrole.Build(db, provider)
	ticketMod := ticket.Build(db, guildMod.Repo, provider)

	return &Modules{
		Settings:     settingsMod,
		Auth:         authMod,
		Guild:        guildMod,
		Welcome:      welcomeMod,
		ReactionRole: reactionRoleMod,
		Ticket:       ticketMod,
	}
}

// Models aggregates and returns all GORM schema models across all submodules.
func (m *Modules) Models() []any {
	var all []any
	all = append(all, m.Settings.Models()...)
	all = append(all, m.Auth.Models()...)
	all = append(all, m.Guild.Models()...)
	all = append(all, m.Welcome.Models()...)
	all = append(all, m.ReactionRole.Models()...)
	all = append(all, m.Ticket.Models()...)
	return all
}

