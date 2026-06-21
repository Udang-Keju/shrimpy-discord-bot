package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotFound is returned when the bot_settings row does not exist yet.
var ErrNotFound = errors.New("bot_settings: record not found")

// SettingsRepo handles database operations for the bot_settings singleton table.
type SettingsRepo struct{ db *gorm.DB }

// NewSettingsRepo creates a new concrete SettingsRepo instance.
func NewSettingsRepo(db *gorm.DB) *SettingsRepo { return &SettingsRepo{db: db} }

// Get returns the singleton bot_settings row.
func (r *SettingsRepo) Get(ctx context.Context) (*model.BotSettings, error) {
	var s model.BotSettings
	result := r.db.WithContext(ctx).First(&s, "id = ?", 1)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &s, result.Error
}

// Upsert inserts or replaces the singleton bot_settings row (id always = 1).
func (r *SettingsRepo) Upsert(ctx context.Context, s *model.BotSettings) error {
	s.ID = 1 // enforce singleton constraint
	s.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"discord_token_enc":          s.DiscordTokenEnc,
				"discord_client_id":          s.DiscordClientID,
				"discord_client_secret_enc":  s.DiscordClientSecretEnc,
				"discord_redirect_uri":       s.DiscordRedirectURI,
				"updated_at":                 s.UpdatedAt,
			}),
		}).
		Create(s).Error
}
