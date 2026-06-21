package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/crypto"
)

// DiscordAppDTO is the public representation of a bot application.
// Sensitive fields are masked.
type DiscordAppDTO struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	DiscordToken        string    `json:"discord_token"` // masked as "***"
	DiscordClientID     string    `json:"discord_client_id"`
	DiscordClientSecret string    `json:"discord_client_secret"` // masked as "***"
	DiscordRedirectURI  string    `json:"discord_redirect_uri"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// CreateRequest carries fields to add a new application.
type CreateRequest struct {
	Name                string `json:"name"`
	DiscordToken        string `json:"discord_token"`
	DiscordClientID     string `json:"discord_client_id"`
	DiscordClientSecret string `json:"discord_client_secret"`
	DiscordRedirectURI  string `json:"discord_redirect_uri"`
}

// UpdateRequest carries fields to modify an existing application.
type UpdateRequest struct {
	Name                string `json:"name"`
	DiscordToken        string `json:"discord_token"`
	DiscordClientID     string `json:"discord_client_id"`
	DiscordClientSecret string `json:"discord_client_secret"`
	DiscordRedirectURI  string `json:"discord_redirect_uri"`
}

// SettingsRepo defines DB operations for discord_apps.
type SettingsRepo interface {
	GetAll(ctx context.Context) ([]model.DiscordApp, error)
	GetByID(ctx context.Context, id string) (*model.DiscordApp, error)
	GetByClientID(ctx context.Context, clientID string) (*model.DiscordApp, error)
	Create(ctx context.Context, app *model.DiscordApp) error
	Update(ctx context.Context, app *model.DiscordApp) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

// BotSessionController triggers session connection operations in the gateway registry.
type BotSessionController interface {
	StartSession(appID string, token string) error
	StopSession(appID string) error
}

type cachedSecrets struct {
	token        string
	clientSecret string
	redirectURI  string
	clientID     string
	expiresAt    time.Time
}

// SettingsService manages bot application configurations.
type SettingsService struct {
	repo        SettingsRepo
	tokenEncKey []byte
	controller  BotSessionController

	mu    sync.RWMutex
	cache map[string]cachedSecrets
}

// NewSettingsService constructs a new SettingsService.
func NewSettingsService(
	repo SettingsRepo,
	tokenEncKey []byte,
	controller BotSessionController,
) *SettingsService {
	return &SettingsService{
		repo:        repo,
		tokenEncKey: tokenEncKey,
		controller:  controller,
		cache:       make(map[string]cachedSecrets),
	}
}

// List returns all configured bot applications.
func (s *SettingsService) List(ctx context.Context) ([]DiscordAppDTO, error) {
	rows, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]DiscordAppDTO, len(rows))
	for i, row := range rows {
		dtos[i] = DiscordAppDTO{
			ID:                  row.ID,
			Name:                row.Name,
			DiscordToken:        "***",
			DiscordClientID:     row.DiscordClientID,
			DiscordClientSecret: "***",
			DiscordRedirectURI:  row.DiscordRedirectURI,
			CreatedAt:           row.CreatedAt,
			UpdatedAt:           row.UpdatedAt,
		}
	}
	return dtos, nil
}

// GetByID returns a single app DTO by ID.
func (s *SettingsService) GetByID(ctx context.Context, id string) (*DiscordAppDTO, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &DiscordAppDTO{
		ID:                  row.ID,
		Name:                row.Name,
		DiscordToken:        "***",
		DiscordClientID:     row.DiscordClientID,
		DiscordClientSecret: "***",
		DiscordRedirectURI:  row.DiscordRedirectURI,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}, nil
}

// Create registers a new bot application and starts its Discord connection.
func (s *SettingsService) Create(ctx context.Context, req CreateRequest) (*DiscordAppDTO, error) {
	if req.Name == "" || req.DiscordToken == "" || req.DiscordClientID == "" || req.DiscordClientSecret == "" || req.DiscordRedirectURI == "" {
		return nil, errors.New("settings: missing required fields")
	}

	encToken, err := crypto.Encrypt([]byte(req.DiscordToken), s.tokenEncKey)
	if err != nil {
		return nil, fmt.Errorf("settings: encrypt token: %w", err)
	}

	encSecret, err := crypto.Encrypt([]byte(req.DiscordClientSecret), s.tokenEncKey)
	if err != nil {
		return nil, fmt.Errorf("settings: encrypt client secret: %w", err)
	}

	app := &model.DiscordApp{
		Name:                   req.Name,
		DiscordTokenEnc:        encToken,
		DiscordClientID:        req.DiscordClientID,
		DiscordClientSecretEnc: encSecret,
		DiscordRedirectURI:     req.DiscordRedirectURI,
	}

	if err := s.repo.Create(ctx, app); err != nil {
		return nil, err
	}

	// Spin up dynamic bot connection
	if s.controller != nil {
		if err := s.controller.StartSession(app.ID, req.DiscordToken); err != nil {
			fmt.Printf("settings error: failed to start bot session during creation: %v\n", err)
		}
	}

	return &DiscordAppDTO{
		ID:                  app.ID,
		Name:                app.Name,
		DiscordToken:        "***",
		DiscordClientID:     app.DiscordClientID,
		DiscordClientSecret: "***",
		DiscordRedirectURI:  app.DiscordRedirectURI,
		CreatedAt:           app.CreatedAt,
		UpdatedAt:           app.UpdatedAt,
	}, nil
}

// Update modifies an existing bot application.
func (s *SettingsService) Update(ctx context.Context, id string, req UpdateRequest) (tokenChanged bool, newToken string, err error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, "", err
	}

	if req.Name != "" {
		existing.Name = req.Name
	}

	if req.DiscordToken != "" {
		enc, encErr := crypto.Encrypt([]byte(req.DiscordToken), s.tokenEncKey)
		if encErr != nil {
			return false, "", fmt.Errorf("settings: encrypt token: %w", encErr)
		}
		existing.DiscordTokenEnc = enc
		tokenChanged = true
		newToken = req.DiscordToken
	}

	if req.DiscordClientID != "" {
		existing.DiscordClientID = req.DiscordClientID
	}

	if req.DiscordClientSecret != "" {
		enc, encErr := crypto.Encrypt([]byte(req.DiscordClientSecret), s.tokenEncKey)
		if encErr != nil {
			return false, "", fmt.Errorf("settings: encrypt secret: %w", encErr)
		}
		existing.DiscordClientSecretEnc = enc
	}

	if req.DiscordRedirectURI != "" {
		existing.DiscordRedirectURI = req.DiscordRedirectURI
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return false, "", err
	}

	// Invalidate cache
	s.mu.Lock()
	delete(s.cache, id)
	s.mu.Unlock()

	// If the token changed, restart the dynamic session
	if tokenChanged && s.controller != nil {
		if err := s.controller.StartSession(existing.ID, newToken); err != nil {
			return true, newToken, fmt.Errorf("settings: failed to restart session: %w", err)
		}
	}

	return tokenChanged, newToken, nil
}

// Delete removes an application and terminates its bot gateway connection.
func (s *SettingsService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Terminate gateway session
	if s.controller != nil {
		_ = s.controller.StopSession(id)
	}

	s.mu.Lock()
	delete(s.cache, id)
	s.mu.Unlock()

	return nil
}

// GetDecryptedCredentials returns all plaintext credentials for an app ID (UUID).
func (s *SettingsService) GetDecryptedCredentials(ctx context.Context, id string) (token, clientID, clientSecret, redirectURI string, err error) {
	s.mu.RLock()
	if c, found := s.cache[id]; found && time.Now().Before(c.expiresAt) {
		s.mu.RUnlock()
		return c.token, c.clientID, c.clientSecret, c.redirectURI, nil
	}
	s.mu.RUnlock()

	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", "", "", "", err
	}

	plainToken, err := crypto.Decrypt(row.DiscordTokenEnc, s.tokenEncKey)
	if err != nil {
		return "", "", "", "", fmt.Errorf("settings: decrypt token: %w", err)
	}

	plainSecret, err := crypto.Decrypt(row.DiscordClientSecretEnc, s.tokenEncKey)
	if err != nil {
		return "", "", "", "", fmt.Errorf("settings: decrypt secret: %w", err)
	}

	s.mu.Lock()
	s.cache[id] = cachedSecrets{
		token:        string(plainToken),
		clientID:     row.DiscordClientID,
		clientSecret: string(plainSecret),
		redirectURI:  row.DiscordRedirectURI,
		expiresAt:    time.Now().Add(30 * time.Second),
	}
	s.mu.Unlock()

	return string(plainToken), row.DiscordClientID, string(plainSecret), row.DiscordRedirectURI, nil
}

// SeedFromEnv seeds a default application on first boot if no applications exist in the database.
func (s *SettingsService) SeedFromEnv(ctx context.Context, token, clientID, clientSecret, redirectURI string) error {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already seeded
	}

	encToken, err := crypto.Encrypt([]byte(token), s.tokenEncKey)
	if err != nil {
		return err
	}

	encSecret, err := crypto.Encrypt([]byte(clientSecret), s.tokenEncKey)
	if err != nil {
		return err
	}

	app := &model.DiscordApp{
		Name:                   "First Boot App",
		DiscordTokenEnc:        encToken,
		DiscordClientID:        clientID,
		DiscordClientSecretEnc: encSecret,
		DiscordRedirectURI:     redirectURI,
	}

	return s.repo.Create(ctx, app)
}

// Reconnect forces the controller to re-establish the connection.
func (s *SettingsService) Reconnect(ctx context.Context, id string) error {
	token, _, _, _, err := s.GetDecryptedCredentials(ctx, id)
	if err != nil {
		return err
	}

	if s.controller != nil {
		return s.controller.StartSession(id, token)
	}
	return nil
}
