package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Discord
	DiscordToken        string
	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURI  string

	// Database
	DatabaseURL string

	// Security
	JWTSecret          string
	TokenEncryptionKey []byte // 32 bytes for AES-256

	// API
	APIPort            string
	CORSAllowedOrigins string

	// Bot
	BotPrefix   string
	Environment string
	LogLevel    string

	// Performance
	CacheTTLSeconds        int
}

// Load reads configuration from environment variables.
// A .env file is loaded automatically in development.
func Load() (*Config, error) {
	// Load .env file if present — silently ignored in production
	_ = godotenv.Load()

	cacheTTL, _ := strconv.Atoi(getEnv("CACHE_TTL_SECONDS", "300"))

	encKeyHex := getEnv("TOKEN_ENCRYPTION_KEY", "")
	encKey, err := decodeHexKey(encKeyHex)
	if err != nil && encKeyHex != "" {
		return nil, fmt.Errorf("config: invalid TOKEN_ENCRYPTION_KEY: %w", err)
	}

	return &Config{
		DiscordToken:        mustGetEnv("DISCORD_TOKEN"),
		DiscordClientID:     mustGetEnv("DISCORD_CLIENT_ID"),
		DiscordClientSecret: mustGetEnv("DISCORD_CLIENT_SECRET"),
		DiscordRedirectURI:  mustGetEnv("DISCORD_REDIRECT_URI"),

		DatabaseURL: mustGetEnv("DATABASE_URL"),

		JWTSecret:          mustGetEnv("JWT_SECRET"),
		TokenEncryptionKey: encKey,

		APIPort:            getEnvFallback("PORT", "API_PORT", "8080"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),

		BotPrefix:   getEnv("BOT_PREFIX", "!"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		CacheTTLSeconds:        cacheTTL,
	}, nil
}

// IsDevelopment returns true when running in a non-production environment.
func (c *Config) IsDevelopment() bool {
	return c.Environment != "production"
}

// getEnv returns the environment variable value or a fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvFallback tries each key in order and returns the first non-empty value,
// falling back to the default if none are set. Used to handle Railway's PORT vs API_PORT.
func getEnvFallback(keys ...string) string {
	// Last element is the default value
	if len(keys) == 0 {
		return ""
	}
	defaultVal := keys[len(keys)-1]
	for _, key := range keys[:len(keys)-1] {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return defaultVal
}


// mustGetEnv panics if the required environment variable is not set.
func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable not set: %s", key))
	}
	return v
}

// decodeHexKey decodes a hex string into a 32-byte key.
func decodeHexKey(hex string) ([]byte, error) {
	if hex == "" {
		return nil, nil
	}
	if len(hex) != 64 {
		return nil, fmt.Errorf("key must be 64 hex characters (32 bytes), got %d", len(hex))
	}
	key := make([]byte, 32)
	for i := 0; i < 32; i++ {
		b, err := strconv.ParseUint(hex[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex at position %d: %w", i*2, err)
		}
		key[i] = byte(b)
	}
	return key, nil
}
