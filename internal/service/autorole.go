package service

import (
	"context"
	"fmt"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/bwmarrin/discordgo"
)

// AutoRoleRepository defines the database operations consumed by AutoRoleService.
type AutoRoleRepository interface {
	ListAutoRoles(ctx context.Context, guildID int64) ([]repository.AutoRole, error)
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
			// We log the error but keep trying other roles if multiple roles are configured
			// in case the bot lacks permissions for a specific role.
			fmt.Printf("failed to assign auto role %s to user %s in guild %s: %v\n", roleIDStr, userIDStr, guildIDStr, err)
		}
	}

	return nil
}
