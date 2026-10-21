package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
)

// maxEmojiRunes bounds the optional custom emoji prefix.
const maxEmojiRunes = 8

// renameTimeout bounds a single rename so one rate-limited channel can't
// stall the caller; discordgo otherwise blocks until the bucket resets.
const renameTimeout = 10 * time.Second

// ErrNotFound is returned when a clock channel does not exist in the guild.
var ErrNotFound = errors.New("clock channel not found")

// ValidationError reports a mistake in caller-supplied settings.
type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

func invalid(format string, args ...any) error {
	return &ValidationError{Msg: fmt.Sprintf(format, args...)}
}

// ClockRepository is the persistence the service depends on.
type ClockRepository interface {
	List(ctx context.Context, guildID int64) ([]model.ClockChannel, error)
	ListAll(ctx context.Context) ([]model.ClockChannel, error)
	Get(ctx context.Context, guildID int64, id string) (*model.ClockChannel, error)
	Create(ctx context.Context, c *model.ClockChannel) (*model.ClockChannel, error)
	Update(ctx context.Context, c *model.ClockChannel) error
	SetLastName(ctx context.Context, id, name string) error
	Delete(ctx context.Context, guildID int64, id string) error
}

// channelEditor is the slice of *discordgo.Session used to rename channels.
type channelEditor interface {
	ChannelEdit(channelID string, data *discordgo.ChannelEdit, options ...discordgo.RequestOption) (*discordgo.Channel, error)
}

// ClockService manages clock channels and renames them on demand.
type ClockService struct {
	repo     ClockRepository
	provider discordutil.DiscordSessionProvider
	now      func() time.Time
}

// NewClockService creates a new ClockService.
func NewClockService(repo ClockRepository, provider discordutil.DiscordSessionProvider) *ClockService {
	return &ClockService{repo: repo, provider: provider, now: time.Now}
}

// CreateInput describes a clock to add. When ChannelID is nil the bot creates
// a locked voice channel (under CategoryID, if set); otherwise the existing
// voice channel is linked and renamed.
type CreateInput struct {
	Format     string
	Timezone   string
	Emoji      string
	ChannelID  *int64
	CategoryID *int64
}

// UpdateInput changes how an existing clock is displayed.
type UpdateInput struct {
	Format   string
	Timezone string
	Emoji    string
}

// List returns the clock channels configured for a guild.
func (s *ClockService) List(ctx context.Context, guildID int64) ([]model.ClockChannel, error) {
	return s.repo.List(ctx, guildID)
}

// validateDisplay checks the shared display settings and returns the clock
// kind implied by the format and the normalized emoji (nil = default).
func validateDisplay(format, tz, emoji string) (kind string, emojiPtr *string, err error) {
	kind, ok := FormatKind(format)
	if !ok {
		return "", nil, invalid("unknown format %q", format)
	}
	if tz == "" {
		tz = "UTC"
	}
	if !IsValidTimezone(tz) {
		return "", nil, invalid("unsupported timezone %q", tz)
	}
	if emoji != "" {
		if len([]rune(emoji)) > maxEmojiRunes {
			return "", nil, invalid("emoji must be at most %d characters", maxEmojiRunes)
		}
		emojiPtr = &emoji
	}
	return kind, emojiPtr, nil
}

// Create adds a clock channel, creating the Discord channel if none is linked,
// and renames it right away so it doesn't show a stale name until the next tick.
func (s *ClockService) Create(ctx context.Context, guildID int64, in CreateInput) (*model.ClockChannel, error) {
	kind, emoji, err := validateDisplay(in.Format, in.Timezone, in.Emoji)
	if err != nil {
		return nil, err
	}
	tz := in.Timezone
	if tz == "" {
		tz = "UTC"
	}

	session, err := s.provider.GetSessionForGuild(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("get session for guild %d: %w", guildID, err)
	}

	clock := &model.ClockChannel{
		GuildID:  guildID,
		Kind:     kind,
		Format:   in.Format,
		Timezone: tz,
		Emoji:    emoji,
	}
	name, err := FormatName(kind, in.Format, tz, in.Emoji, s.now())
	if err != nil {
		return nil, fmt.Errorf("format clock name: %w", err)
	}

	if in.ChannelID != nil {
		if err := s.checkLinkable(session, guildID, *in.ChannelID); err != nil {
			return nil, err
		}
		clock.ChannelID = *in.ChannelID
	} else {
		ch, err := s.createLockedVoiceChannel(session, guildID, name, in.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("create clock channel: %w", err)
		}
		id, err := discordutil.ParseID(ch.ID)
		if err != nil {
			return nil, fmt.Errorf("parse created channel id: %w", err)
		}
		clock.ChannelID = id
		clock.Managed = true
		clock.LastName = name
	}

	saved, err := s.repo.Create(ctx, clock)
	if err != nil {
		if clock.Managed {
			// Don't leave an orphaned channel behind that nothing will rename.
			if _, delErr := session.ChannelDelete(discordutil.FormatID(clock.ChannelID)); delErr != nil {
				fmt.Printf("Clock Error: failed to clean up channel %d after save error: %v\n", clock.ChannelID, delErr)
			}
		}
		return nil, fmt.Errorf("save clock channel: %w", err)
	}

	if !clock.Managed {
		if err := s.refresh(ctx, session, saved); err != nil {
			// The clock is saved; the scheduler will retry on its next tick.
			fmt.Printf("Clock Error: initial rename of channel %d failed: %v\n", saved.ChannelID, err)
		}
	}
	return saved, nil
}

// Update changes the display settings of a clock and renames its channel.
func (s *ClockService) Update(ctx context.Context, guildID int64, id string, in UpdateInput) (*model.ClockChannel, error) {
	kind, emoji, err := validateDisplay(in.Format, in.Timezone, in.Emoji)
	if err != nil {
		return nil, err
	}

	clock, err := s.get(ctx, guildID, id)
	if err != nil {
		return nil, err
	}
	clock.Kind = kind
	clock.Format = in.Format
	clock.Timezone = in.Timezone
	if clock.Timezone == "" {
		clock.Timezone = "UTC"
	}
	clock.Emoji = emoji
	clock.LastName = "" // force the next refresh to write the new name

	if err := s.repo.Update(ctx, clock); err != nil {
		return nil, fmt.Errorf("update clock channel: %w", err)
	}

	if session, err := s.provider.GetSessionForGuild(ctx, guildID); err != nil {
		fmt.Printf("Clock Error: no session for guild %d, rename deferred: %v\n", guildID, err)
	} else if err := s.refresh(ctx, session, clock); err != nil {
		fmt.Printf("Clock Error: rename of channel %d failed: %v\n", clock.ChannelID, err)
	}
	return clock, nil
}

// Delete removes a clock. The Discord channel is only deleted when the bot
// created it and deleteChannel is set; linked channels are never deleted.
func (s *ClockService) Delete(ctx context.Context, guildID int64, id string, deleteChannel bool) error {
	clock, err := s.get(ctx, guildID, id)
	if err != nil {
		return err
	}

	if deleteChannel && clock.Managed {
		session, err := s.provider.GetSessionForGuild(ctx, guildID)
		if err != nil {
			return fmt.Errorf("get session for guild %d: %w", guildID, err)
		}
		if _, err := session.ChannelDelete(discordutil.FormatID(clock.ChannelID)); err != nil && !isNotFound(err) {
			return fmt.Errorf("delete clock channel: %w", err)
		}
	}

	if err := s.repo.Delete(ctx, guildID, id); err != nil {
		return fmt.Errorf("delete clock channel: %w", err)
	}
	return nil
}

func (s *ClockService) get(ctx context.Context, guildID int64, id string) (*model.ClockChannel, error) {
	clock, err := s.repo.Get(ctx, guildID, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get clock channel: %w", err)
	}
	return clock, nil
}

// checkLinkable verifies an existing channel is a voice channel of this guild.
func (s *ClockService) checkLinkable(session *discordgo.Session, guildID, channelID int64) error {
	ch, err := session.Channel(discordutil.FormatID(channelID))
	if err != nil {
		if isNotFound(err) {
			return invalid("channel %d does not exist", channelID)
		}
		return fmt.Errorf("fetch channel: %w", err)
	}
	if ch.GuildID != discordutil.FormatID(guildID) {
		return invalid("channel %d does not belong to this server", channelID)
	}
	if ch.Type != discordgo.ChannelTypeGuildVoice {
		return invalid("channel %d is not a voice channel", channelID)
	}
	return nil
}

// createLockedVoiceChannel makes a voice channel members can see but not join,
// so it reads as a display rather than a place to talk.
func (s *ClockService) createLockedVoiceChannel(session *discordgo.Session, guildID int64, name string, categoryID *int64) (*discordgo.Channel, error) {
	data := discordgo.GuildChannelCreateData{
		Name: name,
		Type: discordgo.ChannelTypeGuildVoice,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{{
			ID:    discordutil.FormatID(guildID), // @everyone role shares the guild ID
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionViewChannel,
			Deny:  discordgo.PermissionVoiceConnect,
		}},
	}
	if categoryID != nil {
		data.ParentID = discordutil.FormatID(*categoryID)
	}
	return session.GuildChannelCreateComplex(discordutil.FormatID(guildID), data)
}

// refresh renames a clock's channel to the current name unless it already
// shows it. The renaming is bounded by renameTimeout.
func (s *ClockService) refresh(ctx context.Context, editor channelEditor, c *model.ClockChannel) error {
	return refreshClock(ctx, s.repo, editor, c, s.now())
}

// refreshClock is shared by the service and the scheduler.
func refreshClock(ctx context.Context, repo ClockRepository, editor channelEditor, c *model.ClockChannel, now time.Time) error {
	emoji := ""
	if c.Emoji != nil {
		emoji = *c.Emoji
	}
	name, err := FormatName(c.Kind, c.Format, c.Timezone, emoji, now)
	if err != nil {
		return fmt.Errorf("format clock name: %w", err)
	}
	if name == c.LastName {
		return nil
	}

	editCtx, cancel := context.WithTimeout(ctx, renameTimeout)
	defer cancel()
	if _, err := editor.ChannelEdit(
		strconv.FormatInt(c.ChannelID, 10),
		&discordgo.ChannelEdit{Name: name},
		discordgo.WithContext(editCtx),
	); err != nil {
		return fmt.Errorf("rename channel %d: %w", c.ChannelID, err)
	}

	c.LastName = name
	if err := repo.SetLastName(ctx, c.ID, name); err != nil {
		return fmt.Errorf("save last name: %w", err)
	}
	return nil
}

// isNotFound reports whether err is a Discord 404 (channel already gone).
func isNotFound(err error) bool {
	var restErr *discordgo.RESTError
	return errors.As(err, &restErr) && restErr.Response != nil && restErr.Response.StatusCode == 404
}
