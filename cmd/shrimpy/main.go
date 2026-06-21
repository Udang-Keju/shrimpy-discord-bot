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
	settings_model "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/model"
	settings_repo "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/repository"
	settings_svc "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/service"
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

	// 3. Migrate bot_settings early so we can seed/read credentials before the session is built.
	if err := db.AutoMigrate(&settings_model.BotSettings{}); err != nil {
		log.Fatalf("Fatal: failed to migrate bot_settings table: %v", err)
	}

	// 4. Bootstrap settings repo/service (before full app.Build wiring)
	bootRepo := settings_repo.NewSettingsRepo(db)
	bootSvc := settings_svc.NewSettingsService(bootRepo, cfg.TokenEncryptionKey)

	// 5. Seed bot_settings from env vars if the row doesn't exist yet (first boot)
	ctx := context.Background()
	_, seedErr := bootRepo.Get(ctx)
	if seedErr != nil {
		if cfg.HasDiscordSeed() {
			fmt.Println("DB: Seeding bot_settings from environment variables (first boot)...")
			if err := bootSvc.SeedFromEnv(ctx,
				cfg.DiscordToken,
				cfg.DiscordClientID,
				cfg.DiscordClientSecret,
				cfg.DiscordRedirectURI,
			); err != nil {
				log.Fatalf("Fatal: failed to seed bot_settings: %v", err)
			}
			fmt.Println("DB: bot_settings seeded successfully.")
		} else {
			log.Fatalf("Fatal: bot_settings row not found and no DISCORD_* env vars set for seeding.\n" +
				"Set DISCORD_TOKEN, DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET, DISCORD_REDIRECT_URI\n" +
				"in your environment for the first boot.")
		}
	}

	// 6. Load the Discord bot token from DB
	discordToken, _, _, _, err := bootSvc.GetDecryptedCredentials(ctx)
	if err != nil {
		log.Fatalf("Fatal: failed to load Discord token from bot_settings: %v", err)
	}

	// 7. Initialize DiscordGo Session
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatalf("Fatal: failed to construct Discord session: %v", err)
	}
	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentMessageContent

	// 8. Build Bot struct early so we can pass bot.Reconnect as a dependency to the settings module.
	//    Handlers are wired in step 10 after all modules are built.
	discordBot := &bot.Bot{Session: dg}

	// 9. Build Business Feature Modules
	modules := app.Build(
		db,
		dg,
		[]byte(cfg.JWTSecret),
		cfg.TokenEncryptionKey,
		time.Duration(cfg.CacheTTLSeconds)*time.Second,
		discordBot.Reconnect,
	)

	// 10. Perform automatic database migrations for all remaining models
	fmt.Println("DB: Running migrations/schema auto-sync...")
	if err := db.AutoMigrate(modules.Models()...); err != nil {
		log.Fatalf("Fatal: failed to auto-migrate database schema: %v", err)
	}

	// 11. Wire bot handler context and dev guild
	handlerCtx := handlers.NewHandlerContext(modules)
	discordBot.Ctx = handlerCtx
	discordBot.DevGuildID = os.Getenv("DEV_GUILD_ID")

	// 12. Initialize REST API Server
	apiServer := api.NewServer(
		cfg.APIPort,
		[]byte(cfg.JWTSecret),
		cfg.TokenEncryptionKey,
		modules,
		dg,
	)
	apiServer.SetupRoutes(cfg.CORSAllowedOrigins)

	// 13. Start services
	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start auto-close background worker
	go modules.Ticket.SchedulerSvc.Start(shutdownCtx, dg)

	// Start HTTP server FIRST — Railway health checks hit /health immediately after deploy
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("Fatal: REST API server crashed: %v", err)
		}
	}()

	// Connect to Discord Gateway
	fmt.Println("Bot: Connecting to Discord Gateway...")
	if err := discordBot.Start(); err != nil {
		log.Fatalf("Fatal: failed to start Discord bot: %v", err)
	}

	fmt.Println("Shrimpy Backend is now fully operational! 🦐")

	// 14. Graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("\nShrimpy Backend shutting down gracefully...")
	cancel()

	if err := discordBot.Stop(); err != nil {
		fmt.Printf("Error during bot shutdown: %v\n", err)
	}

	fmt.Println("Shrimpy Backend successfully stopped. Goodbye! 🦐")
}
