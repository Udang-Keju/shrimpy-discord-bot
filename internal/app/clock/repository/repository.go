package repository

import (
	"context"
	"errors"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotFound is returned when a clock channel is not found in the database.
var ErrNotFound = errors.New("record not found")

// ClockRepo handles database operations for the clock_channels table.
type ClockRepo struct{ db *gorm.DB }

// NewClockRepo creates a new concrete ClockRepo instance.
func NewClockRepo(db *gorm.DB) *ClockRepo { return &ClockRepo{db: db} }

// List returns the clock channels configured for a guild.
func (r *ClockRepo) List(ctx context.Context, guildID int64) ([]model.ClockChannel, error) {
	var clocks []model.ClockChannel
	err := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Order("created_at ASC").
		Find(&clocks).Error
	return clocks, err
}

// ListAll returns every configured clock channel across all guilds, for the
// refresh scheduler.
func (r *ClockRepo) ListAll(ctx context.Context) ([]model.ClockChannel, error) {
	var clocks []model.ClockChannel
	err := r.db.WithContext(ctx).Find(&clocks).Error
	return clocks, err
}

// Get returns a single clock channel scoped to its guild.
func (r *ClockRepo) Get(ctx context.Context, guildID int64, id string) (*model.ClockChannel, error) {
	var c model.ClockChannel
	result := r.db.WithContext(ctx).First(&c, "guild_id = ? AND id = ?", guildID, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, result.Error
}

// Create inserts a new clock channel. A channel can only have one clock, so a
// duplicate (guild_id, channel_id) replaces the display settings of the
// existing row.
func (r *ClockRepo) Create(ctx context.Context, c *model.ClockChannel) (*model.ClockChannel, error) {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "guild_id"}, {Name: "channel_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"kind", "format", "timezone", "emoji", "refresh_minutes"}),
		}).
		Create(c)
	return c, result.Error
}

// Update saves the display settings of an existing clock channel.
func (r *ClockRepo) Update(ctx context.Context, c *model.ClockChannel) error {
	return r.db.WithContext(ctx).
		Model(&model.ClockChannel{}).
		Where("guild_id = ? AND id = ?", c.GuildID, c.ID).
		Select("kind", "format", "timezone", "emoji", "refresh_minutes").
		Updates(c).Error
}

// SetLastName records the channel name most recently written to Discord.
func (r *ClockRepo) SetLastName(ctx context.Context, id, name string) error {
	return r.db.WithContext(ctx).
		Model(&model.ClockChannel{}).
		Where("id = ?", id).
		Update("last_name", name).Error
}

// Delete removes a clock channel for a guild.
func (r *ClockRepo) Delete(ctx context.Context, guildID int64, id string) error {
	return r.db.WithContext(ctx).
		Where("guild_id = ? AND id = ?", guildID, id).
		Delete(&model.ClockChannel{}).Error
}
