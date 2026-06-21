package model

import "time"

// User maps to the users table in database.
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

// TableName returns the table name GORM uses.
func (User) TableName() string { return "users" }
