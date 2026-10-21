package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	ownerCacheTTL   = 10 * time.Minute
	ownerRetryAfter = time.Minute // after a failed lookup, keep the previous answer and retry soon
)

type ownerEntry struct {
	ids     map[string]struct{}
	expires time.Time
}

// ownerIDs extracts the user IDs that own a Discord application: its owner and, for
// team-owned apps, the team's owner. (For team apps Discord reports the team itself as
// `owner`, so the team owner's user ID is what identifies the actual person.)
func ownerIDs(app *discordgo.Application) map[string]struct{} {
	ids := make(map[string]struct{})
	if app == nil {
		return ids
	}
	if app.Owner != nil && app.Owner.ID != "" {
		ids[app.Owner.ID] = struct{}{}
	}
	if app.Team != nil && app.Team.OwnerID != "" {
		ids[app.Team.OwnerID] = struct{}{}
	}
	return ids
}

// IsBotOwner reports whether userID owns any Discord application managed by this registry,
// as reported by Discord itself (GET /oauth2/applications/@me with the bot's token). This
// needs no configuration and can't drift from who actually owns the app.
func (r *Registry) IsBotOwner(userID string) bool {
	if userID == "" {
		return false
	}
	for appID, s := range r.snapshot() {
		if _, ok := r.ownersOf(appID, s)[userID]; ok {
			return true
		}
	}
	return false
}

// ownersOf returns the (cached) owners of an application, refreshing from Discord when stale.
// If Discord can't be reached the last known answer is kept so a blip doesn't lock the owner out.
func (r *Registry) ownersOf(appID string, s *discordgo.Session) map[string]struct{} {
	r.ownerMu.Lock()
	entry, cached := r.owners[appID]
	r.ownerMu.Unlock()
	if cached && time.Now().Before(entry.expires) {
		return entry.ids
	}

	app, err := s.Application("@me")
	next := ownerEntry{ids: entry.ids, expires: time.Now().Add(ownerRetryAfter)}
	if err != nil {
		fmt.Printf("Registry: failed to look up owner of application %s: %v\n", appID, err)
	} else {
		next = ownerEntry{ids: ownerIDs(app), expires: time.Now().Add(ownerCacheTTL)}
	}

	r.ownerMu.Lock()
	r.owners[appID] = next
	r.ownerMu.Unlock()
	return next.ids
}
