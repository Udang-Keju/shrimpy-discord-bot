package discordutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptr(s string) *string { return &s }

func TestEmbedFields_HasContent(t *testing.T) {
	assert.False(t, EmbedFields{}.HasContent())
	assert.True(t, EmbedFields{Title: ptr("t")}.HasContent())
	assert.True(t, EmbedFields{Description: ptr("d")}.HasContent())
	color := int32(5)
	assert.True(t, EmbedFields{Color: &color}.HasContent())
	assert.True(t, EmbedFields{Media: &EmbedMedia{}}.HasContent())
	assert.False(t, EmbedFields{Title: ptr(""), Description: ptr("")}.HasContent())
}

func identity(s string) string { return s }

func TestBuildContentAndEmbed_TextOnly(t *testing.T) {
	content, embed := BuildContentAndEmbed(ptr("hello"), EmbedFields{}, identity)
	assert.Equal(t, "hello", content)
	assert.Nil(t, embed)
}

func TestBuildContentAndEmbed_EmbedOnly(t *testing.T) {
	fields := EmbedFields{Title: ptr("Title"), Description: ptr("Desc")}
	content, embed := BuildContentAndEmbed(nil, fields, identity)
	assert.Equal(t, "", content)
	assert.NotNil(t, embed)
	assert.Equal(t, "Title", embed.Title)
	assert.Equal(t, "Desc", embed.Description)
}

func TestBuildContentAndEmbed_TextAndEmbed(t *testing.T) {
	fields := EmbedFields{Title: ptr("Title")}
	content, embed := BuildContentAndEmbed(ptr("hello"), fields, identity)
	assert.Equal(t, "hello", content)
	assert.NotNil(t, embed)
	assert.Equal(t, "Title", embed.Title)
}

func TestBuildContentAndEmbed_Neither(t *testing.T) {
	content, embed := BuildContentAndEmbed(nil, EmbedFields{}, identity)
	assert.Equal(t, "", content)
	assert.Nil(t, embed)
}

func TestBuildContentAndEmbed_ReplaceVarsApplied(t *testing.T) {
	upper := func(s string) string { return strings.ToUpper(s) }
	icon := "icon-url"
	fields := EmbedFields{
		Title:       ptr("title"),
		Description: ptr("desc"),
		Media: &EmbedMedia{
			Author: &EmbedAuthor{Name: "author", IconURL: &icon},
			Footer: &EmbedFooter{Text: "footer"},
		},
	}
	content, embed := BuildContentAndEmbed(ptr("text"), fields, upper)
	assert.Equal(t, "TEXT", content)
	assert.Equal(t, "TITLE", embed.Title)
	assert.Equal(t, "DESC", embed.Description)
	assert.Equal(t, "AUTHOR", embed.Author.Name)
	assert.Equal(t, "FOOTER", embed.Footer.Text)
}
