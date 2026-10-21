package bot

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestOwnerIDs(t *testing.T) {
	tests := []struct {
		name string
		app  *discordgo.Application
		want []string
	}{
		{"nil app", nil, nil},
		{"personal app", &discordgo.Application{Owner: &discordgo.User{ID: "111"}}, []string{"111"}},
		{
			"team app: team id as owner plus team owner",
			&discordgo.Application{
				Owner: &discordgo.User{ID: "team-1"},
				Team:  &discordgo.Team{ID: "team-1", OwnerID: "222"},
			},
			[]string{"team-1", "222"},
		},
		{"no owner info", &discordgo.Application{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ownerIDs(tt.app)
			assert.Len(t, got, len(tt.want))
			for _, id := range tt.want {
				assert.Contains(t, got, id)
			}
		})
	}
}

func TestIsBotOwner_UsesCachedOwners(t *testing.T) {
	r := NewRegistry(nil, "", nil)
	// A session with no network access: the cached entry must satisfy the lookup so Discord is not called.
	r.sessions["app"] = &discordgo.Session{}
	r.owners["app"] = ownerEntry{ids: map[string]struct{}{"111": {}}, expires: time.Now().Add(time.Hour)}

	assert.True(t, r.IsBotOwner("111"))
	assert.False(t, r.IsBotOwner("222"))
	assert.False(t, r.IsBotOwner(""), "empty user ID must never match")
}

func TestIsBotOwner_NoSessions(t *testing.T) {
	r := NewRegistry(nil, "", nil)
	assert.False(t, r.IsBotOwner("111"))
}
