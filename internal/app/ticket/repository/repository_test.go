package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/repository"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE ticket_panels (
		id TEXT PRIMARY KEY,
		guild_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		channel_id INTEGER NOT NULL,
		message_id INTEGER,
		panel_style TEXT DEFAULT 'buttons',
		content TEXT,
		embed_title TEXT,
		embed_description TEXT,
		embed_color INTEGER,
		embed_media TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE ticket_categories (
		id TEXT PRIMARY KEY,
		panel_id TEXT NOT NULL,
		name TEXT NOT NULL,
		emoji TEXT,
		button_label TEXT NOT NULL,
		button_style TEXT DEFAULT 'primary',
		button_description TEXT,
		button_order INTEGER DEFAULT 0,
		ticket_destination TEXT DEFAULT 'thread',
		thread_parent_channel_id INTEGER,
		channel_category_id INTEGER,
		ticket_name_template TEXT DEFAULT '{category}-{number}',
		ticket_open_title TEXT,
		ticket_open_message TEXT,
		ticket_open_color INTEGER,
		ticket_open_media TEXT,
		ticket_open_content TEXT,
		max_tickets_per_user INTEGER DEFAULT 1,
		auto_close_hours INTEGER,
		transcript_channel_id INTEGER,
		allow_user_close BOOLEAN DEFAULT true,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE tickets (
		id TEXT PRIMARY KEY,
		guild_id INTEGER NOT NULL,
		category_id TEXT NOT NULL,
		channel_id INTEGER,
		thread_id INTEGER,
		opened_by INTEGER NOT NULL,
		claimed_by INTEGER,
		status TEXT DEFAULT 'open',
		priority TEXT DEFAULT 'medium',
		close_reason TEXT,
		auto_close_at DATETIME,
		closed_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE ticket_messages (
		id TEXT PRIMARY KEY,
		ticket_id TEXT NOT NULL,
		author_id INTEGER NOT NULL,
		author_username TEXT NOT NULL,
		content TEXT,
		is_staff_note BOOLEAN DEFAULT false,
		attachments TEXT,
		sent_at DATETIME
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

func TestPanelAndCategoryRepo(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)

	t.Run("Panels and Categories Operations Flow", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewCategoryRepo(db)

		// 1. Create Panel
		p := &model.TicketPanel{
			GuildID:    guildID,
			Name:       "Main Support",
			ChannelID:  999888,
			PanelStyle: "buttons",
			Content:    stringPtr("Need help? Click below."),
		}
		createdPanel, err := repo.CreatePanel(ctx, p)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdPanel.ID)

		// 2. Fetch Panel
		fetchedPanel, err := repo.GetPanel(ctx, createdPanel.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Main Support", fetchedPanel.Name)
		assert.Equal(t, "Need help? Click below.", *fetchedPanel.Content)

		// 3. Set Panel Message ID
		err = repo.SetPanelMessage(ctx, createdPanel.ID, 555)
		assert.NoError(t, err)

		fetchedPanel, err = repo.GetPanelByGuild(ctx, createdPanel.ID, guildID)
		assert.NoError(t, err)
		assert.Equal(t, int64(555), *fetchedPanel.MessageID)

		// 3b. Update panel content
		fetchedPanel.Content = stringPtr("Updated text")
		updatedPanel, err := repo.UpdatePanel(ctx, fetchedPanel)
		assert.NoError(t, err)
		assert.Equal(t, "Updated text", *updatedPanel.Content)

		// 3c. Clear panel message
		err = repo.ClearPanelMessage(ctx, createdPanel.ID)
		assert.NoError(t, err)
		fetchedPanel, err = repo.GetPanel(ctx, createdPanel.ID)
		assert.NoError(t, err)
		assert.Nil(t, fetchedPanel.MessageID)

		// 4. Create Category
		cat := &model.TicketCategory{
			PanelID:           createdPanel.ID,
			Name:              "billing",
			ButtonLabel:       "Open Billing Ticket",
			MaxTicketsPerUser: 3,
			TicketOpenContent: stringPtr("Welcome, we'll be right with you."),
		}
		createdCat, err := repo.CreateCategory(ctx, cat)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdCat.ID)

		// 5. Fetch Category
		fetchedCat, err := repo.GetCategory(ctx, createdCat.ID)
		assert.NoError(t, err)
		assert.Equal(t, "billing", fetchedCat.Name)
		assert.Equal(t, "Welcome, we'll be right with you.", *fetchedCat.TicketOpenContent)

		// 6. List categories by panel
		cats, err := repo.ListCategoriesByPanel(ctx, createdPanel.ID)
		assert.NoError(t, err)
		assert.Len(t, cats, 1)

		// 7. Update Category
		fetchedCat.Name = "billing-updated"
		fetchedCat.TicketOpenContent = stringPtr("Updated greeting")
		updatedCat, err := repo.UpdateCategory(ctx, fetchedCat)
		assert.NoError(t, err)
		assert.Equal(t, "billing-updated", updatedCat.Name)
		assert.Equal(t, "Updated greeting", *updatedCat.TicketOpenContent)

		// 8. Delete Category
		err = repo.DeleteCategory(ctx, createdCat.ID)
		assert.NoError(t, err)

		fetchedCat, err = repo.GetCategory(ctx, createdCat.ID)
		assert.ErrorIs(t, err, model.ErrNotFound)
		assert.Nil(t, fetchedCat)

		// 9. Delete Panel
		err = repo.DeletePanel(ctx, createdPanel.ID)
		assert.NoError(t, err)
	})
}

func TestTicketRepo(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	categoryID := "cat-uuid"
	userID := int64(67890)

	t.Run("Ticket Operations Lifecycle", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewTicketRepo(db)

		tkt := &model.Ticket{
			GuildID:    guildID,
			CategoryID: categoryID,
			OpenedBy:   userID,
			Status:     model.TicketStatusOpen,
			Priority:   model.TicketPriorityMedium,
		}

		// 1. Create ticket
		created, err := repo.Create(ctx, tkt)
		assert.NoError(t, err)
		assert.NotEmpty(t, created.ID)

		// 2. Fetch ticket
		fetched, err := repo.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, model.TicketStatusOpen, fetched.Status)

		// 3. Count Open By User
		count, err := repo.CountOpenByUser(ctx, guildID, categoryID, userID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// 4. Set Channel
		err = repo.SetChannel(ctx, created.ID, int64Ptr(888999), nil)
		assert.NoError(t, err)

		fetched, err = repo.GetByChannelID(ctx, 888999)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, fetched.ID)

		// 5. Update Status
		reason := "issue resolved"
		updated, err := repo.UpdateStatus(ctx, created.ID, model.TicketStatusClosed, &reason)
		assert.NoError(t, err)
		assert.Equal(t, model.TicketStatusClosed, updated.Status)
		assert.Equal(t, reason, *updated.CloseReason)

		// 6. Update Claim
		staffID := int64(111222)
		updated, err = repo.UpdateClaim(ctx, created.ID, &staffID)
		assert.NoError(t, err)
		assert.Equal(t, staffID, *updated.ClaimedBy)

		// 7. Update Priority
		updated, err = repo.UpdatePriority(ctx, created.ID, model.TicketPriorityHigh)
		assert.NoError(t, err)
		assert.Equal(t, model.TicketPriorityHigh, updated.Priority)

		// 8. Reset Auto Close
		now := time.Now().Truncate(time.Second)
		err = repo.ResetAutoClose(ctx, created.ID, &now)
		assert.NoError(t, err)

		fetched, err = repo.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.True(t, fetched.AutoCloseAt.Equal(now))
	})
}

func TestMessageRepo(t *testing.T) {
	ctx := context.Background()
	ticketID := "ticket-uuid"

	t.Run("Add and List Messages", func(t *testing.T) {
		db := setupTestDB(t)
		repo := repository.NewMessageRepo(db)

		msgContent := "hello"
		noteContent := "staff note"

		m1 := &model.TicketMessage{
			TicketID:       ticketID,
			AuthorID:       111,
			AuthorUsername: "user1",
			Content:        &msgContent,
			IsStaffNote:    false,
			SentAt:         time.Now(),
		}
		m2 := &model.TicketMessage{
			TicketID:       ticketID,
			AuthorID:       222,
			AuthorUsername: "staff1",
			Content:        &noteContent,
			IsStaffNote:    true,
			SentAt:         time.Now(),
		}

		_, err := repo.Add(ctx, m1)
		assert.NoError(t, err)
		_, err = repo.Add(ctx, m2)
		assert.NoError(t, err)

		// List all (including staff notes)
		all, err := repo.ListByTicket(ctx, ticketID)
		assert.NoError(t, err)
		assert.Len(t, all, 2)

		// List non-notes
		nonNotes, err := repo.ListNonNotesByTicket(ctx, ticketID)
		assert.NoError(t, err)
		assert.Len(t, nonNotes, 1)
		assert.Equal(t, "hello", *nonNotes[0].Content)
	})
}
