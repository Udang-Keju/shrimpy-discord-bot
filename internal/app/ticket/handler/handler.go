package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
)

// Handler coordinates panel/category CRUD, dashboard ticket lists, actions, stats, and transcripts.
type Handler struct {
	ticketSvc    *service.TicketService
	categoryRepo *repository.CategoryRepo
	transcript   *service.TranscriptService
	dg           *discordgo.Session
}

// NewHandler constructs a new Handler.
func NewHandler(ticketSvc *service.TicketService, categoryRepo *repository.CategoryRepo, transcript *service.TranscriptService, dg *discordgo.Session) *Handler {
	return &Handler{
		ticketSvc:    ticketSvc,
		categoryRepo: categoryRepo,
		transcript:   transcript,
		dg:           dg,
	}
}

// ─── Panel Handlers ───────────────────────────────────────────────────────────

// ListPanels lists all panels for a guild.
func (h *Handler) ListPanels(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	panels, err := h.categoryRepo.ListPanels(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch panels")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, panels)
}

// CreatePanel creates a ticket panel config in the database.
func (h *Handler) CreatePanel(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var p model.TicketPanel
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid panel payload")
		return
	}

	p.GuildID = guildID

	created, err := h.categoryRepo.CreatePanel(r.Context(), &p)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create panel: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusCreated, created)
}

// UpdatePanel updates and saves changes to a ticket panel.
func (h *Handler) UpdatePanel(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	existing, err := h.categoryRepo.GetPanelByGuild(r.Context(), panelID, guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Panel not found")
		return
	}

	var p model.TicketPanel
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	existing.Name = p.Name
	existing.ChannelID = p.ChannelID
	existing.PanelStyle = p.PanelStyle
	existing.EmbedTitle = p.EmbedTitle
	existing.EmbedDescription = p.EmbedDescription
	existing.EmbedColor = p.EmbedColor
	existing.EmbedMedia = p.EmbedMedia

	updated, err := h.categoryRepo.UpdatePanel(r.Context(), existing)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update panel")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, updated)
}

// DeletePanel deletes a panel and its categories from DB.
func (h *Handler) DeletePanel(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")

	err := h.categoryRepo.DeletePanel(r.Context(), panelID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete panel")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// ─── Category Handlers ────────────────────────────────────────────────────────

// ListCategories lists categories for a panel.
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")

	cats, err := h.categoryRepo.ListCategoriesByPanel(r.Context(), panelID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch categories")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, cats)
}

// CreateCategory adds a ticket category under a panel.
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")

	var c model.TicketCategory
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid category payload")
		return
	}

	c.PanelID = panelID

	created, err := h.categoryRepo.CreateCategory(r.Context(), &c)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create category")
		return
	}

	apiutil.WriteJSON(w, http.StatusCreated, created)
}

// UpdateCategory updates configuration for an existing category.
func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catId")

	existing, err := h.categoryRepo.GetCategory(r.Context(), catID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	var c model.TicketCategory
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	existing.Name = c.Name
	existing.Emoji = c.Emoji
	existing.ButtonLabel = c.ButtonLabel
	existing.ButtonStyle = c.ButtonStyle
	existing.ButtonDescription = c.ButtonDescription
	existing.ButtonOrder = c.ButtonOrder
	existing.TicketDestination = c.TicketDestination
	existing.TicketNameTemplate = c.TicketNameTemplate
	existing.TicketOpenTitle = c.TicketOpenTitle
	existing.TicketOpenMessage = c.TicketOpenMessage
	existing.TicketOpenColor = c.TicketOpenColor
	existing.TicketOpenMedia = c.TicketOpenMedia
	existing.MaxTicketsPerUser = c.MaxTicketsPerUser
	existing.AutoCloseHours = c.AutoCloseHours
	existing.TranscriptChannelID = c.TranscriptChannelID
	existing.AllowUserClose = c.AllowUserClose

	updated, err := h.categoryRepo.UpdateCategory(r.Context(), existing)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update category")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, updated)
}

// DeleteCategory deletes a category.
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catId")

	err := h.categoryRepo.DeleteCategory(r.Context(), catID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete category")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// ─── Ticket Handlers ──────────────────────────────────────────────────────────

// List returns a paginated list of tickets for a guild.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	q := r.URL.Query()
	f := model.TicketFilter{}

	if status := q.Get("status"); status != "" {
		s := model.TicketStatus(status)
		f.Status = &s
	}
	if priority := q.Get("priority"); priority != "" {
		p := model.TicketPriority(priority)
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
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query tickets: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{
		"tickets": tickets,
		"total":   total,
		"page":    f.Page,
		"limit":   f.Limit,
	})
}

// Get returns the details of a specific ticket.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")

	ticket, err := h.ticketSvc.GetByID(r.Context(), ticketID)
	if err != nil {
		if err == model.ErrNotFound {
			apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Ticket not found")
		} else {
			apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch ticket")
		}
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, ticket)
}

type updateTicketPayload struct {
	Priority *string `json:"priority"`
	Claimed  *bool   `json:"claimed"`
}

// Update handles priority updates and manual dashboard claiming.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	callerUserIDStr := apiutil.GetUserID(r.Context())
	callerUserID, _ := strconv.ParseInt(callerUserIDStr, 10, 64)

	var payload updateTicketPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid update payload")
		return
	}

	var ticket *model.Ticket
	var err error

	if payload.Claimed != nil {
		if *payload.Claimed {
			ticket, err = h.ticketSvc.Claim(r.Context(), h.dg, ticketID, callerUserID)
		} else {
			ticket, err = h.ticketSvc.Unclaim(r.Context(), h.dg, ticketID)
		}
		if err != nil {
			apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Claim operation failed: "+err.Error())
			return
		}
	}

	if payload.Priority != nil {
		prio := model.TicketPriority(*payload.Priority)
		ticket, err = h.ticketSvc.UpdatePriority(r.Context(), ticketID, prio)
		if err != nil {
			apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update priority: "+err.Error())
			return
		}
	}

	if ticket == nil {
		ticket, err = h.ticketSvc.GetByID(r.Context(), ticketID)
		if err != nil {
			apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch updated record")
			return
		}
	}

	apiutil.WriteJSON(w, http.StatusOK, ticket)
}

type closeTicketPayload struct {
	Reason *string `json:"reason"`
}

// Close closes a ticket.
func (h *Handler) Close(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	callerUserIDStr := apiutil.GetUserID(r.Context())
	callerUserID, _ := strconv.ParseInt(callerUserIDStr, 10, 64)

	var payload closeTicketPayload
	_ = json.NewDecoder(r.Body).Decode(&payload)

	ticket, err := h.ticketSvc.Close(r.Context(), h.dg, ticketID, payload.Reason, callerUserID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to close ticket: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, ticket)
}

// Reopen reopens a closed ticket.
func (h *Handler) Reopen(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")

	ticket, err := h.ticketSvc.Reopen(r.Context(), h.dg, ticketID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reopen ticket: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, ticket)
}

// Archive archives a ticket (deletes Discord channel).
func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")

	err := h.ticketSvc.Archive(r.Context(), h.dg, ticketID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to archive ticket: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// DownloadTranscript generates and returns the ticket transcript file as a attachment stream.
func (h *Handler) DownloadTranscript(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")
	format := r.URL.Query().Get("format")

	ticket, err := h.ticketSvc.GetByID(r.Context(), ticketID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Ticket not found")
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
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate transcript: "+err.Error())
		return
	}

	fileName := fmt.Sprintf("transcript-%s.%s", ticket.ID[:8], fileExt)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(content))
}

// ─── Stats Handlers ───────────────────────────────────────────────────────────

// GetStats returns aggregated support panel metrics and guild membership counts.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	memberCount := 0
	guild, err := h.dg.State.Guild(guildIDStr)
	if err == nil {
		memberCount = guild.MemberCount
	} else {
		guild, err = h.dg.Guild(guildIDStr)
		if err == nil {
			memberCount = guild.MemberCount
		}
	}

	stats, err := h.ticketSvc.GetStats(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute ticket stats: "+err.Error())
		return
	}

	topCategoryName := "None"
	if stats.TopCategoryID != "" {
		cat, err := h.categoryRepo.GetCategory(r.Context(), stats.TopCategoryID)
		if err == nil {
			topCategoryName = cat.Name
		}
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{
		"memberCount": memberCount,
		"tickets": apiutil.JSONResponse{
			"open":            stats.Open,
			"claimed":         stats.Claimed,
			"closedThisMonth": stats.ClosedThisMonth,
			"archivedTotal":   stats.ArchivedTotal,
		},
		"avgResolutionMinutes": stats.AvgResolutionMin,
		"topCategory":          topCategoryName,
	})
}
