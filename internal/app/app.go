package app

import (
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings"
	settings_handler "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/handler"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome"
	"github.com/bwmarrin/discordgo"
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
// reconnectFn is passed to the settings module so the dashboard can trigger live bot reconnects.
func Build(
	db *gorm.DB,
	dg *discordgo.Session,
	jwtSecret []byte,
	tokenEncKey []byte,
	guildCacheTTL time.Duration,
	reconnectFn settings_handler.ReconnectFn,
) *Modules {
	settingsMod := settings.Build(db, tokenEncKey, reconnectFn)
	authMod := auth.Build(db, jwtSecret, tokenEncKey, settingsMod.Service)
	guildMod := guild.Build(db, guildCacheTTL, dg)
	welcomeMod := welcome.Build(db)
	reactionRoleMod := reactionrole.Build(db, dg)
	ticketMod := ticket.Build(db, guildMod.Repo, dg)

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

