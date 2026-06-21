package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotFound is returned when a user record is not found.
var ErrNotFound = errors.New("record not found")

// UserRepo handles database operations for the users table.
type UserRepo struct{ db *gorm.DB }

// NewUserRepo creates a new concrete UserRepo instance.
func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

// Upsert inserts or updates a user's public Discord profile.
func (r *UserRepo) Upsert(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"username":      u.Username,
				"discriminator": u.Discriminator,
				"avatar_hash":   u.AvatarHash,
				"last_seen":     time.Now(),
			}),
		}).
		Create(u).Error
}

// UpdateTokens persists the AES-256-GCM–encrypted OAuth2 tokens.
func (r *UserRepo) UpdateTokens(ctx context.Context, userID int64, accessEnc, refreshEnc []byte, expiresAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"discord_access_token_enc":  accessEnc,
			"discord_refresh_token_enc": refreshEnc,
			"token_expires_at":          expiresAt,
		}).Error
}

// GetByID returns a user record by their Discord ID.
func (r *UserRepo) GetByID(ctx context.Context, userID int64) (*model.User, error) {
	var u model.User
	result := r.db.WithContext(ctx).First(&u, "user_id = ?", userID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, result.Error
}
