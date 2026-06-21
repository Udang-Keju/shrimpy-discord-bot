package model

import "time"

// DiscordApp maps to the discord_apps table.
type DiscordApp struct {
	ID                     string    `gorm:"primaryKey;column:id;type:uuid;default:gen_random_uuid()"`
	Name                   string    `gorm:"column:name"`
	DiscordTokenEnc        []byte    `gorm:"column:discord_token_enc"`
	DiscordClientID        string    `gorm:"column:discord_client_id;unique"`
	DiscordClientSecretEnc []byte    `gorm:"column:discord_client_secret_enc"`
	DiscordRedirectURI     string    `gorm:"column:discord_redirect_uri"`
	CreatedAt              time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt              time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName returns the GORM table name.
func (DiscordApp) TableName() string { return "discord_apps" }
