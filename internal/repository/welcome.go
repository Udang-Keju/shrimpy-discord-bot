package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WelcomeRepo handles database operations for the welcome_config table.
type WelcomeRepo struct{ db *gorm.DB }

func NewWelcomeRepo(db *gorm.DB) *WelcomeRepo { return &WelcomeRepo{db: db} }

// Get returns the welcome configuration for a guild.
func (r *WelcomeRepo) Get(ctx context.Context, guildID int64) (*WelcomeConfig, error) {
	var cfg WelcomeConfig
	result := r.db.WithContext(ctx).First(&cfg, "guild_id = ?", guildID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &cfg, result.Error
}

// Upsert creates or fully replaces a guild's welcome configuration.
func (r *WelcomeRepo) Upsert(ctx context.Context, cfg *WelcomeConfig) (*WelcomeConfig, error) {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "guild_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"enabled", "dm_message", "channel_id", "channel_message",
				"embed_color", "embed_media",
			}),
		}).
		Create(cfg)
	return cfg, result.Error
}

// SetEnabled toggles the welcome system without changing other settings.
func (r *WelcomeRepo) SetEnabled(ctx context.Context, guildID int64, enabled bool) error {
	return r.db.WithContext(ctx).
		Model(&WelcomeConfig{}).
		Where("guild_id = ?", guildID).
		Update("enabled", enabled).Error
}

// Delete removes the welcome configuration for a guild.
func (r *WelcomeRepo) Delete(ctx context.Context, guildID int64) error {
	return r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Delete(&WelcomeConfig{}).Error
}
