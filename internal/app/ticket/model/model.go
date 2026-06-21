package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"gorm.io/datatypes"
)

// ─── Status / Priority types ──────────────────────────────────────────────────

// TicketStatus represents the lifecycle state of a ticket.
type TicketStatus string

const (
	TicketStatusOpen     TicketStatus = "open"
	TicketStatusClaimed  TicketStatus = "claimed"
	TicketStatusClosed   TicketStatus = "closed"
	TicketStatusArchived TicketStatus = "archived"
)

// TicketPriority represents the urgency level of a ticket.
type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityMedium TicketPriority = "medium"
	TicketPriorityHigh   TicketPriority = "high"
	TicketPriorityUrgent TicketPriority = "urgent"
)

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var (
	ErrNotFound     = fmt.Errorf("record not found")
	ErrLimitReached = fmt.Errorf("limit reached")
)

// ─── GORM Domain Models ───────────────────────────────────────────────────────

// TicketPanel maps to the ticket_panels table.
type TicketPanel struct {
	ID               string         `gorm:"primaryKey;type:uuid"`
	GuildID          int64          `gorm:"column:guild_id;not null"`
	Name             string         `gorm:"column:name;not null"`
	ChannelID        int64          `gorm:"column:channel_id;not null"`
	MessageID        *int64         `gorm:"column:message_id"`
	PanelStyle       string         `gorm:"column:panel_style;default:'buttons'"`
	EmbedTitle       *string        `gorm:"column:embed_title"`
	EmbedDescription *string        `gorm:"column:embed_description"`
	EmbedColor       *int32         `gorm:"column:embed_color"`
	EmbedMedia       datatypes.JSON `gorm:"column:embed_media;type:jsonb"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	// Associations
	Categories []TicketCategory `gorm:"foreignKey:PanelID"`
}

// TableName overrides the default table name mapping.
func (TicketPanel) TableName() string { return "ticket_panels" }

// GetMedia deserializes the EmbedMedia JSONB column.
func (p *TicketPanel) GetMedia() (*discordutil.EmbedMedia, error) {
	return discordutil.DecodeMedia(p.EmbedMedia)
}

// SetMedia serializes an EmbedMedia into the JSONB column.
func (p *TicketPanel) SetMedia(m *discordutil.EmbedMedia) (err error) {
	p.EmbedMedia, err = discordutil.EncodeMedia(m)
	return
}

// TicketCategory maps to the ticket_categories table.
type TicketCategory struct {
	ID                  string         `gorm:"primaryKey;type:uuid"`
	PanelID             string         `gorm:"column:panel_id;not null;type:uuid"`
	Name                string         `gorm:"column:name;not null"`
	Emoji               *string        `gorm:"column:emoji"`
	ButtonLabel         string         `gorm:"column:button_label;not null"`
	ButtonStyle         string         `gorm:"column:button_style;default:'primary'"`
	ButtonDescription   *string        `gorm:"column:button_description"`
	ButtonOrder         int16          `gorm:"column:button_order;default:0"`
	TicketDestination   string         `gorm:"column:ticket_destination;default:'thread'"`
	TicketNameTemplate  string         `gorm:"column:ticket_name_template;default:'{category}-{number}'"`
	TicketOpenTitle     *string        `gorm:"column:ticket_open_title"`
	TicketOpenMessage   *string        `gorm:"column:ticket_open_message"`
	TicketOpenColor     *int32         `gorm:"column:ticket_open_color"`
	TicketOpenMedia     datatypes.JSON `gorm:"column:ticket_open_media;type:jsonb"`
	MaxTicketsPerUser   int            `gorm:"column:max_tickets_per_user;default:1"`
	AutoCloseHours      *int           `gorm:"column:auto_close_hours"`
	TranscriptChannelID *int64         `gorm:"column:transcript_channel_id"`
	AllowUserClose      bool           `gorm:"column:allow_user_close;default:true"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TableName overrides the default table name mapping.
func (TicketCategory) TableName() string { return "ticket_categories" }

// GetOpenMedia deserializes the EmbedMedia JSONB column.
func (c *TicketCategory) GetOpenMedia() (*discordutil.EmbedMedia, error) {
	return discordutil.DecodeMedia(c.TicketOpenMedia)
}

// SetOpenMedia serializes an EmbedMedia into the JSONB column.
func (c *TicketCategory) SetOpenMedia(m *discordutil.EmbedMedia) (err error) {
	c.TicketOpenMedia, err = discordutil.EncodeMedia(m)
	return
}

// Ticket maps to the tickets table.
type Ticket struct {
	ID          string         `gorm:"primaryKey;type:uuid"`
	GuildID     int64          `gorm:"column:guild_id;not null"`
	CategoryID  string         `gorm:"column:category_id;not null;type:uuid"`
	ChannelID   *int64         `gorm:"column:channel_id"`
	ThreadID    *int64         `gorm:"column:thread_id"`
	OpenedBy    int64          `gorm:"column:opened_by;not null"`
	ClaimedBy   *int64         `gorm:"column:claimed_by"`
	Status      TicketStatus   `gorm:"column:status;default:'open'"`
	Priority    TicketPriority `gorm:"column:priority;default:'medium'"`
	CloseReason *string        `gorm:"column:close_reason"`
	AutoCloseAt *time.Time     `gorm:"column:auto_close_at"`
	ClosedAt    *time.Time     `gorm:"column:closed_at"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// Associations
	Category TicketCategory `gorm:"foreignKey:CategoryID"`
}

// TableName overrides the default table name mapping.
func (Ticket) TableName() string { return "tickets" }

// TicketMessage maps to the ticket_messages table.
type TicketMessage struct {
	ID             string         `gorm:"primaryKey;type:uuid"`
	TicketID       string         `gorm:"column:ticket_id;not null;type:uuid"`
	AuthorID       int64          `gorm:"column:author_id;not null"`
	AuthorUsername string         `gorm:"column:author_username;not null"`
	Content        *string        `gorm:"column:content"`
	IsStaffNote    bool           `gorm:"column:is_staff_note;default:false"`
	Attachments    datatypes.JSON `gorm:"column:attachments;type:jsonb"`
	SentAt         time.Time      `gorm:"column:sent_at;autoCreateTime:false;autoUpdateTime:false"`
}

// TableName overrides the default table name mapping.
func (TicketMessage) TableName() string { return "ticket_messages" }

// GetAttachments deserializes the JSONB attachments column.
func (m *TicketMessage) GetAttachments() ([]discordutil.Attachment, error) {
	if len(m.Attachments) == 0 {
		return nil, nil
	}
	var attachments []discordutil.Attachment
	return attachments, json.Unmarshal(m.Attachments, &attachments)
}

// ─── Query structs ────────────────────────────────────────────────────────────

// TicketFilter holds optional filters for listing tickets.
type TicketFilter struct {
	Status     *TicketStatus
	Priority   *TicketPriority
	CategoryID *string
	OpenedBy   *int64
	Page       int
	Limit      int
}

// TicketStats holds computed stats for the dashboard.
type TicketStats struct {
	Open             int64
	Claimed          int64
	ClosedThisMonth  int64
	ArchivedTotal    int64
	AvgResolutionMin int64
	TopCategoryID    string
}
