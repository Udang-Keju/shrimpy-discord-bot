package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/apiutil"
	"github.com/stretchr/testify/assert"
)

func TestAdminMiddleware(t *testing.T) {
	const ownerID = "111"

	isOwner := func(userID string) bool { return userID == ownerID }
	// Bot is installed only in guild "joined".
	botJoined := func(_ context.Context, guildID string) bool { return guildID == "joined" }

	tests := []struct {
		name   string
		userID string
		guilds []apiutil.Guild
		want   int
	}{
		{"owner with no guilds", ownerID, nil, http.StatusOK},
		{"admin of guild with bot", "222", []apiutil.Guild{{ID: "joined", IsAdmin: true}}, http.StatusOK},
		{"admin of guild without bot", "222", []apiutil.Guild{{ID: "other", IsAdmin: true}}, http.StatusForbidden},
		{"manage-guild only on guild with bot", "222", []apiutil.Guild{{ID: "joined", IsAdmin: false}}, http.StatusForbidden},
		{"admin elsewhere plus manage-guild on joined", "222", []apiutil.Guild{{ID: "other", IsAdmin: true}, {ID: "joined"}}, http.StatusForbidden},
		{"no guilds", "222", nil, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			req := httptest.NewRequest(http.MethodGet, "/admin/apps", nil)
			ctx := context.WithValue(req.Context(), apiutil.UserIDKey, tt.userID)
			ctx = context.WithValue(ctx, apiutil.ManagedGuildsKey, tt.guilds)

			rec := httptest.NewRecorder()
			AdminMiddleware(isOwner, botJoined)(next).ServeHTTP(rec, req.WithContext(ctx))

			assert.Equal(t, tt.want, rec.Code)
		})
	}
}
