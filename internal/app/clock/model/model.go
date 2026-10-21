package model

import "time"

// Clock kinds: what a clock channel displays.
const (
	KindDate = "date"
	KindTime = "time"
)

// ClockChannel maps to the clock_channels table — a channel whose name is
// periodically rewritten to show the current date or time in a timezone.
type ClockChannel struct {
	ID        string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	GuildID   int64   `gorm:"column:guild_id;not null"`
	ChannelID int64   `gorm:"column:channel_id;not null"`
	Kind      string  `gorm:"column:kind;not null"`
	Format    string  `gorm:"column:format;not null"`
	Timezone  string  `gorm:"column:timezone;not null;default:'UTC'"`
	Emoji     *string `gorm:"column:emoji"`     // nil = default emoji for the kind
	Managed   bool    `gorm:"column:managed"`   // true if the bot created the channel
	LastName  string  `gorm:"column:last_name"` // last name written, to skip no-op renames
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName overrides the GORM default table name mapping.
func (ClockChannel) TableName() string { return "clock_channels" }
