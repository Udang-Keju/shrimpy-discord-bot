package discordutil

import "testing"

func TestParseComponentEmoji(t *testing.T) {
	cases := []struct {
		in            string
		name, id      string
		animated      bool
	}{
		{"🦐", "🦐", "", false},
		{"<:shrimp:123>", "shrimp", "123", false},
		{"<a:party:456>", "party", "456", true},
		{"shrimp:123", "shrimp", "123", false},
	}
	for _, c := range cases {
		got := ParseComponentEmoji(c.in)
		if got.Name != c.name || got.ID != c.id || got.Animated != c.animated {
			t.Errorf("ParseComponentEmoji(%q) = {Name:%q ID:%q Animated:%v}, want {Name:%q ID:%q Animated:%v}",
				c.in, got.Name, got.ID, got.Animated, c.name, c.id, c.animated)
		}
	}
}

func TestReactionEmojiAPIName(t *testing.T) {
	cases := map[string]string{
		"🦐":              "🦐",
		"<:shrimp:123>":  "shrimp:123",
		"<a:party:456>":  "party:456",
		"shrimp:123":     "shrimp:123",
	}
	for in, want := range cases {
		if got := ReactionEmojiAPIName(in); got != want {
			t.Errorf("ReactionEmojiAPIName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsCustomEmoji(t *testing.T) {
	custom := []string{"<:a:1>", "<a:b:2>", "name:3"}
	for _, s := range custom {
		if !IsCustomEmoji(s) {
			t.Errorf("IsCustomEmoji(%q) = false, want true", s)
		}
	}
	if IsCustomEmoji("🦐") {
		t.Errorf("IsCustomEmoji(unicode) = true, want false")
	}
}
