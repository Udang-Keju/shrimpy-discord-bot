package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/api"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/bot"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/bot/handlers"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/cache"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/config"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🦐 Shrimpy Backend starting up...")

	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Fatal: failed to load config: %v", err)
	}

	// 2. Connect to PostgreSQL via GORM
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Fatal: failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}

	// 3. Perform automatic database migrations
	fmt.Println("DB: Running migrations/schema auto-sync...")
	err = db.AutoMigrate(
		&repository.Guild{},
		&repository.User{},
		&repository.StaffRole{},
		&repository.AutoRole{},
		&repository.WelcomeConfig{},
		&repository.TicketPanel{},
		&repository.TicketCategory{},
		&repository.Ticket{},
		&repository.TicketMessage{},
		&repository.ReactionRoleMessage{},
		&repository.ReactionRoleEmoji{},
	)
	if err != nil {
		log.Fatalf("Fatal: failed to auto-migrate database schema: %v", err)
	}

	// 4. Initialize Config Cache
	guildCache := cache.NewGuildCache[*repository.Guild](time.Duration(cfg.CacheTTLSeconds) * time.Second)

	// 5. Initialize Repositories
	guildRepo := repository.NewGuildRepo(db)
	userRepo := repository.NewUserRepo(db)
	welcomeRepo := repository.NewWelcomeRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	ticketRepo := repository.NewTicketRepo(db)
	messageRepo := repository.NewMessageRepo(db)
	reactionRoleRepo := repository.NewReactionRoleRepo(db)

	// 6. Initialize Services
	guildSvc := service.NewGuildService(guildRepo, guildCache)
	welcomeSvc := service.NewWelcomeService(welcomeRepo)
	autoRoleSvc := service.NewAutoRoleService(guildRepo)
	reactionRoleSvc := service.NewReactionRoleService(reactionRoleRepo)
	transcriptSvc := service.NewTranscriptService(messageRepo)
	ticketSvc := service.NewTicketService(ticketRepo, categoryRepo, guildRepo, messageRepo, transcriptSvc)
	schedulerSvc := service.NewScheduler(ticketRepo, ticketSvc, cfg.AutoCloseCheckInterval)

	// 7. Initialize Bot
	handlerCtx := handlers.NewHandlerContext(guildSvc, welcomeSvc, autoRoleSvc, reactionRoleSvc, ticketSvc)
	devGuildID := os.Getenv("DEV_GUILD_ID") // Optional, for instant slash command dev testing
	discordBot, err := bot.New(cfg.DiscordToken, handlerCtx, devGuildID)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize Discord bot: %v", err)
	}

	// 8. Initialize REST API Server
	apiServer := api.NewServer(
		cfg.APIPort,
		[]byte(cfg.JWTSecret),
		cfg.TokenEncryptionKey,
		categoryRepo,
		userRepo,
		guildSvc,
		welcomeSvc,
		autoRoleSvc,
		reactionRoleSvc,
		ticketSvc,
		transcriptSvc,
		discordBot.Session,
	)
	apiServer.SetupRoutes(cfg.CORSAllowedOrigins)

	// 9. Startup Services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start auto-close background worker
	go schedulerSvc.Start(ctx, discordBot.Session)

	// Start Discord Gateway connection
	fmt.Println("Bot: Connecting to Discord Gateway...")
	if err := discordBot.Start(); err != nil {
		log.Fatalf("Fatal: failed to start Discord bot: %v", err)
	}

	// Start HTTP REST API server
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("Fatal: REST API server crashed: %v", err)
		}
	}()

	fmt.Println("Shrimpy Backend is now fully operational!")

	// 10. Handle Graceful Shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("\nShrimpy Backend shutting down gracefully...")
	cancel() // Stop the background scheduler

	if err := discordBot.Stop(); err != nil {
		fmt.Printf("Error during bot shutdown: %v\n", err)
	}

	fmt.Println("Shrimpy Backend successfully stopped. Goodbye! 🦐")
}
