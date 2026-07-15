package repository

import (
	"context"
	"errors"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotFound is returned when a translation config is not found in database.
var ErrNotFound = errors.New("record not found")

// TranslationRepo handles database operations for the translation tables.
type TranslationRepo struct{ db *gorm.DB }

// NewTranslationRepo creates a new concrete TranslationRepo instance.
func NewTranslationRepo(db *gorm.DB) *TranslationRepo { return &TranslationRepo{db: db} }

// GetConfig returns the translation configuration for a guild.
func (r *TranslationRepo) GetConfig(ctx context.Context, guildID int64) (*model.TranslationConfig, error) {
	var cfg model.TranslationConfig
	result := r.db.WithContext(ctx).First(&cfg, "guild_id = ?", guildID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &cfg, result.Error
}

// UpsertConfig creates or updates a guild's translation configuration.
func (r *TranslationRepo) UpsertConfig(ctx context.Context, cfg *model.TranslationConfig) (*model.TranslationConfig, error) {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "guild_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"enabled", "auto_enabled", "reaction_enabled",
				"provider", "api_key_enc", "endpoint_url", "target_lang",
			}),
		}).
		Create(cfg)
	return cfg, result.Error
}

// SetEnabled toggles the master translation switch without touching other settings.
func (r *TranslationRepo) SetEnabled(ctx context.Context, guildID int64, enabled bool) error {
	return r.db.WithContext(ctx).
		Model(&model.TranslationConfig{}).
		Where("guild_id = ?", guildID).
		Update("enabled", enabled).Error
}

// ListChannels returns the configured auto-translate channels for a guild.
func (r *TranslationRepo) ListChannels(ctx context.Context, guildID int64) ([]model.TranslationChannel, error) {
	var channels []model.TranslationChannel
	err := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Order("created_at ASC").
		Find(&channels).Error
	return channels, err
}

// AddChannel adds (or updates the override of) an auto-translate channel.
func (r *TranslationRepo) AddChannel(ctx context.Context, ch *model.TranslationChannel) (*model.TranslationChannel, error) {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "guild_id"}, {Name: "channel_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"target_lang_override"}),
		}).
		Create(ch)
	return ch, result.Error
}

// RemoveChannel deletes an auto-translate channel for a guild.
func (r *TranslationRepo) RemoveChannel(ctx context.Context, guildID, channelID int64) error {
	return r.db.WithContext(ctx).
		Where("guild_id = ? AND channel_id = ?", guildID, channelID).
		Delete(&model.TranslationChannel{}).Error
}

// ListEmojis returns the configured trigger emojis for a guild.
func (r *TranslationRepo) ListEmojis(ctx context.Context, guildID int64) ([]model.TranslationReactionEmoji, error) {
	var emojis []model.TranslationReactionEmoji
	err := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Order("created_at ASC").
		Find(&emojis).Error
	return emojis, err
}

// AddEmoji adds (or updates the override of) a trigger emoji.
func (r *TranslationRepo) AddEmoji(ctx context.Context, e *model.TranslationReactionEmoji) (*model.TranslationReactionEmoji, error) {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "guild_id"}, {Name: "emoji"}},
			DoUpdates: clause.AssignmentColumns([]string{"target_lang_override"}),
		}).
		Create(e)
	return e, result.Error
}

// RemoveEmoji deletes a trigger emoji for a guild.
func (r *TranslationRepo) RemoveEmoji(ctx context.Context, guildID int64, emoji string) error {
	return r.db.WithContext(ctx).
		Where("guild_id = ? AND emoji = ?", guildID, emoji).
		Delete(&model.TranslationReactionEmoji{}).Error
}
