package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/crypto"
)

// BotSettingsDTO is the public-facing shape returned by the API.
// Sensitive values (token, client secret) are always masked.
type BotSettingsDTO struct {
	DiscordToken        string    `json:"discord_token"` // always "***"
	DiscordClientID     string    `json:"discord_client_id"`
	DiscordClientSecret string    `json:"discord_client_secret"` // always "***"
	DiscordRedirectURI  string    `json:"discord_redirect_uri"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// UpdateRequest carries the fields a caller wants to change.
// Empty string means "keep existing value".
type UpdateRequest struct {
	DiscordToken        string `json:"discord_token"`
	DiscordClientID     string `json:"discord_client_id"`
	DiscordClientSecret string `json:"discord_client_secret"`
	DiscordRedirectURI  string `json:"discord_redirect_uri"`
}

// SettingsRepo defines the DB operations needed by SettingsService.
type SettingsRepo interface {
	Get(ctx context.Context) (*model.BotSettings, error)
	Upsert(ctx context.Context, s *model.BotSettings) error
}

// cachedSecrets is a short-lived in-process cache for decrypted credentials,
// to avoid a DB round-trip on every OAuth2 callback.
type cachedSecrets struct {
	token        string
	clientSecret string
	redirectURI  string
	clientID     string
	expiresAt    time.Time
}

// SettingsService provides business logic for bot credential management.
type SettingsService struct {
	repo        SettingsRepo
	tokenEncKey []byte

	mu    sync.RWMutex
	cache *cachedSecrets
}

// NewSettingsService creates a new SettingsService.
func NewSettingsService(repo SettingsRepo, tokenEncKey []byte) *SettingsService {
	return &SettingsService{
		repo:        repo,
		tokenEncKey: tokenEncKey,
	}
}

// GetBotSettings returns the current settings with all secrets masked.
func (s *SettingsService) GetBotSettings(ctx context.Context) (*BotSettingsDTO, error) {
	row, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &BotSettingsDTO{
		DiscordToken:        "***",
		DiscordClientID:     row.DiscordClientID,
		DiscordClientSecret: "***",
		DiscordRedirectURI:  row.DiscordRedirectURI,
		UpdatedAt:           row.UpdatedAt,
	}, nil
}

// GetDecryptedCredentials returns all plaintext credentials (used internally at startup/reconnect).
// Results are cached for 30 seconds to reduce DB load during busy OAuth2 callback periods.
func (s *SettingsService) GetDecryptedCredentials(ctx context.Context) (token, clientID, clientSecret, redirectURI string, err error) {
	s.mu.RLock()
	if s.cache != nil && time.Now().Before(s.cache.expiresAt) {
		c := s.cache
		s.mu.RUnlock()
		return c.token, c.clientID, c.clientSecret, c.redirectURI, nil
	}
	s.mu.RUnlock()

	// Cache miss — fetch from DB
	row, err := s.repo.Get(ctx)
	if err != nil {
		return "", "", "", "", fmt.Errorf("settings: failed to load credentials: %w", err)
	}

	plainToken, err := crypto.Decrypt(row.DiscordTokenEnc, s.tokenEncKey)
	if err != nil {
		return "", "", "", "", fmt.Errorf("settings: failed to decrypt token: %w", err)
	}

	plainSecret, err := crypto.Decrypt(row.DiscordClientSecretEnc, s.tokenEncKey)
	if err != nil {
		return "", "", "", "", fmt.Errorf("settings: failed to decrypt client secret: %w", err)
	}

	s.mu.Lock()
	s.cache = &cachedSecrets{
		token:        string(plainToken),
		clientID:     row.DiscordClientID,
		clientSecret: string(plainSecret),
		redirectURI:  row.DiscordRedirectURI,
		expiresAt:    time.Now().Add(30 * time.Second),
	}
	s.mu.Unlock()

	return string(plainToken), row.DiscordClientID, string(plainSecret), row.DiscordRedirectURI, nil
}

// UpdateBotSettings saves new credential values. Any field left empty keeps its existing value.
// Returns (tokenChanged bool, newToken string, error).
func (s *SettingsService) UpdateBotSettings(ctx context.Context, req UpdateRequest) (tokenChanged bool, newToken string, err error) {
	// Load existing row so we only update the fields that changed
	existing, err := s.repo.Get(ctx)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return false, "", fmt.Errorf("settings: failed to load existing settings: %w", err)
	}
	if existing == nil {
		existing = &model.BotSettings{}
	}

	updated := &model.BotSettings{
		DiscordTokenEnc:        existing.DiscordTokenEnc,
		DiscordClientID:        existing.DiscordClientID,
		DiscordClientSecretEnc: existing.DiscordClientSecretEnc,
		DiscordRedirectURI:     existing.DiscordRedirectURI,
	}

	if req.DiscordToken != "" {
		enc, encErr := crypto.Encrypt([]byte(req.DiscordToken), s.tokenEncKey)
		if encErr != nil {
			return false, "", fmt.Errorf("settings: failed to encrypt token: %w", encErr)
		}
		updated.DiscordTokenEnc = enc
		tokenChanged = true
		newToken = req.DiscordToken
	}

	if req.DiscordClientID != "" {
		updated.DiscordClientID = req.DiscordClientID
	}

	if req.DiscordClientSecret != "" {
		enc, encErr := crypto.Encrypt([]byte(req.DiscordClientSecret), s.tokenEncKey)
		if encErr != nil {
			return false, "", fmt.Errorf("settings: failed to encrypt client secret: %w", encErr)
		}
		updated.DiscordClientSecretEnc = enc
	}

	if req.DiscordRedirectURI != "" {
		updated.DiscordRedirectURI = req.DiscordRedirectURI
	}

	if err := s.repo.Upsert(ctx, updated); err != nil {
		return false, "", fmt.Errorf("settings: failed to save settings: %w", err)
	}

	// Invalidate the credentials cache so next request fetches fresh values
	s.mu.Lock()
	s.cache = nil
	s.mu.Unlock()

	return tokenChanged, newToken, nil
}

// SeedFromEnv saves initial credentials from environment variable values.
// Should only be called on first boot when the row does not yet exist.
func (s *SettingsService) SeedFromEnv(ctx context.Context, token, clientID, clientSecret, redirectURI string) error {
	encToken, err := crypto.Encrypt([]byte(token), s.tokenEncKey)
	if err != nil {
		return fmt.Errorf("settings: seed encrypt token: %w", err)
	}
	encSecret, err := crypto.Encrypt([]byte(clientSecret), s.tokenEncKey)
	if err != nil {
		return fmt.Errorf("settings: seed encrypt secret: %w", err)
	}
	return s.repo.Upsert(ctx, &model.BotSettings{
		DiscordTokenEnc:        encToken,
		DiscordClientID:        clientID,
		DiscordClientSecretEnc: encSecret,
		DiscordRedirectURI:     redirectURI,
	})
}
