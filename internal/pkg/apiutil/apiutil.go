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

// GetUserID retrieves the Discord user ID string from the request context.
func GetUserID(ctx context.Context) string {
	if val, ok := ctx.Value(UserIDKey).(string); ok {
		return val
	}
	return ""
}

// GetManagedGuilds retrieves the list of guild IDs the user has dashboard access permissions for.
func GetManagedGuilds(ctx context.Context) []string {
	if val, ok := ctx.Value(ManagedGuildsKey).([]string); ok {
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
