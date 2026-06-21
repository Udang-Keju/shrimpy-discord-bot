package service

import (
	"context"
	"fmt"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/cache"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/bwmarrin/discordgo"
)

// GuildRepository defines the database operations consumed by GuildService.
type GuildRepository interface {
	Upsert(ctx context.Context, guildID int64) (*repository.Guild, error)
	GetByID(ctx context.Context, guildID int64) (*repository.Guild, error)
	Update(ctx context.Context, guildID int64, updates map[string]interface{}) (*repository.Guild, error)
	Deactivate(ctx context.Context, guildID int64) error

	ListStaffRoles(ctx context.Context, guildID int64) ([]repository.StaffRole, error)
	AddStaffRole(ctx context.Context, guildID, roleID int64) (*repository.StaffRole, error)
	RemoveStaffRole(ctx context.Context, guildID, roleID int64) error
	IsStaffRole(ctx context.Context, guildID int64, roleIDs []int64) (bool, error)

	ListAutoRoles(ctx context.Context, guildID int64) ([]repository.AutoRole, error)
	AddAutoRole(ctx context.Context, guildID, roleID int64) (*repository.AutoRole, error)
	RemoveAutoRole(ctx context.Context, guildID, roleID int64) error
}

// GuildService manages server configuration, support staff roles, auto-roles, and bot nicknames.
type GuildService struct {
	repo  GuildRepository
	cache *cache.GuildCache[*repository.Guild]
}

// NewGuildService constructs a new GuildService with the given repository and cache.
func NewGuildService(repo GuildRepository, cache *cache.GuildCache[*repository.Guild]) *GuildService {
	return &GuildService{
		repo:  repo,
		cache: cache,
	}
}

// GetConfig returns the guild configuration, serving from the cache if available.
func (s *GuildService) GetConfig(ctx context.Context, guildID int64) (*repository.Guild, error) {
	if cfg, found := s.cache.Get(guildID); found {
		return cfg, nil
	}

	cfg, err := s.repo.GetByID(ctx, guildID)
	if err != nil {
		if err == repository.ErrNotFound {
			// If not found in DB, auto-register (upsert) the guild.
			cfg, err = s.repo.Upsert(ctx, guildID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	s.cache.Set(guildID, cfg)
	return cfg, nil
}

// UpdateConfig updates the guild configuration in the DB and invalidates the cache.
func (s *GuildService) UpdateConfig(ctx context.Context, guildID int64, updates map[string]interface{}) (*repository.Guild, error) {
	cfg, err := s.repo.Update(ctx, guildID, updates)
	if err != nil {
		return nil, err
	}
	s.cache.Set(guildID, cfg)
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

func (s *GuildService) ListStaffRoles(ctx context.Context, guildID int64) ([]repository.StaffRole, error) {
	return s.repo.ListStaffRoles(ctx, guildID)
}

func (s *GuildService) AddStaffRole(ctx context.Context, guildID, roleID int64) (*repository.StaffRole, error) {
	return s.repo.AddStaffRole(ctx, guildID, roleID)
}

func (s *GuildService) RemoveStaffRole(ctx context.Context, guildID, roleID int64) error {
	return s.repo.RemoveStaffRole(ctx, guildID, roleID)
}

func (s *GuildService) IsStaff(ctx context.Context, guildID int64, roleIDs []int64) (bool, error) {
	return s.repo.IsStaffRole(ctx, guildID, roleIDs)
}

// ─── Auto Roles ───────────────────────────────────────────────────────────────

func (s *GuildService) ListAutoRoles(ctx context.Context, guildID int64) ([]repository.AutoRole, error) {
	return s.repo.ListAutoRoles(ctx, guildID)
}

func (s *GuildService) AddAutoRole(ctx context.Context, guildID, roleID int64) (*repository.AutoRole, error) {
	return s.repo.AddAutoRole(ctx, guildID, roleID)
}

func (s *GuildService) RemoveAutoRole(ctx context.Context, guildID, roleID int64) error {
	return s.repo.RemoveAutoRole(ctx, guildID, roleID)
}
