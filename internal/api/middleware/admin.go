package middleware

import (
	"context"
	"net/http"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
)

// AdminMiddleware restricts access to bot-wide admin endpoints (/admin/apps). A request is
// allowed if the authenticated user:
//   - is the bot owner (isOwner: owner of a managed Discord application, as reported by Discord), OR
//   - is the owner or holds ADMINISTRATOR of at least one guild where the bot is installed
//     (botJoined reports installation). Admins of guilds without the bot, and users with only
//     MANAGE_GUILD, are denied — otherwise anyone could create a server and administer the bot.
//
// This middleware must be chained AFTER AuthMiddleware.
func AdminMiddleware(
	isOwner func(userID string) bool,
	botJoined func(ctx context.Context, guildID string) bool,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isOwner(apiutil.GetUserID(r.Context())) {
				next.ServeHTTP(w, r)
				return
			}

			for _, g := range apiutil.GetManagedGuilds(r.Context()) {
				if g.IsAdmin && botJoined(r.Context(), g.ID) {
					next.ServeHTTP(w, r)
					return
				}
			}

			apiutil.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Bot owner or server administrator access required")
		})
	}
}
