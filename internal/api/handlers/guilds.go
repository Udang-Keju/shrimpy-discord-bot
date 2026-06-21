package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/api/middleware"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
)

// GuildService defines the operations consumed by GuildHandler.
type GuildService interface {
	GetConfig(ctx context.Context, guildID int64) (*repository.Guild, error)
	UpdateConfig(ctx context.Context, guildID int64, updates map[string]interface{}) (*repository.Guild, error)
	UpdateNickname(ctx context.Context, dg *discordgo.Session, guildID int64, nickname *string) error
}

// GuildHandler handles dashboard requests for server configuration and bot nickname setting.
type GuildHandler struct {
	guildSvc GuildService
	dg       *discordgo.Session
}

// NewGuildHandler constructs a new GuildHandler.
func NewGuildHandler(guildSvc GuildService, dg *discordgo.Session) *GuildHandler {
	return &GuildHandler{
		guildSvc: guildSvc,
		dg:       dg,
	}
}

// List returns a list of guilds managed by the user, annotating whether Shrimpy has joined them.
func (h *GuildHandler) List(w http.ResponseWriter, r *http.Request) {
	managedGuildIDs := middleware.GetManagedGuilds(r.Context())
	if len(managedGuildIDs) == 0 {
		WriteJSON(w, http.StatusOK, []interface{}{})
		return
	}

	type guildResponse struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Icon      string `json:"icon"`
		BotJoined bool   `json:"bot_joined"`
	}

	var list []guildResponse
	for _, idStr := range managedGuildIDs {
		// Attempt to fetch guild info from bot state
		botJoined := true
		guild, err := h.dg.State.Guild(idStr)
		if err != nil {
			// Fallback: check if bot is in the guild via REST API (handles rare state misses)
			guild, err = h.dg.Guild(idStr)
			if err != nil {
				botJoined = false
			}
		}

		if botJoined && guild != nil {
			list = append(list, guildResponse{
				ID:        guild.ID,
				Name:      guild.Name,
				Icon:      guild.Icon,
				BotJoined: true,
			})
		} else {
			// Guild is in user's list but the bot hasn't joined it yet
			list = append(list, guildResponse{
				ID:        idStr,
				Name:      "Server " + idStr,
				Icon:      "",
				BotJoined: false,
			})
		}
	}

	WriteJSON(w, http.StatusOK, list)
}

// GetConfig returns the Shrimpy database configuration for a specific guild.
func (h *GuildHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	cfg, err := h.guildSvc.GetConfig(r.Context(), guildID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch guild configuration")
		return
	}

	WriteJSON(w, http.StatusOK, cfg)
}

// UpdateConfig updates fields on the guild configuration.
func (h *GuildHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON updates")
		return
	}

	// Permitted config updates
	allowedUpdates := make(map[string]interface{})
	if prefix, ok := updates["prefix"].(string); ok && len(prefix) <= 10 {
		allowedUpdates["prefix"] = prefix
	}
	if language, ok := updates["language"].(string); ok && len(language) <= 10 {
		allowedUpdates["language"] = language
	}
	if logCh, ok := updates["log_channel_id"]; ok {
		if logCh == nil {
			allowedUpdates["log_channel_id"] = nil
		} else if logChVal, err := strconv.ParseInt(fmt.Sprintf("%v", logCh), 10, 64); err == nil {
			allowedUpdates["log_channel_id"] = logChVal
		}
	}

	cfg, err := h.guildSvc.UpdateConfig(r.Context(), guildID, allowedUpdates)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update configuration")
		return
	}

	WriteJSON(w, http.StatusOK, cfg)
}

type nicknamePayload struct {
	Nickname *string `json:"nickname"`
}

// UpdateNickname sets the custom per-server display nickname of the bot.
func (h *GuildHandler) UpdateNickname(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var payload nicknamePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload")
		return
	}

	if payload.Nickname != nil && len(*payload.Nickname) > 32 {
		WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Nickname cannot exceed 32 characters")
		return
	}

	err := h.guildSvc.UpdateNickname(r.Context(), h.dg, guildID, payload.Nickname)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set nickname: "+err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, JSONResponse{"success": true})
}

// GetDiscordChannels returns text channels in the guild for dropdown selectors.
func (h *GuildHandler) GetDiscordChannels(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")

	channels, err := h.dg.GuildChannels(guildIDStr)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "DISCORD_ERROR", "Failed to fetch Discord channels: "+err.Error())
		return
	}

	type channelResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type int    `json:"type"`
	}

	var list []channelResponse
	for _, ch := range channels {
		// Only list GuildText channels
		if ch.Type == discordgo.ChannelTypeGuildText {
			list = append(list, channelResponse{
				ID:   ch.ID,
				Name: ch.Name,
				Type: int(ch.Type),
			})
		}
	}

	WriteJSON(w, http.StatusOK, list)
}

// GetDiscordRoles returns roles in the guild for role pickers.
func (h *GuildHandler) GetDiscordRoles(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")

	roles, err := h.dg.GuildRoles(guildIDStr)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "DISCORD_ERROR", "Failed to fetch Discord roles: "+err.Error())
		return
	}

	type roleResponse struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Color int    `json:"color"`
	}

	var list []roleResponse
	for _, role := range roles {
		// Ignore the global @everyone role from picker lists
		if role.Name != "@everyone" {
			list = append(list, roleResponse{
				ID:    role.ID,
				Name:  role.Name,
				Color: role.Color,
			})
		}
	}

	WriteJSON(w, http.StatusOK, list)
}
