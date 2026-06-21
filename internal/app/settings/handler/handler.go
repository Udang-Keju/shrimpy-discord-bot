package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/settings/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
)

// SettingsSvc defines the service operations used by this handler.
type SettingsSvc interface {
	GetBotSettings(ctx context.Context) (*service.BotSettingsDTO, error)
	UpdateBotSettings(ctx context.Context, req service.UpdateRequest) (tokenChanged bool, newToken string, err error)
	GetDecryptedCredentials(ctx context.Context) (token, clientID, clientSecret, redirectURI string, err error)
}

// ReconnectFn is the function signature for triggering a bot reconnect.
// It receives the new plaintext token and re-opens the Discord gateway.
type ReconnectFn func(newToken string) error

// SettingsHandler handles admin API endpoints for managing bot credentials.
type SettingsHandler struct {
	svc       SettingsSvc
	reconnect ReconnectFn
}

// NewSettingsHandler constructs a new SettingsHandler.
func NewSettingsHandler(svc SettingsSvc, reconnect ReconnectFn) *SettingsHandler {
	return &SettingsHandler{svc: svc, reconnect: reconnect}
}

// Get returns the current bot settings with secrets masked.
// GET /api/v1/admin/settings
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	dto, err := h.svc.GetBotSettings(r.Context())
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load bot settings")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, dto)
}

// Update saves new credential values. Only non-empty fields are changed.
// If the Discord token changed it automatically triggers a live reconnect.
// PUT /api/v1/admin/settings
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req service.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	tokenChanged, newToken, err := h.svc.UpdateBotSettings(r.Context(), req)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save settings")
		return
	}

	// If the token changed, reconnect the bot automatically
	if tokenChanged {
		if err := h.reconnect(newToken); err != nil {
			// Settings are saved — warn but don't fail the request
			apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{
				"success": true,
				"warning": "Settings saved but bot reconnect failed: " + err.Error(),
			})
			return
		}
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// Reconnect manually re-opens the Discord gateway using the current DB token.
// POST /api/v1/admin/settings/reconnect
func (h *SettingsHandler) Reconnect(w http.ResponseWriter, r *http.Request) {
	token, _, _, _, err := h.svc.GetDecryptedCredentials(r.Context())
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to load bot token")
		return
	}

	if err := h.reconnect(token); err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "RECONNECT_FAILED", "Bot reconnect failed: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true, "message": "Bot reconnected successfully"})
}
