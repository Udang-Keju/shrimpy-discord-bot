package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/api/middleware"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
)

// TicketService defines business operations consumed by TicketHandler.
type TicketService interface {
	Claim(ctx context.Context, dg *discordgo.Session, ticketID string, staffUserID int64) (*repository.Ticket, error)
	Unclaim(ctx context.Context, dg *discordgo.Session, ticketID string) (*repository.Ticket, error)
	Close(ctx context.Context, dg *discordgo.Session, ticketID string, reason *string, closedByUserID int64) (*repository.Ticket, error)
	Reopen(ctx context.Context, dg *discordgo.Session, ticketID string) (*repository.Ticket, error)
	Archive(ctx context.Context, dg *discordgo.Session, ticketID string) error
	GetByID(ctx context.Context, ticketID string) (*repository.Ticket, error)
	List(ctx context.Context, guildID int64, f repository.TicketFilter) ([]repository.Ticket, int64, error)
	UpdatePriority(ctx context.Context, ticketID string, priority repository.TicketPriority) (*repository.Ticket, error)
}

// TicketTranscriptService generates transcript contents.
type TicketTranscriptService interface {
	GenerateHTML(ctx context.Context, ticket *repository.Ticket, categoryName string, openerUsername string, includeStaffNotes bool) (string, error)
	GenerateText(ctx context.Context, ticketID string, includeStaffNotes bool) (string, error)
}

// TicketCategoryGetter gets category details for metadata mapping.
type TicketCategoryGetter interface {
	GetCategory(ctx context.Context, categoryID string) (*repository.TicketCategory, error)
}

// TicketHandler manages ticket search, claiming/closing via REST, and downloading HTML transcripts.
type TicketHandler struct {
	ticketSvc    TicketService
	categoryRepo TicketCategoryGetter
	transcript   TicketTranscriptService
	dg           *discordgo.Session
}

// NewTicketHandler constructs a new TicketHandler.
func NewTicketHandler(ticketSvc TicketService, categoryRepo TicketCategoryGetter, transcript TicketTranscriptService, dg *discordgo.Session) *TicketHandler {
	return &TicketHandler{
		ticketSvc:    ticketSvc,
		categoryRepo: categoryRepo,
		transcript:   transcript,
		dg:           dg,
	}
}

// List returns a paginated list of tickets for a guild.
func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	// Read filters from query params
	q := r.URL.Query()
	f := repository.TicketFilter{}

	if status := q.Get("status"); status != "" {
		s := repository.TicketStatus(status)
		f.Status = &s
	}
	if priority := q.Get("priority"); priority != "" {
		p := repository.TicketPriority(priority)
		f.Priority = &p
	}
	if catID := q.Get("categoryId"); catID != "" {
		f.CategoryID = &catID
	}
	if openedBy := q.Get("openedBy"); openedBy != "" {
		if idVal, err := strconv.ParseInt(openedBy, 10, 64); err == nil {
			f.OpenedBy = &idVal
		}
	}

	f.Page, _ = strconv.Atoi(q.Get("page"))
	f.Limit, _ = strconv.Atoi(q.Get("limit"))

	tickets, total, err := h.ticketSvc.List(r.Context(), guildID, f)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query tickets: "+err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, JSONResponse{
		"tickets": tickets,
		"total":   total,
		"page":    f.Page,
		"limit":   f.Limit,
	})
}

// Get returns the details of a specific ticket.
func (h *TicketHandler) Get(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")

	ticket, err := h.ticketSvc.GetByID(r.Context(), ticketID)
	if err != nil {
		if err == repository.ErrNotFound {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "Ticket not found")
		} else {
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch ticket")
		}
		return
	}

	WriteJSON(w, http.StatusOK, ticket)
}

type updateTicketPayload struct {
	Priority *string `json:"priority"`
	Claimed  *bool   `json:"claimed"` // true = claim by caller, false = unclaim
}

// Update handles priority updates and manual dashboard claiming.
func (h *TicketHandler) Update(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	callerUserIDStr := middleware.GetUserID(r.Context())
	callerUserID, _ := strconv.ParseInt(callerUserIDStr, 10, 64)

	var payload updateTicketPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid update payload")
		return
	}

	var ticket *repository.Ticket
	var err error

	// Handle Claim/Unclaim toggle
	if payload.Claimed != nil {
		if *payload.Claimed {
			ticket, err = h.ticketSvc.Claim(r.Context(), h.dg, ticketID, callerUserID)
		} else {
			ticket, err = h.ticketSvc.Unclaim(r.Context(), h.dg, ticketID)
		}
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Claim operation failed: "+err.Error())
			return
		}
	}

	// Handle Priority update
	if payload.Priority != nil {
		prio := repository.TicketPriority(*payload.Priority)
		ticket, err = h.ticketSvc.UpdatePriority(r.Context(), ticketID, prio)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update priority: "+err.Error())
			return
		}
	}

	if ticket == nil {
		ticket, err = h.ticketSvc.GetByID(r.Context(), ticketID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch updated record")
			return
		}
	}

	WriteJSON(w, http.StatusOK, ticket)
}

type closeTicketPayload struct {
	Reason *string `json:"reason"`
}

// Close closes a ticket.
func (h *TicketHandler) Close(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	callerUserIDStr := middleware.GetUserID(r.Context())
	callerUserID, _ := strconv.ParseInt(callerUserIDStr, 10, 64)

	var payload closeTicketPayload
	_ = json.NewDecoder(r.Body).Decode(&payload) // reason is optional

	ticket, err := h.ticketSvc.Close(r.Context(), h.dg, ticketID, payload.Reason, callerUserID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to close ticket: "+err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, ticket)
}

// Reopen reopens a closed ticket.
func (h *TicketHandler) Reopen(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")

	ticket, err := h.ticketSvc.Reopen(r.Context(), h.dg, ticketID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reopen ticket: "+err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, ticket)
}

// Archive archives a ticket (deletes Discord channel).
func (h *TicketHandler) Archive(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")

	err := h.ticketSvc.Archive(r.Context(), h.dg, ticketID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to archive ticket: "+err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, JSONResponse{"success": true})
}

// DownloadTranscript generates and returns the ticket transcript file as a attachment stream.
func (h *TicketHandler) DownloadTranscript(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	format := r.URL.Query().Get("format") // "html" or "text"

	ticket, err := h.ticketSvc.GetByID(r.Context(), ticketID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Ticket not found")
		return
	}

	var content string
	var contentType string
	var fileExt string

	if format == "text" {
		content, err = h.transcript.GenerateText(r.Context(), ticketID, true)
		contentType = "text/plain"
		fileExt = "txt"
	} else {
		// Default to HTML
		catName := "General"
		cat, catErr := h.categoryRepo.GetCategory(r.Context(), ticket.CategoryID)
		if catErr == nil {
			catName = cat.Name
		}

		openerName := "unknown"
		openerUser, uErr := h.dg.User(fmt.Sprintf("%d", ticket.OpenedBy))
		if uErr == nil {
			openerName = openerUser.Username
		}

		content, err = h.transcript.GenerateHTML(r.Context(), ticket, catName, openerName, true)
		contentType = "text/html"
		fileExt = "html"
	}

	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate transcript: "+err.Error())
		return
	}

	fileName := fmt.Sprintf("transcript-%s.%s", ticket.ID[:8], fileExt)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(content))
}
