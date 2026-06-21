package repository

import (
	"context"
	"errors"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/model"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a requested Discord application is not found.
var ErrNotFound = errors.New("settings: app not found")

// SettingsRepo handles database operations for the discord_apps table.
type SettingsRepo struct{ db *gorm.DB }

// NewSettingsRepo creates a new concrete SettingsRepo instance.
func NewSettingsRepo(db *gorm.DB) *SettingsRepo { return &SettingsRepo{db: db} }

// GetAll returns all configured Discord applications.
func (r *SettingsRepo) GetAll(ctx context.Context) ([]model.DiscordApp, error) {
	var apps []model.DiscordApp
	err := r.db.WithContext(ctx).Order("created_at").Find(&apps).Error
	return apps, err
}

// GetByID fetches a Discord application by its database ID (UUID).
func (r *SettingsRepo) GetByID(ctx context.Context, id string) (*model.DiscordApp, error) {
	var app model.DiscordApp
	err := r.db.WithContext(ctx).First(&app, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &app, err
}

// GetByClientID fetches a Discord application by its client ID.
func (r *SettingsRepo) GetByClientID(ctx context.Context, clientID string) (*model.DiscordApp, error) {
	var app model.DiscordApp
	err := r.db.WithContext(ctx).First(&app, "discord_client_id = ?", clientID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &app, err
}

// Create inserts a new Discord application record.
func (r *SettingsRepo) Create(ctx context.Context, app *model.DiscordApp) error {
	return r.db.WithContext(ctx).Create(app).Error
}

// Update saves an existing Discord application record.
func (r *SettingsRepo) Update(ctx context.Context, app *model.DiscordApp) error {
	return r.db.WithContext(ctx).Save(app).Error
}

// Delete removes a Discord application record by ID.
func (r *SettingsRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.DiscordApp{}, "id = ?", id).Error
}

// Count returns the total number of registered bot applications.
func (r *SettingsRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.DiscordApp{}).Count(&count).Error
	return count, err
}
