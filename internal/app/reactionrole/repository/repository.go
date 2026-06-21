package repository

import (
	"context"
	"errors"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotFound is returned when a reaction role record is not found.
var ErrNotFound = errors.New("record not found")

// ReactionRoleRepo handles database operations for reaction role messages and emojis.
type ReactionRoleRepo struct{ db *gorm.DB }

// NewReactionRoleRepo creates a new concrete ReactionRoleRepo instance.
func NewReactionRoleRepo(db *gorm.DB) *ReactionRoleRepo { return &ReactionRoleRepo{db: db} }

// ─── Reaction Role Messages ───────────────────────────────────────────────────

// ListByGuild returns all reaction role messages for a guild with emojis loaded.
func (r *ReactionRoleRepo) ListByGuild(ctx context.Context, guildID int64) ([]model.ReactionRoleMessage, error) {
	var msgs []model.ReactionRoleMessage
	result := r.db.WithContext(ctx).
		Preload("Emojis", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at")
		}).
		Where("guild_id = ?", guildID).
		Order("created_at").
		Find(&msgs)
	return msgs, result.Error
}

// GetMessage returns a single reaction role message with its emojis.
func (r *ReactionRoleRepo) GetMessage(ctx context.Context, messageID string) (*model.ReactionRoleMessage, error) {
	var msg model.ReactionRoleMessage
	result := r.db.WithContext(ctx).
		Preload("Emojis").
		First(&msg, "id = ?", messageID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &msg, result.Error
}

// GetByDiscordMessageID finds a reaction role message by the Discord message ID.
func (r *ReactionRoleRepo) GetByDiscordMessageID(ctx context.Context, discordMsgID int64) (*model.ReactionRoleMessage, error) {
	var msg model.ReactionRoleMessage
	result := r.db.WithContext(ctx).
		Preload("Emojis").
		First(&msg, "message_id = ?", discordMsgID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &msg, result.Error
}

// CreateMessage inserts a new reaction role message record.
func (r *ReactionRoleRepo) CreateMessage(ctx context.Context, msg *model.ReactionRoleMessage) (*model.ReactionRoleMessage, error) {
	msg.ID = uuid.NewString()
	return msg, r.db.WithContext(ctx).Create(msg).Error
}

// UpdateMessage saves changes to an existing reaction role message.
func (r *ReactionRoleRepo) UpdateMessage(ctx context.Context, msg *model.ReactionRoleMessage) (*model.ReactionRoleMessage, error) {
	result := r.db.WithContext(ctx).
		Model(msg).
		Updates(map[string]interface{}{
			"embed_title":       msg.EmbedTitle,
			"embed_description": msg.EmbedDescription,
			"embed_color":       msg.EmbedColor,
			"embed_media":       msg.EmbedMedia,
		})
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return msg, result.Error
}

// SetDiscordMessageID stores the Discord message ID after the embed is posted.
func (r *ReactionRoleRepo) SetDiscordMessageID(ctx context.Context, id string, discordMsgID int64) error {
	return r.db.WithContext(ctx).
		Model(&model.ReactionRoleMessage{}).
		Where("id = ?", id).
		Update("message_id", discordMsgID).Error
}

// DeleteMessage removes a reaction role message and its emojis (CASCADE).
func (r *ReactionRoleRepo) DeleteMessage(ctx context.Context, messageID string) error {
	return r.db.WithContext(ctx).Where("id = ?", messageID).Delete(&model.ReactionRoleMessage{}).Error
}

// ─── Reaction Role Emojis ─────────────────────────────────────────────────────

// AddEmoji adds an emoji→role mapping to a reaction role message (idempotent).
func (r *ReactionRoleRepo) AddEmoji(ctx context.Context, e *model.ReactionRoleEmoji) (*model.ReactionRoleEmoji, error) {
	e.ID = uuid.NewString()
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_id"}, {Name: "emoji"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"role_id": e.RoleID}),
		}).
		Create(e)
	return e, result.Error
}

// RemoveEmoji deletes an emoji mapping from a reaction role message.
func (r *ReactionRoleRepo) RemoveEmoji(ctx context.Context, messageID, emoji string) error {
	return r.db.WithContext(ctx).
		Where("message_id = ? AND emoji = ?", messageID, emoji).
		Delete(&model.ReactionRoleEmoji{}).Error
}

// GetEmojiRole finds the role ID mapped to a specific emoji on a specific message.
func (r *ReactionRoleRepo) GetEmojiRole(ctx context.Context, discordMsgID int64, emoji string) (*model.ReactionRoleEmoji, error) {
	var e model.ReactionRoleEmoji
	result := r.db.WithContext(ctx).
		Joins("JOIN reaction_role_messages rrm ON rrm.id = reaction_role_emojis.message_id").
		Where("rrm.message_id = ? AND reaction_role_emojis.emoji = ?", discordMsgID, emoji).
		First(&e)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &e, result.Error
}
