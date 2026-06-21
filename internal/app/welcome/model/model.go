package model

import (
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"gorm.io/datatypes"
)

// WelcomeConfig maps to the welcome_config table.
type WelcomeConfig struct {
	GuildID        int64          `gorm:"primaryKey;column:guild_id;autoIncrement:false"`
	Enabled        bool           `gorm:"column:enabled"`
	DMMessage      *string        `gorm:"column:dm_message"`
	ChannelID      *int64         `gorm:"column:channel_id"`
	ChannelMessage *string        `gorm:"column:channel_message"`
	EmbedColor     *int32         `gorm:"column:embed_color"`
	EmbedMedia     datatypes.JSON `gorm:"column:embed_media;type:jsonb"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TableName overrides the default table name mapping.
func (WelcomeConfig) TableName() string { return "welcome_config" }

// GetMedia deserializes the EmbedMedia JSONB column.
func (w *WelcomeConfig) GetMedia() (*discordutil.EmbedMedia, error) {
	return discordutil.DecodeMedia(w.EmbedMedia)
}

// SetMedia serializes an EmbedMedia into the JSONB column.
func (w *WelcomeConfig) SetMedia(m *discordutil.EmbedMedia) (err error) {
	w.EmbedMedia, err = discordutil.EncodeMedia(m)
	return
}
