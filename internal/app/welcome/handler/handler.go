package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
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

// Get returns the welcome config for a server.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	cfg, err := h.welcomeSvc.Get(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve configuration")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, cfg)
}

// Save saves or updates the welcome config for a server.
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var cfg model.WelcomeConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload")
		return
	}

	cfg.GuildID = guildID

	saved, err := h.welcomeSvc.Save(r.Context(), &cfg)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save configuration")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, saved)
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
