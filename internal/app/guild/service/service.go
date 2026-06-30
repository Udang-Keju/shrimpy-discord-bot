package service

import (
	"context"
	"fmt"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/repository"
	"github.com/bwmarrin/discordgo"
)

// GuildRepository defines the database operations consumed by GuildService.
type GuildRepository interface {
	Upsert(ctx context.Context, guildID int64, appID *string) (*model.Guild, error)
	GetByID(ctx context.Context, guildID int64) (*model.Guild, error)
	Update(ctx context.Context, guildID int64, updates map[string]interface{}) (*model.Guild, error)
	Deactivate(ctx context.Context, guildID int64) error
	GetAppIDByClientID(ctx context.Context, clientID string) (string, error)

	ListStaffRoles(ctx context.Context, guildID int64) ([]model.StaffRole, error)
	AddStaffRole(ctx context.Context, guildID, roleID int64) (*model.StaffRole, error)
	RemoveStaffRole(ctx context.Context, guildID, roleID int64) error
	IsStaffRole(ctx context.Context, guildID int64, roleIDs []int64) (bool, error)

	ListAutoRoles(ctx context.Context, guildID int64) ([]model.AutoRole, error)
	AddAutoRole(ctx context.Context, guildID, roleID int64) (*model.AutoRole, error)
	RemoveAutoRole(ctx context.Context, guildID, roleID int64) error
}

// AutoRoleRepository defines the database operations consumed by AutoRoleService.
type AutoRoleRepository interface {
	ListAutoRoles(ctx context.Context, guildID int64) ([]model.AutoRole, error)
}

// GuildService manages server configuration, support staff roles, auto-roles, and bot nicknames.
type GuildService struct {
	repo  GuildRepository
	cache *repository.GuildCache[*model.Guild]
}

// NewGuildService constructs a new GuildService with the given repository and cache.
func NewGuildService(repo GuildRepository, cache *repository.GuildCache[*model.Guild]) *GuildService {
	return &GuildService{
		repo:  repo,
		cache: cache,
	}
}

// GetConfig returns the guild configuration, serving from the cache if available.
func (s *GuildService) GetConfig(ctx context.Context, guildID int64) (*model.Guild, error) {
	if cfg, found := s.cache.Get(guildID); found {
		return cfg, nil
	}

	cfg, err := s.repo.GetByID(ctx, guildID)
	if err != nil {
		if err == repository.ErrNotFound {
			// No persisted row yet — return an in-memory default WITHOUT writing one.
			// Persisting here would force is_active=true (see repo.Upsert), which is the
			// membership flag maintained solely by gateway events (RegisterGuild on
			// GUILD_CREATE, Deactivate on GUILD_DELETE). A dashboard/bot *read* of an
			// uninvited guild must not mark it as joined. The default is not cached so
			// that a later GUILD_CREATE registration is picked up immediately.
			return &model.Guild{GuildID: guildID, Prefix: "!", Language: "en", IsActive: false}, nil
		}
		return nil, err
	}

	s.cache.Set(guildID, cfg)
	return cfg, nil
}

// IsJoined reports whether the bot has ever joined this guild and hasn't since left it,
// per the persisted DB row. Used as a fallback when the bot's gateway session is offline,
// so a temporary disconnect doesn't make an already-invited server look uninvited.
func (s *GuildService) IsJoined(ctx context.Context, guildID int64) bool {
	cfg, err := s.repo.GetByID(ctx, guildID)
	if err != nil {
		return false
	}
	return cfg.IsActive
}

// RegisterGuild is called by the Gateway event handler when a bot joins a server or reconnects.
func (s *GuildService) RegisterGuild(ctx context.Context, guildID int64, clientID string) (*model.Guild, error) {
	appID, err := s.repo.GetAppIDByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("guild: failed to lookup application for client ID %s: %w", clientID, err)
	}

	var pAppID *string
	if appID != "" {
		pAppID = &appID
	}

	cfg, err := s.repo.Upsert(ctx, guildID, pAppID)
	if err != nil {
		return nil, fmt.Errorf("guild: failed to upsert config: %w", err)
	}

	s.cache.Set(guildID, cfg)
	return cfg, nil
}

// UpdateConfig updates the guild configuration in the DB and invalidates the cache.
func (s *GuildService) UpdateConfig(ctx context.Context, guildID int64, updates map[string]interface{}) (*model.Guild, error) {
	cfg, err := s.repo.Update(ctx, guildID, updates)
	if err != nil {
		return nil, err
	}
	s.cache.Invalidate(guildID)
	return cfg, nil
}

// Deactivate marks the guild as inactive and removes it from the cache.
func (s *GuildService) Deactivate(ctx context.Context, guildID int64) error {
	if err := s.repo.Deactivate(ctx, guildID); err != nil {
		return err
	}
	s.cache.Invalidate(guildID)
	return nil
}

// UpdateNickname updates the bot's nickname in the target guild on Discord and in the DB.
func (s *GuildService) UpdateNickname(ctx context.Context, dg *discordgo.Session, guildID int64, nickname *string) error {
	// 1. Update on Discord
	nickStr := ""
	if nickname != nil {
		nickStr = *nickname
	}
	err := dg.GuildMemberNickname(fmt.Sprintf("%d", guildID), "@me", nickStr)
	if err != nil {
		return fmt.Errorf("failed to update Discord nickname: %w", err)
	}

	// 2. Persist in Database
	_, err = s.UpdateConfig(ctx, guildID, map[string]interface{}{
		"bot_nickname": nickname,
	})
	return err
}

// ─── Support Staff Roles ──────────────────────────────────────────────────────

func (s *GuildService) ListStaffRoles(ctx context.Context, guildID int64) ([]model.StaffRole, error) {
	return s.repo.ListStaffRoles(ctx, guildID)
}

func (s *GuildService) AddStaffRole(ctx context.Context, guildID, roleID int64) (*model.StaffRole, error) {
	return s.repo.AddStaffRole(ctx, guildID, roleID)
}

func (s *GuildService) RemoveStaffRole(ctx context.Context, guildID, roleID int64) error {
	return s.repo.RemoveStaffRole(ctx, guildID, roleID)
}

func (s *GuildService) IsStaff(ctx context.Context, guildID int64, roleIDs []int64) (bool, error) {
	return s.repo.IsStaffRole(ctx, guildID, roleIDs)
}

// ─── Auto Roles ───────────────────────────────────────────────────────────────

func (s *GuildService) ListAutoRoles(ctx context.Context, guildID int64) ([]model.AutoRole, error) {
	return s.repo.ListAutoRoles(ctx, guildID)
}

func (s *GuildService) AddAutoRole(ctx context.Context, guildID, roleID int64) (*model.AutoRole, error) {
	return s.repo.AddAutoRole(ctx, guildID, roleID)
}

func (s *GuildService) RemoveAutoRole(ctx context.Context, guildID, roleID int64) error {
	return s.repo.RemoveAutoRole(ctx, guildID, roleID)
}

// AutoRoleService manages assigning roles automatically when a user joins a server.
type AutoRoleService struct {
	repo AutoRoleRepository
}

// NewAutoRoleService constructs a new AutoRoleService.
func NewAutoRoleService(repo AutoRoleRepository) *AutoRoleService {
	return &AutoRoleService{repo: repo}
}

// AssignRoles assigns all configured auto-roles to the given user in the target guild.
func (s *AutoRoleService) AssignRoles(ctx context.Context, dg *discordgo.Session, guildID int64, userID int64) error {
	roles, err := s.repo.ListAutoRoles(ctx, guildID)
	if err != nil {
		return fmt.Errorf("failed to fetch auto roles: %w", err)
	}

	if len(roles) == 0 {
		return nil
	}

	guildIDStr := fmt.Sprintf("%d", guildID)
	userIDStr := fmt.Sprintf("%d", userID)

	for _, r := range roles {
		roleIDStr := fmt.Sprintf("%d", r.RoleID)
		err = dg.GuildMemberRoleAdd(guildIDStr, userIDStr, roleIDStr)
		if err != nil {
			fmt.Printf("failed to assign auto role %s to user %s in guild %s: %v\n", roleIDStr, userIDStr, guildIDStr, err)
		}
	}

	return nil
}
