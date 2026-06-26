package apiutil

import (
	"context"
	"encoding/json"
	"net/http"
)

type contextKey string

// Context keys used to store and retrieve authentication details.
const (
	UserIDKey        contextKey = "user_id"
	ManagedGuildsKey contextKey = "managed_guilds"
)

// Guild carries the display data for a guild the user manages. Name/Icon are
// populated from Discord's OAuth2 /users/@me/guilds response at login/refresh
// time; BotJoined is filled in live by handlers that need it.
type Guild struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	BotJoined bool   `json:"bot_joined"`
}

// GetUserID retrieves the Discord user ID string from the request context.
func GetUserID(ctx context.Context) string {
	if val, ok := ctx.Value(UserIDKey).(string); ok {
		return val
	}
	return ""
}

// GetManagedGuilds retrieves the guilds the user has dashboard access permissions for.
func GetManagedGuilds(ctx context.Context) []Guild {
	if val, ok := ctx.Value(ManagedGuildsKey).([]Guild); ok {
		return val
	}
	return nil
}

// JSONResponse is a helper type for structured API responses.
type JSONResponse map[string]interface{}

// WriteJSON sends a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteError sends a standard RFC-like error response structure.
func WriteError(w http.ResponseWriter, status int, code string, msg string) {
	WriteJSON(w, status, JSONResponse{
		"error": JSONResponse{
			"code":    code,
			"message": msg,
		},
	})
}
