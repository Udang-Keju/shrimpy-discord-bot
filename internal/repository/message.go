package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageRepo handles database operations for the ticket_messages table.
type MessageRepo struct{ db *gorm.DB }

// NewMessageRepo creates a new concrete MessageRepo instance.
func NewMessageRepo(db *gorm.DB) *MessageRepo { return &MessageRepo{db: db} }

// Add appends a message (or staff note) to a ticket transcript.
func (r *MessageRepo) Add(ctx context.Context, m *TicketMessage) (*TicketMessage, error) {
	m.ID = uuid.NewString()
	return m, r.db.WithContext(ctx).Create(m).Error
}

// ListByTicket returns all messages for a ticket ordered chronologically.
func (r *MessageRepo) ListByTicket(ctx context.Context, ticketID string) ([]TicketMessage, error) {
	var msgs []TicketMessage
	result := r.db.WithContext(ctx).
		Where("ticket_id = ?", ticketID).
		Order("sent_at").
		Find(&msgs)
	return msgs, result.Error
}

// ListNonNotesByTicket returns only the public (non-staff-note) messages.
// Used for generating user-facing transcripts.
func (r *MessageRepo) ListNonNotesByTicket(ctx context.Context, ticketID string) ([]TicketMessage, error) {
	var msgs []TicketMessage
	result := r.db.WithContext(ctx).
		Where("ticket_id = ? AND is_staff_note = false", ticketID).
		Order("sent_at").
		Find(&msgs)
	return msgs, result.Error
}
