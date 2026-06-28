package discordutil

import "github.com/bwmarrin/discordgo"

// EmbedFields holds the raw (unreplaced) embed configuration fields shared by
// any feature that lets admins configure an optional embed alongside plain text
// (ticket panels, ticket category greetings, etc).
type EmbedFields struct {
	Title       *string
	Description *string
	Color       *int32
	Media       *EmbedMedia
}

// HasContent reports whether any of the embed fields are actually set.
func (f EmbedFields) HasContent() bool {
	return (f.Title != nil && *f.Title != "") ||
		(f.Description != nil && *f.Description != "") ||
		f.Color != nil ||
		f.Media != nil
}

// BuildContentAndEmbed implements the shared "plain text + optional embed" rules:
//   - text only, no embed fields  -> (text, nil)
//   - text + embed fields         -> (text, embed)
//   - no text, embed fields only  -> ("", embed)
//   - neither                     -> ("", nil)
//
// replaceVars is applied to text, title, description, and the media author/footer text.
func BuildContentAndEmbed(plainText *string, fields EmbedFields, replaceVars func(string) string) (content string, embed *discordgo.MessageEmbed) {
	if plainText != nil && *plainText != "" {
		content = replaceVars(*plainText)
	}

	if !fields.HasContent() {
		return content, nil
	}

	embed = &discordgo.MessageEmbed{}
	if fields.Title != nil && *fields.Title != "" {
		embed.Title = replaceVars(*fields.Title)
	}
	if fields.Description != nil && *fields.Description != "" {
		embed.Description = replaceVars(*fields.Description)
	}
	if fields.Color != nil {
		embed.Color = int(*fields.Color)
	}
	if fields.Media != nil {
		m := fields.Media
		if m.Author != nil {
			embed.Author = &discordgo.MessageEmbedAuthor{Name: replaceVars(m.Author.Name)}
			if m.Author.IconURL != nil {
				embed.Author.IconURL = *m.Author.IconURL
			}
			if m.Author.URL != nil {
				embed.Author.URL = *m.Author.URL
			}
		}
		if m.Thumbnail != nil {
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: m.Thumbnail.URL}
		}
		if m.Image != nil {
			embed.Image = &discordgo.MessageEmbedImage{URL: m.Image.URL}
		}
		if m.Footer != nil {
			embed.Footer = &discordgo.MessageEmbedFooter{Text: replaceVars(m.Footer.Text)}
			if m.Footer.IconURL != nil {
				embed.Footer.IconURL = *m.Footer.IconURL
			}
		}
	}

	return content, embed
}
