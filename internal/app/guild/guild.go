package guild

import (
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/bot"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/handler"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/service"
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

// Module wraps all layers of the guild business feature.
type Module struct {
	Repo        *repository.GuildRepo
	Service     *service.GuildService
	AutoRoleSvc *service.AutoRoleService
	Handler     *handler.Handler
	Bot         *bot.BotHandler
}

// Build compiles all layers of the guild feature.
func Build(db *gorm.DB, cacheTTL time.Duration, dg *discordgo.Session) *Module {
	guildCache := repository.NewGuildCache[*model.Guild](cacheTTL)
	repo := repository.NewGuildRepo(db)
	svc := service.NewGuildService(repo, guildCache)
	autoRoleSvc := service.NewAutoRoleService(repo)
	h := handler.NewHandler(svc, dg)
	b := bot.NewBotHandler(svc)

	return &Module{
		Repo:        repo,
		Service:     svc,
		AutoRoleSvc: autoRoleSvc,
		Handler:     h,
		Bot:         b,
	}
}

// Models returns all GORM models utilized by the guild feature.
func (m *Module) Models() []any {
	return []any{
		&model.Guild{},
		&model.StaffRole{},
		&model.AutoRole{},
	}
}

