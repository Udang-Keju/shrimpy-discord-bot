package bot

import (
	"errors"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

// newTestWatchdog returns a watchdog over one fake session whose health and reopen calls the
// test controls.
func newTestWatchdog(healthy *bool, reopenErr *error, reopens *int) *watchdog {
	s := &discordgo.Session{}
	return &watchdog{
		sessions: func() map[string]*discordgo.Session { return map[string]*discordgo.Session{"app": s} },
		healthy:  func(*discordgo.Session) bool { return *healthy },
		reopen: func(string, *discordgo.Session) error {
			*reopens++
			return *reopenErr
		},
		state: make(map[string]*watchState),
	}
}

func TestWatchdog_HealthySessionIsLeftAlone(t *testing.T) {
	healthy, reopens := true, 0
	var err error
	w := newTestWatchdog(&healthy, &err, &reopens)

	now := time.Now()
	for i := 0; i < 20; i++ {
		w.check(now.Add(time.Duration(i) * watchdogInterval))
	}
	assert.Equal(t, 0, reopens)
}

func TestWatchdog_WaitsOutGraceBeforeRestarting(t *testing.T) {
	healthy, reopens := false, 0
	var err error
	w := newTestWatchdog(&healthy, &err, &reopens)

	start := time.Now()
	w.check(start) // first sighting
	w.check(start.Add(watchdogGrace - time.Second))
	assert.Equal(t, 0, reopens, "must not restart before the grace period elapses")

	w.check(start.Add(watchdogGrace))
	assert.Equal(t, 1, reopens)
}

func TestWatchdog_BacksOffWhileFailing(t *testing.T) {
	healthy, reopens := false, 0
	err := errors.New("gateway rejected")
	w := newTestWatchdog(&healthy, &err, &reopens)

	start := time.Now()
	w.check(start)
	w.check(start.Add(watchdogGrace)) // attempt 1; next allowed after +3m
	assert.Equal(t, 1, reopens)

	w.check(start.Add(watchdogGrace + 2*time.Minute))
	assert.Equal(t, 1, reopens, "must wait out the backoff")

	second := start.Add(watchdogGrace + 3*time.Minute)
	w.check(second) // attempt 2; next allowed after +6m
	assert.Equal(t, 2, reopens)

	w.check(second.Add(5 * time.Minute))
	assert.Equal(t, 2, reopens)
	w.check(second.Add(6 * time.Minute))
	assert.Equal(t, 3, reopens)
}

func TestWatchdog_BackoffIsCapped(t *testing.T) {
	healthy, reopens := false, 0
	err := errors.New("gateway rejected")
	w := newTestWatchdog(&healthy, &err, &reopens)

	now := time.Now()
	w.check(now)
	now = now.Add(watchdogGrace)
	for i := 0; i < 12; i++ { // plenty to saturate the doubling
		w.check(now)
		now = w.state["app"].nextAttempt
	}
	assert.Equal(t, watchdogMaxBackoff, w.state["app"].backoff)
}

func TestWatchdog_RecoveryResetsState(t *testing.T) {
	healthy, reopens := false, 0
	var err error
	w := newTestWatchdog(&healthy, &err, &reopens)

	start := time.Now()
	w.check(start)
	w.check(start.Add(watchdogGrace))
	assert.Equal(t, 1, reopens)

	healthy = true
	w.check(start.Add(watchdogGrace + watchdogInterval))
	assert.NotContains(t, w.state, "app")

	// A later outage starts a fresh grace period rather than restarting immediately.
	healthy = false
	later := start.Add(time.Hour)
	w.check(later)
	w.check(later.Add(time.Minute))
	assert.Equal(t, 1, reopens)
}

func TestWatchdog_ForgetsRemovedSessions(t *testing.T) {
	healthy, reopens := false, 0
	var err error
	w := newTestWatchdog(&healthy, &err, &reopens)

	w.check(time.Now())
	assert.Contains(t, w.state, "app")

	w.sessions = func() map[string]*discordgo.Session { return map[string]*discordgo.Session{} }
	w.check(time.Now())
	assert.NotContains(t, w.state, "app")
}

func TestSessionHealthy(t *testing.T) {
	fresh := &discordgo.Session{DataReady: true, LastHeartbeatAck: time.Now()}
	assert.True(t, sessionHealthy(fresh))

	notReady := &discordgo.Session{DataReady: false, LastHeartbeatAck: time.Now()}
	assert.False(t, sessionHealthy(notReady))

	stale := &discordgo.Session{DataReady: true, LastHeartbeatAck: time.Now().Add(-gatewayStaleAfter - time.Second)}
	assert.False(t, sessionHealthy(stale))
}
