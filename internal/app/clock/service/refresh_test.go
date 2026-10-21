package service

import (
	"context"
	"testing"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/model"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeRefresh(t *testing.T) {
	tests := []struct {
		name    string
		in      int
		want    int
		wantErr bool
	}{
		{"zero uses the default", 0, 5, false},
		{"5 minutes", 5, 5, false},
		{"10 minutes", 10, 10, false},
		{"15 minutes", 15, 15, false},
		{"30 minutes", 30, 30, false},
		{"1 hour", 60, 60, false},
		{"below the minimum", 1, 0, true},
		{"between options", 20, 0, true},
		{"above the maximum", 120, 0, true},
		{"negative", -5, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeRefresh(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRefreshDue(t *testing.T) {
	now := time.Date(2024, time.April, 27, 3, 17, 0, 0, time.UTC)

	tests := []struct {
		name    string
		lastRun time.Time
		minutes int
		want    bool
	}{
		{"never ran", time.Time{}, 60, true},
		{"5m clock, 5m ago", now.Add(-5 * time.Minute), 5, true},
		{"5m clock, 5m ago with ticker jitter", now.Add(-4*time.Minute - 50*time.Second), 5, true},
		{"10m clock, 5m ago", now.Add(-5 * time.Minute), 10, false},
		{"10m clock, 10m ago", now.Add(-10 * time.Minute), 10, true},
		{"1h clock, 55m ago", now.Add(-55 * time.Minute), 60, false},
		{"1h clock, 1h ago", now.Add(-time.Hour), 60, true},
		{"zero minutes behaves like the default", now.Add(-5 * time.Minute), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, refreshDue(tt.lastRun, now, tt.minutes))
		})
	}
}

func TestScheduler_RespectsRefreshInterval(t *testing.T) {
	start := time.Date(2024, time.April, 27, 3, 0, 0, 0, time.UTC)
	repo := &fakeRepo{
		clocks: []model.ClockChannel{
			{ID: "fast", GuildID: 1, ChannelID: 10, Kind: model.KindTime, Format: FormatTime24h, Timezone: "UTC", RefreshMinutes: 5},
			{ID: "slow", GuildID: 1, ChannelID: 11, Kind: model.KindTime, Format: FormatTime24h, Timezone: "UTC", RefreshMinutes: 15},
		},
		lastNames: map[string]string{},
	}
	editor := &fakeEditor{renames: map[string]string{}}
	sched := NewScheduler(repo, DefaultRefreshInterval)
	sessionFor := func(context.Context, int64) (channelEditor, error) { return editor, nil }

	// Offset from start -> channels expected to be renamed at that tick.
	tests := []struct {
		offset time.Duration
		want   map[string]string
	}{
		{0, map[string]string{"10": "🕐 03:00 UTC", "11": "🕐 03:00 UTC"}},
		{5 * time.Minute, map[string]string{"10": "🕐 03:05 UTC"}},
		{10 * time.Minute, map[string]string{"10": "🕐 03:10 UTC"}},
		{15 * time.Minute, map[string]string{"10": "🕐 03:15 UTC", "11": "🕐 03:15 UTC"}},
	}
	for _, tt := range tests {
		editor.renames = map[string]string{}
		sched.now = func() time.Time { return start.Add(tt.offset) }
		sched.runCheck(context.Background(), sessionFor)
		assert.Equal(t, tt.want, editor.renames, "at +%s", tt.offset)
	}
}
