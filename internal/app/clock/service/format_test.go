package service

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/model"
	"github.com/stretchr/testify/assert"
)

func TestFormatName(t *testing.T) {
	// Saturday, 27 Apr 2024 03:17 UTC.
	now := time.Date(2024, time.April, 27, 3, 17, 0, 0, time.UTC)

	tests := []struct {
		name    string
		kind    string
		format  string
		tz      string
		emoji   string
		want    string
		wantErr bool
	}{
		{name: "date long utc", kind: model.KindDate, format: FormatDateLong, tz: "UTC", want: "📅 Saturday Apr 27th"},
		{name: "date medium", kind: model.KindDate, format: FormatDateMedium, tz: "UTC", want: "📅 Apr 27th, 2024"},
		{name: "date iso", kind: model.KindDate, format: FormatDateISO, tz: "UTC", want: "📅 2024-04-27"},
		{name: "time 12h utc", kind: model.KindTime, format: FormatTime12h, tz: "UTC", want: "🕐 3:17am UTC"},
		{name: "time 24h utc", kind: model.KindTime, format: FormatTime24h, tz: "UTC", want: "🕐 03:17 UTC"},
		{name: "date stays same day ahead of utc", kind: model.KindDate, format: FormatDateLong, tz: "Asia/Tokyo", want: "📅 Saturday Apr 27th"},
		{name: "date rolls back behind utc", kind: model.KindDate, format: FormatDateLong, tz: "America/New_York", want: "📅 Friday Apr 26th"},
		{name: "time uses dst abbreviation", kind: model.KindTime, format: FormatTime12h, tz: "America/New_York", want: "🕐 11:17pm EDT"},
		{name: "custom emoji", kind: model.KindTime, format: FormatTime24h, tz: "UTC", emoji: "⏰", want: "⏰ 03:17 UTC"},
		{name: "format of the wrong kind", kind: model.KindDate, format: FormatTime12h, tz: "UTC", wantErr: true},
		{name: "unknown format", kind: model.KindDate, format: "nope", tz: "UTC", wantErr: true},
		{name: "unknown timezone", kind: model.KindDate, format: FormatDateLong, tz: "Not/AZone", wantErr: true},
		{
			name: "name clipped to discord limit", kind: model.KindDate, format: FormatDateISO, tz: "UTC",
			emoji: strings.Repeat("x", 200), want: strings.Repeat("x", maxChannelName),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatName(tt.kind, tt.format, tt.tz, tt.emoji, now)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOrdinal(t *testing.T) {
	tests := []struct {
		day  int
		want string
	}{
		{1, "1st"}, {2, "2nd"}, {3, "3rd"}, {4, "4th"},
		{11, "11th"}, {12, "12th"}, {13, "13th"},
		{21, "21st"}, {22, "22nd"}, {23, "23rd"},
		{30, "30th"}, {31, "31st"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, ordinal(tt.day))
		})
	}
}

func TestTimezoneOptions(t *testing.T) {
	tests := []struct {
		name      string
		now       time.Time
		wantLabel map[string]string // zone ID -> expected label
	}{
		{
			name: "northern summer",
			now:  time.Date(2024, time.April, 27, 3, 17, 0, 0, time.UTC),
			wantLabel: map[string]string{
				"UTC":              "(GMT+00:00) UTC",
				"Asia/Jakarta":     "(GMT+07:00) Asia/Jakarta",
				"Asia/Kolkata":     "(GMT+05:30) Asia/Kolkata",
				"America/New_York": "(GMT-04:00) America/New_York",
			},
		},
		{
			name: "northern winter",
			now:  time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC),
			wantLabel: map[string]string{
				"UTC":              "(GMT+00:00) UTC",
				"Asia/Jakarta":     "(GMT+07:00) Asia/Jakarta",
				"America/New_York": "(GMT-05:00) America/New_York",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := TimezoneOptions(tt.now)

			assert.Len(t, opts, len(timezones))
			assert.True(t, sort.SliceIsSorted(opts, func(i, j int) bool {
				return opts[i].OffsetMinutes < opts[j].OffsetMinutes
			}), "options must be sorted by GMT offset")

			labels := make(map[string]string, len(opts))
			for _, o := range opts {
				labels[o.ID] = o.Label
			}
			for id, want := range tt.wantLabel {
				assert.Equal(t, want, labels[id], id)
			}
		})
	}
}

func TestIsValidTimezone(t *testing.T) {
	tests := []struct {
		tz   string
		want bool
	}{
		{"UTC", true},
		{"Asia/Jakarta", true},
		{"Mars/Olympus", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.tz, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidTimezone(tt.tz))
		})
	}
}
