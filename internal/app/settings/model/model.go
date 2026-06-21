package model

import "time"

// BotSettings maps to the bot_settings singleton table.
// There is always exactly one row (id = 1) in this table.
// All sensitive values are AES-256-GCM encrypted at rest.
type BotSettings struct {
	ID                       int16     `gorm:"primaryKey;column:id;default:1"`
	DiscordTokenEnc          []byte    `gorm:"column:discord_token_enc"`
	DiscordClientID          string    `gorm:"column:discord_client_id"`
	DiscordClientSecretEnc   []byte    `gorm:"column:discord_client_secret_enc"`
	DiscordRedirectURI       string    `gorm:"column:discord_redirect_uri"`
	UpdatedAt                time.Time `gorm:"column:updated_at;autoUpdateTime:false"`
}

// TableName returns the GORM table name.
func (BotSettings) TableName() string { return "bot_settings" }
