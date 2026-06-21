package service

import (
	"context"
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

// Create creates a reaction role message in the DB, posts the embed on Discord, and stores the message ID.
func (s *ReactionRoleService) Create(ctx context.Context, dg *discordgo.Session, guildID int64, channelID int64, title, desc string, color *int32, media *discordutil.EmbedMedia) (*model.ReactionRoleMessage, error) {
	msg := &model.ReactionRoleMessage{
		GuildID:          guildID,
		ChannelID:        channelID,
		EmbedTitle:       &title,
		EmbedDescription: &desc,
		EmbedColor:       color,
	}
	if err := msg.SetMedia(media); err != nil {
		return nil, err
	}

	created, err := s.repo.CreateMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
	}
	if color != nil {
		embed.Color = int(*color)
	}

	if media != nil {
		if media.Author != nil {
			embed.Author = &discordgo.MessageEmbedAuthor{Name: media.Author.Name}
			if media.Author.IconURL != nil {
				embed.Author.IconURL = *media.Author.IconURL
			}
			if media.Author.URL != nil {
				embed.Author.URL = *media.Author.URL
			}
		}
		if media.Thumbnail != nil {
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: media.Thumbnail.URL}
		}
		if media.Image != nil {
			embed.Image = &discordgo.MessageEmbedImage{URL: media.Image.URL}
		}
		if media.Footer != nil {
			embed.Footer = &discordgo.MessageEmbedFooter{Text: media.Footer.Text}
			if media.Footer.IconURL != nil {
				embed.Footer.IconURL = *media.Footer.IconURL
			}
		}
	}

	dgMsg, err := dg.ChannelMessageSendEmbed(fmt.Sprintf("%d", channelID), embed)
	if err != nil {
		_ = s.repo.DeleteMessage(ctx, created.ID)
		return nil, fmt.Errorf("failed to post Discord embed: %w", err)
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

// AddEmoji adds a reaction emoji and maps it to a role.
func (s *ReactionRoleService) AddEmoji(ctx context.Context, dg *discordgo.Session, messageID string, emoji string, roleID int64) (*model.ReactionRoleEmoji, error) {
	msg, err := s.repo.GetMessage(ctx, messageID)
	if err != nil {
		return nil, err
	}

	if msg.MessageID == nil {
		return nil, fmt.Errorf("cannot add emoji to a message that hasn't been posted to Discord")
	}

	mapping := &model.ReactionRoleEmoji{
		MessageID: messageID,
		Emoji:     emoji,
		RoleID:    roleID,
	}
	created, err := s.repo.AddEmoji(ctx, mapping)
	if err != nil {
		return nil, err
	}

	channelIDStr := fmt.Sprintf("%d", msg.ChannelID)
	msgIDStr := fmt.Sprintf("%d", *msg.MessageID)
	err = dg.MessageReactionAdd(channelIDStr, msgIDStr, emoji)
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

	if err := s.repo.RemoveEmoji(ctx, messageID, emoji); err != nil {
		return err
	}

	if msg.MessageID != nil {
		channelIDStr := fmt.Sprintf("%d", msg.ChannelID)
		msgIDStr := fmt.Sprintf("%d", *msg.MessageID)
		_ = dg.MessageReactionRemove(channelIDStr, msgIDStr, emoji, "@me")
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
