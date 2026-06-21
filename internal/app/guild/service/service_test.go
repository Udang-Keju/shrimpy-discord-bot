package service_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/service"
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

// MockGuildRepository implements service.GuildRepository and service.AutoRoleRepository
type MockGuildRepository struct {
	UpsertFunc          func(ctx context.Context, guildID int64) (*model.Guild, error)
	GetByIDFunc         func(ctx context.Context, guildID int64) (*model.Guild, error)
	UpdateFunc          func(ctx context.Context, guildID int64, updates map[string]interface{}) (*model.Guild, error)
	DeactivateFunc      func(ctx context.Context, guildID int64) error
	ListStaffRolesFunc  func(ctx context.Context, guildID int64) ([]model.StaffRole, error)
	AddStaffRoleFunc    func(ctx context.Context, guildID, roleID int64) (*model.StaffRole, error)
	RemoveStaffRoleFunc func(ctx context.Context, guildID, roleID int64) error
	IsStaffRoleFunc     func(ctx context.Context, guildID int64, roleIDs []int64) (bool, error)
	ListAutoRolesFunc   func(ctx context.Context, guildID int64) ([]model.AutoRole, error)
	AddAutoRoleFunc     func(ctx context.Context, guildID, roleID int64) (*model.AutoRole, error)
	RemoveAutoRoleFunc  func(ctx context.Context, guildID, roleID int64) error
}

func (m *MockGuildRepository) Upsert(ctx context.Context, guildID int64) (*model.Guild, error) {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, guildID)
	}
	return nil, nil
}

func (m *MockGuildRepository) GetByID(ctx context.Context, guildID int64) (*model.Guild, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, guildID)
	}
	return nil, nil
}

func (m *MockGuildRepository) Update(ctx context.Context, guildID int64, updates map[string]interface{}) (*model.Guild, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, guildID, updates)
	}
	return nil, nil
}

func (m *MockGuildRepository) Deactivate(ctx context.Context, guildID int64) error {
	if m.DeactivateFunc != nil {
		return m.DeactivateFunc(ctx, guildID)
	}
	return nil
}

func (m *MockGuildRepository) ListStaffRoles(ctx context.Context, guildID int64) ([]model.StaffRole, error) {
	if m.ListStaffRolesFunc != nil {
		return m.ListStaffRolesFunc(ctx, guildID)
	}
	return nil, nil
}

func (m *MockGuildRepository) AddStaffRole(ctx context.Context, guildID, roleID int64) (*model.StaffRole, error) {
	if m.AddStaffRoleFunc != nil {
		return m.AddStaffRoleFunc(ctx, guildID, roleID)
	}
	return nil, nil
}

func (m *MockGuildRepository) RemoveStaffRole(ctx context.Context, guildID, roleID int64) error {
	if m.RemoveStaffRoleFunc != nil {
		return m.RemoveStaffRoleFunc(ctx, guildID, roleID)
	}
	return nil
}

func (m *MockGuildRepository) IsStaffRole(ctx context.Context, guildID int64, roleIDs []int64) (bool, error) {
	if m.IsStaffRoleFunc != nil {
		return m.IsStaffRoleFunc(ctx, guildID, roleIDs)
	}
	return false, nil
}

func (m *MockGuildRepository) ListAutoRoles(ctx context.Context, guildID int64) ([]model.AutoRole, error) {
	if m.ListAutoRolesFunc != nil {
		return m.ListAutoRolesFunc(ctx, guildID)
	}
	return nil, nil
}

func (m *MockGuildRepository) AddAutoRole(ctx context.Context, guildID, roleID int64) (*model.AutoRole, error) {
	if m.AddAutoRoleFunc != nil {
		return m.AddAutoRoleFunc(ctx, guildID, roleID)
	}
	return nil, nil
}

func (m *MockGuildRepository) RemoveAutoRole(ctx context.Context, guildID, roleID int64) error {
	if m.RemoveAutoRoleFunc != nil {
		return m.RemoveAutoRoleFunc(ctx, guildID, roleID)
	}
	return nil
}

func TestGuildService_GetConfig(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	expectedGuild := &model.Guild{
		GuildID:     guildID,
		IsActive:    true,
		BotNickname: nil,
	}
	dbErr := errors.New("db error")

	tests := []struct {
		name         string
		cacheSetup   func(cache *repository.GuildCache[*model.Guild])
		repoSetup    func(repo *MockGuildRepository, called *bool)
		expectError  error
		expectResult *model.Guild
	}{
		{
			name: "Cache Hit",
			cacheSetup: func(cache *repository.GuildCache[*model.Guild]) {
				cache.Set(guildID, expectedGuild)
			},
			repoSetup: func(repo *MockGuildRepository, called *bool) {
				// No repo setup needed (will panic if called)
			},
			expectResult: expectedGuild,
		},
		{
			name: "Cache Miss - Record Found in DB",
			repoSetup: func(repo *MockGuildRepository, called *bool) {
				repo.GetByIDFunc = func(c context.Context, id int64) (*model.Guild, error) {
					*called = true
					return expectedGuild, nil
				}
			},
			expectResult: expectedGuild,
		},
		{
			name: "Cache Miss - Record Not Found, Auto-Register",
			repoSetup: func(repo *MockGuildRepository, called *bool) {
				repo.GetByIDFunc = func(c context.Context, id int64) (*model.Guild, error) {
					return nil, repository.ErrNotFound
				}
				repo.UpsertFunc = func(c context.Context, id int64) (*model.Guild, error) {
					*called = true
					return expectedGuild, nil
				}
			},
			expectResult: expectedGuild,
		},
		{
			name: "Cache Miss - DB Error",
			repoSetup: func(repo *MockGuildRepository, called *bool) {
				repo.GetByIDFunc = func(c context.Context, id int64) (*model.Guild, error) {
					*called = true
					return nil, dbErr
				}
			},
			expectError: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := repository.NewGuildCache[*model.Guild](time.Minute)
			if tt.cacheSetup != nil {
				tt.cacheSetup(cache)
			}
			repo := &MockGuildRepository{}
			called := false
			if tt.repoSetup != nil {
				tt.repoSetup(repo, &called)
			}

			svc := service.NewGuildService(repo, cache)
			res, err := svc.GetConfig(ctx, guildID)

			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResult, res)
				if tt.name != "Cache Hit" {
					assert.True(t, called)
				}
			}
		})
	}
}

func TestGuildService_UpdatesAndDeactivation(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	updates := map[string]interface{}{"is_active": false}
	updatedGuild := &model.Guild{GuildID: guildID, IsActive: false}

	t.Run("UpdateConfig", func(t *testing.T) {
		cache := repository.NewGuildCache[*model.Guild](time.Minute)
		cache.Set(guildID, &model.Guild{GuildID: guildID, IsActive: true})

		repo := &MockGuildRepository{
			UpdateFunc: func(c context.Context, id int64, u map[string]interface{}) (*model.Guild, error) {
				assert.Equal(t, guildID, id)
				assert.Equal(t, updates, u)
				return updatedGuild, nil
			},
		}

		svc := service.NewGuildService(repo, cache)
		cfg, err := svc.UpdateConfig(ctx, guildID, updates)
		assert.NoError(t, err)
		assert.Equal(t, updatedGuild, cfg)

		cachedVal, found := cache.Get(guildID)
		assert.True(t, found)
		assert.Equal(t, updatedGuild, cachedVal)
	})

	t.Run("Deactivate", func(t *testing.T) {
		cache := repository.NewGuildCache[*model.Guild](time.Minute)
		cache.Set(guildID, &model.Guild{GuildID: guildID, IsActive: true})

		repoCalled := false
		repo := &MockGuildRepository{
			DeactivateFunc: func(c context.Context, id int64) error {
				assert.Equal(t, guildID, id)
				repoCalled = true
				return nil
			},
		}

		svc := service.NewGuildService(repo, cache)
		err := svc.Deactivate(ctx, guildID)
		assert.NoError(t, err)
		assert.True(t, repoCalled)

		_, found := cache.Get(guildID)
		assert.False(t, found)
	})
}

func TestGuildService_UpdateNickname(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	nick := "NewNick"

	tests := []struct {
		name          string
		discordStatus int
		discordBody   string
		repoSetup     func(repo *MockGuildRepository, called *bool)
		expectError   bool
	}{
		{
			name:          "Success",
			discordStatus: http.StatusOK,
			discordBody:   "{}",
			repoSetup: func(repo *MockGuildRepository, called *bool) {
				repo.UpdateFunc = func(c context.Context, id int64, updates map[string]interface{}) (*model.Guild, error) {
					*called = true
					assert.Equal(t, guildID, id)
					assert.Equal(t, &nick, updates["bot_nickname"])
					return &model.Guild{GuildID: guildID, BotNickname: &nick}, nil
				}
			},
			expectError: false,
		},
		{
			name:          "Discord API Error",
			discordStatus: http.StatusBadRequest,
			discordBody:   `{"message": "Invalid nickname"}`,
			repoSetup: func(repo *MockGuildRepository, called *bool) {
				// Repo should not be called due to Discord API error
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := repository.NewGuildCache[*model.Guild](time.Minute)
			repo := &MockGuildRepository{}
			repoCalled := false
			if tt.repoSetup != nil {
				tt.repoSetup(repo, &repoCalled)
			}
			svc := service.NewGuildService(repo, cache)

			dg, _ := discordgo.New("Bot Token")
			discordCalled := false
			dg.Client.Transport = &mockTransport{
				roundTrip: func(req *http.Request) (*http.Response, error) {
					discordCalled = true
					assert.Contains(t, req.URL.Path, "/guilds/12345/members/@me/nick")
					assert.Equal(t, "PATCH", req.Method)
					return &http.Response{
						StatusCode: tt.discordStatus,
						Body:       io.NopCloser(strings.NewReader(tt.discordBody)),
					}, nil
				},
			}

			err := svc.UpdateNickname(ctx, dg, guildID, &nick)
			if tt.expectError {
				assert.Error(t, err)
				assert.False(t, repoCalled)
			} else {
				assert.NoError(t, err)
				assert.True(t, discordCalled)
				assert.True(t, repoCalled)
			}
		})
	}
}

func TestGuildService_RolesManagement(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	roleID := int64(98765)
	cache := repository.NewGuildCache[*model.Guild](time.Minute)

	tests := []struct {
		name      string
		repoSetup func(repo *MockGuildRepository)
		testFunc  func(svc *service.GuildService)
	}{
		{
			name: "ListStaffRoles",
			repoSetup: func(repo *MockGuildRepository) {
				repo.ListStaffRolesFunc = func(c context.Context, id int64) ([]model.StaffRole, error) {
					return []model.StaffRole{{ID: "uuid1", GuildID: guildID, RoleID: roleID}}, nil
				}
			},
			testFunc: func(svc *service.GuildService) {
				roles, err := svc.ListStaffRoles(ctx, guildID)
				assert.NoError(t, err)
				assert.Len(t, roles, 1)
				assert.Equal(t, roleID, roles[0].RoleID)
			},
		},
		{
			name: "AddStaffRole",
			repoSetup: func(repo *MockGuildRepository) {
				repo.AddStaffRoleFunc = func(c context.Context, gID, rID int64) (*model.StaffRole, error) {
					return &model.StaffRole{ID: "uuid1", GuildID: gID, RoleID: rID}, nil
				}
			},
			testFunc: func(svc *service.GuildService) {
				role, err := svc.AddStaffRole(ctx, guildID, roleID)
				assert.NoError(t, err)
				assert.Equal(t, roleID, role.RoleID)
			},
		},
		{
			name: "RemoveStaffRole",
			repoSetup: func(repo *MockGuildRepository) {
				repo.RemoveStaffRoleFunc = func(c context.Context, gID, rID int64) error {
					return nil
				}
			},
			testFunc: func(svc *service.GuildService) {
				err := svc.RemoveStaffRole(ctx, guildID, roleID)
				assert.NoError(t, err)
			},
		},
		{
			name: "IsStaff",
			repoSetup: func(repo *MockGuildRepository) {
				repo.IsStaffRoleFunc = func(c context.Context, gID int64, rIDs []int64) (bool, error) {
					return true, nil
				}
			},
			testFunc: func(svc *service.GuildService) {
				isStaff, err := svc.IsStaff(ctx, guildID, []int64{roleID})
				assert.NoError(t, err)
				assert.True(t, isStaff)
			},
		},
		{
			name: "ListAutoRoles",
			repoSetup: func(repo *MockGuildRepository) {
				repo.ListAutoRolesFunc = func(c context.Context, id int64) ([]model.AutoRole, error) {
					return []model.AutoRole{{ID: "uuid1", GuildID: guildID, RoleID: roleID}}, nil
				}
			},
			testFunc: func(svc *service.GuildService) {
				roles, err := svc.ListAutoRoles(ctx, guildID)
				assert.NoError(t, err)
				assert.Len(t, roles, 1)
				assert.Equal(t, roleID, roles[0].RoleID)
			},
		},
		{
			name: "AddAutoRole",
			repoSetup: func(repo *MockGuildRepository) {
				repo.AddAutoRoleFunc = func(c context.Context, gID, rID int64) (*model.AutoRole, error) {
					return &model.AutoRole{ID: "uuid1", GuildID: gID, RoleID: rID}, nil
				}
			},
			testFunc: func(svc *service.GuildService) {
				role, err := svc.AddAutoRole(ctx, guildID, roleID)
				assert.NoError(t, err)
				assert.Equal(t, roleID, role.RoleID)
			},
		},
		{
			name: "RemoveAutoRole",
			repoSetup: func(repo *MockGuildRepository) {
				repo.RemoveAutoRoleFunc = func(c context.Context, gID, rID int64) error {
					return nil
				}
			},
			testFunc: func(svc *service.GuildService) {
				err := svc.RemoveAutoRole(ctx, guildID, roleID)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockGuildRepository{}
			tt.repoSetup(repo)
			svc := service.NewGuildService(repo, cache)
			tt.testFunc(svc)
		})
	}
}

func TestAutoRoleService_AssignRoles(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	userID := int64(67890)

	tests := []struct {
		name          string
		repoSetup     func(repo *MockGuildRepository)
		expectAPIKeys []string
	}{
		{
			name: "Assign Multiple Roles",
			repoSetup: func(repo *MockGuildRepository) {
				repo.ListAutoRolesFunc = func(c context.Context, id int64) ([]model.AutoRole, error) {
					return []model.AutoRole{
						{ID: "uuid1", GuildID: guildID, RoleID: 111},
						{ID: "uuid2", GuildID: guildID, RoleID: 222},
					}, nil
				}
			},
			expectAPIKeys: []string{"111", "222"},
		},
		{
			name: "No Roles Configured",
			repoSetup: func(repo *MockGuildRepository) {
				repo.ListAutoRolesFunc = func(c context.Context, id int64) ([]model.AutoRole, error) {
					return nil, nil
				}
			},
			expectAPIKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockGuildRepository{}
			tt.repoSetup(repo)
			svc := service.NewAutoRoleService(repo)

			dg, _ := discordgo.New("Bot Token")
			assignedRoles := make(map[string]bool)
			dg.Client.Transport = &mockTransport{
				roundTrip: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "PUT", req.Method)
					parts := strings.Split(req.URL.Path, "/")
					roleID := parts[len(parts)-1]
					assignedRoles[roleID] = true
					return &http.Response{
						StatusCode: http.StatusNoContent,
						Body:       io.NopCloser(strings.NewReader("")),
					}, nil
				},
			}

			err := svc.AssignRoles(ctx, dg, guildID, userID)
			assert.NoError(t, err)
			assert.Equal(t, len(tt.expectAPIKeys), len(assignedRoles))
			for _, k := range tt.expectAPIKeys {
				assert.True(t, assignedRoles[k])
			}
		})
	}
}
