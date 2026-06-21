package handlers

import (
	"encoding/json"
	"net/http"
)

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
