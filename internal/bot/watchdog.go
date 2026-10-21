package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	// gatewayStaleAfter is how long a session may go without a heartbeat ACK before it is
	// reported as disconnected. Discord's heartbeat interval is ~41s, so this tolerates two misses.
	gatewayStaleAfter = 2 * time.Minute

	watchdogInterval   = 30 * time.Second
	watchdogGrace      = 3 * time.Minute  // how long a session must stay unhealthy before we intervene
	watchdogMaxBackoff = 30 * time.Minute // cap between restart attempts for a persistently failing session
)

// sessionHealthy reports whether a session is open and heartbeating: discordgo has marked it
// ready and Discord has ACKed a heartbeat recently.
func sessionHealthy(s *discordgo.Session) bool {
	s.RLock()
	defer s.RUnlock()
	return s.DataReady && time.Since(s.LastHeartbeatAck) < gatewayStaleAfter
}

// reopenGateway closes and re-opens a session's websocket in place. The same *discordgo.Session
// is reused so only one connection can exist for it even if discordgo's own reconnect loop is
// running concurrently (Open then returns ErrWSAlreadyOpen, which is success for our purposes).
func reopenGateway(s *discordgo.Session) error {
	_ = s.Close()
	if err := s.Open(); err != nil && !errors.Is(err, discordgo.ErrWSAlreadyOpen) {
		return err
	}
	return nil
}

// RunWatchdog restarts bot gateway sessions that stay unhealthy, so a silently dropped
// connection heals without anyone pressing Reconnect. It blocks until ctx is cancelled.
func (r *Registry) RunWatchdog(ctx context.Context) {
	w := &watchdog{
		sessions: r.snapshot,
		healthy:  sessionHealthy,
		reopen:   r.restartSession,
		state:    make(map[string]*watchState),
	}

	ticker := time.NewTicker(watchdogInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			w.check(now)
		}
	}
}

type watchState struct {
	downSince   time.Time
	nextAttempt time.Time
	backoff     time.Duration
}

// watchdog holds the restart policy; its collaborators are injected so the policy is testable.
type watchdog struct {
	sessions func() map[string]*discordgo.Session
	healthy  func(*discordgo.Session) bool
	reopen   func(appID string, s *discordgo.Session) error

	state map[string]*watchState
}

// check evaluates every session once. A session unhealthy for longer than watchdogGrace is
// restarted, then retried with exponentially growing delay (capped) while it keeps failing, so a
// permanent fault (e.g. a disallowed intent) can't hammer Discord's identify rate limit.
func (w *watchdog) check(now time.Time) {
	sessions := w.sessions()

	for id := range w.state {
		if _, ok := sessions[id]; !ok {
			delete(w.state, id)
		}
	}

	for id, s := range sessions {
		if w.healthy(s) {
			delete(w.state, id)
			continue
		}

		st, seen := w.state[id]
		if !seen {
			w.state[id] = &watchState{downSince: now, backoff: watchdogGrace}
			continue
		}
		if now.Sub(st.downSince) < watchdogGrace || now.Before(st.nextAttempt) {
			continue
		}

		fmt.Printf("Watchdog: application %s unhealthy for %s, restarting gateway...\n", id, now.Sub(st.downSince).Round(time.Second))
		if err := w.reopen(id, s); err != nil {
			fmt.Printf("Watchdog: restart of application %s failed: %v\n", id, err)
		}

		st.nextAttempt = now.Add(st.backoff)
		st.backoff *= 2
		if st.backoff > watchdogMaxBackoff {
			st.backoff = watchdogMaxBackoff
		}
	}
}
