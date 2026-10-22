package service

import (
	"fmt"
	"sort"
	"strconv"
	"time"
	_ "time/tzdata" // embed zone data so lookups work without system tz files (slim containers)

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/model"
)

// maxChannelName is Discord's channel name length limit.
const maxChannelName = 100

// Format presets, keyed by the value stored in clock_channels.format.
const (
	FormatDateLong   = "date_long"   // Saturday Apr 27th
	FormatDateMedium = "date_medium" // Apr 27th, 2024
	FormatDateISO    = "date_iso"    // 2024-04-27
	FormatTime12h    = "time_12h"    // 3:17am UTC
	FormatTime24h    = "time_24h"    // 15:17 UTC
)

// formatKinds maps each preset to the clock kind it belongs to.
var formatKinds = map[string]string{
	FormatDateLong:   model.KindDate,
	FormatDateMedium: model.KindDate,
	FormatDateISO:    model.KindDate,
	FormatTime12h:    model.KindTime,
	FormatTime24h:    model.KindTime,
}

// defaultEmoji is used when a clock has no custom emoji.
var defaultEmoji = map[string]string{
	model.KindDate: "📅",
	model.KindTime: "🕐",
}

// timezones is the curated set offered in the dashboard dropdown. Only these
// zones are accepted, so a request can't make the worker load arbitrary names.
var timezones = []string{
	"UTC",
	"Pacific/Honolulu",
	"America/Anchorage",
	"America/Los_Angeles",
	"America/Denver",
	"America/Chicago",
	"America/New_York",
	"America/Halifax",
	"America/Sao_Paulo",
	"Atlantic/Reykjavik",
	"Europe/London",
	"Europe/Paris",
	"Europe/Berlin",
	"Europe/Athens",
	"Europe/Istanbul",
	"Europe/Moscow",
	"Africa/Cairo",
	"Africa/Lagos",
	"Africa/Johannesburg",
	"Asia/Dubai",
	"Asia/Karachi",
	"Asia/Kolkata",
	"Asia/Dhaka",
	"Asia/Bangkok",
	"Asia/Jakarta",
	"Asia/Singapore",
	"Asia/Manila",
	"Asia/Shanghai",
	"Asia/Tokyo",
	"Asia/Seoul",
	"Australia/Perth",
	"Australia/Sydney",
	"Pacific/Auckland",
}

// TimezoneOption is one dropdown entry: an IANA zone with its GMT offset at a
// given instant.
type TimezoneOption struct {
	ID            string `json:"id"`            // IANA name, e.g. "Asia/Jakarta"
	OffsetMinutes int    `json:"offset_minutes"` // minutes east of GMT at the reference instant
	Label         string `json:"label"`         // "(GMT+07:00) Asia/Jakarta"
}

// TimezoneOptions returns the allowed zones with their GMT offset at now,
// sorted from the most negative offset to the most positive (ties by name).
// The offset is computed per call so DST shifts are reflected.
func TimezoneOptions(now time.Time) []TimezoneOption {
	opts := make([]TimezoneOption, 0, len(timezones))
	for _, tz := range timezones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			continue
		}
		_, offsetSec := now.In(loc).Zone()
		opts = append(opts, TimezoneOption{
			ID:            tz,
			OffsetMinutes: offsetSec / 60,
			Label:         fmt.Sprintf("(%s) %s", gmtLabel(offsetSec/60), tz),
		})
	}
	sort.SliceStable(opts, func(i, j int) bool {
		if opts[i].OffsetMinutes != opts[j].OffsetMinutes {
			return opts[i].OffsetMinutes < opts[j].OffsetMinutes
		}
		return opts[i].ID < opts[j].ID
	})
	return opts
}

// gmtLabel renders minutes east of GMT as "GMT+07:00" / "GMT-03:30".
func gmtLabel(offsetMinutes int) string {
	sign := '+'
	if offsetMinutes < 0 {
		sign = '-'
		offsetMinutes = -offsetMinutes
	}
	return fmt.Sprintf("GMT%c%02d:%02d", sign, offsetMinutes/60, offsetMinutes%60)
}

// IsValidTimezone reports whether tz is one of the allowed dropdown zones.
func IsValidTimezone(tz string) bool {
	for _, z := range timezones {
		if z == tz {
			return true
		}
	}
	return false
}

// FormatKind returns the clock kind a format preset belongs to.
func FormatKind(format string) (string, bool) {
	kind, ok := formatKinds[format]
	return kind, ok
}

// FormatName builds the channel name for a clock at the given instant. emoji
// may be empty to use the default for the kind.
func FormatName(kind, format, tz, emoji string, now time.Time) (string, error) {
	if k, ok := formatKinds[format]; !ok || k != kind {
		return "", fmt.Errorf("format %q is not valid for kind %q", format, kind)
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", fmt.Errorf("load timezone %q: %w", tz, err)
	}
	t := now.In(loc)

	var text string
	switch format {
	case FormatDateLong:
		text = fmt.Sprintf("%s %s %s", t.Format("Monday"), t.Format("Jan"), ordinal(t.Day()))
	case FormatDateMedium:
		text = fmt.Sprintf("%s %s, %d", t.Format("Jan"), ordinal(t.Day()), t.Year())
	case FormatDateISO:
		text = t.Format("2006-01-02")
	case FormatTime12h:
		text = fmt.Sprintf("%s %s", lowerMeridiem(t.Format("3:04PM")), t.Format("MST"))
	case FormatTime24h:
		text = fmt.Sprintf("%s %s", t.Format("15:04"), t.Format("MST"))
	}

	if emoji == "" {
		emoji = defaultEmoji[kind]
	}
	return truncateName(emoji + " " + text), nil
}

// ordinal renders a day of month with its English suffix (1st, 2nd, 11th...).
func ordinal(n int) string {
	suffix := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return strconv.Itoa(n) + suffix
}

// lowerMeridiem turns "3:17PM" into "3:17pm".
func lowerMeridiem(s string) string {
	if len(s) < 2 {
		return s
	}
	return s[:len(s)-2] + string([]byte{s[len(s)-2] + 32, s[len(s)-1] + 32})
}

// truncateName clips a name to Discord's limit without splitting a rune.
func truncateName(name string) string {
	r := []rune(name)
	if len(r) <= maxChannelName {
		return name
	}
	return string(r[:maxChannelName])
}
