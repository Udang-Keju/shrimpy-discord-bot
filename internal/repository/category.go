package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CategoryRepo handles database operations for ticket_panels and ticket_categories.
type CategoryRepo struct{ db *gorm.DB }

// NewCategoryRepo creates a new concrete CategoryRepo instance.
func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

// ─── Ticket Panels ────────────────────────────────────────────────────────────

// ListPanels returns all panels for a guild with their categories eagerly loaded.
func (r *CategoryRepo) ListPanels(ctx context.Context, guildID int64) ([]TicketPanel, error) {
	var panels []TicketPanel
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
func (r *CategoryRepo) GetPanel(ctx context.Context, panelID string) (*TicketPanel, error) {
	var p TicketPanel
	result := r.db.WithContext(ctx).
		Preload("Categories", func(db *gorm.DB) *gorm.DB {
			return db.Order("button_order, created_at")
		}).
		First(&p, "id = ?", panelID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, result.Error
}

// GetPanelByGuild ensures a panel belongs to the expected guild (authorization guard).
func (r *CategoryRepo) GetPanelByGuild(ctx context.Context, panelID string, guildID int64) (*TicketPanel, error) {
	var p TicketPanel
	result := r.db.WithContext(ctx).
		First(&p, "id = ? AND guild_id = ?", panelID, guildID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, result.Error
}

// CreatePanel inserts a new ticket panel.
func (r *CategoryRepo) CreatePanel(ctx context.Context, p *TicketPanel) (*TicketPanel, error) {
	p.ID = uuid.NewString()
	return p, r.db.WithContext(ctx).Create(p).Error
}

// UpdatePanel saves changes to an existing panel.
func (r *CategoryRepo) UpdatePanel(ctx context.Context, p *TicketPanel) (*TicketPanel, error) {
	result := r.db.WithContext(ctx).
		Model(p).
		Clauses(clause.Returning{}).
		Updates(map[string]interface{}{
			"name":              p.Name,
			"channel_id":        p.ChannelID,
			"panel_style":       p.PanelStyle,
			"embed_title":       p.EmbedTitle,
			"embed_description": p.EmbedDescription,
			"embed_color":       p.EmbedColor,
			"embed_media":       p.EmbedMedia,
		})
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return p, result.Error
}

// SetPanelMessage stores the Discord message ID of the posted panel embed.
func (r *CategoryRepo) SetPanelMessage(ctx context.Context, panelID string, messageID int64) error {
	return r.db.WithContext(ctx).
		Model(&TicketPanel{}).
		Where("id = ?", panelID).
		Update("message_id", messageID).Error
}

// DeletePanel removes a panel and its categories (CASCADE in DB).
func (r *CategoryRepo) DeletePanel(ctx context.Context, panelID string) error {
	return r.db.WithContext(ctx).Where("id = ?", panelID).Delete(&TicketPanel{}).Error
}

// ─── Ticket Categories ────────────────────────────────────────────────────────

// ListCategoriesByPanel returns all categories for a panel, ordered by button_order.
func (r *CategoryRepo) ListCategoriesByPanel(ctx context.Context, panelID string) ([]TicketCategory, error) {
	var cats []TicketCategory
	result := r.db.WithContext(ctx).
		Where("panel_id = ?", panelID).
		Order("button_order, created_at").
		Find(&cats)
	return cats, result.Error
}

// GetCategory returns a single category by UUID.
func (r *CategoryRepo) GetCategory(ctx context.Context, categoryID string) (*TicketCategory, error) {
	var c TicketCategory
	result := r.db.WithContext(ctx).First(&c, "id = ?", categoryID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, result.Error
}

// CreateCategory inserts a new ticket category.
func (r *CategoryRepo) CreateCategory(ctx context.Context, c *TicketCategory) (*TicketCategory, error) {
	c.ID = uuid.NewString()
	return c, r.db.WithContext(ctx).Create(c).Error
}

// UpdateCategory saves changes to an existing category.
func (r *CategoryRepo) UpdateCategory(ctx context.Context, c *TicketCategory) (*TicketCategory, error) {
	result := r.db.WithContext(ctx).
		Model(c).
		Clauses(clause.Returning{}).
		Updates(map[string]interface{}{
			"name":                   c.Name,
			"emoji":                  c.Emoji,
			"button_label":           c.ButtonLabel,
			"button_style":           c.ButtonStyle,
			"button_description":     c.ButtonDescription,
			"button_order":           c.ButtonOrder,
			"ticket_destination":     c.TicketDestination,
			"ticket_name_template":   c.TicketNameTemplate,
			"ticket_open_title":      c.TicketOpenTitle,
			"ticket_open_message":    c.TicketOpenMessage,
			"ticket_open_color":      c.TicketOpenColor,
			"ticket_open_media":      c.TicketOpenMedia,
			"max_tickets_per_user":   c.MaxTicketsPerUser,
			"auto_close_hours":       c.AutoCloseHours,
			"transcript_channel_id":  c.TranscriptChannelID,
			"allow_user_close":       c.AllowUserClose,
		})
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return c, result.Error
}

// DeleteCategory removes a ticket category.
func (r *CategoryRepo) DeleteCategory(ctx context.Context, categoryID string) error {
	return r.db.WithContext(ctx).Where("id = ?", categoryID).Delete(&TicketCategory{}).Error
}
