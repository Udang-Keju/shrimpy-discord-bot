package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/go-chi/chi/v5"
)

// Handler manages clock channel dashboard configuration.
type Handler struct {
	clockSvc *service.ClockService
}

// NewHandler constructs a new Handler.
func NewHandler(clockSvc *service.ClockService) *Handler {
	return &Handler{clockSvc: clockSvc}
}

// clockDTO shapes the API response. Snowflakes are strings in JSON.
type clockDTO struct {
	ID             string  `json:"id"`
	ChannelID      string  `json:"channelId"`
	Kind           string  `json:"kind"`
	Format         string  `json:"format"`
	Timezone       string  `json:"timezone"`
	RefreshMinutes int     `json:"refreshMinutes"`
	Emoji          *string `json:"emoji"`
	Managed        bool    `json:"managed"`
	Name           string  `json:"name"` // channel name last written; "" until first rename
}

func toDTO(c model.ClockChannel) clockDTO {
	return clockDTO{
		ID:             c.ID,
		ChannelID:      discordutil.FormatID(c.ChannelID),
		Kind:           c.Kind,
		Format:         c.Format,
		Timezone:       c.Timezone,
		RefreshMinutes: c.RefreshMinutes,
		Emoji:          c.Emoji,
		Managed:        c.Managed,
		Name:           c.LastName,
	}
}

// List returns the clock channels configured for a guild.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	guildID, _ := strconv.ParseInt(chi.URLParam(r, "guildId"), 10, 64)

	clocks, err := h.clockSvc.List(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve clock channels")
		return
	}

	out := make([]clockDTO, 0, len(clocks))
	for _, c := range clocks {
		out = append(out, toDTO(c))
	}
	apiutil.WriteJSON(w, http.StatusOK, out)
}

// Timezones returns the allowed timezones with their current GMT offset,
// sorted by offset, for the dashboard dropdown.
func (h *Handler) Timezones(w http.ResponseWriter, r *http.Request) {
	apiutil.WriteJSON(w, http.StatusOK, service.TimezoneOptions(time.Now()))
}

// createRequest is the payload for adding a clock. Omit channelId to have the
// bot create the channel (under categoryId, if given).
type createRequest struct {
	Format         string  `json:"format"`
	Timezone       string  `json:"timezone"`
	RefreshMinutes int     `json:"refreshMinutes"`
	Emoji          string  `json:"emoji"`
	ChannelID      *string `json:"channelId"`
	CategoryID     *string `json:"categoryId"`
}

// Create adds a clock channel.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	guildID, _ := strconv.ParseInt(chi.URLParam(r, "guildId"), 10, 64)

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload")
		return
	}

	channelID, err := parseOptionalID(req.ChannelID)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid channel ID")
		return
	}
	categoryID, err := parseOptionalID(req.CategoryID)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid category ID")
		return
	}

	clock, err := h.clockSvc.Create(r.Context(), guildID, service.CreateInput{
		Format:         req.Format,
		Timezone:       req.Timezone,
		RefreshMinutes: req.RefreshMinutes,
		Emoji:          req.Emoji,
		ChannelID:      channelID,
		CategoryID:     categoryID,
	})
	if err != nil {
		writeServiceError(w, err, "Failed to create clock channel")
		return
	}
	apiutil.WriteJSON(w, http.StatusCreated, toDTO(*clock))
}

// updateRequest is the payload for changing how a clock is displayed.
type updateRequest struct {
	Format         string `json:"format"`
	Timezone       string `json:"timezone"`
	RefreshMinutes int    `json:"refreshMinutes"`
	Emoji          string `json:"emoji"`
}

// Update changes a clock's format, timezone or emoji.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	guildID, _ := strconv.ParseInt(chi.URLParam(r, "guildId"), 10, 64)

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload")
		return
	}

	clock, err := h.clockSvc.Update(r.Context(), guildID, chi.URLParam(r, "id"), service.UpdateInput{
		Format:         req.Format,
		Timezone:       req.Timezone,
		RefreshMinutes: req.RefreshMinutes,
		Emoji:          req.Emoji,
	})
	if err != nil {
		writeServiceError(w, err, "Failed to update clock channel")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, toDTO(*clock))
}

// Delete removes a clock. With ?deleteChannel=true the Discord channel is also
// deleted, but only if the bot created it.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	guildID, _ := strconv.ParseInt(chi.URLParam(r, "guildId"), 10, 64)
	deleteChannel := r.URL.Query().Get("deleteChannel") == "true"

	if err := h.clockSvc.Delete(r.Context(), guildID, chi.URLParam(r, "id"), deleteChannel); err != nil {
		writeServiceError(w, err, "Failed to delete clock channel")
		return
	}
	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// parseOptionalID parses a snowflake string, treating nil and "" as absent.
func parseOptionalID(s *string) (*int64, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	id, err := discordutil.ParseID(*s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// writeServiceError maps service errors to HTTP responses. Only validation
// and not-found errors reach the client; anything else is a generic 500.
func writeServiceError(w http.ResponseWriter, err error, internalMsg string) {
	var validation *service.ValidationError
	switch {
	case errors.As(err, &validation):
		apiutil.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", validation.Msg)
	case errors.Is(err, service.ErrNotFound):
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Clock channel not found")
	default:
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", internalMsg)
	}
}
