package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
)

// CategoryRepository defines database operations consumed by CategoryHandler.
type CategoryRepository interface {
	ListPanels(ctx context.Context, guildID int64) ([]repository.TicketPanel, error)
	GetPanel(ctx context.Context, panelID string) (*repository.TicketPanel, error)
	GetPanelByGuild(ctx context.Context, panelID string, guildID int64) (*repository.TicketPanel, error)
	CreatePanel(ctx context.Context, p *repository.TicketPanel) (*repository.TicketPanel, error)
	UpdatePanel(ctx context.Context, p *repository.TicketPanel) (*repository.TicketPanel, error)
	SetPanelMessage(ctx context.Context, panelID string, messageID int64) error
	DeletePanel(ctx context.Context, panelID string) error

	ListCategoriesByPanel(ctx context.Context, panelID string) ([]repository.TicketCategory, error)
	GetCategory(ctx context.Context, categoryID string) (*repository.TicketCategory, error)
	CreateCategory(ctx context.Context, c *repository.TicketCategory) (*repository.TicketCategory, error)
	UpdateCategory(ctx context.Context, c *repository.TicketCategory) (*repository.TicketCategory, error)
	DeleteCategory(ctx context.Context, categoryID string) error
}

// CategoryHandler handles REST routes for creating, updating, and posting ticket panels and their categories.
type CategoryHandler struct {
	repo CategoryRepository
	dg   *discordgo.Session
}

// NewCategoryHandler constructs a new CategoryHandler.
func NewCategoryHandler(repo CategoryRepository, dg *discordgo.Session) *CategoryHandler {
	return &CategoryHandler{
		repo: repo,
		dg:   dg,
	}
}

// ListPanels lists all panels for a guild.
func (h *CategoryHandler) ListPanels(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	panels, err := h.repo.ListPanels(r.Context(), guildID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch panels")
		return
	}

	WriteJSON(w, http.StatusOK, panels)
}

// CreatePanel creates a ticket panel config in the database.
func (h *CategoryHandler) CreatePanel(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var p repository.TicketPanel
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid panel payload")
		return
	}

	p.GuildID = guildID

	created, err := h.repo.CreatePanel(r.Context(), &p)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create panel: "+err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, created)
}

// UpdatePanel updates and saves changes to a ticket panel.
func (h *CategoryHandler) UpdatePanel(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	existing, err := h.repo.GetPanelByGuild(r.Context(), panelID, guildID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Panel not found")
		return
	}

	var p repository.TicketPanel
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	existing.Name = p.Name
	existing.ChannelID = p.ChannelID
	existing.PanelStyle = p.PanelStyle
	existing.EmbedTitle = p.EmbedTitle
	existing.EmbedDescription = p.EmbedDescription
	existing.EmbedColor = p.EmbedColor
	existing.EmbedMedia = p.EmbedMedia

	updated, err := h.repo.UpdatePanel(r.Context(), existing)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update panel")
		return
	}

	WriteJSON(w, http.StatusOK, updated)
}

// DeletePanel deletes a panel and its categories from DB.
func (h *CategoryHandler) DeletePanel(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")

	err := h.repo.DeletePanel(r.Context(), panelID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete panel")
		return
	}

	WriteJSON(w, http.StatusOK, JSONResponse{"success": true})
}

// ListCategories lists categories for a panel.
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")

	cats, err := h.repo.ListCategoriesByPanel(r.Context(), panelID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch categories")
		return
	}

	WriteJSON(w, http.StatusOK, cats)
}

// CreateCategory adds a ticket category under a panel.
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	panelID := chi.URLParam(r, "panelId")

	var c repository.TicketCategory
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid category payload")
		return
	}

	c.PanelID = panelID

	created, err := h.repo.CreateCategory(r.Context(), &c)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create category")
		return
	}

	WriteJSON(w, http.StatusCreated, created)
}

// UpdateCategory updates configuration for an existing category.
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catId")

	existing, err := h.repo.GetCategory(r.Context(), catID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	var c repository.TicketCategory
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
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

	updated, err := h.repo.UpdateCategory(r.Context(), existing)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update category")
		return
	}

	WriteJSON(w, http.StatusOK, updated)
}

// DeleteCategory deletes a category.
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	catID := chi.URLParam(r, "catId")

	err := h.repo.DeleteCategory(r.Context(), catID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete category")
		return
	}

	WriteJSON(w, http.StatusOK, JSONResponse{"success": true})
}
