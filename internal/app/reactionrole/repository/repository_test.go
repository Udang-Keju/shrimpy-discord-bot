package repository_test

import (
	"context"
	"testing"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/repository"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE reaction_role_messages (
		id TEXT PRIMARY KEY,
		guild_id INTEGER NOT NULL,
		channel_id INTEGER NOT NULL,
		message_id INTEGER,
		embed_title TEXT,
		embed_description TEXT,
		embed_color INTEGER,
		embed_media TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE reaction_role_emojis (
		id TEXT PRIMARY KEY,
		message_id TEXT NOT NULL,
		emoji TEXT NOT NULL,
		role_id INTEGER NOT NULL,
		created_at DATETIME,
		UNIQUE(message_id, emoji)
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

func TestReactionRoleRepo_MessagesLifecycle(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	channelID := int64(999888)

	t.Run("Create, Fetch, Update, and Delete Message", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewReactionRoleRepo(db)

		msg := &model.ReactionRoleMessage{
			GuildID:          guildID,
			ChannelID:        channelID,
			EmbedTitle:       stringPtr("Rules"),
			EmbedDescription: stringPtr("Role assignment"),
		}

		// 1. Create Message
		created, err := repo.CreateMessage(ctx, msg)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)

		// 2. Fetch and Verify
		fetched, err := repo.GetMessage(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Rules", *fetched.EmbedTitle)
		assert.Nil(t, fetched.MessageID)

		// 3. Set Discord Message ID
		err = repo.SetDiscordMessageID(ctx, created.ID, 777)
		assert.NoError(t, err)

		fetched, err = repo.GetByDiscordMessageID(ctx, 777)
		assert.NoError(t, err)
		assert.Equal(t, int64(777), *fetched.MessageID)

		// 4. Update Message Embed details
		fetched.EmbedTitle = stringPtr("New Title")
		updated, err := repo.UpdateMessage(ctx, fetched)
		assert.NoError(t, err)
		assert.Equal(t, "New Title", *updated.EmbedTitle)

		// 5. List messages by guild
		list, err := repo.ListByGuild(ctx, guildID)
		assert.NoError(t, err)
		assert.Len(t, list, 1)

		// 6. Delete message
		err = repo.DeleteMessage(ctx, created.ID)
		assert.NoError(t, err)

		fetched, err = repo.GetMessage(ctx, created.ID)
		assert.ErrorIs(t, err, repository.ErrNotFound)
		assert.Nil(t, fetched)
	})
}

func TestReactionRoleRepo_EmojisMapping(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	channelID := int64(999888)
	db := setupTestDB(t)
	repo := repository.NewReactionRoleRepo(db)

	msg := &model.ReactionRoleMessage{
		GuildID:   guildID,
		ChannelID: channelID,
		MessageID: int64Ptr(777),
	}
	created, err := repo.CreateMessage(ctx, msg)
	require.NoError(t, err)

	t.Run("Add, Get, and Remove Emoji Mappings", func(t *testing.T) {
		e := &model.ReactionRoleEmoji{
			MessageID: created.ID,
			Emoji:     "🦀",
			RoleID:    555,
		}

		// 1. Add Emoji Mapping
		added, err := repo.AddEmoji(ctx, e)
		assert.NoError(t, err)
		assert.NotEmpty(t, added.ID)

		// 2. Add Same Emoji with different role (Idempotent update check)
		e.RoleID = 666
		addedDup, err := repo.AddEmoji(ctx, e)
		assert.NoError(t, err)
		assert.Equal(t, added.ID, addedDup.ID)

		// 3. Get Emoji Role (Join check)
		emojiRole, err := repo.GetEmojiRole(ctx, 777, "🦀")
		assert.NoError(t, err)
		assert.Equal(t, int64(666), emojiRole.RoleID)

		// 4. Preload check on GetMessage
		fetchedMsg, err := repo.GetMessage(ctx, created.ID)
		assert.NoError(t, err)
		assert.Len(t, fetchedMsg.Emojis, 1)
		assert.Equal(t, "🦀", fetchedMsg.Emojis[0].Emoji)

		// 5. Remove Emoji mapping
		err = repo.RemoveEmoji(ctx, created.ID, "🦀")
		assert.NoError(t, err)

		emojiRole, err = repo.GetEmojiRole(ctx, 777, "🦀")
		assert.ErrorIs(t, err, repository.ErrNotFound)
		assert.Nil(t, emojiRole)
	})
}
