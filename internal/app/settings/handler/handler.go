package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/go-chi/chi/v5"
)

// SettingsSvc defines service operations for handlers.
type SettingsSvc interface {
	List(ctx context.Context) ([]service.DiscordAppDTO, error)
	GetByID(ctx context.Context, id string) (*service.DiscordAppDTO, error)
	Create(ctx context.Context, req service.CreateRequest) (*service.DiscordAppDTO, error)
	Update(ctx context.Context, id string, req service.UpdateRequest) (tokenChanged bool, newToken string, err error)
	Delete(ctx context.Context, id string) error
	Reconnect(ctx context.Context, id string) error
}

// SettingsHandler exposes HTTP endpoints to manage bot applications.
type SettingsHandler struct {
	svc SettingsSvc
}

// NewSettingsHandler constructs a new SettingsHandler.
func NewSettingsHandler(svc SettingsSvc) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

// List handles listing all configured applications.
// GET /api/v1/admin/apps
func (h *SettingsHandler) List(w http.ResponseWriter, r *http.Request) {
	dtos, err := h.svc.List(r.Context())
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list bot applications")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, dtos)
}

// Get handles fetching a single application's details.
// GET /api/v1/admin/apps/{id}
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	dto, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot application not found")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, dto)
}

// Create registers a new application and connects its gateway bot.
// POST /api/v1/admin/apps
func (h *SettingsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	dto, err := h.svc.Create(r.Context(), req)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusCreated, dto)
}

// Update modifies an existing application configuration.
// PUT /api/v1/admin/apps/{id}
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req service.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	_, _, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update configuration: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// Delete deletes a Discord application and terminates its gateway session.
// DELETE /api/v1/admin/apps/{id}
func (h *SettingsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Delete(r.Context(), id); err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete bot application")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// Reconnect manually re-establishes connection for a specific application.
// POST /api/v1/admin/apps/{id}/reconnect
func (h *SettingsHandler) Reconnect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Reconnect(r.Context(), id); err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "RECONNECT_FAILED", "Bot reconnect failed: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true, "message": "Bot reconnected successfully"})
}
