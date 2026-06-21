package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds feature-specific settings for support tickets.
type Config struct {
	TranscriptWorkerCount  int
	AutoCloseCheckInterval time.Duration
}

// Load retrieves ticket settings from environment variables.
func Load() *Config {
	workers, err := strconv.Atoi(getEnv("TRANSCRIPT_WORKER_COUNT", "4"))
	if err != nil {
		workers = 4
	}
	interval, err := time.ParseDuration(getEnv("AUTO_CLOSE_CHECK_INTERVAL", "15m"))
	if err != nil {
		interval = 15 * time.Minute
	}
	return &Config{
		TranscriptWorkerCount:  workers,
		AutoCloseCheckInterval: interval,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
