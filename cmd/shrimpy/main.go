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
	settings_repo "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/repository"
	settings_svc "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/bot"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/bot/handlers"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/config"
	"github.com/bwmarrin/discordgo"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: %v", err)
	}

	// 2b. Run database schema migrations using golang-migrate
	if err := runMigrations(cfg); err != nil {
		log.Fatalf("Fatal: %v", err)
	}

	// 4. Instantiate Registry with nil handlers first to break the circular dependency.
	registry := bot.NewRegistry(db, os.Getenv("DEV_GUILD_ID"), nil)

	// 5. Bootstrap settings repo/service
	bootRepo := settings_repo.NewSettingsRepo(db)
	bootSvc := settings_svc.NewSettingsService(bootRepo, cfg.TokenEncryptionKey, registry)

	// 6. Seed discord_apps from env vars if the count is 0 (first boot)
	ctx := context.Background()
	if err := seedDiscordApps(ctx, cfg, bootRepo, bootSvc); err != nil {
		log.Fatalf("Fatal: %v", err)
	}

	// 7. Build Business Feature Modules using Registry as Provider and Controller
	modules := app.Build(
		db,
		registry, // provider
		registry, // controller
		[]byte(cfg.JWTSecret),
		cfg.TokenEncryptionKey,
		time.Duration(cfg.CacheTTLSeconds)*time.Second,
	)

	// 9. Wire bot handler context and register handlers post-instantiation
	handlerCtx := handlers.NewHandlerContext(modules)
	registerHandlers := func(s *discordgo.Session) {
		s.AddHandler(handlerCtx.OnReady)
		s.AddHandler(handlerCtx.OnGuildCreate)
		s.AddHandler(handlerCtx.OnGuildDelete)
		s.AddHandler(handlerCtx.OnGuildMemberAdd)
		s.AddHandler(handlerCtx.OnMessageCreate)
		s.AddHandler(handlerCtx.OnMessageReactionAdd)
		s.AddHandler(handlerCtx.OnMessageReactionRemove)
		s.AddHandler(handlerCtx.OnInteractionCreate)
	}
	registry.SetHandlers(registerHandlers)

	// 10. Initialize REST API Server
	apiServer := api.NewServer(
		cfg.APIPort,
		[]byte(cfg.JWTSecret),
		cfg.TokenEncryptionKey,
		modules,
		registry,
	)
	apiServer.SetupRoutes(cfg.CORSAllowedOrigins)

	// 11. Start services
	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start auto-close background worker
	go modules.Ticket.SchedulerSvc.Start(shutdownCtx, registry)

	// Start HTTP server FIRST — Railway health checks hit /health immediately after deploy
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("Fatal: REST API server crashed: %v", err)
		}
	}()

	// Connect to Discord Gateway for all registered applications
	if err := startDiscordSessions(ctx, bootRepo, bootSvc, registry); err != nil {
		log.Fatalf("Fatal: %v", err)
	}

	fmt.Println("Shrimpy Backend is now fully operational! 🦐")

	// 12. Graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("\nShrimpy Backend shutting down gracefully...")
	cancel()

	registry.Clear()

	fmt.Println("Shrimpy Backend successfully stopped. Goodbye! 🦐")
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DatabaseURL,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}
	return db, nil
}

func runMigrations(cfg *config.Config) error {
	migrationURL := os.Getenv("DIRECT_DATABASE_URL")
	if migrationURL == "" {
		migrationURL = cfg.DatabaseURL
	}
	fmt.Println("DB: Running database migrations (golang-migrate)...")
	m, err := migrate.New("file://migrations", migrationURL)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply database migrations: %w", err)
	}
	fmt.Println("DB: Database migrations completed successfully.")
	return nil
}

func seedDiscordApps(ctx context.Context, cfg *config.Config, repo *settings_repo.SettingsRepo, svc *settings_svc.SettingsService) error {
	count, countErr := repo.Count(ctx)
	if countErr != nil {
		return countErr
	}

	if count == 0 {
		if cfg.HasDiscordSeed() {
			fmt.Println("DB: Seeding discord_apps from environment variables (first boot)...")
			if err := svc.SeedFromEnv(ctx,
				cfg.DiscordToken,
				cfg.DiscordClientID,
				cfg.DiscordClientSecret,
				cfg.DiscordRedirectURI,
			); err != nil {
				return fmt.Errorf("failed to seed discord_apps: %w", err)
			}
			fmt.Println("DB: discord_apps seeded successfully.")
		} else {
			return fmt.Errorf("no bot applications found in DB and no DISCORD_* env vars set for seeding.\n" +
				"Set DISCORD_TOKEN, DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET, DISCORD_REDIRECT_URI\n" +
				"in your environment for the first boot.")
		}
	} else if count == 1 {
		// If there is exactly one app, check if it was seeded with placeholder credentials and we have new ones in env vars.
		apps, err := repo.GetAll(ctx)
		if err == nil && len(apps) == 1 {
			app := apps[0]
			if app.DiscordClientID == "your_application_client_id" && cfg.DiscordClientID != "your_application_client_id" && cfg.DiscordClientID != "" {
				fmt.Println("DB: Upgrading placeholder discord_app with actual environment variables...")
				_, _, err = svc.Update(ctx, app.ID, settings_svc.UpdateRequest{
					Name:                "First Boot App",
					DiscordToken:        cfg.DiscordToken,
					DiscordClientID:     cfg.DiscordClientID,
					DiscordClientSecret: cfg.DiscordClientSecret,
					DiscordRedirectURI:  cfg.DiscordRedirectURI,
				})
				if err != nil {
					return fmt.Errorf("failed to upgrade placeholder discord_app: %w", err)
				}
				fmt.Println("DB: discord_app upgraded successfully.")
			}
		}
	}

	return nil
}

func startDiscordSessions(ctx context.Context, repo *settings_repo.SettingsRepo, svc *settings_svc.SettingsService, registry *bot.Registry) error {
	fmt.Println("Bot: Starting session gateways for registered applications...")
	apps, err := repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load bot applications from DB: %w", err)
	}
	for _, app := range apps {
		decryptedToken, _, _, _, err := svc.GetDecryptedCredentials(ctx, app.ID)
		if err != nil {
			fmt.Printf("Warning: failed to decrypt token for app %s (%s): %v\n", app.Name, app.ID, err)
			continue
		}
		if err := registry.StartSession(app.ID, decryptedToken); err != nil {
			fmt.Printf("Warning: failed to start session gateway for app %s (%s): %v\n", app.Name, app.ID, err)
		}
	}
	return nil
}
