package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/go-chi/chi/v5"
)

// Handler coordinates panel/category CRUD, dashboard ticket lists, actions, stats, and transcripts.
type Handler struct {
	ticketSvc    *service.TicketService
	categoryRepo *repository.CategoryRepo
	transcript   *service.TranscriptService
	provider     discordutil.DiscordSessionProvider
}

// NewHandler constructs a new Handler.
func NewHandler(ticketSvc *service.TicketService, categoryRepo *repository.CategoryRepo, transcript *service.TranscriptService, provider discordutil.DiscordSessionProvider) *Handler {
	return &Handler{
		ticketSvc:    ticketSvc,
		categoryRepo: categoryRepo,
		transcript:   transcript,
		provider:     provider,
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

	if p.HandlerRoleIDs != nil {
		if err := h.categoryRepo.SetPanelHandlerRoles(r.Context(), created.ID, parseRoleIDs(*p.HandlerRoleIDs)); err != nil {
			fmt.Printf("warning: failed to set handler roles for panel %s: %v\n", created.ID, err)
		}
	}

	h.syncPanelMessage(r.Context(), guildID, created.ID)

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

	oldChannelID := existing.ChannelID

	existing.Name = p.Name
	existing.ChannelID = p.ChannelID
	existing.PanelStyle = p.PanelStyle
	existing.Content = p.Content
	existing.EmbedTitle = p.EmbedTitle
	existing.EmbedDescription = p.EmbedDescription
	existing.EmbedColor = p.EmbedColor
	existing.EmbedMedia = p.EmbedMedia

	updated, err := h.categoryRepo.UpdatePanel(r.Context(), existing)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update panel")
		return
	}

	if p.HandlerRoleIDs != nil {
		if err := h.categoryRepo.SetPanelHandlerRoles(r.Context(), updated.ID, parseRoleIDs(*p.HandlerRoleIDs)); err != nil {
			fmt.Printf("warning: failed to set handler roles for panel %s: %v\n", updated.ID, err)
		}
	}

	if oldChannelID != updated.ChannelID && updated.MessageID != nil {
		if dg, dgErr := h.provider.GetSessionForGuild(r.Context(), guildID); dgErr == nil {
			_ = dg.ChannelMessageDelete(discordutil.FormatID(oldChannelID), discordutil.FormatID(*updated.MessageID))
		}
		_ = h.categoryRepo.ClearPanelMessage(r.Context(), updated.ID)
	}

	h.syncPanelMessage(r.Context(), guildID, updated.ID)

	apiutil.WriteJSON(w, http.StatusOK, updated)
}

// DeletePanel deletes a panel and its categories from DB.
func (h *Handler) DeletePanel(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")

	if panel, pErr := h.categoryRepo.GetPanel(r.Context(), panelID); pErr == nil {
		if dg, dgErr := h.provider.GetSessionForGuild(r.Context(), panel.GuildID); dgErr == nil {
			_ = h.ticketSvc.DeletePanelMessage(r.Context(), dg, panel)
		}
	}

	err := h.categoryRepo.DeletePanel(r.Context(), panelID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete panel")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// syncPanelMessage fetches a live Discord session for the guild and republishes the panel
// message. Failures are logged and swallowed: the DB write that triggered this already
// succeeded, and the Discord-side message will catch up next time the panel is touched.
func (h *Handler) syncPanelMessage(ctx context.Context, guildID int64, panelID string) {
	dg, err := h.provider.GetSessionForGuild(ctx, guildID)
	if err != nil {
		fmt.Printf("warning: no bot session for guild %d, skipping panel sync: %v\n", guildID, err)
		return
	}
	if err := h.ticketSvc.SyncPanelMessage(ctx, dg, panelID); err != nil {
		fmt.Printf("warning: failed to sync panel message for panel %s: %v\n", panelID, err)
	}
}

// ─── Panel Handler Role Endpoints ─────────────────────────────────────────────

type handlerRolePayload struct {
	RoleID string `json:"role_id"`
}

// parseRoleIDs converts Discord snowflake role ID strings to int64, silently skipping
// any that fail to parse (the payload comes from the dashboard's own role dropdown, so
// malformed entries aren't expected in practice).
func parseRoleIDs(ids []string) []int64 {
	parsed := make([]int64, 0, len(ids))
	for _, id := range ids {
		if roleID, err := strconv.ParseInt(id, 10, 64); err == nil {
			parsed = append(parsed, roleID)
		}
	}
	return parsed
}

// ListPanelHandlerRoles lists the roles invited into tickets created from this panel.
func (h *Handler) ListPanelHandlerRoles(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")

	roles, err := h.categoryRepo.ListPanelHandlerRoles(r.Context(), panelID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list panel handler roles")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, roles)
}

// AddPanelHandlerRole adds a Discord role to a panel's ticket handler list.
func (h *Handler) AddPanelHandlerRole(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")

	var payload handlerRolePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	roleID, err := strconv.ParseInt(payload.RoleID, 10, 64)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Role ID format")
		return
	}

	created, err := h.categoryRepo.AddPanelHandlerRole(r.Context(), panelID, roleID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save panel handler role")
		return
	}

	apiutil.WriteJSON(w, http.StatusCreated, created)
}

// RemovePanelHandlerRole removes a role from a panel's ticket handler list.
func (h *Handler) RemovePanelHandlerRole(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")

	roleIDStr := chi.URLParam(r, "roleId")
	roleID, _ := strconv.ParseInt(roleIDStr, 10, 64)

	err := h.categoryRepo.RemovePanelHandlerRole(r.Context(), panelID, roleID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove panel handler role")
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

	if c.HandlerRoleIDs != nil {
		if err := h.categoryRepo.SetCategoryHandlerRoles(r.Context(), created.ID, parseRoleIDs(*c.HandlerRoleIDs)); err != nil {
			fmt.Printf("warning: failed to set handler roles for category %s: %v\n", created.ID, err)
		}
	}

	h.syncPanelMessageForGuildOfPanel(r.Context(), panelID)

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
	existing.ThreadParentChannelID = c.ThreadParentChannelID
	existing.ChannelCategoryID = c.ChannelCategoryID
	existing.TicketNameTemplate = c.TicketNameTemplate
	existing.TicketOpenTitle = c.TicketOpenTitle
	existing.TicketOpenMessage = c.TicketOpenMessage
	existing.TicketOpenColor = c.TicketOpenColor
	existing.TicketOpenMedia = c.TicketOpenMedia
	existing.TicketOpenContent = c.TicketOpenContent
	existing.MaxTicketsPerUser = c.MaxTicketsPerUser
	existing.AutoCloseHours = c.AutoCloseHours
	existing.TranscriptChannelID = c.TranscriptChannelID
	existing.AllowUserClose = c.AllowUserClose

	updated, err := h.categoryRepo.UpdateCategory(r.Context(), existing)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update category")
		return
	}

	if c.HandlerRoleIDs != nil {
		if err := h.categoryRepo.SetCategoryHandlerRoles(r.Context(), updated.ID, parseRoleIDs(*c.HandlerRoleIDs)); err != nil {
			fmt.Printf("warning: failed to set handler roles for category %s: %v\n", updated.ID, err)
		}
	}

	h.syncPanelMessageForGuildOfPanel(r.Context(), updated.PanelID)

	apiutil.WriteJSON(w, http.StatusOK, updated)
}

// DeleteCategory deletes a category.
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catId")

	cat, catErr := h.categoryRepo.GetCategory(r.Context(), catID)

	err := h.categoryRepo.DeleteCategory(r.Context(), catID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete category")
		return
	}

	if catErr == nil {
		h.syncPanelMessageForGuildOfPanel(r.Context(), cat.PanelID)
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// syncPanelMessageForGuildOfPanel looks up the panel's guild (categories only carry
// panelId, not guildId) and republishes the panel message to reflect category changes.
func (h *Handler) syncPanelMessageForGuildOfPanel(ctx context.Context, panelID string) {
	panel, err := h.categoryRepo.GetPanel(ctx, panelID)
	if err != nil {
		fmt.Printf("warning: failed to load panel %s for message sync: %v\n", panelID, err)
		return
	}
	h.syncPanelMessage(ctx, panel.GuildID, panelID)
}

// ─── Category Handler Role Endpoints ──────────────────────────────────────────

// ListCategoryHandlerRoles lists the roles invited into tickets created from this category.
func (h *Handler) ListCategoryHandlerRoles(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catId")

	roles, err := h.categoryRepo.ListCategoryHandlerRoles(r.Context(), catID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list category handler roles")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, roles)
}

// AddCategoryHandlerRole adds a Discord role to a category's ticket handler list.
func (h *Handler) AddCategoryHandlerRole(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catId")

	var payload handlerRolePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	roleID, err := strconv.ParseInt(payload.RoleID, 10, 64)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Role ID format")
		return
	}

	created, err := h.categoryRepo.AddCategoryHandlerRole(r.Context(), catID, roleID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save category handler role")
		return
	}

	apiutil.WriteJSON(w, http.StatusCreated, created)
}

// RemoveCategoryHandlerRole removes a role from a category's ticket handler list.
func (h *Handler) RemoveCategoryHandlerRole(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catId")

	roleIDStr := chi.URLParam(r, "roleId")
	roleID, _ := strconv.ParseInt(roleIDStr, 10, 64)

	err := h.categoryRepo.RemoveCategoryHandlerRole(r.Context(), catID, roleID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove category handler role")
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

	ticket, err := h.ticketSvc.GetByID(r.Context(), ticketID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Ticket not found")
		return
	}

	dg, err := h.provider.GetSessionForGuild(r.Context(), ticket.GuildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	if payload.Claimed != nil {
		if *payload.Claimed {
			ticket, err = h.ticketSvc.Claim(r.Context(), dg, ticketID, callerUserID)
		} else {
			ticket, err = h.ticketSvc.Unclaim(r.Context(), dg, ticketID)
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

	ticket, err := h.ticketSvc.GetByID(r.Context(), ticketID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Ticket not found")
		return
	}

	dg, err := h.provider.GetSessionForGuild(r.Context(), ticket.GuildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	ticket, err = h.ticketSvc.Close(r.Context(), dg, ticketID, payload.Reason, callerUserID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to close ticket: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, ticket)
}

// Reopen reopens a closed ticket.
func (h *Handler) Reopen(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")

	ticket, err := h.ticketSvc.GetByID(r.Context(), ticketID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Ticket not found")
		return
	}

	dg, err := h.provider.GetSessionForGuild(r.Context(), ticket.GuildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	ticket, err = h.ticketSvc.Reopen(r.Context(), dg, ticketID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reopen ticket: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, ticket)
}

// Archive archives a ticket (deletes Discord channel).
func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	ticketID := chi.URLParam(r, "ticketId")

	ticket, err := h.ticketSvc.GetByID(r.Context(), ticketID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Ticket not found")
		return
	}

	dg, err := h.provider.GetSessionForGuild(r.Context(), ticket.GuildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	err = h.ticketSvc.Archive(r.Context(), dg, ticketID)
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

	dg, err := h.provider.GetSessionForGuild(r.Context(), ticket.GuildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
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
		openerUser, uErr := dg.User(fmt.Sprintf("%d", ticket.OpenedBy))
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

	dg, err := h.provider.GetSessionForGuild(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	memberCount := 0
	guild, err := dg.State.Guild(guildIDStr)
	if err == nil {
		memberCount = guild.MemberCount
	} else {
		guild, err = dg.Guild(guildIDStr)
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
