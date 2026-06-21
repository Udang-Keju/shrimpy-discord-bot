package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/go-chi/chi/v5"
)

// WelcomeService defines the business operations consumed by WelcomeHandler.
type WelcomeService interface {
	Get(ctx context.Context, guildID int64) (*repository.WelcomeConfig, error)
	Save(ctx context.Context, cfg *repository.WelcomeConfig) (*repository.WelcomeConfig, error)
	Disable(ctx context.Context, guildID int64) error
}

// WelcomeHandler manages onboarding messages dashboard configurations.
type WelcomeHandler struct {
	welcomeSvc WelcomeService
}

// NewWelcomeHandler constructs a new WelcomeHandler.
func NewWelcomeHandler(welcomeSvc WelcomeService) *WelcomeHandler {
	return &WelcomeHandler{
		welcomeSvc: welcomeSvc,
	}
}

// Get returns the welcome config for a server.
func (h *WelcomeHandler) Get(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	cfg, err := h.welcomeSvc.Get(r.Context(), guildID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve configuration")
		return
	}

	WriteJSON(w, http.StatusOK, cfg)
}

// Save saves or updates the welcome config for a server.
func (h *WelcomeHandler) Save(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var cfg repository.WelcomeConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload")
		return
	}

	cfg.GuildID = guildID

	saved, err := h.welcomeSvc.Save(r.Context(), &cfg)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save configuration")
		return
	}

	WriteJSON(w, http.StatusOK, saved)
}

// Delete disables and cleans up the welcome config for a server.
func (h *WelcomeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	err := h.welcomeSvc.Disable(r.Context(), guildID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to disable onboarding configuration")
		return
	}

	WriteJSON(w, http.StatusOK, JSONResponse{"success": true})
}
