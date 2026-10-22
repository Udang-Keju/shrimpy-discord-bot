package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
)

// ReactionRoleRepository defines the database operations consumed by ReactionRoleService.
type ReactionRoleRepository interface {
	ListByGuild(ctx context.Context, guildID int64) ([]model.ReactionRoleMessage, error)
	GetMessage(ctx context.Context, messageID string) (*model.ReactionRoleMessage, error)
	GetByDiscordMessageID(ctx context.Context, discordMsgID int64) (*model.ReactionRoleMessage, error)
	CreateMessage(ctx context.Context, msg *model.ReactionRoleMessage) (*model.ReactionRoleMessage, error)
	UpdateMessage(ctx context.Context, msg *model.ReactionRoleMessage) (*model.ReactionRoleMessage, error)
	SetDiscordMessageID(ctx context.Context, id string, discordMsgID int64) error
	DeleteMessage(ctx context.Context, messageID string) error

	AddEmoji(ctx context.Context, e *model.ReactionRoleEmoji) (*model.ReactionRoleEmoji, error)
	RemoveEmoji(ctx context.Context, messageID, emoji string) error
	GetEmojiRole(ctx context.Context, discordMsgID int64, emoji string) (*model.ReactionRoleEmoji, error)
}

// ReactionRoleService coordinates posting reaction role embeds, managing mappings, and granting/revoking roles.
type ReactionRoleService struct {
	repo ReactionRoleRepository
}

// NewReactionRoleService constructs a new ReactionRoleService.
func NewReactionRoleService(repo ReactionRoleRepository) *ReactionRoleService {
	return &ReactionRoleService{repo: repo}
}

func (s *ReactionRoleService) List(ctx context.Context, guildID int64) ([]model.ReactionRoleMessage, error) {
	return s.repo.ListByGuild(ctx, guildID)
}

func (s *ReactionRoleService) Get(ctx context.Context, messageID string) (*model.ReactionRoleMessage, error) {
	return s.repo.GetMessage(ctx, messageID)
}

// ErrEmptyMessage is returned when a reaction role message has neither plain text nor embed content.
var ErrEmptyMessage = errors.New("message needs plain text or embed content")

// CreateParams holds the message content for a new reaction role message. Content is plain
// text (Discord markdown); the embed fields are all optional, but at least one of Content or
// an embed field must be set.
type CreateParams struct {
	GuildID     int64
	ChannelID   int64
	Content     string
	Title       string
	Description string
	Color       *int32
	Media       *discordutil.EmbedMedia
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// buildMessage validates the params and renders them into the Discord message parts.
func buildMessage(p CreateParams) (content string, embed *discordgo.MessageEmbed, err error) {
	fields := discordutil.EmbedFields{
		Title:       nilIfEmpty(p.Title),
		Description: nilIfEmpty(p.Description),
		Color:       p.Color,
		Media:       p.Media,
	}
	content, embed = discordutil.BuildContentAndEmbed(nilIfEmpty(p.Content), fields, func(s string) string { return s })
	// An embed with only a color has nothing to show, and Discord rejects it.
	if embed != nil && embed.Title == "" && embed.Description == "" && embed.Author == nil &&
		embed.Thumbnail == nil && embed.Image == nil && embed.Footer == nil {
		embed = nil
	}
	if content == "" && embed == nil {
		return "", nil, ErrEmptyMessage
	}
	return content, embed, nil
}

// applyParams copies the editable message fields from p onto msg.
func applyParams(msg *model.ReactionRoleMessage, p CreateParams) error {
	msg.ChannelID = p.ChannelID
	msg.Content = nilIfEmpty(p.Content)
	msg.EmbedTitle = nilIfEmpty(p.Title)
	msg.EmbedDescription = nilIfEmpty(p.Description)
	msg.EmbedColor = p.Color
	return msg.SetMedia(p.Media)
}

func messageSend(content string, embed *discordgo.MessageEmbed) *discordgo.MessageSend {
	send := &discordgo.MessageSend{Content: content}
	if embed != nil {
		send.Embeds = []*discordgo.MessageEmbed{embed}
	}
	return send
}

// Create creates a reaction role message in the DB, posts it on Discord (plain text, embed, or both), and stores the message ID.
func (s *ReactionRoleService) Create(ctx context.Context, dg *discordgo.Session, p CreateParams) (*model.ReactionRoleMessage, error) {
	content, embed, err := buildMessage(p)
	if err != nil {
		return nil, err
	}

	msg := &model.ReactionRoleMessage{GuildID: p.GuildID}
	if err := applyParams(msg, p); err != nil {
		return nil, err
	}

	created, err := s.repo.CreateMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	dgMsg, err := dg.ChannelMessageSendComplex(fmt.Sprintf("%d", p.ChannelID), messageSend(content, embed))
	if err != nil {
		_ = s.repo.DeleteMessage(ctx, created.ID)
		return nil, fmt.Errorf("failed to post Discord message: %w", err)
	}

	discordMsgID, err := discordutil.ParseID(dgMsg.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid Discord message ID returned: %w", err)
	}

	err = s.repo.SetDiscordMessageID(ctx, created.ID, discordMsgID)
	if err != nil {
		return nil, err
	}

	created.MessageID = &discordMsgID
	return created, nil
}

// Update changes a reaction role message's content and re-renders it on Discord. Editing in
// place keeps the existing reactions; if the channel changed, the message is reposted in the
// new channel, the existing emoji reactions are re-added, and the old message is deleted.
func (s *ReactionRoleService) Update(ctx context.Context, dg *discordgo.Session, id string, p CreateParams) (*model.ReactionRoleMessage, error) {
	content, embed, err := buildMessage(p)
	if err != nil {
		return nil, err
	}

	msg, err := s.repo.GetMessage(ctx, id)
	if err != nil {
		return nil, err
	}
	oldChannelID := msg.ChannelID
	var oldDiscordMsgID *int64 = msg.MessageID

	if err := applyParams(msg, p); err != nil {
		return nil, err
	}

	switch {
	case oldDiscordMsgID == nil || oldChannelID != p.ChannelID:
		dgMsg, err := dg.ChannelMessageSendComplex(fmt.Sprintf("%d", p.ChannelID), messageSend(content, embed))
		if err != nil {
			return nil, fmt.Errorf("failed to post Discord message: %w", err)
		}
		newID, err := discordutil.ParseID(dgMsg.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid Discord message ID returned: %w", err)
		}
		for _, e := range msg.Emojis {
			if err := dg.MessageReactionAdd(dgMsg.ChannelID, dgMsg.ID, e.Emoji); err != nil {
				fmt.Printf("warning: failed to re-add reaction %q: %v\n", e.Emoji, err)
			}
		}
		if oldDiscordMsgID != nil {
			_ = dg.ChannelMessageDelete(fmt.Sprintf("%d", oldChannelID), fmt.Sprintf("%d", *oldDiscordMsgID))
		}
		msg.MessageID = &newID
	default:
		// Pointers (and a non-nil empty slice) so that clearing the text or the embed is sent.
		embeds := []*discordgo.MessageEmbed{}
		if embed != nil {
			embeds = append(embeds, embed)
		}
		edit := discordgo.NewMessageEdit(fmt.Sprintf("%d", p.ChannelID), fmt.Sprintf("%d", *oldDiscordMsgID))
		edit.Content = &content
		edit.Embeds = &embeds
		if _, err := dg.ChannelMessageEditComplex(edit); err != nil {
			return nil, fmt.Errorf("failed to edit Discord message: %w", err)
		}
	}

	updated, err := s.repo.UpdateMessage(ctx, msg)
	if err != nil {
		return nil, err
	}
	if msg.MessageID != oldDiscordMsgID {
		if err := s.repo.SetDiscordMessageID(ctx, id, *msg.MessageID); err != nil {
			return nil, err
		}
	}
	return updated, nil
}

// AddEmoji adds a reaction emoji and maps it to a role.
func (s *ReactionRoleService) AddEmoji(ctx context.Context, dg *discordgo.Session, messageID string, emoji string, roleID int64) (*model.ReactionRoleEmoji, error) {
	msg, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return nil, err
	}

	if msg.MessageID == nil {
		return nil, fmt.Errorf("cannot add emoji to a message that hasn't been posted to Discord")
	}

	// Store and react using the reaction API form (name:id for custom emoji, or the
	// raw unicode char). This is the same key gateway reaction events resolve to via
	// (*Emoji).APIName, so grants/revokes in HandleReactionAdd/Remove match on lookup.
	key := discordutil.ReactionEmojiAPIName(emoji)

	mapping := &model.ReactionRoleEmoji{
		MessageID: messageID,
		Emoji:     key,
		RoleID:    roleID,
	}
	created, err := s.repo.AddEmoji(ctx, mapping)
	if err != nil {
		return nil, err
	}

	channelIDStr := fmt.Sprintf("%d", msg.ChannelID)
	msgIDStr := fmt.Sprintf("%d", *msg.MessageID)
	err = dg.MessageReactionAdd(channelIDStr, msgIDStr, key)
	if err != nil {
		fmt.Printf("warning: failed to add reaction reaction to message: %v\n", err)
	}

	return created, nil
}

// RemoveEmoji removes an emoji mapping.
func (s *ReactionRoleService) RemoveEmoji(ctx context.Context, dg *discordgo.Session, messageID string, emoji string) error {
	msg, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return err
	}

	key := discordutil.ReactionEmojiAPIName(emoji)
	if err := s.repo.RemoveEmoji(ctx, messageID, key); err != nil {
		return err
	}

	if msg.MessageID != nil {
		channelIDStr := fmt.Sprintf("%d", msg.ChannelID)
		msgIDStr := fmt.Sprintf("%d", *msg.MessageID)
		_ = dg.MessageReactionRemove(channelIDStr, msgIDStr, key, "@me")
	}

	return nil
}

// Delete removes the reaction role message and deletes the Discord message.
func (s *ReactionRoleService) Delete(ctx context.Context, dg *discordgo.Session, messageID string) error {
	msg, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return err
	}

	if msg.MessageID != nil {
		_ = dg.ChannelMessageDelete(fmt.Sprintf("%d", msg.ChannelID), fmt.Sprintf("%d", *msg.MessageID))
	}

	return s.repo.DeleteMessage(ctx, messageID)
}

// HandleReactionAdd grants the role corresponding to the reaction.
func (s *ReactionRoleService) HandleReactionAdd(ctx context.Context, dg *discordgo.Session, discordMsgID int64, guildID int64, userID int64, emojiNameOrID string) error {
	emoji, err := s.repo.GetEmojiRole(ctx, discordMsgID, emojiNameOrID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil
		}
		return err
	}

	guildIDStr := fmt.Sprintf("%d", guildID)
	userIDStr := fmt.Sprintf("%d", userID)
	roleIDStr := fmt.Sprintf("%d", emoji.RoleID)

	return dg.GuildMemberRoleAdd(guildIDStr, userIDStr, roleIDStr)
}

// HandleReactionRemove revokes the role corresponding to the reaction.
func (s *ReactionRoleService) HandleReactionRemove(ctx context.Context, dg *discordgo.Session, discordMsgID int64, guildID int64, userID int64, emojiNameOrID string) error {
	emoji, err := s.repo.GetEmojiRole(ctx, discordMsgID, emojiNameOrID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil
		}
		return err
	}

	guildIDStr := fmt.Sprintf("%d", guildID)
	userIDStr := fmt.Sprintf("%d", userID)
	roleIDStr := fmt.Sprintf("%d", emoji.RoleID)

	return dg.GuildMemberRoleRemove(guildIDStr, userIDStr, roleIDStr)
}
