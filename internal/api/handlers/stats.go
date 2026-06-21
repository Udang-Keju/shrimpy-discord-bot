package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
)

// StatsTicketService defines business stats operations.
type StatsTicketService interface {
	GetStats(ctx context.Context, guildID int64) (*repository.TicketStats, error)
}

// StatsCategoryRepo gets category names.
type StatsCategoryRepo interface {
	GetCategory(ctx context.Context, categoryID string) (*repository.TicketCategory, error)
}

// StatsHandler aggregates database ticket stats and Discord server membership numbers.
type StatsHandler struct {
	ticketSvc StatsTicketService
	catRepo   StatsCategoryRepo
	dg        *discordgo.Session
}

// NewStatsHandler constructs a new StatsHandler.
func NewStatsHandler(ticketSvc StatsTicketService, catRepo StatsCategoryRepo, dg *discordgo.Session) *StatsHandler {
	return &StatsHandler{
		ticketSvc: ticketSvc,
		catRepo:   catRepo,
		dg:        dg,
	}
}

// GetStats returns aggregated support panel metrics and guild membership counts.
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	// 1. Fetch member count from Discord
	memberCount := 0
	guild, err := h.dg.State.Guild(guildIDStr)
	if err == nil {
		memberCount = guild.MemberCount
	} else {
		// Fallback REST call
		guild, err = h.dg.Guild(guildIDStr)
		if err == nil {
			memberCount = guild.MemberCount
		}
	}

	// 2. Fetch ticket stats from Database
	stats, err := h.ticketSvc.GetStats(r.Context(), guildID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute ticket stats: "+err.Error())
		return
	}

	// 3. Resolve category name for the top category ID
	topCategoryName := "None"
	if stats.TopCategoryID != "" {
		cat, err := h.catRepo.GetCategory(r.Context(), stats.TopCategoryID)
		if err == nil {
			topCategoryName = cat.Name
		}
	}

	WriteJSON(w, http.StatusOK, JSONResponse{
		"memberCount": memberCount,
		"tickets": JSONResponse{
			"open":            stats.Open,
			"claimed":         stats.Claimed,
			"closedThisMonth": stats.ClosedThisMonth,
			"archivedTotal":   stats.ArchivedTotal,
		},
		"avgResolutionMinutes": stats.AvgResolutionMin,
		"topCategory":          topCategoryName,
	})
}
