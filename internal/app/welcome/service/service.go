package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/repository"
	"github.com/bwmarrin/discordgo"
)

// WelcomeRepository defines the database operations consumed by WelcomeService.
type WelcomeRepository interface {
	Get(ctx context.Context, guildID int64) (*model.WelcomeConfig, error)
	Upsert(ctx context.Context, cfg *model.WelcomeConfig) (*model.WelcomeConfig, error)
	SetEnabled(ctx context.Context, guildID int64, enabled bool) error
	Delete(ctx context.Context, guildID int64) error
}

// WelcomeService handles retrieving, saving, and executing welcome/onboarding workflows.
type WelcomeService struct {
	repo WelcomeRepository
}

// NewWelcomeService constructs a new WelcomeService.
func NewWelcomeService(repo WelcomeRepository) *WelcomeService {
	return &WelcomeService{repo: repo}
}

// Get returns the welcome config for a server. If not found, returns an disabled config.
func (s *WelcomeService) Get(ctx context.Context, guildID int64) (*model.WelcomeConfig, error) {
	cfg, err := s.repo.Get(ctx, guildID)
	if err == repository.ErrNotFound {
		return &model.WelcomeConfig{GuildID: guildID, Enabled: false}, nil
	}
	return cfg, err
}

// Save upserts the welcome configuration.
func (s *WelcomeService) Save(ctx context.Context, cfg *model.WelcomeConfig) (*model.WelcomeConfig, error) {
	return s.repo.Upsert(ctx, cfg)
}

// SetEnabled toggles the welcome service.
func (s *WelcomeService) SetEnabled(ctx context.Context, guildID int64, enabled bool) error {
	return s.repo.SetEnabled(ctx, guildID, enabled)
}

// Disable welcome config (deletes or disables it).
func (s *WelcomeService) Disable(ctx context.Context, guildID int64) error {
	return s.repo.Delete(ctx, guildID)
}

// SendWelcome triggers onboarding messages for a new member joining a guild.
func (s *WelcomeService) SendWelcome(ctx context.Context, dg *discordgo.Session, guildID int64, member *discordgo.Member) error {
	cfg, err := s.repo.Get(ctx, guildID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil // Welcome config not initialized
		}
		return err
	}

	if !cfg.Enabled {
		return nil
	}

	// Fetch guild info for template replacement (name, member count)
	guild, err := dg.State.Guild(fmt.Sprintf("%d", guildID))
	if err != nil {
		guild, err = dg.Guild(fmt.Sprintf("%d", guildID))
		if err != nil {
			return fmt.Errorf("failed to fetch guild info: %w", err)
		}
	}

	guildName := guild.Name
	memberCount := guild.MemberCount
	username := member.User.Username
	mention := member.User.Mention()

	// Replace template variables
	replaceVars := func(text string) string {
		r := strings.NewReplacer(
			"{user}", mention,
			"{mention}", mention,
			"{username}", username,
			"{server}", guildName,
			"{membercount}", fmt.Sprintf("%d", memberCount),
		)
		return r.Replace(text)
	}

	// 1. Send DM Greeting if configured
	if cfg.DMMessage != nil && *cfg.DMMessage != "" {
		dmChannel, err := dg.UserChannelCreate(member.User.ID)
		if err == nil {
			_, _ = dg.ChannelMessageSend(dmChannel.ID, replaceVars(*cfg.DMMessage))
		}
	}

	// 2. Send Channel Greeting if channel and message/embed are configured
	if cfg.ChannelID != nil {
		channelIDStr := fmt.Sprintf("%d", *cfg.ChannelID)
		var textContent string
		if cfg.ChannelMessage != nil {
			textContent = replaceVars(*cfg.ChannelMessage)
		}

		var embed *discordgo.MessageEmbed
		media, err := cfg.GetMedia()
		if err == nil && media != nil {
			embed = &discordgo.MessageEmbed{}
			if cfg.EmbedColor != nil {
				embed.Color = int(*cfg.EmbedColor)
			}

			if media.Author != nil {
				embed.Author = &discordgo.MessageEmbedAuthor{
					Name: replaceVars(media.Author.Name),
				}
				if media.Author.IconURL != nil {
					embed.Author.IconURL = *media.Author.IconURL
				}
				if media.Author.URL != nil {
					embed.Author.URL = *media.Author.URL
				}
			}

			if media.Thumbnail != nil {
				embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
					URL: media.Thumbnail.URL,
				}
			}

			if media.Image != nil {
				embed.Image = &discordgo.MessageEmbedImage{
					URL: media.Image.URL,
				}
			}

			if media.Footer != nil {
				embed.Footer = &discordgo.MessageEmbedFooter{
					Text: replaceVars(media.Footer.Text),
				}
				if media.Footer.IconURL != nil {
					embed.Footer.IconURL = *media.Footer.IconURL
				}
			}
		}

		if textContent != "" || embed != nil {
			params := &discordgo.MessageSend{
				Content: textContent,
				Embed:   embed,
			}
			_, err = dg.ChannelMessageSendComplex(channelIDStr, params)
			if err != nil {
				return fmt.Errorf("failed to send welcome message to channel: %w", err)
			}
		}
	}

	return nil
}
