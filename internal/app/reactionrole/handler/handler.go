package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/go-chi/chi/v5"
)

// Handler manages reaction role message dashboard configurations.
type Handler struct {
	reactionSvc *service.ReactionRoleService
	provider    discordutil.DiscordSessionProvider
}

// NewHandler constructs a new Handler.
func NewHandler(reactionSvc *service.ReactionRoleService, provider discordutil.DiscordSessionProvider) *Handler {
	return &Handler{
		reactionSvc: reactionSvc,
		provider:    provider,
	}
}

type emojiMappingResponse struct {
	Emoji  string `json:"emoji"`
	RoleID string `json:"role_id"`
}

// reactionRoleResponse is the dashboard-facing shape of a reaction role message.
// Snowflake IDs are strings because JS numbers cannot hold 64-bit integers.
type reactionRoleResponse struct {
	ID          string                  `json:"id"`
	GuildID     string                  `json:"guild_id"`
	ChannelID   string                  `json:"channel_id"`
	Content     string                  `json:"content"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Color       *int32                  `json:"color,omitempty"`
	Media       *discordutil.EmbedMedia `json:"media,omitempty"`
	Mappings    []emojiMappingResponse  `json:"mappings"`
}

func toReactionRoleResponse(m *model.ReactionRoleMessage) reactionRoleResponse {
	resp := reactionRoleResponse{
		ID:        m.ID,
		GuildID:   strconv.FormatInt(m.GuildID, 10),
		ChannelID: strconv.FormatInt(m.ChannelID, 10),
		Mappings:  make([]emojiMappingResponse, 0, len(m.Emojis)),
	}
	if m.Content != nil {
		resp.Content = *m.Content
	}
	resp.Color = m.EmbedColor
	if media, err := m.GetMedia(); err == nil {
		resp.Media = media
	}
	if m.EmbedTitle != nil {
		resp.Title = *m.EmbedTitle
	}
	if m.EmbedDescription != nil {
		resp.Description = *m.EmbedDescription
	}
	for _, e := range m.Emojis {
		resp.Mappings = append(resp.Mappings, emojiMappingResponse{
			Emoji:  e.Emoji,
			RoleID: strconv.FormatInt(e.RoleID, 10),
		})
	}
	return resp
}

func (h *Handler) ListReactionRoles(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	msgs, err := h.reactionSvc.List(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list reaction roles")
		return
	}

	resp := make([]reactionRoleResponse, 0, len(msgs))
	for i := range msgs {
		resp = append(resp, toReactionRoleResponse(&msgs[i]))
	}
	apiutil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) GetReactionRole(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")

	msg, err := h.reactionSvc.Get(r.Context(), msgID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Reaction role message not found")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, toReactionRoleResponse(msg))
}

type createReactionRolePayload struct {
	ChannelID   string                  `json:"channel_id"`
	Content     string                  `json:"content"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Color       *int32                  `json:"color"`
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

	dg, err := h.provider.GetSessionForGuild(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	msg, err := h.reactionSvc.Create(r.Context(), dg, service.CreateParams{
		GuildID:     guildID,
		ChannelID:   chID,
		Content:     payload.Content,
		Title:       payload.Title,
		Description: payload.Description,
		Color:       payload.Color,
		Media:       payload.Media,
	})
	if errors.Is(err, service.ErrEmptyMessage) {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Add plain text or embed content to the message")
		return
	}
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create reaction role: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusCreated, toReactionRoleResponse(msg))
}

// UpdateReactionRole edits a reaction role message's content and re-renders it on Discord.
func (h *Handler) UpdateReactionRole(w http.ResponseWriter, r *http.Request) {
	guildID, _ := strconv.ParseInt(chi.URLParam(r, "guildId"), 10, 64)
	msgID := chi.URLParam(r, "msgId")

	existing, err := h.reactionSvc.Get(r.Context(), msgID)
	if err != nil || existing.GuildID != guildID {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Reaction role message not found")
		return
	}

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

	dg, err := h.provider.GetSessionForGuild(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	updated, err := h.reactionSvc.Update(r.Context(), dg, msgID, service.CreateParams{
		GuildID:     guildID,
		ChannelID:   chID,
		Content:     payload.Content,
		Title:       payload.Title,
		Description: payload.Description,
		Color:       payload.Color,
		Media:       payload.Media,
	})
	if errors.Is(err, service.ErrEmptyMessage) {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Add plain text or embed content to the message")
		return
	}
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update reaction role: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, toReactionRoleResponse(updated))
}

func (h *Handler) DeleteReactionRole(w http.ResponseWriter, r *http.Request) {
	msgID := chi.URLParam(r, "msgId")

	msg, err := h.reactionSvc.Get(r.Context(), msgID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Reaction role message not found")
		return
	}

	dg, err := h.provider.GetSessionForGuild(r.Context(), msg.GuildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	err = h.reactionSvc.Delete(r.Context(), dg, msgID)
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

	msg, err := h.reactionSvc.Get(r.Context(), msgID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Reaction role message not found")
		return
	}

	dg, err := h.provider.GetSessionForGuild(r.Context(), msg.GuildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	created, err := h.reactionSvc.AddEmoji(r.Context(), dg, msgID, payload.Emoji, roleID)
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

	msg, err := h.reactionSvc.Get(r.Context(), msgID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Reaction role message not found")
		return
	}

	dg, err := h.provider.GetSessionForGuild(r.Context(), msg.GuildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	err = h.reactionSvc.RemoveEmoji(r.Context(), dg, msgID, emoji)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove emoji mapping")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}
