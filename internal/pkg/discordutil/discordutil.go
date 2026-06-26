package discordutil

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/datatypes"
)

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

// GuildIconURL builds the CDN URL for a guild icon hash, as returned by
// Discord's REST/OAuth2 APIs. Returns "" if hash is empty.
func GuildIconURL(guildID, hash string) string {
	if hash == "" {
		return ""
	}
	ext := "png"
	if strings.HasPrefix(hash, "a_") {
		ext = "gif"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.%s", guildID, hash, ext)
}

// UserAvatarURL builds the CDN URL for a user avatar hash, as returned by
// Discord's REST/OAuth2 APIs. Returns "" if hash is empty.
func UserAvatarURL(userID, hash string) string {
	if hash == "" {
		return ""
	}
	ext := "png"
	if strings.HasPrefix(hash, "a_") {
		ext = "gif"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.%s", userID, hash, ext)
}

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
