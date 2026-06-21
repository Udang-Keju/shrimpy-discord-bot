package model

import "time"

// Guild maps to the guilds table.
type Guild struct {
	GuildID      int64   `gorm:"primaryKey;column:guild_id;autoIncrement:false"`
	DiscordAppID *string `gorm:"column:discord_app_id;type:uuid"`
	Prefix       string  `gorm:"column:prefix;default:'!'"`
	Language     string  `gorm:"column:language;default:'en'"`
	BotNickname  *string `gorm:"column:bot_nickname"`
	LogChannelID *int64  `gorm:"column:log_channel_id"`
	IsActive     bool    `gorm:"column:is_active;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName overrides the GORM default table name mapping.
func (Guild) TableName() string { return "guilds" }

// StaffRole maps to the staff_roles table.
type StaffRole struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	GuildID   int64     `gorm:"column:guild_id;not null"`
	RoleID    int64     `gorm:"column:role_id;not null"`
	CreatedAt time.Time
}

// TableName overrides the GORM default table name mapping.
func (StaffRole) TableName() string { return "staff_roles" }

// AutoRole maps to the auto_roles table.
type AutoRole struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	GuildID   int64     `gorm:"column:guild_id;not null"`
	RoleID    int64     `gorm:"column:role_id;not null"`
	CreatedAt time.Time
}

// TableName overrides the GORM default table name mapping.
func (AutoRole) TableName() string { return "auto_roles" }
