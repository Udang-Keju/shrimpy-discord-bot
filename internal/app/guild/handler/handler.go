package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
)

// Handler handles dashboard requests for server configuration, roles and bot nickname.
type Handler struct {
	guildSvc *service.GuildService
	provider discordutil.DiscordSessionProvider
}

// NewHandler constructs a new Handler.
func NewHandler(guildSvc *service.GuildService, provider discordutil.DiscordSessionProvider) *Handler {
	return &Handler{
		guildSvc: guildSvc,
		provider: provider,
	}
}

// List returns a list of guilds managed by the user, annotating whether Shrimpy has joined them.
// Name/icon come from the OAuth2 guild list cached at login (so unjoined guilds still display
// correctly); when the bot has joined we prefer the live Discord session data since it's fresher.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	managedGuilds := apiutil.GetManagedGuilds(r.Context())
	if len(managedGuilds) == 0 {
		apiutil.WriteJSON(w, http.StatusOK, []interface{}{})
		return
	}

	list := make([]apiutil.Guild, 0, len(managedGuilds))
	for _, g := range managedGuilds {
		g.BotJoined = h.provider.IsBotInGuild(g.ID)

		if g.BotJoined {
			if gID, err := strconv.ParseInt(g.ID, 10, 64); err == nil {
				if dg, sErr := h.provider.GetSessionForGuild(r.Context(), gID); sErr == nil {
					guild, err := dg.State.Guild(g.ID)
					if err != nil {
						guild, err = dg.Guild(g.ID)
					}
					if err == nil && guild != nil {
						g.Name = guild.Name
						if guild.Icon != "" {
							g.Icon = discordutil.GuildIconURL(g.ID, guild.Icon)
						}
					}
				}
			}
		}

		list = append(list, g)
	}

	apiutil.WriteJSON(w, http.StatusOK, list)
}

// GetConfig returns the Shrimpy database configuration for a specific guild.
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	cfg, err := h.guildSvc.GetConfig(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch guild configuration")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, cfg)
}

// UpdateConfig updates fields on the guild configuration.
func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON updates")
		return
	}

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
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update configuration")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, cfg)
}

type nicknamePayload struct {
	Nickname *string `json:"nickname"`
}

// UpdateNickname sets the custom per-server display nickname of the bot.
func (h *Handler) UpdateNickname(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var payload nicknamePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload")
		return
	}

	if payload.Nickname != nil && len(*payload.Nickname) > 32 {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Nickname cannot exceed 32 characters")
		return
	}

	dg, err := h.provider.GetSessionForGuild(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	err = h.guildSvc.UpdateNickname(r.Context(), dg, guildID, payload.Nickname)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set nickname: "+err.Error())
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// GetDiscordChannels returns text channels in the guild for dropdown selectors.
func (h *Handler) GetDiscordChannels(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	dg, err := h.provider.GetSessionForGuild(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	channels, err := dg.GuildChannels(guildIDStr)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "DISCORD_ERROR", "Failed to fetch Discord channels: "+err.Error())
		return
	}

	type channelResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type int    `json:"type"`
	}

	var list []channelResponse
	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildText {
			list = append(list, channelResponse{
				ID:   ch.ID,
				Name: ch.Name,
				Type: int(ch.Type),
			})
		}
	}

	apiutil.WriteJSON(w, http.StatusOK, list)
}

// GetDiscordRoles returns roles in the guild for role pickers.
func (h *Handler) GetDiscordRoles(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	dg, err := h.provider.GetSessionForGuild(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Bot session not found: "+err.Error())
		return
	}

	roles, err := dg.GuildRoles(guildIDStr)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "DISCORD_ERROR", "Failed to fetch Discord roles: "+err.Error())
		return
	}

	type roleResponse struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Color int    `json:"color"`
	}

	var list []roleResponse
	for _, role := range roles {
		if role.Name != "@everyone" {
			list = append(list, roleResponse{
				ID:    role.ID,
				Name:  role.Name,
				Color: role.Color,
			})
		}
	}

	apiutil.WriteJSON(w, http.StatusOK, list)
}

// ─── Auto Roles ───────────────────────────────────────────────────────────────

func (h *Handler) ListAutoRoles(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	roles, err := h.guildSvc.ListAutoRoles(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list auto roles")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, roles)
}

type rolePayload struct {
	RoleID string `json:"role_id"`
}

func (h *Handler) AddAutoRole(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var payload rolePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	roleID, err := strconv.ParseInt(payload.RoleID, 10, 64)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Role ID format")
		return
	}

	created, err := h.guildSvc.AddAutoRole(r.Context(), guildID, roleID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save auto role")
		return
	}

	apiutil.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) RemoveAutoRole(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	roleIDStr := chi.URLParam(r, "roleId")
	roleID, _ := strconv.ParseInt(roleIDStr, 10, 64)

	err := h.guildSvc.RemoveAutoRole(r.Context(), guildID, roleID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove auto role")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}

// ─── Staff Roles ──────────────────────────────────────────────────────────────

func (h *Handler) ListStaffRoles(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	roles, err := h.guildSvc.ListStaffRoles(r.Context(), guildID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list staff roles")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, roles)
}

func (h *Handler) AddStaffRole(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	var payload rolePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid payload")
		return
	}

	roleID, err := strconv.ParseInt(payload.RoleID, 10, 64)
	if err != nil {
		apiutil.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid Role ID format")
		return
	}

	created, err := h.guildSvc.AddStaffRole(r.Context(), guildID, roleID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save staff role")
		return
	}

	apiutil.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) RemoveStaffRole(w http.ResponseWriter, r *http.Request) {
	guildIDStr := chi.URLParam(r, "guildId")
	guildID, _ := strconv.ParseInt(guildIDStr, 10, 64)

	roleIDStr := chi.URLParam(r, "roleId")
	roleID, _ := strconv.ParseInt(roleIDStr, 10, 64)

	err := h.guildSvc.RemoveStaffRole(r.Context(), guildID, roleID)
	if err != nil {
		apiutil.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove staff role")
		return
	}

	apiutil.WriteJSON(w, http.StatusOK, apiutil.JSONResponse{"success": true})
}
