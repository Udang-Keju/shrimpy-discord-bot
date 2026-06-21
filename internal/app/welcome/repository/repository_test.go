package repository_test

import (
	"context"
	"testing"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/repository"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE guilds (
		guild_id INTEGER PRIMARY KEY
	)`).Error
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE welcome_config (
		guild_id INTEGER PRIMARY KEY,
		enabled BOOLEAN DEFAULT true,
		dm_message TEXT,
		channel_id INTEGER,
		channel_message TEXT,
		embed_color INTEGER,
		embed_media TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)

	return db
}

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestWelcomeRepo_GetUpsertDelete(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)

	t.Run("Get - Not Found", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewWelcomeRepo(db)

		fetched, err := repo.Get(ctx, guildID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
		assert.Nil(t, fetched)
	})

	t.Run("Full Configuration Lifecycle Flow", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewWelcomeRepo(db)

		// Create guild first for foreign key constraint if verified by driver,
		// though sqlite memory by default doesn't enforce FK unless PRAGMA foreign_keys = ON is executed.
		err := db.Exec("INSERT INTO guilds (guild_id) VALUES (?)", guildID).Error
		require.NoError(t, err)

		cfg := &model.WelcomeConfig{
			GuildID:        guildID,
			Enabled:        true,
			DMMessage:      stringPtr("Welcome!"),
			ChannelID:      int64Ptr(999888),
			ChannelMessage: stringPtr("Welcome to server!"),
		}

		// 1. Upsert initial configuration
		inserted, err := repo.Upsert(ctx, cfg)
		assert.NoError(t, err)
		assert.Equal(t, guildID, inserted.GuildID)

		fetched, err := repo.Get(ctx, guildID)
		assert.NoError(t, err)
		assert.Equal(t, "Welcome!", *fetched.DMMessage)
		assert.True(t, fetched.Enabled)

		// 2. Conflict Upsert (Update existing configuration details)
		cfg.DMMessage = stringPtr("New DM Welcome!")
		cfg.Enabled = false
		updated, err := repo.Upsert(ctx, cfg)
		assert.NoError(t, err)
		assert.False(t, updated.Enabled)

		fetched, err = repo.Get(ctx, guildID)
		assert.NoError(t, err)
		assert.Equal(t, "New DM Welcome!", *fetched.DMMessage)
		assert.False(t, fetched.Enabled)

		// 3. SetEnabled Status Toggle
		err = repo.SetEnabled(ctx, guildID, true)
		assert.NoError(t, err)

		fetched, err = repo.Get(ctx, guildID)
		assert.NoError(t, err)
		assert.True(t, fetched.Enabled)

		// 4. Delete Configuration
		err = repo.Delete(ctx, guildID)
		assert.NoError(t, err)

		fetched, err = repo.Get(ctx, guildID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
		assert.Nil(t, fetched)
	})
}
