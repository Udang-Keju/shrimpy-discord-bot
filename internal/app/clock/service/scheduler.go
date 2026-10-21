package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
)

// DefaultRefreshInterval is how often the scheduler wakes up: the shortest
// refresh option. Each clock is then refreshed only once its own interval
// has passed.
const DefaultRefreshInterval = DefaultRefreshMinutes * time.Minute

// sessionFunc resolves the editor for a guild's bot session.
type sessionFunc func(ctx context.Context, guildID int64) (channelEditor, error)

// Scheduler periodically renames every configured clock channel.
type Scheduler struct {
	repo     ClockRepository
	interval time.Duration
	now      func() time.Time
	lastRun  map[string]time.Time // clock ID -> last attempt; resets on restart
}

// NewScheduler constructs a new Scheduler.
func NewScheduler(repo ClockRepository, interval time.Duration) *Scheduler {
	return &Scheduler{repo: repo, interval: interval, now: time.Now, lastRun: map[string]time.Time{}}
}

// Start renames all clocks immediately, then on every interval. Blocks until
// the context is cancelled.
func (s *Scheduler) Start(ctx context.Context, provider discordutil.DiscordSessionProvider) {
	sessionFor := func(ctx context.Context, guildID int64) (channelEditor, error) {
		return provider.GetSessionForGuild(ctx, guildID)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	fmt.Printf("Scheduler: Clock channel refresh active (every %s)\n", s.interval)
	s.runCheck(ctx, sessionFor)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Scheduler: Clock channel refresh stopped.")
			return
		case <-ticker.C:
			s.runCheck(ctx, sessionFor)
		}
	}
}

// runCheck refreshes every clock. A failure on one channel is logged and does
// not stop the others.
func (s *Scheduler) runCheck(ctx context.Context, sessionFor sessionFunc) {
	clocks, err := s.repo.ListAll(ctx)
	if err != nil {
		fmt.Printf("Scheduler Error: failed to list clock channels: %v\n", err)
		return
	}

	now := s.now()
	live := make(map[string]bool, len(clocks))
	for i := range clocks {
		live[clocks[i].ID] = true
	}
	for id := range s.lastRun {
		if !live[id] {
			delete(s.lastRun, id)
		}
	}

	for i := range clocks {
		if ctx.Err() != nil {
			return
		}
		c := &clocks[i]
		if !refreshDue(s.lastRun[c.ID], now, c.RefreshMinutes) {
			continue
		}
		// Record the attempt even on failure so a broken channel isn't retried
		// on every tick.
		s.lastRun[c.ID] = now

		editor, err := sessionFor(ctx, c.GuildID)
		if err != nil {
			fmt.Printf("Scheduler Error: no session for guild %d: %v\n", c.GuildID, err)
			continue
		}
		if err := refreshClock(ctx, s.repo, editor, c, now); err != nil {
			fmt.Printf("Scheduler Error: clock %s: %v\n", c.ID, err)
		}
	}
}
