package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth/repository"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE users (
		user_id INTEGER PRIMARY KEY,
		username TEXT NOT NULL,
		discriminator TEXT,
		avatar_hash TEXT,
		discord_access_token_enc BLOB,
		discord_refresh_token_enc BLOB,
		token_expires_at DATETIME,
		last_seen DATETIME,
		created_at DATETIME
	)`).Error
	require.NoError(t, err)

	return db
}

func stringPtr(s string) *string {
	return &s
}

func TestUserRepo_UpsertAndGetByID(t *testing.T) {
	ctx := context.Background()
	userID := int64(67890)

	t.Run("Upsert and Fetch Flow", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewUserRepo(db)

		u := &model.User{
			UserID:        userID,
			Username:      "oldname",
			Discriminator: stringPtr("0001"),
			AvatarHash:    stringPtr("avatar1"),
			LastSeen:      time.Now(),
		}

		// 1. Initial Upsert
		err := repo.Upsert(ctx, u)
		assert.NoError(t, err)

		fetched, err := repo.GetByID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, "oldname", fetched.Username)
		assert.Equal(t, "0001", *fetched.Discriminator)

		// 2. Conflicting Upsert (Updates profile details)
		u.Username = "newname"
		u.AvatarHash = stringPtr("avatar2")

		err = repo.Upsert(ctx, u)
		assert.NoError(t, err)

		fetched, err = repo.GetByID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, "newname", fetched.Username)
		assert.Equal(t, "avatar2", *fetched.AvatarHash)
	})

	t.Run("GetByID - Not Found", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewUserRepo(db)

		fetched, err := repo.GetByID(ctx, 99999)
		assert.ErrorIs(t, err, repository.ErrNotFound)
		assert.Nil(t, fetched)
	})
}

func TestUserRepo_UpdateTokens(t *testing.T) {
	ctx := context.Background()
	userID := int64(67890)
	db := setupTestDB(t)
	repo := repository.NewUserRepo(db)

	u := &model.User{
		UserID:   userID,
		Username: "user1",
		LastSeen: time.Now(),
	}
	err := repo.Upsert(ctx, u)
	require.NoError(t, err)

	accessEnc := []byte("access_encrypted")
	refreshEnc := []byte("refresh_encrypted")
	expiresAt := time.Now().Add(time.Hour).Truncate(time.Second)

	err = repo.UpdateTokens(ctx, userID, accessEnc, refreshEnc, expiresAt)
	assert.NoError(t, err)

	fetched, err := repo.GetByID(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, accessEnc, fetched.DiscordAccessTokenEnc)
	assert.Equal(t, refreshEnc, fetched.DiscordRefreshTokenEnc)
	assert.True(t, fetched.TokenExpiresAt.Equal(expiresAt))
}
