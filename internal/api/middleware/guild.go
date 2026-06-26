package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/go-chi/chi/v5"
)

type cacheEntry struct {
	roles     []string
	expiresAt time.Time
}

var (
	roleCache sync.Map
	cacheTTL  = 5 * time.Minute
)

// GuildPermissionMiddleware checks if the authenticated user has dashboard access to the requested guild.
func GuildPermissionMiddleware(guildSvc *service.GuildService, provider discordutil.DiscordSessionProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			guildIDStr := chi.URLParam(r, "guildId")
			if guildIDStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			guildID, err := strconv.ParseInt(guildIDStr, 10, 64)
			if err != nil {
				http.Error(w, `{"error": {"code": "BAD_REQUEST", "message": "Invalid guild ID"}}`, http.StatusBadRequest)
				return
			}

			userIDStr := apiutil.GetUserID(r.Context())
			if userIDStr == "" {
				http.Error(w, `{"error": {"code": "UNAUTHORIZED", "message": "User not authenticated"}}`, http.StatusUnauthorized)
				return
			}

			// 1. Level 1 Access: Check if guild is in JWT managed_guilds claim
			managedGuilds := apiutil.GetManagedGuilds(r.Context())
			isManaged := false
			for _, g := range managedGuilds {
				if g.ID == guildIDStr {
					isManaged = true
					break
				}
			}

			if isManaged {
				next.ServeHTTP(w, r)
				return
			}

			// 2. Level 2 Access: Check if user has a configured support staff role in this server
			// Check memory cache first to avoid hitting Discord rate limits
			cacheKey := guildIDStr + ":" + userIDStr
			var userRoles []string

			if val, ok := roleCache.Load(cacheKey); ok {
				entry := val.(cacheEntry)
				if time.Now().Before(entry.expiresAt) {
					userRoles = entry.roles
				}
			}

			if userRoles == nil {
				dg, err := provider.GetSessionForGuild(r.Context(), guildID)
				if err != nil {
					http.Error(w, `{"error": {"code": "NOT_FOUND", "message": "Bot session not found for this server: `+err.Error()+`"}}`, http.StatusNotFound)
					return
				}
				// Query member roles from Discord
				member, err := dg.GuildMember(guildIDStr, userIDStr)
				if err != nil {
					// Failed to fetch member (user is not in the guild or bot lacks access)
					http.Error(w, `{"error": {"code": "FORBIDDEN", "message": "You are not a member of this server"}}`, http.StatusForbidden)
					return
				}
				userRoles = member.Roles
				// Store in cache
				roleCache.Store(cacheKey, cacheEntry{
					roles:     userRoles,
					expiresAt: time.Now().Add(cacheTTL),
				})
			}

			// Convert role IDs to int64 slice
			var memberRoleIDs []int64
			for _, rIDStr := range userRoles {
				rID, err := strconv.ParseInt(rIDStr, 10, 64)
				if err == nil {
					memberRoleIDs = append(memberRoleIDs, rID)
				}
			}

			// Query DB staff roles
			isStaff, err := guildSvc.IsStaff(r.Context(), guildID, memberRoleIDs)
			if err != nil {
				http.Error(w, `{"error": {"code": "INTERNAL_ERROR", "message": "Failed to verify staff roles"}}`, http.StatusInternalServerError)
				return
			}

			if isStaff {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, `{"error": {"code": "FORBIDDEN", "message": "You do not have access to this server's dashboard"}}`, http.StatusForbidden)
		})
	}
}
