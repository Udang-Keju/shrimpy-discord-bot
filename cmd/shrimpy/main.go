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
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth"
	auth_model "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild"
	guild_model "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole"
	rr_model "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket"
	ticket_config "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/config"
	ticket_model "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome"
	welcome_model "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/bot"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/bot/handlers"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/cache"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/config"
	"github.com/bwmarrin/discordgo"
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
		&guild_model.Guild{},
		&auth_model.User{},
		&guild_model.StaffRole{},
		&guild_model.AutoRole{},
		&welcome_model.WelcomeConfig{},
		&ticket_model.TicketPanel{},
		&ticket_model.TicketCategory{},
		&ticket_model.Ticket{},
		&ticket_model.TicketMessage{},
		&rr_model.ReactionRoleMessage{},
		&rr_model.ReactionRoleEmoji{},
	)
	if err != nil {
		log.Fatalf("Fatal: failed to auto-migrate database schema: %v", err)
	}

	// 4. Initialize Config Cache
	guildCache := cache.NewGuildCache[*guild_model.Guild](time.Duration(cfg.CacheTTLSeconds) * time.Second)

	// 5. Load feature configs
	ticketCfg := ticket_config.Load()

	// 6. Initialize DiscordGo Session Early (solves chicken-and-egg initialization)
	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("Fatal: failed to construct Discord session: %v", err)
	}
	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentMessageContent

	// 7. Build Business Feature Modules
	authModule := auth.Build(db, []byte(cfg.JWTSecret), cfg.TokenEncryptionKey)
	guildModule := guild.Build(db, guildCache, dg)
	welcomeModule := welcome.Build(db)
	reactionRoleModule := reactionrole.Build(db, dg)
	ticketModule := ticket.Build(db, guildModule.Repo, ticketCfg, dg)

	// 8. Initialize Bot Handler Context
	handlerCtx := handlers.NewHandlerContext(
		guildModule.Service,
		guildModule.AutoRoleSvc,
		welcomeModule.Service,
		reactionRoleModule.Service,
		ticketModule.TicketSvc,
		guildModule.Bot,
		welcomeModule.Bot,
		reactionRoleModule.Bot,
		ticketModule.Bot,
	)

	// 9. Initialize Bot
	devGuildID := os.Getenv("DEV_GUILD_ID") // Optional, for instant slash command dev testing
	discordBot := bot.New(dg, handlerCtx, devGuildID)

	// 10. Initialize REST API Server
	apiServer := api.NewServer(
		cfg.APIPort,
		[]byte(cfg.JWTSecret),
		cfg.TokenEncryptionKey,
		authModule.Handler,
		guildModule.Handler,
		welcomeModule.Handler,
		ticketModule.Handler,
		reactionRoleModule.Handler,
		guildModule.Service,
		dg,
	)
	apiServer.SetupRoutes(cfg.CORSAllowedOrigins)

	// 11. Startup Services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start auto-close background worker
	go ticketModule.SchedulerSvc.Start(ctx, dg)

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

	// 12. Handle Graceful Shutdown
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
