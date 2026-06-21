package model

import (
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"gorm.io/datatypes"
)

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

// TableName overrides the GORM default table name mapping.
func (ReactionRoleMessage) TableName() string { return "reaction_role_messages" }

// GetMedia deserializes the EmbedMedia JSONB column.
func (r *ReactionRoleMessage) GetMedia() (*discordutil.EmbedMedia, error) {
	return discordutil.DecodeMedia(r.EmbedMedia)
}

// SetMedia serializes an EmbedMedia into the JSONB column.
func (r *ReactionRoleMessage) SetMedia(m *discordutil.EmbedMedia) (err error) {
	r.EmbedMedia, err = discordutil.EncodeMedia(m)
	return
}

// ReactionRoleEmoji maps to the reaction_role_emojis table.
type ReactionRoleEmoji struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	MessageID string    `gorm:"column:message_id;not null;type:uuid"`
	Emoji     string    `gorm:"column:emoji;not null"`
	RoleID    int64     `gorm:"column:role_id;not null"`
	CreatedAt time.Time
}

// TableName overrides the GORM default table name mapping.
func (ReactionRoleEmoji) TableName() string { return "reaction_role_emojis" }
