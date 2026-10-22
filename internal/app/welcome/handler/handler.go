package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/go-chi/chi/v5"
)

// Handler manages onboarding messages dashboard configurations.
type Handler struct {
	welcomeSvc *service.WelcomeService
}

// NewHandler constructs a new Handler.
func NewHandler(welcomeSvc *service.WelcomeService) *Handler {
	return &Handler{
		welcomeSvc: welcomeSvc,
	}
}

// welcomeConfigResponse is the dashboard-facing shape of a welcome config. Snowflake IDs are
// strings because JS numbers cannot hold 64-bit integers; an unset channel is "".
type welcomeConfigResponse struct {
	GuildID     string                  `json:"guild_id"`
	Enabled     bool                    `json:"enabled"`
	DMMessage   string                  `json:"dm_message"`
	ChannelID   string                  `json:"channel_id"`
	Content     string                  `json:"content"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Color       *int32                  `json:"color,omitempty"`
	Media       *discordutil.EmbedMedia `json:"media,omitempty"`
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toResponse(c *model.WelcomeConfig) welcomeConfigResponse {
	resp := welcomeConfigResponse{
		GuildID:     strconv.FormatInt(c.GuildID, 10),
		Enabled:     c.Enabled,
		DMMessage:   deref(c.DMMessage),
		Content:     deref(c.ChannelMessage),
		Title:       deref(c.EmbedTitle),
		Description: deref(c.EmbedDesc),
		Color:       c.EmbedColor,
	}
	if c.ChannelID != nil {
		resp.ChannelID = strconv.FormatInt(*c.ChannelID, 10)
	}
	if media, err := c.GetMedia(); err == nil {
		resp.Media = media
	}
	return resp
}

// Get returns the welcome config for a server.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	cfg, err := h.welcomeSvc.Get(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve configuration")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, toResponse(cfg))
}

type savePayload struct {
	Enabled     bool                    `json:"enabled"`
	DMMessage   string                  `json:"dm_message"`
	ChannelID   string                  `json:"channel_id"`
	Content     string                  `json:"content"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Color       *int32                  `json:"color"`
	Media       *discordutil.EmbedMedia `json:"media"`
}

// Save saves or updates the welcome config for a server.
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var payload savePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload")
		return
	}

	cfg := model.WelcomeConfig{
		GuildID:        guildID,
		Enabled:        payload.Enabled,
		DMMessage:      nilIfEmpty(payload.DMMessage),
		ChannelMessage: nilIfEmpty(payload.Content),
		EmbedTitle:     nilIfEmpty(payload.Title),
		EmbedDesc:      nilIfEmpty(payload.Description),
		EmbedColor:     payload.Color,
	}
	if payload.ChannelID != "" {
		chID, err := discordutil.ParseID(payload.ChannelID)
		if err != nil {
			apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid channel ID format")
			return
		}
		cfg.ChannelID = &chID
	}
	if err := cfg.SetMedia(payload.Media); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid embed media")
		return
	}

	saved, err := h.welcomeSvc.Save(r.Context(), &cfg)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save configuration")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, toResponse(saved))
}

// Delete disables and cleans up the welcome config for a server.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	err := h.welcomeSvc.Disable(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to disable onboarding configuration")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}
