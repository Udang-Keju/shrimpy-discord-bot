package main

import (
	"context"
	"fmt"
	"log"
	"os"

	settings_repo "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/repository"
	settings_svc "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: shrimpy-db <migrate-up|migrate-down|seed>")
	}

	cmd := os.Args[1]

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Fatal: failed to load config: %v", err)
	}

	switch cmd {
	case "migrate-up":
		if err := runMigrationsUp(cfg); err != nil {
			log.Fatalf("Fatal: %v", err)
		}
	case "migrate-down":
		if err := runMigrationsDown(cfg); err != nil {
			log.Fatalf("Fatal: %v", err)
		}
	case "seed":
		if err := seedDatabase(cfg); err != nil {
			log.Fatalf("Fatal: %v", err)
		}
	default:
		log.Fatalf("Unknown command: %s. Use migrate-up, migrate-down, or seed.", cmd)
	}
}

func runMigrationsUp(cfg *config.Config) error {
	migrationURL := os.Getenv("DIRECT_DATABASE_URL")
	if migrationURL == "" {
		migrationURL = cfg.DatabaseURL
	}
	fmt.Println("DB: Running database migrations up...")
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

func runMigrationsDown(cfg *config.Config) error {
	migrationURL := os.Getenv("DIRECT_DATABASE_URL")
	if migrationURL == "" {
		migrationURL = cfg.DatabaseURL
	}
	fmt.Println("DB: Rolling back database migrations...")
	m, err := migrate.New("file://migrations", migrationURL)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback database migrations: %w", err)
	}
	fmt.Println("DB: Database migrations rolled back successfully.")
	return nil
}

func seedDatabase(cfg *config.Config) error {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DatabaseURL,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	bootRepo := settings_repo.NewSettingsRepo(db)
	bootSvc := settings_svc.NewSettingsService(bootRepo, cfg.TokenEncryptionKey, nil)

	ctx := context.Background()
	count, err := bootRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count applications in database: %w", err)
	}

	if count > 0 {
		fmt.Println("DB: Database already seeded (discord_apps is not empty). Skipping.")
		return nil
	}

	if !cfg.HasDiscordSeed() {
		return fmt.Errorf("missing required environment variables for seeding (DISCORD_TOKEN, etc.)")
	}

	fmt.Println("DB: Seeding discord_apps from environment variables...")
	if err := bootSvc.SeedFromEnv(ctx,
		cfg.DiscordToken,
		cfg.DiscordClientID,
		cfg.DiscordClientSecret,
		cfg.DiscordRedirectURI,
	); err != nil {
		return fmt.Errorf("failed to seed discord_apps: %w", err)
	}

	fmt.Println("DB: Seeding completed successfully.")
	return nil
}
