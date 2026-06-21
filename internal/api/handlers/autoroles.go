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

// AutoRolesGuildService defines DB config operations for roles.
type AutoRolesGuildService interface {
	ListStaffRoles(ctx context.Context, guildID int64) ([]repository.StaffRole, error)
	AddStaffRole(ctx context.Context, guildID, roleID int64) (*repository.StaffRole, error)
	RemoveStaffRole(ctx context.Context, guildID, roleID int64) error

	ListAutoRoles(ctx context.Context, guildID int64) ([]repository.AutoRole, error)
	AddAutoRole(ctx context.Context, guildID, roleID int64) (*repository.AutoRole, error)
	RemoveAutoRole(ctx context.Context, guildID, roleID int64) error
}

// AutoRolesReactionService defines reaction role operations.
type AutoRolesReactionService interface {
	List(ctx context.Context, guildID int64) ([]repository.ReactionRoleMessage, error)
	Get(ctx context.Context, messageID string) (*repository.ReactionRoleMessage, error)
	Create(ctx context.Context, dg *discordgo.Session, guildID int64, channelID int64, title, desc string, color *int32, media *repository.EmbedMedia) (*repository.ReactionRoleMessage, error)
	Delete(ctx context.Context, dg *discordgo.Session, messageID string) error
	AddEmoji(ctx context.Context, dg *discordgo.Session, messageID string, emoji string, roleID int64) (*repository.ReactionRoleEmoji, error)
	RemoveEmoji(ctx context.Context, dg *discordgo.Session, messageID string, emoji string) error
}

// AutoRolesHandler manages auto-roles, staff roles, and reaction role message dashboard configurations.
type AutoRolesHandler struct {
	guildSvc    AutoRolesGuildService
	reactionSvc AutoRolesReactionService
	dg          *discordgo.Session
}

// NewAutoRolesHandler constructs a new AutoRolesHandler.
func NewAutoRolesHandler(guildSvc AutoRolesGuildService, reactionSvc AutoRolesReactionService, dg *discordgo.Session) *AutoRolesHandler {
	return &AutoRolesHandler{
		guildSvc:    guildSvc,
		reactionSvc: reactionSvc,
		dg:          dg,
	}
}

// ─── Auto Roles ───────────────────────────────────────────────────────────────

func (h *AutoRolesHandler) ListAutoRoles(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	roles, err := h.guildSvc.ListAutoRoles(r.Context(), guildID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list auto roles")
		return
	}

	WriteJSON(w, http.StatusOK, roles)
}

type rolePayload struct {
	RoleID string `json:"role_id"`
}

func (h *AutoRolesHandler) AddAutoRole(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var payload rolePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	roleID, err := strconv.ParseInt(payload.RoleID, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Role ID format")
		return
	}

	created, err := h.guildSvc.AddAutoRole(r.Context(), guildID, roleID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save auto role")
		return
	}

	WriteJSON(w, http.StatusCreated, created)
}

func (h *AutoRolesHandler) RemoveAutoRole(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	roleIDStr := chi.URLParam(r, "roleId")
	roleID, _ := strconv.ParseInt(roleIDStr, 10, 64)

	err := h.guildSvc.RemoveAutoRole(r.Context(), guildID, roleID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove auto role")
		return
	}

	WriteJSON(w, http.StatusOK, JSONResponse{"success": true})
}

// ─── Staff Roles ──────────────────────────────────────────────────────────────

func (h *AutoRolesHandler) ListStaffRoles(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	roles, err := h.guildSvc.ListStaffRoles(r.Context(), guildID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list staff roles")
		return
	}

	WriteJSON(w, http.StatusOK, roles)
}

func (h *AutoRolesHandler) AddStaffRole(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var payload rolePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	roleID, err := strconv.ParseInt(payload.RoleID, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Role ID format")
		return
	}

	created, err := h.guildSvc.AddStaffRole(r.Context(), guildID, roleID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save staff role")
		return
	}

	WriteJSON(w, http.StatusCreated, created)
}

func (h *AutoRolesHandler) RemoveStaffRole(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	roleIDStr := chi.URLParam(r, "roleId")
	roleID, _ := strconv.ParseInt(roleIDStr, 10, 64)

	err := h.guildSvc.RemoveStaffRole(r.Context(), guildID, roleID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove staff role")
		return
	}

	WriteJSON(w, http.StatusOK, JSONResponse{"success": true})
}

// ─── Reaction Roles ───────────────────────────────────────────────────────────

func (h *AutoRolesHandler) ListReactionRoles(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	msgs, err := h.reactionSvc.List(r.Context(), guildID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list reaction roles")
		return
	}

	WriteJSON(w, http.StatusOK, msgs)
}

func (h *AutoRolesHandler) GetReactionRole(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")

	msg, err := h.reactionSvc.Get(r.Context(), msgID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Reaction role message not found")
		return
	}

	WriteJSON(w, http.StatusOK, msg)
}

type createReactionRolePayload struct {
	ChannelID   string                 `json:"channel_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Color       *int32                 `json:"color"`
	Media       *repository.EmbedMedia `json:"media"`
}

func (h *AutoRolesHandler) CreateReactionRole(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var payload createReactionRolePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	chID, err := strconv.ParseInt(payload.ChannelID, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Channel ID format")
		return
	}

	msg, err := h.reactionSvc.Create(r.Context(), h.dg, guildID, chID, payload.Title, payload.Description, payload.Color, payload.Media)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create reaction role: "+err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, msg)
}

func (h *AutoRolesHandler) DeleteReactionRole(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")

	err := h.reactionSvc.Delete(r.Context(), h.dg, msgID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete reaction role message: "+err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, JSONResponse{"success": true})
}

type addEmojiPayload struct {
	Emoji  string `json:"emoji"`
	RoleID string `json:"role_id"`
}

func (h *AutoRolesHandler) AddEmojiMapping(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")

	var payload addEmojiPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	roleID, err := strconv.ParseInt(payload.RoleID, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Role ID format")
		return
	}

	created, err := h.reactionSvc.AddEmoji(r.Context(), h.dg, msgID, payload.Emoji, roleID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add emoji mapping")
		return
	}

	WriteJSON(w, http.StatusOK, created)
}

func (h *AutoRolesHandler) RemoveEmojiMapping(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")
	emoji := r.URL.Query().Get("emoji")

	if emoji == "" {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Missing emoji parameter")
		return
	}

	err := h.reactionSvc.RemoveEmoji(r.Context(), h.dg, msgID, emoji)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove emoji mapping")
		return
	}

	WriteJSON(w, http.StatusOK, JSONResponse{"success": true})
}
