package ticket

import (
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/bot"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/config"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/handler"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/service"
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

// Module wraps all layers of the ticket business feature.
type Module struct {
	CategoryRepo  *repository.CategoryRepo
	TicketRepo    *repository.TicketRepo
	MessageRepo   *repository.MessageRepo
	TicketSvc     *service.TicketService
	TranscriptSvc *service.TranscriptService
	SchedulerSvc  *service.Scheduler
	Handler       *handler.Handler
	Bot           *bot.BotHandler
}

// Build compiles all layers of the ticket feature.
func Build(db *gorm.DB, guildRepo service.TicketGuildRepository, dg *discordgo.Session) *Module {
	categoryRepo := repository.NewCategoryRepo(db)
	ticketRepo := repository.NewTicketRepo(db)
	messageRepo := repository.NewMessageRepo(db)

	ticketCfg := config.Load()

	transcriptSvc := service.NewTranscriptService(messageRepo)
	ticketSvc := service.NewTicketService(ticketRepo, categoryRepo, guildRepo, messageRepo, transcriptSvc)
	schedulerSvc := service.NewScheduler(ticketRepo, ticketSvc, ticketCfg.AutoCloseCheckInterval)

	h := handler.NewHandler(ticketSvc, categoryRepo, transcriptSvc, dg)
	b := bot.NewBotHandler(ticketSvc)

	return &Module{
		CategoryRepo:  categoryRepo,
		TicketRepo:    ticketRepo,
		MessageRepo:   messageRepo,
		TicketSvc:     ticketSvc,
		TranscriptSvc: transcriptSvc,
		SchedulerSvc:  schedulerSvc,
		Handler:       h,
		Bot:           b,
	}
}

// Models returns all GORM models utilized by the ticket feature.
func (m *Module) Models() []any {
	return []any{
		&model.TicketPanel{},
		&model.TicketCategory{},
		&model.Ticket{},
		&model.TicketMessage{},
	}
}

