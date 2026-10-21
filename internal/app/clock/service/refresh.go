package service

import "time"

// DefaultRefreshMinutes is used when a request does not choose an interval.
// It is also the shortest option, which keeps each channel under Discord's
// limit of about 2 renames per 10 minutes.
const DefaultRefreshMinutes = 5

// RefreshOptions are the refresh intervals, in minutes, an admin can choose.
var RefreshOptions = []int{5, 10, 15, 30, 60}

// normalizeRefresh maps 0 to the default and rejects values outside
// RefreshOptions.
func normalizeRefresh(minutes int) (int, error) {
	if minutes == 0 {
		return DefaultRefreshMinutes, nil
	}
	for _, o := range RefreshOptions {
		if o == minutes {
			return minutes, nil
		}
	}
	return 0, invalid("refresh interval must be one of %v minutes", RefreshOptions)
}

// refreshDue reports whether a clock should be refreshed now. lastRun is the
// zero time for a clock that has not run since startup. The slack absorbs
// ticker jitter so a 10-minute clock isn't pushed to the tick after.
func refreshDue(lastRun, now time.Time, minutes int) bool {
	if lastRun.IsZero() {
		return true
	}
	if minutes == 0 {
		minutes = DefaultRefreshMinutes
	}
	interval := time.Duration(minutes) * time.Minute
	return now.Sub(lastRun) >= interval-schedulerSlack
}

const schedulerSlack = 30 * time.Second
