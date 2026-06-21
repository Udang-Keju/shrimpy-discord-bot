package settings

import (
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/handler"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/service"
	"gorm.io/gorm"
)

// Module wraps all layers of the settings feature.
type Module struct {
	Repo    *repository.SettingsRepo
	Service *service.SettingsService
	Handler *handler.SettingsHandler
}

// Build compiles all layers of the settings feature.
// reconnectFn is called with the new plaintext token whenever the bot token is updated.
func Build(db *gorm.DB, tokenEncKey []byte, reconnectFn handler.ReconnectFn) *Module {
	repo := repository.NewSettingsRepo(db)
	svc := service.NewSettingsService(repo, tokenEncKey)
	h := handler.NewSettingsHandler(svc, reconnectFn)
	return &Module{
		Repo:    repo,
		Service: svc,
		Handler: h,
	}
}

// Models returns all GORM models used by the settings feature.
func (m *Module) Models() []any {
	return []any{&model.BotSettings{}}
}
