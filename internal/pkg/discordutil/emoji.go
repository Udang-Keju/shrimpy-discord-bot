package discordutil

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
)

// customEmojiMentionRe matches the Discord custom-emoji mention forms <:name:id>
// and <a:name:id> (animated). Capture groups: 1=animated flag, 2=name, 3=id.
var customEmojiMentionRe = regexp.MustCompile(`^<(a)?:([a-zA-Z0-9_]+):(\d+)>$`)

// reactionKeyRe matches the reaction API form name:id — the shape returned by
// discordgo's (*Emoji).APIName for custom emoji and expected by the reaction
// endpoints. Capture groups: 1=name, 2=id.
var reactionKeyRe = regexp.MustCompile(`^([a-zA-Z0-9_]+):(\d+)$`)

// CustomEmojiURL builds the CDN URL for a custom emoji by ID.
func CustomEmojiURL(id string, animated bool) string {
	ext := "png"
	if animated {
		ext = "gif"
	}
	return "https://cdn.discordapp.com/emojis/" + id + "." + ext
}

// CustomEmojiMention builds the canonical Discord mention form for a custom emoji
// (<:name:id> or <a:name:id>). This is the format the dashboard emoji picker emits
// and that ParseComponentEmoji round-trips.
func CustomEmojiMention(name, id string, animated bool) string {
	if animated {
		return "<a:" + name + ":" + id + ">"
	}
	return "<:" + name + ":" + id + ">"
}

// ParseComponentEmoji converts a stored emoji string into a discordgo.ComponentEmoji
// suitable for buttons and select-menu options. It accepts the custom-emoji mention
// forms (<:name:id> / <a:name:id>), the reaction API form (name:id), or a raw unicode
// emoji, which is returned as a name-only ComponentEmoji.
func ParseComponentEmoji(s string) discordgo.ComponentEmoji {
	if m := customEmojiMentionRe.FindStringSubmatch(s); m != nil {
		return discordgo.ComponentEmoji{Name: m[2], ID: m[3], Animated: m[1] == "a"}
	}
	if m := reactionKeyRe.FindStringSubmatch(s); m != nil {
		return discordgo.ComponentEmoji{Name: m[1], ID: m[2]}
	}
	return discordgo.ComponentEmoji{Name: s}
}

// ReactionEmojiAPIName converts a stored emoji string into the form the reaction
// endpoints (Session.MessageReactionAdd/Remove) expect and that gateway reaction
// events match against (discordgo (*Emoji).APIName): name:id for custom emoji, or
// the raw unicode emoji unchanged. The animated <a:...> prefix is intentionally
// dropped because the reaction API keys custom emoji by name:id regardless of
// animation.
func ReactionEmojiAPIName(s string) string {
	if m := customEmojiMentionRe.FindStringSubmatch(s); m != nil {
		return m[2] + ":" + m[3]
	}
	return s
}

// IsCustomEmoji reports whether s denotes a Discord custom emoji in either the
// mention form (<:name:id> / <a:name:id>) or the reaction API form (name:id).
func IsCustomEmoji(s string) bool {
	return customEmojiMentionRe.MatchString(s) || reactionKeyRe.MatchString(s)
}
