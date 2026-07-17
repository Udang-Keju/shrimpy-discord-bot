package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/go-chi/chi/v5"
)

// Handler manages translation feature dashboard configuration.
type Handler struct {
	translationSvc *service.TranslationService
}

// NewHandler constructs a new Handler.
func NewHandler(translationSvc *service.TranslationService) *Handler {
	return &Handler{translationSvc: translationSvc}
}

// channelDTO / emojiDTO / configDTO shape the API response. The API key is
// never serialized — only whether one is set — so ciphertext never leaves the
// backend.
type channelDTO struct {
	ChannelID          string  `json:"channelId"`
	TargetLangOverride *string `json:"targetLangOverride"`
}

type emojiDTO struct {
	Emoji              string  `json:"emoji"`
	TargetLangOverride *string `json:"targetLangOverride"`
}

type configDTO struct {
	GuildID          string       `json:"guildId"`
	Enabled          bool         `json:"enabled"`
	AutoEnabled      bool         `json:"autoEnabled"`
	ReactionEnabled  bool         `json:"reactionEnabled"`
	ReactionDelivery string       `json:"reactionDelivery"` // "channel" or "dm"
	Provider         string       `json:"provider"`
	APIKey          string       `json:"apiKey"` // masked "***" when set, "" when unset
	HasAPIKey       bool         `json:"hasApiKey"`
	EndpointURL     *string      `json:"endpointUrl"`
	TargetLang      *string      `json:"targetLang"`
	Channels        []channelDTO `json:"channels"`
	Emojis          []emojiDTO   `json:"emojis"`
}

// Get returns the translation config plus channel and emoji lists for a guild.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	guildID, _ := strconv.ParseInt(chi.URLParam(r, "guildId"), 10, 64)

	cfg, err := h.translationSvc.GetConfig(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve translation configuration")
		return
	}

	channels, err := h.translationSvc.ListChannels(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve translation channels")
		return
	}

	emojis, err := h.translationSvc.ListEmojis(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve translation emojis")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, toConfigDTO(cfg, channels, emojis))
}

// saveRequest is the dashboard payload for the config PUT.
type saveRequest struct {
	Enabled          bool    `json:"enabled"`
	AutoEnabled      bool    `json:"autoEnabled"`
	ReactionEnabled  bool    `json:"reactionEnabled"`
	ReactionDelivery string  `json:"reactionDelivery"`
	Provider         string  `json:"provider"`
	APIKey           string  `json:"apiKey"`
	EndpointURL      *string `json:"endpointUrl"`
	TargetLang       *string `json:"targetLang"`
}

// Save upserts the translation config (excluding channel/emoji lists).
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	guildID, _ := strconv.ParseInt(chi.URLParam(r, "guildId"), 10, 64)

	var req saveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload")
		return
	}

	cfg, err := h.translationSvc.SaveConfig(r.Context(), guildID, service.SaveConfigInput{
		Enabled:          req.Enabled,
		AutoEnabled:      req.AutoEnabled,
		ReactionEnabled:  req.ReactionEnabled,
		ReactionDelivery: req.ReactionDelivery,
		Provider:         req.Provider,
		APIKey:           req.APIKey,
		EndpointURL:      req.EndpointURL,
		TargetLang:       req.TargetLang,
	})
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save translation configuration")
		return
	}

	channels, _ := h.translationSvc.ListChannels(r.Context(), guildID)
	emojis, _ := h.translationSvc.ListEmojis(r.Context(), guildID)
	apiutil.WriteJSON(w, http.StatusOK, toConfigDTO(cfg, channels, emojis))
}

// channelRequest is the payload for adding an auto-translate channel.
type channelRequest struct {
	ChannelID          string  `json:"channelId"`
	TargetLangOverride *string `json:"targetLangOverride"`
}

// AddChannel adds an auto-translate channel.
func (h *Handler) AddChannel(w http.ResponseWriter, r *http.Request) {
	guildID, _ := strconv.ParseInt(chi.URLParam(r, "guildId"), 10, 64)

	var req channelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload")
		return
	}
	channelID, err := strconv.ParseInt(req.ChannelID, 10, 64)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid channel ID")
		return
	}

	if _, err := h.translationSvc.AddChannel(r.Context(), guildID, channelID, req.TargetLangOverride); err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add translation channel")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// RemoveChannel removes an auto-translate channel.
func (h *Handler) RemoveChannel(w http.ResponseWriter, r *http.Request) {
	guildID, _ := strconv.ParseInt(chi.URLParam(r, "guildId"), 10, 64)
	channelID, err := strconv.ParseInt(chi.URLParam(r, "channelId"), 10, 64)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid channel ID")
		return
	}

	if err := h.translationSvc.RemoveChannel(r.Context(), guildID, channelID); err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove translation channel")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// emojiRequest is the payload for adding a trigger emoji.
type emojiRequest struct {
	Emoji              string  `json:"emoji"`
	TargetLangOverride *string `json:"targetLangOverride"`
}

// AddEmoji adds a reaction trigger emoji.
func (h *Handler) AddEmoji(w http.ResponseWriter, r *http.Request) {
	guildID, _ := strconv.ParseInt(chi.URLParam(r, "guildId"), 10, 64)

	var req emojiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload")
		return
	}

	if _, err := h.translationSvc.AddEmoji(r.Context(), guildID, req.Emoji, req.TargetLangOverride); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Failed to add trigger emoji")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// RemoveEmoji removes a reaction trigger emoji. The emoji identifier is passed
// as a query parameter to safely carry unicode / "name:id" values.
func (h *Handler) RemoveEmoji(w http.ResponseWriter, r *http.Request) {
	guildID, _ := strconv.ParseInt(chi.URLParam(r, "guildId"), 10, 64)
	emoji := r.URL.Query().Get("emoji")
	if emoji == "" {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Missing emoji parameter")
		return
	}

	if err := h.translationSvc.RemoveEmoji(r.Context(), guildID, emoji); err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove trigger emoji")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

func toConfigDTO(cfg *model.TranslationConfig, channels []model.TranslationChannel, emojis []model.TranslationReactionEmoji) configDTO {
	hasKey := len(cfg.APIKeyEnc) > 0
	apiKey := ""
	if hasKey {
		apiKey = service.MaskedAPIKey
	}

	chDTOs := make([]channelDTO, 0, len(channels))
	for _, ch := range channels {
		chDTOs = append(chDTOs, channelDTO{
			ChannelID:          strconv.FormatInt(ch.ChannelID, 10),
			TargetLangOverride: ch.TargetLangOverride,
		})
	}

	emDTOs := make([]emojiDTO, 0, len(emojis))
	for _, e := range emojis {
		emDTOs = append(emDTOs, emojiDTO{
			Emoji:              e.Emoji,
			TargetLangOverride: e.TargetLangOverride,
		})
	}

	return configDTO{
		GuildID:          strconv.FormatInt(cfg.GuildID, 10),
		Enabled:          cfg.Enabled,
		AutoEnabled:      cfg.AutoEnabled,
		ReactionEnabled:  cfg.ReactionEnabled,
		ReactionDelivery: cfg.ReactionDelivery,
		Provider:         cfg.Provider,
		APIKey:           apiKey,
		HasAPIKey:        hasKey,
		EndpointURL:      cfg.EndpointURL,
		TargetLang:       cfg.TargetLang,
		Channels:         chDTOs,
		Emojis:           emDTOs,
	}
}
