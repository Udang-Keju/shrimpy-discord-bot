package reactionrole

import (
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/bot"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/handler"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"gorm.io/gorm"
)

// Module wraps all layers of the reaction role business feature.
type Module struct {
	Repo    *repository.ReactionRoleRepo
	Service *service.ReactionRoleService
	Handler *handler.Handler
	Bot     *bot.BotHandler
}

// Build compiles all layers of the reaction role feature.
func Build(db *gorm.DB, provider discordutil.DiscordSessionProvider) *Module {
	repo := repository.NewReactionRoleRepo(db)
	svc := service.NewReactionRoleService(repo)
	h := handler.NewHandler(svc, provider)
	b := bot.NewBotHandler(svc)

	return &Module{
		Repo:    repo,
		Service: svc,
		Handler: h,
		Bot:     b,
	}
}

// Models returns all GORM models utilized by the reaction role feature.
func (m *Module) Models() []any {
	return []any{
		&model.ReactionRoleMessage{},
		&model.ReactionRoleEmoji{},
	}
}

