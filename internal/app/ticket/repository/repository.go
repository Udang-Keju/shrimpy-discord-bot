package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TicketRepo handles database operations for the tickets table.
type TicketRepo struct{ db *gorm.DB }

// NewTicketRepo creates a new concrete TicketRepo instance.
func NewTicketRepo(db *gorm.DB) *TicketRepo { return &TicketRepo{db: db} }

// Create inserts a new ticket and returns it with its DB-assigned fields.
func (r *TicketRepo) Create(ctx context.Context, t *model.Ticket) (*model.Ticket, error) {
	t.ID = uuid.NewString()
	return t, r.db.WithContext(ctx).Create(t).Error
}

// GetByID returns a ticket by its UUID.
func (r *TicketRepo) GetByID(ctx context.Context, ticketID string) (*model.Ticket, error) {
	var t model.Ticket
	result := r.db.WithContext(ctx).First(&t, "id = ?", ticketID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, model.ErrNotFound
	}
	return &t, result.Error
}

// GetByChannelID finds the ticket associated with a Discord channel or thread.
func (r *TicketRepo) GetByChannelID(ctx context.Context, channelID int64) (*model.Ticket, error) {
	var t model.Ticket
	result := r.db.WithContext(ctx).
		Where("channel_id = ? OR thread_id = ?", channelID, channelID).
		First(&t)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, model.ErrNotFound
	}
	return &t, result.Error
}

// List returns a paginated, filtered list of tickets for a guild.
func (r *TicketRepo) List(ctx context.Context, guildID int64, f model.TicketFilter) ([]model.Ticket, int64, error) {
	if f.Limit <= 0 {
		f.Limit = 25
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Page <= 0 {
		f.Page = 1
	}

	query := r.db.WithContext(ctx).Model(&model.Ticket{}).Where("guild_id = ?", guildID)

	if f.Status != nil {
		query = query.Where("status = ?", *f.Status)
	}
	if f.Priority != nil {
		query = query.Where("priority = ?", *f.Priority)
	}
	if f.CategoryID != nil {
		query = query.Where("category_id = ?", *f.CategoryID)
	}
	if f.OpenedBy != nil {
		query = query.Where("opened_by = ?", *f.OpenedBy)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tickets []model.Ticket
	offset := (f.Page - 1) * f.Limit
	result := query.Order("created_at DESC").Limit(f.Limit).Offset(offset).Find(&tickets)
	return tickets, total, result.Error
}

// CountOpenByUser returns the number of open/claimed tickets a user has in a category.
func (r *TicketRepo) CountOpenByUser(ctx context.Context, guildID int64, categoryID string, userID int64) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&model.Ticket{}).
		Where("guild_id = ? AND category_id = ? AND opened_by = ? AND status IN ?",
			guildID, categoryID, userID, []model.TicketStatus{model.TicketStatusOpen, model.TicketStatusClaimed}).
		Count(&count)
	return count, result.Error
}

// UpdateStatus transitions the ticket to a new status and optionally sets close reason / closed_at.
func (r *TicketRepo) UpdateStatus(ctx context.Context, ticketID string, status model.TicketStatus, reason *string) (*model.Ticket, error) {
	updates := map[string]interface{}{"status": string(status), "close_reason": reason}
	if status == model.TicketStatusClosed || status == model.TicketStatusArchived {
		updates["closed_at"] = time.Now().UTC()
	}

	result := r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ?", ticketID).
		Updates(updates)
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}
	return r.GetByID(ctx, ticketID)
}

// UpdateClaim sets or clears the claimed_by field and syncs status.
func (r *TicketRepo) UpdateClaim(ctx context.Context, ticketID string, claimedBy *int64) (*model.Ticket, error) {
	status := model.TicketStatusClaimed
	if claimedBy == nil {
		status = model.TicketStatusOpen
	}
	result := r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ?", ticketID).
		Updates(map[string]interface{}{"claimed_by": claimedBy, "status": string(status)})
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}
	return r.GetByID(ctx, ticketID)
}

// UpdatePriority changes the priority of a ticket.
func (r *TicketRepo) UpdatePriority(ctx context.Context, ticketID string, priority model.TicketPriority) (*model.Ticket, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ?", ticketID).
		Update("priority", string(priority))
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}
	return r.GetByID(ctx, ticketID)
}

// SetChannel stores the Discord channel/thread ID after creation.
func (r *TicketRepo) SetChannel(ctx context.Context, ticketID string, channelID, threadID *int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ?", ticketID).
		Updates(map[string]interface{}{"channel_id": channelID, "thread_id": threadID}).Error
}

// ResetAutoClose updates the auto-close deadline (called on new activity in a ticket).
func (r *TicketRepo) ResetAutoClose(ctx context.Context, ticketID string, autoCloseAt *time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ?", ticketID).
		Update("auto_close_at", autoCloseAt).Error
}

// ListDueForAutoClose returns all open/claimed tickets past their auto_close_at time.
func (r *TicketRepo) ListDueForAutoClose(ctx context.Context) ([]model.Ticket, error) {
	var tickets []model.Ticket
	result := r.db.WithContext(ctx).
		Where("auto_close_at IS NOT NULL AND auto_close_at <= NOW() AND status IN ?",
			[]model.TicketStatus{model.TicketStatusOpen, model.TicketStatusClaimed}).
		Find(&tickets)
	return tickets, result.Error
}

// GetStats calculates guild ticket stats directly via SQL aggregations.
func (r *TicketRepo) GetStats(ctx context.Context, guildID int64) (*model.TicketStats, error) {
	var stats model.TicketStats

	// 1. Open count
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("guild_id = ? AND status = ?", guildID, model.TicketStatusOpen).Count(&stats.Open)

	// 2. Claimed count
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("guild_id = ? AND status = ?", guildID, model.TicketStatusClaimed).Count(&stats.Claimed)

	// 3. Closed this month
	firstOfMonth := time.Now().UTC().AddDate(0, 0, -time.Now().UTC().Day()+1)
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("guild_id = ? AND status = ? AND closed_at >= ?", guildID, model.TicketStatusClosed, firstOfMonth).Count(&stats.ClosedThisMonth)

	// 4. Archived count
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("guild_id = ? AND status = ?", guildID, model.TicketStatusArchived).Count(&stats.ArchivedTotal)

	// 5. Avg resolution time in minutes
	var avgSec float64
	row := r.db.WithContext(ctx).Model(&model.Ticket{}).
		Select("COALESCE(AVG(EXTRACT(EPOCH FROM (closed_at - created_at))), 0)").
		Where("guild_id = ? AND closed_at IS NOT NULL", guildID).
		Row()
	_ = row.Scan(&avgSec)
	stats.AvgResolutionMin = int64(avgSec / 60)

	// 6. Top Category ID
	var topCat struct {
		CategoryID string
		Count      int64
	}
	r.db.WithContext(ctx).Model(&model.Ticket{}).
		Select("category_id, count(*) as count").
		Where("guild_id = ?", guildID).
		Group("category_id").
		Order("count DESC").
		Limit(1).
		Scan(&topCat)
	stats.TopCategoryID = topCat.CategoryID

	return &stats, nil
}

// ─── CategoryRepo ─────────────────────────────────────────────────────────────

// CategoryRepo handles database operations for ticket_panels and ticket_categories.
type CategoryRepo struct{ db *gorm.DB }

// NewCategoryRepo creates a new concrete CategoryRepo instance.
func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

// ListPanels returns all panels for a guild with their categories eagerly loaded.
func (r *CategoryRepo) ListPanels(ctx context.Context, guildID int64) ([]model.TicketPanel, error) {
	var panels []model.TicketPanel
	result := r.db.WithContext(ctx).
		Preload("Categories", func(db *gorm.DB) *gorm.DB {
			return db.Order("button_order, created_at")
		}).
		Where("guild_id = ?", guildID).
		Order("created_at").
		Find(&panels)
	return panels, result.Error
}

// GetPanel returns a single panel by UUID with its categories.
func (r *CategoryRepo) GetPanel(ctx context.Context, panelID string) (*model.TicketPanel, error) {
	var p model.TicketPanel
	result := r.db.WithContext(ctx).
		Preload("Categories", func(db *gorm.DB) *gorm.DB {
			return db.Order("button_order, created_at")
		}).
		First(&p, "id = ?", panelID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, model.ErrNotFound
	}
	return &p, result.Error
}

// GetPanelByGuild ensures a panel belongs to the expected guild (authorization guard).
func (r *CategoryRepo) GetPanelByGuild(ctx context.Context, panelID string, guildID int64) (*model.TicketPanel, error) {
	var p model.TicketPanel
	result := r.db.WithContext(ctx).
		First(&p, "id = ? AND guild_id = ?", panelID, guildID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, model.ErrNotFound
	}
	return &p, result.Error
}

// CreatePanel inserts a new ticket panel.
func (r *CategoryRepo) CreatePanel(ctx context.Context, p *model.TicketPanel) (*model.TicketPanel, error) {
	p.ID = uuid.NewString()
	return p, r.db.WithContext(ctx).Create(p).Error
}

// UpdatePanel saves changes to an existing panel.
func (r *CategoryRepo) UpdatePanel(ctx context.Context, p *model.TicketPanel) (*model.TicketPanel, error) {
	result := r.db.WithContext(ctx).
		Model(p).
		Clauses(clause.Returning{}).
		Updates(map[string]interface{}{
			"name":              p.Name,
			"channel_id":        p.ChannelID,
			"panel_style":       p.PanelStyle,
			"content":           p.Content,
			"embed_title":       p.EmbedTitle,
			"embed_description": p.EmbedDescription,
			"embed_color":       p.EmbedColor,
			"embed_media":       p.EmbedMedia,
		})
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}
	return p, result.Error
}

// SetPanelMessage stores the Discord message ID of the posted panel embed.
func (r *CategoryRepo) SetPanelMessage(ctx context.Context, panelID string, messageID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.TicketPanel{}).
		Where("id = ?", panelID).
		Update("message_id", messageID).Error
}

// ClearPanelMessage nulls out the stored Discord message ID, e.g. after the panel's
// destination channel changes and the old message no longer applies.
func (r *CategoryRepo) ClearPanelMessage(ctx context.Context, panelID string) error {
	return r.db.WithContext(ctx).
		Model(&model.TicketPanel{}).
		Where("id = ?", panelID).
		Update("message_id", nil).Error
}

// DeletePanel removes a panel and its categories (CASCADE in DB).
func (r *CategoryRepo) DeletePanel(ctx context.Context, panelID string) error {
	return r.db.WithContext(ctx).Where("id = ?", panelID).Delete(&model.TicketPanel{}).Error
}

// ListCategoriesByPanel returns all categories for a panel, ordered by button_order.
func (r *CategoryRepo) ListCategoriesByPanel(ctx context.Context, panelID string) ([]model.TicketCategory, error) {
	var cats []model.TicketCategory
	result := r.db.WithContext(ctx).
		Where("panel_id = ?", panelID).
		Order("button_order, created_at").
		Find(&cats)
	return cats, result.Error
}

// GetCategory returns a single category by UUID.
func (r *CategoryRepo) GetCategory(ctx context.Context, categoryID string) (*model.TicketCategory, error) {
	var c model.TicketCategory
	result := r.db.WithContext(ctx).First(&c, "id = ?", categoryID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, model.ErrNotFound
	}
	return &c, result.Error
}

// CreateCategory inserts a new ticket category.
func (r *CategoryRepo) CreateCategory(ctx context.Context, c *model.TicketCategory) (*model.TicketCategory, error) {
	c.ID = uuid.NewString()
	return c, r.db.WithContext(ctx).Create(c).Error
}

// UpdateCategory saves changes to an existing category.
func (r *CategoryRepo) UpdateCategory(ctx context.Context, c *model.TicketCategory) (*model.TicketCategory, error) {
	result := r.db.WithContext(ctx).
		Model(c).
		Clauses(clause.Returning{}).
		Updates(map[string]interface{}{
			"name":                  c.Name,
			"emoji":                 c.Emoji,
			"button_label":          c.ButtonLabel,
			"button_style":          c.ButtonStyle,
			"button_description":    c.ButtonDescription,
			"button_order":          c.ButtonOrder,
			"ticket_destination":       c.TicketDestination,
			"thread_parent_channel_id": c.ThreadParentChannelID,
			"channel_category_id":      c.ChannelCategoryID,
			"ticket_name_template":     c.TicketNameTemplate,
			"ticket_open_title":     c.TicketOpenTitle,
			"ticket_open_message":   c.TicketOpenMessage,
			"ticket_open_color":     c.TicketOpenColor,
			"ticket_open_media":     c.TicketOpenMedia,
			"ticket_open_content":   c.TicketOpenContent,
			"max_tickets_per_user":  c.MaxTicketsPerUser,
			"auto_close_hours":      c.AutoCloseHours,
			"transcript_channel_id": c.TranscriptChannelID,
			"allow_user_close":      c.AllowUserClose,
		})
	if result.RowsAffected == 0 {
		return nil, model.ErrNotFound
	}
	return c, result.Error
}

// DeleteCategory removes a ticket category.
func (r *CategoryRepo) DeleteCategory(ctx context.Context, categoryID string) error {
	return r.db.WithContext(ctx).Where("id = ?", categoryID).Delete(&model.TicketCategory{}).Error
}

// ─── Panel Handler Roles ──────────────────────────────────────────────────────

// ListPanelHandlerRoles returns all handler roles configured for a panel.
func (r *CategoryRepo) ListPanelHandlerRoles(ctx context.Context, panelID string) ([]model.PanelHandlerRole, error) {
	var roles []model.PanelHandlerRole
	result := r.db.WithContext(ctx).
		Where("panel_id = ?", panelID).
		Order("created_at").
		Find(&roles)
	return roles, result.Error
}

// AddPanelHandlerRole idempotently adds a Discord role as a panel handler role.
func (r *CategoryRepo) AddPanelHandlerRole(ctx context.Context, panelID string, roleID int64) (*model.PanelHandlerRole, error) {
	hr := model.PanelHandlerRole{ID: uuid.NewString(), PanelID: panelID, RoleID: roleID}
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&hr)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		hr.ID = ""
		result = r.db.WithContext(ctx).
			Where("panel_id = ? AND role_id = ?", panelID, roleID).
			First(&hr)
	}
	return &hr, result.Error
}

// RemovePanelHandlerRole removes a role from the panel's handler list.
func (r *CategoryRepo) RemovePanelHandlerRole(ctx context.Context, panelID string, roleID int64) error {
	return r.db.WithContext(ctx).
		Where("panel_id = ? AND role_id = ?", panelID, roleID).
		Delete(&model.PanelHandlerRole{}).Error
}

// SetPanelHandlerRoles reconciles the panel's handler roles to exactly match roleIDs:
// removes rows not in the set, then idempotently inserts the rest.
func (r *CategoryRepo) SetPanelHandlerRoles(ctx context.Context, panelID string, roleIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		del := tx.Where("panel_id = ?", panelID)
		if len(roleIDs) > 0 {
			del = del.Where("role_id NOT IN ?", roleIDs)
		}
		if err := del.Delete(&model.PanelHandlerRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			hr := model.PanelHandlerRole{ID: uuid.NewString(), PanelID: panelID, RoleID: roleID}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&hr).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ─── Category Handler Roles ───────────────────────────────────────────────────

// ListCategoryHandlerRoles returns all handler roles configured for a category.
func (r *CategoryRepo) ListCategoryHandlerRoles(ctx context.Context, categoryID string) ([]model.CategoryHandlerRole, error) {
	var roles []model.CategoryHandlerRole
	result := r.db.WithContext(ctx).
		Where("category_id = ?", categoryID).
		Order("created_at").
		Find(&roles)
	return roles, result.Error
}

// AddCategoryHandlerRole idempotently adds a Discord role as a category handler role.
func (r *CategoryRepo) AddCategoryHandlerRole(ctx context.Context, categoryID string, roleID int64) (*model.CategoryHandlerRole, error) {
	hr := model.CategoryHandlerRole{ID: uuid.NewString(), CategoryID: categoryID, RoleID: roleID}
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&hr)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		hr.ID = ""
		result = r.db.WithContext(ctx).
			Where("category_id = ? AND role_id = ?", categoryID, roleID).
			First(&hr)
	}
	return &hr, result.Error
}

// RemoveCategoryHandlerRole removes a role from the category's handler list.
func (r *CategoryRepo) RemoveCategoryHandlerRole(ctx context.Context, categoryID string, roleID int64) error {
	return r.db.WithContext(ctx).
		Where("category_id = ? AND role_id = ?", categoryID, roleID).
		Delete(&model.CategoryHandlerRole{}).Error
}

// SetCategoryHandlerRoles reconciles the category's handler roles to exactly match
// roleIDs: removes rows not in the set, then idempotently inserts the rest.
func (r *CategoryRepo) SetCategoryHandlerRoles(ctx context.Context, categoryID string, roleIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		del := tx.Where("category_id = ?", categoryID)
		if len(roleIDs) > 0 {
			del = del.Where("role_id NOT IN ?", roleIDs)
		}
		if err := del.Delete(&model.CategoryHandlerRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			hr := model.CategoryHandlerRole{ID: uuid.NewString(), CategoryID: categoryID, RoleID: roleID}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&hr).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ─── MessageRepo ──────────────────────────────────────────────────────────────

// MessageRepo handles database operations for the ticket_messages table.
type MessageRepo struct{ db *gorm.DB }

// NewMessageRepo creates a new concrete MessageRepo instance.
func NewMessageRepo(db *gorm.DB) *MessageRepo { return &MessageRepo{db: db} }

// Add appends a message (or staff note) to a ticket transcript.
func (r *MessageRepo) Add(ctx context.Context, m *model.TicketMessage) (*model.TicketMessage, error) {
	m.ID = uuid.NewString()
	return m, r.db.WithContext(ctx).Create(m).Error
}

// ListByTicket returns all messages for a ticket ordered chronologically.
func (r *MessageRepo) ListByTicket(ctx context.Context, ticketID string) ([]model.TicketMessage, error) {
	var msgs []model.TicketMessage
	result := r.db.WithContext(ctx).
		Where("ticket_id = ?", ticketID).
		Order("sent_at").
		Find(&msgs)
	return msgs, result.Error
}

// ListNonNotesByTicket returns only the public (non-staff-note) messages.
func (r *MessageRepo) ListNonNotesByTicket(ctx context.Context, ticketID string) ([]model.TicketMessage, error) {
	var msgs []model.TicketMessage
	result := r.db.WithContext(ctx).
		Where("ticket_id = ? AND is_staff_note = false", ticketID).
		Order("sent_at").
		Find(&msgs)
	return msgs, result.Error
}
