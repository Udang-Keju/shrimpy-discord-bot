package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/bwmarrin/discordgo"
)

// SchedulerRepository defines database operations consumed by the background scheduler.
type SchedulerRepository interface {
	ListDueForAutoClose(ctx context.Context) ([]repository.Ticket, error)
}

// TicketCloser defines the interface for closing tickets (satisfied by TicketService).
type TicketCloser interface {
	Close(ctx context.Context, dg *discordgo.Session, ticketID string, reason *string, closedByUserID int64) (*repository.Ticket, error)
}

// Scheduler runs a background worker to close inactive tickets.
type Scheduler struct {
	repo          SchedulerRepository
	closer        TicketCloser
	checkInterval time.Duration
}

// NewScheduler constructs a new Scheduler instance.
func NewScheduler(repo SchedulerRepository, closer TicketCloser, checkInterval time.Duration) *Scheduler {
	return &Scheduler{
		repo:          repo,
		closer:        closer,
		checkInterval: checkInterval,
	}
}

// Start starts the auto-close background loop. It blocks until the context is cancelled.
func (s *Scheduler) Start(ctx context.Context, dg *discordgo.Session) {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	fmt.Printf("Scheduler: Started background ticket auto-close worker (interval: %s)\n", s.checkInterval)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Scheduler: Stopping background ticket auto-close worker...")
			return
		case <-ticker.C:
			s.runCheck(ctx, dg)
		}
	}
}

func (s *Scheduler) runCheck(ctx context.Context, dg *discordgo.Session) {
	// Query tickets that are due for auto-close
	tickets, err := s.repo.ListDueForAutoClose(ctx)
	if err != nil {
		fmt.Printf("Scheduler Error: failed to list tickets due for auto-close: %v\n", err)
		return
	}

	if len(tickets) == 0 {
		return
	}

	fmt.Printf("Scheduler: Found %d ticket(s) due for auto-close\n", len(tickets))

	reason := "AUTO_CLOSE_INACTIVITY"
	systemUserID := int64(0) // Representing system/bot close

	for _, t := range tickets {
		fmt.Printf("Scheduler: Auto-closing ticket %s due to inactivity...\n", t.ID)
		_, err := s.closer.Close(ctx, dg, t.ID, &reason, systemUserID)
		if err != nil {
			fmt.Printf("Scheduler Error: failed to close ticket %s: %v\n", t.ID, err)
		}
	}
}
