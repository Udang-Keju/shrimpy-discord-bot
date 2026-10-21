package clock

import (
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/handler"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"gorm.io/gorm"
)

// Module wraps all layers of the clock channels feature. It has no bot layer:
// clocks are configured from the dashboard and refreshed by the Scheduler.
type Module struct {
	Repo      *repository.ClockRepo
	Service   *service.ClockService
	Handler   *handler.Handler
	Scheduler *service.Scheduler
}

// Build compiles all layers of the clock channels feature.
func Build(db *gorm.DB, provider discordutil.DiscordSessionProvider) *Module {
	repo := repository.NewClockRepo(db)
	svc := service.NewClockService(repo, provider)
	h := handler.NewHandler(svc)
	sched := service.NewScheduler(repo, service.DefaultRefreshInterval)

	return &Module{
		Repo:      repo,
		Service:   svc,
		Handler:   h,
		Scheduler: sched,
	}
}

// Models returns all GORM models utilized by the clock channels feature.
func (m *Module) Models() []any {
	return []any{&model.ClockChannel{}}
}
