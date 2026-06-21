package repository_test

import (
	"context"
	"testing"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/repository"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE discord_apps (
		id TEXT PRIMARY KEY,
		name TEXT,
		discord_token_enc BLOB,
		discord_client_id TEXT UNIQUE,
		discord_client_secret_enc BLOB,
		discord_redirect_uri TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE guilds (
		guild_id INTEGER PRIMARY KEY,
		discord_app_id TEXT,
		prefix TEXT DEFAULT '!',
		language TEXT DEFAULT 'en',
		bot_nickname TEXT,
		log_channel_id INTEGER,
		is_active BOOLEAN DEFAULT true,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE staff_roles (
		id TEXT PRIMARY KEY,
		guild_id INTEGER NOT NULL,
		role_id INTEGER NOT NULL,
		created_at DATETIME,
		UNIQUE(guild_id, role_id)
	)`).Error
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE auto_roles (
		id TEXT PRIMARY KEY,
		guild_id INTEGER NOT NULL,
		role_id INTEGER NOT NULL,
		created_at DATETIME,
		UNIQUE(guild_id, role_id)
	)`).Error
	require.NoError(t, err)

	return db
}

func TestGuildRepo_UpsertAndGetByID(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)

	t.Run("Upsert and GetByID Flow", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewGuildRepo(db)

		// 1. Initial Upsert
		g, err := repo.Upsert(ctx, guildID, nil)
		assert.NoError(t, err)
		assert.Equal(t, guildID, g.GuildID)
		assert.True(t, g.IsActive)

		// 2. Fetch and verify
		fetched, err := repo.GetByID(ctx, guildID)
		assert.NoError(t, err)
		assert.Equal(t, guildID, fetched.GuildID)
		assert.True(t, fetched.IsActive)

		// 3. Deactivate
		err = repo.Deactivate(ctx, guildID)
		assert.NoError(t, err)

		fetched, err = repo.GetByID(ctx, guildID)
		assert.NoError(t, err)
		assert.False(t, fetched.IsActive)

		// 4. Re-activate via Upsert (On Conflict)
		g, err = repo.Upsert(ctx, guildID, nil)
		assert.NoError(t, err)
		assert.True(t, g.IsActive)

		fetched, err = repo.GetByID(ctx, guildID)
		assert.NoError(t, err)
		assert.True(t, fetched.IsActive)
	})

	t.Run("GetByID - Not Found", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewGuildRepo(db)

		fetched, err := repo.GetByID(ctx, 99999)
		assert.ErrorIs(t, err, repository.ErrNotFound)
		assert.Nil(t, fetched)
	})
}

func TestGuildRepo_Update(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	db := setupTestDB(t)
	repo := repository.NewGuildRepo(db)

	_, err := repo.Upsert(ctx, guildID, nil)
	require.NoError(t, err)

	nickname := "MyBotNick"
	logChanID := int64(888999)

	updates := map[string]interface{}{
		"bot_nickname":   &nickname,
		"log_channel_id": &logChanID,
	}

	updated, err := repo.Update(ctx, guildID, updates)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, nickname, *updated.BotNickname)
	assert.Equal(t, logChanID, *updated.LogChannelID)

	// Fetch back to verify persistence
	fetched, err := repo.GetByID(ctx, guildID)
	assert.NoError(t, err)
	assert.Equal(t, nickname, *fetched.BotNickname)
	assert.Equal(t, logChanID, *fetched.LogChannelID)
}

func TestGuildRepo_GetAppIDByClientID(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := repository.NewGuildRepo(db)

	// Insert test app
	err := db.Exec(`INSERT INTO discord_apps (id, name, discord_client_id) VALUES ('app-uuid-123', 'Test App', 'client-id-123')`).Error
	require.NoError(t, err)

	appID, err := repo.GetAppIDByClientID(ctx, "client-id-123")
	assert.NoError(t, err)
	assert.Equal(t, "app-uuid-123", appID)

	// Query for non-existent client ID
	appID, err = repo.GetAppIDByClientID(ctx, "non-existent")
	assert.NoError(t, err)
	assert.Empty(t, appID)
}

func TestGuildRepo_StaffRoles(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)

	tests := []struct {
		name string
		run  func(t *testing.T, repo *repository.GuildRepo)
	}{
		{
			name: "Add and List Staff Roles",
			run: func(t *testing.T, repo *repository.GuildRepo) {
				role1, err := repo.AddStaffRole(ctx, guildID, 111)
				assert.NoError(t, err)
				assert.Equal(t, int64(111), role1.RoleID)

				role2, err := repo.AddStaffRole(ctx, guildID, 222)
				assert.NoError(t, err)

				// Verify Idempotent adding does not create duplicate
				role2Dup, err := repo.AddStaffRole(ctx, guildID, 222)
				assert.NoError(t, err)
				assert.Equal(t, role2.ID, role2Dup.ID)

				roles, err := repo.ListStaffRoles(ctx, guildID)
				assert.NoError(t, err)
				assert.Len(t, roles, 2)
				assert.Equal(t, int64(111), roles[0].RoleID)
				assert.Equal(t, int64(222), roles[1].RoleID)
			},
		},
		{
			name: "IsStaffRole Check",
			run: func(t *testing.T, repo *repository.GuildRepo) {
				_, _ = repo.AddStaffRole(ctx, guildID, 111)

				// Match exists
				ok, err := repo.IsStaffRole(ctx, guildID, []int64{111, 333})
				assert.NoError(t, err)
				assert.True(t, ok)

				// No match
				ok, err = repo.IsStaffRole(ctx, guildID, []int64{222, 333})
				assert.NoError(t, err)
				assert.False(t, ok)
			},
		},
		{
			name: "Remove Staff Role",
			run: func(t *testing.T, repo *repository.GuildRepo) {
				_, _ = repo.AddStaffRole(ctx, guildID, 111)
				_, _ = repo.AddStaffRole(ctx, guildID, 222)

				err := repo.RemoveStaffRole(ctx, guildID, 111)
				assert.NoError(t, err)

				roles, err := repo.ListStaffRoles(ctx, guildID)
				assert.NoError(t, err)
				assert.Len(t, roles, 1)
				assert.Equal(t, int64(222), roles[0].RoleID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := repository.NewGuildRepo(db)
			tt.run(t, repo)
		})
	}
}

func TestGuildRepo_AutoRoles(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)

	tests := []struct {
		name string
		run  func(t *testing.T, repo *repository.GuildRepo)
	}{
		{
			name: "Add and List Auto Roles",
			run: func(t *testing.T, repo *repository.GuildRepo) {
				role1, err := repo.AddAutoRole(ctx, guildID, 111)
				assert.NoError(t, err)
				assert.Equal(t, int64(111), role1.RoleID)

				role2, err := repo.AddAutoRole(ctx, guildID, 222)
				assert.NoError(t, err)

				// Verify Idempotency
				role2Dup, err := repo.AddAutoRole(ctx, guildID, 222)
				assert.NoError(t, err)
				assert.Equal(t, role2.ID, role2Dup.ID)

				roles, err := repo.ListAutoRoles(ctx, guildID)
				assert.NoError(t, err)
				assert.Len(t, roles, 2)
				assert.Equal(t, int64(111), roles[0].RoleID)
				assert.Equal(t, int64(222), roles[1].RoleID)
			},
		},
		{
			name: "Remove Auto Role",
			run: func(t *testing.T, repo *repository.GuildRepo) {
				_, _ = repo.AddAutoRole(ctx, guildID, 111)
				_, _ = repo.AddAutoRole(ctx, guildID, 222)

				err := repo.RemoveAutoRole(ctx, guildID, 111)
				assert.NoError(t, err)

				roles, err := repo.ListAutoRoles(ctx, guildID)
				assert.NoError(t, err)
				assert.Len(t, roles, 1)
				assert.Equal(t, int64(222), roles[0].RoleID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := repository.NewGuildRepo(db)
			tt.run(t, repo)
		})
	}
}
