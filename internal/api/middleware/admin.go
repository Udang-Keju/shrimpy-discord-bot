package middleware

import (
	"net/http"
	"os"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
)

// AdminMiddleware restricts access to bot admin endpoints.
// A request is considered admin if the authenticated user:
//   - Has a non-empty managed_guilds list in their JWT (meaning they have
//     ADMINISTRATOR or MANAGE_GUILD on at least one bot guild), OR
//   - Their Discord user ID matches the OWNER_DISCORD_ID environment variable.
//
// This middleware must be chained AFTER AuthMiddleware.
func AdminMiddleware(next http.Handler) http.Handler {
	ownerID := os.Getenv("OWNER_DISCORD_ID")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := apiutil.GetUserID(r.Context())
		managedGuilds := apiutil.GetManagedGuilds(r.Context())

		// Superadmin override via env var
		if ownerID != "" && userID == ownerID {
			next.ServeHTTP(w, r)
			return
		}

		// Any user who manages at least one guild (has ADMINISTRATOR or MANAGE_GUILD)
		// is considered an admin for bot-wide settings purposes.
		if len(managedGuilds) > 0 {
			next.ServeHTTP(w, r)
			return
		}

		apiutil.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Admin access required")
	})
}
