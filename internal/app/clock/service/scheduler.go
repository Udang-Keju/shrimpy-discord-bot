package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
)

// DefaultRefreshInterval keeps each channel well under Discord's limit of
// about 2 renames per 10 minutes.
const DefaultRefreshInterval = 5 * time.Minute

// sessionFunc resolves the editor for a guild's bot session.
type sessionFunc func(ctx context.Context, guildID int64) (channelEditor, error)

// Scheduler periodically renames every configured clock channel.
type Scheduler struct {
	repo     ClockRepository
	interval time.Duration
	now      func() time.Time
}

// NewScheduler constructs a new Scheduler.
func NewScheduler(repo ClockRepository, interval time.Duration) *Scheduler {
	return &Scheduler{repo: repo, interval: interval, now: time.Now}
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
	for i := range clocks {
		if ctx.Err() != nil {
			return
		}
		c := &clocks[i]

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
