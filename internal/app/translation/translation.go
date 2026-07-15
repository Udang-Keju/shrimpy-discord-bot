package translation

import (
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/bot"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/handler"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/service"
	"gorm.io/gorm"
)

// Module wraps all layers of the translation feature.
type Module struct {
	Repo    *repository.TranslationRepo
	Service *service.TranslationService
	Handler *handler.Handler
	Bot     *bot.BotHandler
}

// Build compiles all layers of the translation feature. guildProvider resolves
// the per-guild fallback target language; tokenEncKey encrypts the engine key.
func Build(db *gorm.DB, guildProvider service.GuildConfigProvider, tokenEncKey []byte) *Module {
	repo := repository.NewTranslationRepo(db)
	svc := service.NewTranslationService(repo, guildProvider, tokenEncKey)
	h := handler.NewHandler(svc)
	b := bot.NewBotHandler(svc)

	return &Module{
		Repo:    repo,
		Service: svc,
		Handler: h,
		Bot:     b,
	}
}

// Models returns all GORM models utilized by the translation feature.
func (m *Module) Models() []any {
	return []any{
		&model.TranslationConfig{},
		&model.TranslationChannel{},
		&model.TranslationReactionEmoji{},
	}
}
