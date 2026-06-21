package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TicketRepo handles database operations for the tickets table.
type TicketRepo struct{ db *gorm.DB }

func NewTicketRepo(db *gorm.DB) *TicketRepo { return &TicketRepo{db: db} }

// TicketFilter holds optional filters for listing tickets.
type TicketFilter struct {
	Status     *TicketStatus
	Priority   *TicketPriority
	CategoryID *string
	OpenedBy   *int64
	Page       int
	Limit      int
}

// Create inserts a new ticket and returns it with its DB-assigned fields.
func (r *TicketRepo) Create(ctx context.Context, t *Ticket) (*Ticket, error) {
	t.ID = uuid.NewString()
	return t, r.db.WithContext(ctx).Create(t).Error
}

// GetByID returns a ticket by its UUID.
func (r *TicketRepo) GetByID(ctx context.Context, ticketID string) (*Ticket, error) {
	var t Ticket
	result := r.db.WithContext(ctx).First(&t, "id = ?", ticketID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &t, result.Error
}

// GetByChannelID finds the ticket associated with a Discord channel or thread.
func (r *TicketRepo) GetByChannelID(ctx context.Context, channelID int64) (*Ticket, error) {
	var t Ticket
	result := r.db.WithContext(ctx).
		Where("channel_id = ? OR thread_id = ?", channelID, channelID).
		First(&t)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &t, result.Error
}

// List returns a paginated, filtered list of tickets for a guild.
func (r *TicketRepo) List(ctx context.Context, guildID int64, f TicketFilter) ([]Ticket, int64, error) {
	if f.Limit <= 0 {
		f.Limit = 25
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Page <= 0 {
		f.Page = 1
	}

	query := r.db.WithContext(ctx).Model(&Ticket{}).Where("guild_id = ?", guildID)

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

	var tickets []Ticket
	offset := (f.Page - 1) * f.Limit
	result := query.Order("created_at DESC").Limit(f.Limit).Offset(offset).Find(&tickets)
	return tickets, total, result.Error
}

// CountOpenByUser returns the number of open/claimed tickets a user has in a category.
func (r *TicketRepo) CountOpenByUser(ctx context.Context, guildID int64, categoryID string, userID int64) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&Ticket{}).
		Where("guild_id = ? AND category_id = ? AND opened_by = ? AND status IN ?",
			guildID, categoryID, userID, []TicketStatus{TicketStatusOpen, TicketStatusClaimed}).
		Count(&count)
	return count, result.Error
}

// UpdateStatus transitions the ticket to a new status and optionally sets close reason / closed_at.
func (r *TicketRepo) UpdateStatus(ctx context.Context, ticketID string, status TicketStatus, reason *string) (*Ticket, error) {
	updates := map[string]interface{}{"status": string(status), "close_reason": reason}
	if status == TicketStatusClosed || status == TicketStatusArchived {
		updates["closed_at"] = time.Now().UTC()
	}

	result := r.db.WithContext(ctx).
		Model(&Ticket{}).
		Where("id = ?", ticketID).
		Updates(updates)
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, ticketID)
}

// UpdateClaim sets or clears the claimed_by field and syncs status.
func (r *TicketRepo) UpdateClaim(ctx context.Context, ticketID string, claimedBy *int64) (*Ticket, error) {
	status := TicketStatusClaimed
	if claimedBy == nil {
		status = TicketStatusOpen
	}
	result := r.db.WithContext(ctx).
		Model(&Ticket{}).
		Where("id = ?", ticketID).
		Updates(map[string]interface{}{"claimed_by": claimedBy, "status": string(status)})
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, ticketID)
}

// UpdatePriority changes the priority of a ticket.
func (r *TicketRepo) UpdatePriority(ctx context.Context, ticketID string, priority TicketPriority) (*Ticket, error) {
	result := r.db.WithContext(ctx).
		Model(&Ticket{}).
		Where("id = ?", ticketID).
		Update("priority", string(priority))
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, ticketID)
}

// SetChannel stores the Discord channel/thread ID after creation.
func (r *TicketRepo) SetChannel(ctx context.Context, ticketID string, channelID, threadID *int64) error {
	return r.db.WithContext(ctx).
		Model(&Ticket{}).
		Where("id = ?", ticketID).
		Updates(map[string]interface{}{"channel_id": channelID, "thread_id": threadID}).Error
}

// ResetAutoClose updates the auto-close deadline (called on new activity in a ticket).
func (r *TicketRepo) ResetAutoClose(ctx context.Context, ticketID string, autoCloseAt *time.Time) error {
	return r.db.WithContext(ctx).
		Model(&Ticket{}).
		Where("id = ?", ticketID).
		Update("auto_close_at", autoCloseAt).Error
}

// ListDueForAutoClose returns all open/claimed tickets past their auto_close_at time.
func (r *TicketRepo) ListDueForAutoClose(ctx context.Context) ([]Ticket, error) {
	var tickets []Ticket
	result := r.db.WithContext(ctx).
		Where("auto_close_at IS NOT NULL AND auto_close_at <= NOW() AND status IN ?",
			[]TicketStatus{TicketStatusOpen, TicketStatusClaimed}).
		Find(&tickets)
	return tickets, result.Error
}
