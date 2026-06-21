package repository

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ─── Database handle ──────────────────────────────────────────────────────────

// DB is the application-level GORM handle. All repositories receive a pointer
// to this so they can scope queries, use transactions, etc.
type DB = gorm.DB

// ─── Discord ID helpers ───────────────────────────────────────────────────────

// ParseID converts a Discord snowflake string to int64.
func ParseID(s string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid snowflake %q: %w", s, err)
	}
	return n, nil
}

// FormatID converts an int64 snowflake back to its string representation.
func FormatID(n int64) string {
	return strconv.FormatInt(n, 10)
}

// ─── JSONB helpers ────────────────────────────────────────────────────────────

// EmbedMedia holds optional visual fields for Discord embeds, stored as JSONB.
type EmbedMedia struct {
	Author    *EmbedAuthor    `json:"author,omitempty"`
	Thumbnail *EmbedThumbnail `json:"thumbnail,omitempty"`
	Image     *EmbedImage     `json:"image,omitempty"`
	Footer    *EmbedFooter    `json:"footer,omitempty"`
}

type EmbedAuthor struct {
	Name    string  `json:"name"`
	IconURL *string `json:"iconUrl,omitempty"`
	URL     *string `json:"url,omitempty"`
}

type EmbedThumbnail struct {
	URL string `json:"url"`
}

type EmbedImage struct {
	URL string `json:"url"`
}

type EmbedFooter struct {
	Text    string  `json:"text"`
	IconURL *string `json:"iconUrl,omitempty"`
}

// Attachment represents a file attached to a ticket message, stored as JSONB.
type Attachment struct {
	Filename string `json:"filename"`
	URL      string `json:"url"`
	Size     int    `json:"size"`
}

// DecodeMedia unmarshals a datatypes.JSON JSONB column into an *EmbedMedia.
// Returns nil, nil when the column is NULL or empty.
func DecodeMedia(raw datatypes.JSON) (*EmbedMedia, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var m EmbedMedia
	return &m, json.Unmarshal(raw, &m)
}

// EncodeMedia marshals an *EmbedMedia into datatypes.JSON for GORM storage.
// Returns nil when m is nil.
func EncodeMedia(m *EmbedMedia) (datatypes.JSON, error) {
	if m == nil {
		return nil, nil
	}
	b, err := json.Marshal(m)
	return datatypes.JSON(b), err
}

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

// ─── GORM Domain Models ───────────────────────────────────────────────────────

// Guild maps to the guilds table.
type Guild struct {
	GuildID      int64   `gorm:"primaryKey;column:guild_id;autoIncrement:false"`
	Prefix       string  `gorm:"column:prefix;default:'!'"`
	Language     string  `gorm:"column:language;default:'en'"`
	BotNickname  *string `gorm:"column:bot_nickname"`
	LogChannelID *int64  `gorm:"column:log_channel_id"`
	IsActive     bool    `gorm:"column:is_active;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Guild) TableName() string { return "guilds" }

// User maps to the users table.
type User struct {
	UserID                 int64      `gorm:"primaryKey;column:user_id;autoIncrement:false"`
	Username               string     `gorm:"column:username"`
	Discriminator          *string    `gorm:"column:discriminator"`
	AvatarHash             *string    `gorm:"column:avatar_hash"`
	DiscordAccessTokenEnc  []byte     `gorm:"column:discord_access_token_enc"`
	DiscordRefreshTokenEnc []byte     `gorm:"column:discord_refresh_token_enc"`
	TokenExpiresAt         *time.Time `gorm:"column:token_expires_at"`
	LastSeen               time.Time  `gorm:"column:last_seen;autoCreateTime:false;autoUpdateTime:false"`
	CreatedAt              time.Time
}

func (User) TableName() string { return "users" }

// StaffRole maps to the staff_roles table.
type StaffRole struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	GuildID   int64     `gorm:"column:guild_id;not null"`
	RoleID    int64     `gorm:"column:role_id;not null"`
	CreatedAt time.Time
}

func (StaffRole) TableName() string { return "staff_roles" }

// AutoRole maps to the auto_roles table.
type AutoRole struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	GuildID   int64     `gorm:"column:guild_id;not null"`
	RoleID    int64     `gorm:"column:role_id;not null"`
	CreatedAt time.Time
}

func (AutoRole) TableName() string { return "auto_roles" }

// WelcomeConfig maps to the welcome_config table.
type WelcomeConfig struct {
	GuildID        int64          `gorm:"primaryKey;column:guild_id;autoIncrement:false"`
	Enabled        bool           `gorm:"column:enabled;default:true"`
	DMMessage      *string        `gorm:"column:dm_message"`
	ChannelID      *int64         `gorm:"column:channel_id"`
	ChannelMessage *string        `gorm:"column:channel_message"`
	EmbedColor     *int32         `gorm:"column:embed_color"`
	EmbedMedia     datatypes.JSON `gorm:"column:embed_media;type:jsonb"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (WelcomeConfig) TableName() string { return "welcome_config" }

// GetMedia deserializes the EmbedMedia JSONB column.
func (w *WelcomeConfig) GetMedia() (*EmbedMedia, error) {
	return DecodeMedia(w.EmbedMedia)
}

// SetMedia serializes an EmbedMedia into the JSONB column.
func (w *WelcomeConfig) SetMedia(m *EmbedMedia) (err error) {
	w.EmbedMedia, err = EncodeMedia(m)
	return
}

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

func (TicketPanel) TableName() string { return "ticket_panels" }

func (p *TicketPanel) GetMedia() (*EmbedMedia, error) { return DecodeMedia(p.EmbedMedia) }
func (p *TicketPanel) SetMedia(m *EmbedMedia) (err error) {
	p.EmbedMedia, err = EncodeMedia(m)
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

func (TicketCategory) TableName() string { return "ticket_categories" }

func (c *TicketCategory) GetOpenMedia() (*EmbedMedia, error) { return DecodeMedia(c.TicketOpenMedia) }
func (c *TicketCategory) SetOpenMedia(m *EmbedMedia) (err error) {
	c.TicketOpenMedia, err = EncodeMedia(m)
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
	// Associations (not loaded by default)
	Category TicketCategory `gorm:"foreignKey:CategoryID"`
}

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
	// No UpdatedAt — messages are append-only
}

func (TicketMessage) TableName() string { return "ticket_messages" }

// GetAttachments deserializes the JSONB attachments column.
func (m *TicketMessage) GetAttachments() ([]Attachment, error) {
	if len(m.Attachments) == 0 {
		return nil, nil
	}
	var attachments []Attachment
	return attachments, json.Unmarshal(m.Attachments, &attachments)
}

// ReactionRoleMessage maps to the reaction_role_messages table.
type ReactionRoleMessage struct {
	ID               string         `gorm:"primaryKey;type:uuid"`
	GuildID          int64          `gorm:"column:guild_id;not null"`
	ChannelID        int64          `gorm:"column:channel_id;not null"`
	MessageID        *int64         `gorm:"column:message_id"`
	EmbedTitle       *string        `gorm:"column:embed_title"`
	EmbedDescription *string        `gorm:"column:embed_description"`
	EmbedColor       *int32         `gorm:"column:embed_color"`
	EmbedMedia       datatypes.JSON `gorm:"column:embed_media;type:jsonb"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	// Associations
	Emojis []ReactionRoleEmoji `gorm:"foreignKey:MessageID"`
}

func (ReactionRoleMessage) TableName() string { return "reaction_role_messages" }

// ReactionRoleEmoji maps to the reaction_role_emojis table.
type ReactionRoleEmoji struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	MessageID string    `gorm:"column:message_id;not null;type:uuid"`
	Emoji     string    `gorm:"column:emoji;not null"`
	RoleID    int64     `gorm:"column:role_id;not null"`
	CreatedAt time.Time
}

func (ReactionRoleEmoji) TableName() string { return "reaction_role_emojis" }

// ─── Sentinel errors ──────────────────────────────────────────────────────────

var (
	ErrNotFound     = fmt.Errorf("record not found")
	ErrLimitReached = fmt.Errorf("limit reached")
)
