package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
)

// Handler manages reaction role message dashboard configurations.
type Handler struct {
	reactionSvc *service.ReactionRoleService
	dg          *discordgo.Session
}

// NewHandler constructs a new Handler.
func NewHandler(reactionSvc *service.ReactionRoleService, dg *discordgo.Session) *Handler {
	return &Handler{
		reactionSvc: reactionSvc,
		dg:          dg,
	}
}

func (h *Handler) ListReactionRoles(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	msgs, err := h.reactionSvc.List(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list reaction roles")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, msgs)
}

func (h *Handler) GetReactionRole(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")

	msg, err := h.reactionSvc.Get(r.Context(), msgID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Reaction role message not found")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, msg)
}

type createReactionRolePayload struct {
	ChannelID   string                 `json:"channel_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Color       *int32                 `json:"color"`
	Media       *discordutil.EmbedMedia `json:"media"`
}

func (h *Handler) CreateReactionRole(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var payload createReactionRolePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	chID, err := strconv.ParseInt(payload.ChannelID, 10, 64)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Channel ID format")
		return
	}

	msg, err := h.reactionSvc.Create(r.Context(), h.dg, guildID, chID, payload.Title, payload.Description, payload.Color, payload.Media)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create reaction role: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusCreated, msg)
}

func (h *Handler) DeleteReactionRole(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")

	err := h.reactionSvc.Delete(r.Context(), h.dg, msgID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete reaction role message: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

type addEmojiPayload struct {
	Emoji  string `json:"emoji"`
	RoleID string `json:"role_id"`
}

func (h *Handler) AddEmojiMapping(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")

	var payload addEmojiPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	roleID, err := strconv.ParseInt(payload.RoleID, 10, 64)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Role ID format")
		return
	}

	created, err := h.reactionSvc.AddEmoji(r.Context(), h.dg, msgID, payload.Emoji, roleID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add emoji mapping")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, created)
}

func (h *Handler) RemoveEmojiMapping(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")
	emoji := r.URL.Query().Get("emoji")

	if emoji == "" {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Missing emoji parameter")
		return
	}

	err := h.reactionSvc.RemoveEmoji(r.Context(), h.dg, msgID, emoji)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove emoji mapping")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}
