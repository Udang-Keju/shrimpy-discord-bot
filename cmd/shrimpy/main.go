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
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/bot"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/bot/handlers"
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

	// 3. Initialize DiscordGo Session Early (solves chicken-and-egg initialization)
	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("Fatal: failed to construct Discord session: %v", err)
	}
	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentMessageContent

	// 4. Build Business Feature Modules
	modules := app.Build(
		db,
		dg,
		[]byte(cfg.JWTSecret),
		cfg.TokenEncryptionKey,
		time.Duration(cfg.CacheTTLSeconds)*time.Second,
	)

	// 5. Perform automatic database migrations
	fmt.Println("DB: Running migrations/schema auto-sync...")
	err = db.AutoMigrate(modules.Models()...)
	if err != nil {
		log.Fatalf("Fatal: failed to auto-migrate database schema: %v", err)
	}

	// 6. Initialize Bot Handler Context
	handlerCtx := handlers.NewHandlerContext(modules)

	// 7. Initialize Bot
	devGuildID := os.Getenv("DEV_GUILD_ID") // Optional, for instant slash command dev testing
	discordBot := bot.New(dg, handlerCtx, devGuildID)

	// 8. Initialize REST API Server
	apiServer := api.NewServer(
		cfg.APIPort,
		[]byte(cfg.JWTSecret),
		cfg.TokenEncryptionKey,
		modules,
		dg,
	)
	apiServer.SetupRoutes(cfg.CORSAllowedOrigins)

	// 9. Startup Services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start auto-close background worker
	go modules.Ticket.SchedulerSvc.Start(ctx, dg)

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
