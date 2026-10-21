package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/model"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

type fakeRepo struct {
	clocks    []model.ClockChannel
	listErr   error
	lastNames map[string]string
}

func (r *fakeRepo) ListAll(context.Context) ([]model.ClockChannel, error) {
	return r.clocks, r.listErr
}
func (r *fakeRepo) SetLastName(_ context.Context, id, name string) error {
	r.lastNames[id] = name
	return nil
}
func (r *fakeRepo) List(context.Context, int64) ([]model.ClockChannel, error) { return nil, nil }
func (r *fakeRepo) Get(context.Context, int64, string) (*model.ClockChannel, error) {
	return nil, nil
}
func (r *fakeRepo) Create(context.Context, *model.ClockChannel) (*model.ClockChannel, error) {
	return nil, nil
}
func (r *fakeRepo) Update(context.Context, *model.ClockChannel) error { return nil }
func (r *fakeRepo) Delete(context.Context, int64, string) error       { return nil }

type fakeEditor struct {
	renames map[string]string // channel ID -> new name
	failFor map[string]bool   // channel IDs whose rename errors
}

func (e *fakeEditor) ChannelEdit(id string, data *discordgo.ChannelEdit, _ ...discordgo.RequestOption) (*discordgo.Channel, error) {
	if e.failFor[id] {
		return nil, errors.New("rate limited")
	}
	e.renames[id] = data.Name
	return &discordgo.Channel{ID: id}, nil
}

func TestScheduler_RunCheck(t *testing.T) {
	// Saturday, 27 Apr 2024 03:17 UTC.
	now := time.Date(2024, time.April, 27, 3, 17, 0, 0, time.UTC)

	tests := []struct {
		name         string
		clocks       []model.ClockChannel
		listErr      error
		sessionErr   map[int64]error
		failFor      map[string]bool
		wantRenames  map[string]string // channel ID -> name
		wantLastName map[string]string // clock ID -> saved name
	}{
		{
			name: "renames a clock whose name changed",
			clocks: []model.ClockChannel{
				{ID: "a", GuildID: 1, ChannelID: 10, Kind: model.KindTime, Format: FormatTime24h, Timezone: "UTC", LastName: "🕐 03:12 UTC"},
			},
			wantRenames:  map[string]string{"10": "🕐 03:17 UTC"},
			wantLastName: map[string]string{"a": "🕐 03:17 UTC"},
		},
		{
			name: "skips a clock that already shows the current name",
			clocks: []model.ClockChannel{
				{ID: "a", GuildID: 1, ChannelID: 10, Kind: model.KindTime, Format: FormatTime24h, Timezone: "UTC", LastName: "🕐 03:17 UTC"},
			},
			wantRenames:  map[string]string{},
			wantLastName: map[string]string{},
		},
		{
			name: "one failing channel does not block the others",
			clocks: []model.ClockChannel{
				{ID: "a", GuildID: 1, ChannelID: 10, Kind: model.KindTime, Format: FormatTime24h, Timezone: "UTC"},
				{ID: "b", GuildID: 1, ChannelID: 11, Kind: model.KindDate, Format: FormatDateISO, Timezone: "UTC"},
			},
			failFor:      map[string]bool{"10": true},
			wantRenames:  map[string]string{"11": "📅 2024-04-27"},
			wantLastName: map[string]string{"b": "📅 2024-04-27"},
		},
		{
			name: "guild without a session is skipped",
			clocks: []model.ClockChannel{
				{ID: "a", GuildID: 1, ChannelID: 10, Kind: model.KindTime, Format: FormatTime24h, Timezone: "UTC"},
				{ID: "b", GuildID: 2, ChannelID: 11, Kind: model.KindDate, Format: FormatDateISO, Timezone: "UTC"},
			},
			sessionErr:   map[int64]error{1: errors.New("bot not in guild")},
			wantRenames:  map[string]string{"11": "📅 2024-04-27"},
			wantLastName: map[string]string{"b": "📅 2024-04-27"},
		},
		{
			name: "clock with an invalid format is skipped",
			clocks: []model.ClockChannel{
				{ID: "a", GuildID: 1, ChannelID: 10, Kind: model.KindTime, Format: "bogus", Timezone: "UTC"},
			},
			wantRenames:  map[string]string{},
			wantLastName: map[string]string{},
		},
		{
			name:         "list failure renames nothing",
			listErr:      errors.New("db down"),
			wantRenames:  map[string]string{},
			wantLastName: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{clocks: tt.clocks, listErr: tt.listErr, lastNames: map[string]string{}}
			editor := &fakeEditor{renames: map[string]string{}, failFor: tt.failFor}
			sched := NewScheduler(repo, DefaultRefreshInterval)
			sched.now = func() time.Time { return now }

			sched.runCheck(context.Background(), func(_ context.Context, guildID int64) (channelEditor, error) {
				if err := tt.sessionErr[guildID]; err != nil {
					return nil, err
				}
				return editor, nil
			})

			assert.Equal(t, tt.wantRenames, editor.renames)
			assert.Equal(t, tt.wantLastName, repo.lastNames)
		})
	}
}
