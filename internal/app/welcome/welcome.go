package welcome

import (
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/bot"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/handler"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/service"
	"gorm.io/gorm"
)

// Module wraps all layers of the welcome business feature.
type Module struct {
	Repo    *repository.WelcomeRepo
	Service *service.WelcomeService
	Handler *handler.Handler
	Bot     *bot.BotHandler
}

// Build compiles all layers of the welcome feature.
func Build(db *gorm.DB) *Module {
	repo := repository.NewWelcomeRepo(db)
	svc := service.NewWelcomeService(repo)
	h := handler.NewHandler(svc)
	b := bot.NewBotHandler(svc)

	return &Module{
		Repo:    repo,
		Service: svc,
		Handler: h,
		Bot:     b,
	}
}
