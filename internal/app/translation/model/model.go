package model

import "time"

// Reaction delivery modes: where a reaction-triggered translation is sent.
const (
	ReactionDeliveryChannel = "channel" // reply in the channel, visible to everyone
	ReactionDeliveryDM      = "dm"      // DM the reacting user; falls back to channel if DMs are closed
)

// TranslationConfig maps to the translation_config table. It holds the
// per-guild translation feature settings, including the (encrypted) engine
// credentials.
type TranslationConfig struct {
	GuildID          int64   `gorm:"primaryKey;column:guild_id;autoIncrement:false"`
	Enabled          bool    `gorm:"column:enabled"`          // master feature toggle
	AutoEnabled      bool    `gorm:"column:auto_enabled"`     // auto-translate in configured channels
	ReactionEnabled  bool    `gorm:"column:reaction_enabled"` // translate on configured emoji reaction
	ReactionDelivery string  `gorm:"column:reaction_delivery;default:'channel'"` // "channel" or "dm"
	Provider         string  `gorm:"column:provider;default:'deepl'"`
	APIKeyEnc        []byte  `gorm:"column:api_key_enc"`  // AES-256-GCM encrypted engine API key
	EndpointURL      *string `gorm:"column:endpoint_url"` // for self-hosted engines (LibreTranslate)
	TargetLang       *string `gorm:"column:target_lang"`  // nil = fall back to guilds.language
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// TableName overrides the GORM default table name mapping.
func (TranslationConfig) TableName() string { return "translation_config" }

// TranslationChannel maps to the translation_channels table — the allowlist of
// channels whose member messages are auto-translated.
type TranslationChannel struct {
	ID                 string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	GuildID            int64   `gorm:"column:guild_id;not null"`
	ChannelID          int64   `gorm:"column:channel_id;not null"`
	TargetLangOverride *string `gorm:"column:target_lang_override"`
	CreatedAt          time.Time
}

// TableName overrides the GORM default table name mapping.
func (TranslationChannel) TableName() string { return "translation_channels" }

// TranslationReactionEmoji maps to the translation_reaction_emojis table — the
// emojis that trigger translation of a message when reacted to it.
type TranslationReactionEmoji struct {
	ID                 string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	GuildID            int64   `gorm:"column:guild_id;not null"`
	Emoji              string  `gorm:"column:emoji;not null"` // unicode char or "name:id" for custom
	TargetLangOverride *string `gorm:"column:target_lang_override"`
	CreatedAt          time.Time
}

// TableName overrides the GORM default table name mapping.
func (TranslationReactionEmoji) TableName() string { return "translation_reaction_emojis" }
