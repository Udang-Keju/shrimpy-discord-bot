package service_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/service"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

// mockTransport intercepts HTTP requests made by discordgo.Session
type mockTransport struct {
	roundTrip func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.roundTrip != nil {
		return m.roundTrip(req)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
	}, nil
}

// MockWelcomeRepository mocks WelcomeRepository
type MockWelcomeRepository struct {
	GetFunc        func(ctx context.Context, guildID int64) (*model.WelcomeConfig, error)
	UpsertFunc     func(ctx context.Context, cfg *model.WelcomeConfig) (*model.WelcomeConfig, error)
	SetEnabledFunc func(ctx context.Context, guildID int64, enabled bool) error
	DeleteFunc     func(ctx context.Context, guildID int64) error
}

func (m *MockWelcomeRepository) Get(ctx context.Context, guildID int64) (*model.WelcomeConfig, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, guildID)
	}
	return nil, nil
}

func (m *MockWelcomeRepository) Upsert(ctx context.Context, cfg *model.WelcomeConfig) (*model.WelcomeConfig, error) {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, cfg)
	}
	return nil, nil
}

func (m *MockWelcomeRepository) SetEnabled(ctx context.Context, guildID int64, enabled bool) error {
	if m.SetEnabledFunc != nil {
		return m.SetEnabledFunc(ctx, guildID, enabled)
	}
	return nil
}

func (m *MockWelcomeRepository) Delete(ctx context.Context, guildID int64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, guildID)
	}
	return nil
}

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestWelcomeService_GetSaveToggle(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	expectedCfg := &model.WelcomeConfig{GuildID: guildID, Enabled: true}

	t.Run("Get Config - Found", func(t *testing.T) {
		repo := &MockWelcomeRepository{
			GetFunc: func(c context.Context, id int64) (*model.WelcomeConfig, error) {
				assert.Equal(t, guildID, id)
				return expectedCfg, nil
			},
		}
		svc := service.NewWelcomeService(repo)
		cfg, err := svc.Get(ctx, guildID)
		assert.NoError(t, err)
		assert.Equal(t, expectedCfg, cfg)
	})

	t.Run("Get Config - Not Found (Returns Disabled Config)", func(t *testing.T) {
		repo := &MockWelcomeRepository{
			GetFunc: func(c context.Context, id int64) (*model.WelcomeConfig, error) {
				return nil, repository.ErrNotFound
			},
		}
		svc := service.NewWelcomeService(repo)
		cfg, err := svc.Get(ctx, guildID)
		assert.NoError(t, err)
		assert.False(t, cfg.Enabled)
		assert.Equal(t, guildID, cfg.GuildID)
	})

	t.Run("Save Config", func(t *testing.T) {
		repo := &MockWelcomeRepository{
			UpsertFunc: func(c context.Context, cfg *model.WelcomeConfig) (*model.WelcomeConfig, error) {
				return cfg, nil
			},
		}
		svc := service.NewWelcomeService(repo)
		cfg, err := svc.Save(ctx, expectedCfg)
		assert.NoError(t, err)
		assert.Equal(t, expectedCfg, cfg)
	})

	t.Run("SetEnabled Toggle", func(t *testing.T) {
		called := false
		repo := &MockWelcomeRepository{
			SetEnabledFunc: func(c context.Context, id int64, enabled bool) error {
				assert.Equal(t, guildID, id)
				assert.True(t, enabled)
				called = true
				return nil
			},
		}
		svc := service.NewWelcomeService(repo)
		err := svc.SetEnabled(ctx, guildID, true)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("Disable Config (Deletes config)", func(t *testing.T) {
		called := false
		repo := &MockWelcomeRepository{
			DeleteFunc: func(c context.Context, id int64) error {
				assert.Equal(t, guildID, id)
				called = true
				return nil
			},
		}
		svc := service.NewWelcomeService(repo)
		err := svc.Disable(ctx, guildID)
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestWelcomeService_SendWelcome(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)

	member := &discordgo.Member{
		User: &discordgo.User{
			ID:       "67890",
			Username: "joininguser",
		},
	}

	tests := []struct {
		name          string
		repoSetup     func(repo *MockWelcomeRepository)
		mockTransport func(req *http.Request, calls *map[string]string) (*http.Response, error)
		expectError   bool
	}{
		{
			name: "Config Not Found",
			repoSetup: func(repo *MockWelcomeRepository) {
				repo.GetFunc = func(c context.Context, id int64) (*model.WelcomeConfig, error) {
					return nil, repository.ErrNotFound
				}
			},
			expectError: false,
		},
		{
			name: "Config Disabled",
			repoSetup: func(repo *MockWelcomeRepository) {
				repo.GetFunc = func(c context.Context, id int64) (*model.WelcomeConfig, error) {
					return &model.WelcomeConfig{GuildID: guildID, Enabled: false}, nil
				}
			},
			expectError: false,
		},
		{
			name: "Send DM and Channel Messages",
			repoSetup: func(repo *MockWelcomeRepository) {
				repo.GetFunc = func(c context.Context, id int64) (*model.WelcomeConfig, error) {
					return &model.WelcomeConfig{
						GuildID:        guildID,
						Enabled:        true,
						DMMessage:      stringPtr("Welcome {username} to {server}! You are member #{membercount}."),
						ChannelID:      int64Ptr(888999),
						ChannelMessage: stringPtr("Hey {user}, welcome!"),
					}, nil
				}
			},
			mockTransport: func(req *http.Request, calls *map[string]string) (*http.Response, error) {
				// 1. Get Guild (GET /guilds/12345)
				if req.Method == "GET" && req.URL.Path == "/api/v9/guilds/12345" {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"id":"12345","name":"Shrimpy Server","member_count":105}`)),
					}, nil
				}
				// 2. Create DM Channel (POST /users/@me/channels)
				if req.Method == "POST" && req.URL.Path == "/api/v9/users/@me/channels" {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"id":"dm_chan_123"}`)),
					}, nil
				}
				// 3. Send DM (POST /channels/dm_chan_123/messages)
				if req.Method == "POST" && req.URL.Path == "/api/v9/channels/dm_chan_123/messages" {
					b, _ := io.ReadAll(req.Body)
					(*calls)["dm"] = string(b)
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("{}"))}, nil
				}
				// 4. Send Channel Message (POST /channels/888999/messages)
				if req.Method == "POST" && req.URL.Path == "/api/v9/channels/888999/messages" {
					b, _ := io.ReadAll(req.Body)
					(*calls)["channel"] = string(b)
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("{}"))}, nil
				}
				return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader("{}"))}, nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockWelcomeRepository{}
			tt.repoSetup(repo)
			svc := service.NewWelcomeService(repo)

			dg, _ := discordgo.New("Bot Token")
			calls := make(map[string]string)
			if tt.mockTransport != nil {
				dg.Client.Transport = &mockTransport{
					roundTrip: func(req *http.Request) (*http.Response, error) {
						return tt.mockTransport(req, &calls)
					},
				}
			}

			err := svc.SendWelcome(ctx, dg, guildID, member)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.name == "Send DM and Channel Messages" {
					assert.Contains(t, calls["dm"], "Welcome joininguser to Shrimpy Server! You are member #105.")
					assert.Contains(t, calls["channel"], "Hey \\u003c@67890\\u003e, welcome!")
				}
			}
		})
	}
}
