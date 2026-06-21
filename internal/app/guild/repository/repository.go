package repository

import (
	"context"
	"errors"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotFound is returned when a record is not found in database.
var ErrNotFound = errors.New("record not found")

// GuildRepo handles database operations for guilds, staff_roles, and auto_roles.
type GuildRepo struct{ db *gorm.DB }

// NewGuildRepo creates a new concrete GuildRepo instance.
func NewGuildRepo(db *gorm.DB) *GuildRepo { return &GuildRepo{db: db} }

// ─── Guild ────────────────────────────────────────────────────────────────────

// Upsert registers a guild on first GUILD_CREATE, or re-activates it.
func (r *GuildRepo) Upsert(ctx context.Context, guildID int64) (*model.Guild, error) {
	g := model.Guild{GuildID: guildID}
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "guild_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"is_active": true}),
		}).
		Create(&g)
	return &g, result.Error
}

// GetByID fetches a guild by its Discord ID.
func (r *GuildRepo) GetByID(ctx context.Context, guildID int64) (*model.Guild, error) {
	var g model.Guild
	result := r.db.WithContext(ctx).First(&g, "guild_id = ?", guildID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &g, result.Error
}

// Update applies partial changes to a guild's settings.
func (r *GuildRepo) Update(ctx context.Context, guildID int64, updates map[string]interface{}) (*model.Guild, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Guild{}).
		Where("guild_id = ?", guildID).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	return r.GetByID(ctx, guildID)
}

// Deactivate marks a guild as inactive (called on GUILD_DELETE).
func (r *GuildRepo) Deactivate(ctx context.Context, guildID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Guild{}).
		Where("guild_id = ?", guildID).
		Update("is_active", false).Error
}

// ─── Staff Roles ──────────────────────────────────────────────────────────────

// ListStaffRoles returns all staff roles for a guild.
func (r *GuildRepo) ListStaffRoles(ctx context.Context, guildID int64) ([]model.StaffRole, error) {
	var roles []model.StaffRole
	result := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Order("created_at").
		Find(&roles)
	return roles, result.Error
}

// AddStaffRole idempotently adds a Discord role as a staff role.
func (r *GuildRepo) AddStaffRole(ctx context.Context, guildID, roleID int64) (*model.StaffRole, error) {
	sr := model.StaffRole{ID: uuid.NewString(), GuildID: guildID, RoleID: roleID}
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&sr)
	if result.Error != nil {
		return nil, result.Error
	}
	// If nothing was inserted (conflict), fetch the existing record
	if result.RowsAffected == 0 {
		result = r.db.WithContext(ctx).
			Where("guild_id = ? AND role_id = ?", guildID, roleID).
			First(&sr)
	}
	return &sr, result.Error
}

// RemoveStaffRole removes a role from the guild's staff list.
func (r *GuildRepo) RemoveStaffRole(ctx context.Context, guildID, roleID int64) error {
	return r.db.WithContext(ctx).
		Where("guild_id = ? AND role_id = ?", guildID, roleID).
		Delete(&model.StaffRole{}).Error
}

// IsStaffRole returns true if any of the provided role IDs are staff roles.
func (r *GuildRepo) IsStaffRole(ctx context.Context, guildID int64, roleIDs []int64) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&model.StaffRole{}).
		Where("guild_id = ? AND role_id IN ?", guildID, roleIDs).
		Count(&count)
	return count > 0, result.Error
}

// ─── Auto Roles ───────────────────────────────────────────────────────────────

// ListAutoRoles returns all auto-roles configured for a guild.
func (r *GuildRepo) ListAutoRoles(ctx context.Context, guildID int64) ([]model.AutoRole, error) {
	var roles []model.AutoRole
	result := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Order("created_at").
		Find(&roles)
	return roles, result.Error
}

// AddAutoRole idempotently adds a Discord role as an auto-role.
func (r *GuildRepo) AddAutoRole(ctx context.Context, guildID, roleID int64) (*model.AutoRole, error) {
	ar := model.AutoRole{ID: uuid.NewString(), GuildID: guildID, RoleID: roleID}
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&ar)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		result = r.db.WithContext(ctx).
			Where("guild_id = ? AND role_id = ?", guildID, roleID).
			First(&ar)
	}
	return &ar, result.Error
}

// RemoveAutoRole removes an auto-role from a guild.
func (r *GuildRepo) RemoveAutoRole(ctx context.Context, guildID, roleID int64) error {
	return r.db.WithContext(ctx).
		Where("guild_id = ? AND role_id = ?", guildID, roleID).
		Delete(&model.AutoRole{}).Error
}
