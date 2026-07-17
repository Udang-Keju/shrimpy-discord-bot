package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	guildmodel "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/crypto"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/translate"
	"github.com/bwmarrin/discordgo"
)

// MaskedAPIKey is the placeholder returned to the dashboard in place of a
// stored key. When the dashboard sends this value back on save, the existing
// key is preserved rather than re-encrypted.
const MaskedAPIKey = "***"

const (
	defaultTargetLang = "en"
	minTranslateChars = 2
	maxTranslateChars = 3000
	embedColor        = 0x5865F2 // Discord blurple
)

// TranslationRepository defines the database operations consumed by TranslationService.
type TranslationRepository interface {
	GetConfig(ctx context.Context, guildID int64) (*model.TranslationConfig, error)
	UpsertConfig(ctx context.Context, cfg *model.TranslationConfig) (*model.TranslationConfig, error)
	SetEnabled(ctx context.Context, guildID int64, enabled bool) error
	ListChannels(ctx context.Context, guildID int64) ([]model.TranslationChannel, error)
	AddChannel(ctx context.Context, ch *model.TranslationChannel) (*model.TranslationChannel, error)
	RemoveChannel(ctx context.Context, guildID, channelID int64) error
	ListEmojis(ctx context.Context, guildID int64) ([]model.TranslationReactionEmoji, error)
	AddEmoji(ctx context.Context, e *model.TranslationReactionEmoji) (*model.TranslationReactionEmoji, error)
	RemoveEmoji(ctx context.Context, guildID int64, emoji string) error
}

// GuildConfigProvider resolves the fallback target language from guild config.
type GuildConfigProvider interface {
	GetConfig(ctx context.Context, guildID int64) (*guildmodel.Guild, error)
}

// TranslatorFactory builds a Translator for a provider/key/endpoint. It is a
// field so tests can inject a fake without hitting the network.
type TranslatorFactory func(provider, apiKey, endpoint string) (translate.Translator, error)

// TranslationService handles retrieving, saving, and executing translations.
type TranslationService struct {
	repo          TranslationRepository
	guildProvider GuildConfigProvider
	tokenEncKey   []byte
	newTranslator TranslatorFactory
}

// NewTranslationService constructs a new TranslationService.
func NewTranslationService(repo TranslationRepository, guildProvider GuildConfigProvider, tokenEncKey []byte) *TranslationService {
	return &TranslationService{
		repo:          repo,
		guildProvider: guildProvider,
		tokenEncKey:   tokenEncKey,
		newTranslator: translate.NewTranslator,
	}
}

// GetConfig returns the translation config for a guild, or a disabled default
// when none exists.
func (s *TranslationService) GetConfig(ctx context.Context, guildID int64) (*model.TranslationConfig, error) {
	cfg, err := s.repo.GetConfig(ctx, guildID)
	if err == repository.ErrNotFound {
		return &model.TranslationConfig{GuildID: guildID, Provider: translate.ProviderDeepL, ReactionDelivery: model.ReactionDeliveryChannel}, nil
	}
	return cfg, err
}

// HasAPIKey reports whether a guild has a stored (encrypted) engine key.
func (s *TranslationService) HasAPIKey(ctx context.Context, guildID int64) bool {
	cfg, err := s.repo.GetConfig(ctx, guildID)
	if err != nil {
		return false
	}
	return len(cfg.APIKeyEnc) > 0
}

// SaveConfigInput carries a dashboard save request.
type SaveConfigInput struct {
	Enabled          bool
	AutoEnabled      bool
	ReactionEnabled  bool
	ReactionDelivery string // "channel" or "dm"; anything else normalizes to "channel"
	Provider         string
	APIKey           string // plaintext; "" or MaskedAPIKey preserves the stored key
	EndpointURL      *string
	TargetLang       *string
}

// SaveConfig upserts the translation configuration, encrypting a newly supplied
// API key and preserving the existing key when the masked placeholder is sent.
func (s *TranslationService) SaveConfig(ctx context.Context, guildID int64, in SaveConfigInput) (*model.TranslationConfig, error) {
	existing, err := s.repo.GetConfig(ctx, guildID)
	if err != nil && err != repository.ErrNotFound {
		return nil, err
	}

	cfg := &model.TranslationConfig{GuildID: guildID}
	if existing != nil {
		cfg.APIKeyEnc = existing.APIKeyEnc // preserve unless replaced below
	}

	cfg.Enabled = in.Enabled
	cfg.AutoEnabled = in.AutoEnabled
	cfg.ReactionEnabled = in.ReactionEnabled
	cfg.ReactionDelivery = model.ReactionDeliveryChannel
	if in.ReactionDelivery == model.ReactionDeliveryDM {
		cfg.ReactionDelivery = model.ReactionDeliveryDM
	}
	cfg.Provider = in.Provider
	if cfg.Provider == "" {
		cfg.Provider = translate.ProviderDeepL
	}
	cfg.EndpointURL = normalizeStrPtr(in.EndpointURL)
	cfg.TargetLang = normalizeStrPtr(in.TargetLang)

	if in.APIKey != "" && in.APIKey != MaskedAPIKey {
		enc, encErr := crypto.Encrypt([]byte(in.APIKey), s.tokenEncKey)
		if encErr != nil {
			return nil, fmt.Errorf("translation: encrypt api key: %w", encErr)
		}
		cfg.APIKeyEnc = enc
	}

	return s.repo.UpsertConfig(ctx, cfg)
}

// SetEnabled toggles the master translation switch.
func (s *TranslationService) SetEnabled(ctx context.Context, guildID int64, enabled bool) error {
	return s.repo.SetEnabled(ctx, guildID, enabled)
}

// --- Channel & emoji management (passthrough to repo) ---

func (s *TranslationService) ListChannels(ctx context.Context, guildID int64) ([]model.TranslationChannel, error) {
	return s.repo.ListChannels(ctx, guildID)
}

func (s *TranslationService) AddChannel(ctx context.Context, guildID, channelID int64, override *string) (*model.TranslationChannel, error) {
	return s.repo.AddChannel(ctx, &model.TranslationChannel{
		GuildID:            guildID,
		ChannelID:          channelID,
		TargetLangOverride: normalizeStrPtr(override),
	})
}

func (s *TranslationService) RemoveChannel(ctx context.Context, guildID, channelID int64) error {
	return s.repo.RemoveChannel(ctx, guildID, channelID)
}

func (s *TranslationService) ListEmojis(ctx context.Context, guildID int64) ([]model.TranslationReactionEmoji, error) {
	return s.repo.ListEmojis(ctx, guildID)
}

func (s *TranslationService) AddEmoji(ctx context.Context, guildID int64, emoji string, override *string) (*model.TranslationReactionEmoji, error) {
	emoji = strings.TrimSpace(emoji)
	if emoji == "" {
		return nil, fmt.Errorf("translation: emoji is required")
	}
	return s.repo.AddEmoji(ctx, &model.TranslationReactionEmoji{
		GuildID:            guildID,
		Emoji:              emoji,
		TargetLangOverride: normalizeStrPtr(override),
	})
}

func (s *TranslationService) RemoveEmoji(ctx context.Context, guildID int64, emoji string) error {
	return s.repo.RemoveEmoji(ctx, guildID, emoji)
}

// --- Runtime translation ---

// TranslateMessage runs the auto-translate path for a newly created message.
func (s *TranslationService) TranslateMessage(ctx context.Context, dg *discordgo.Session, m *discordgo.MessageCreate) error {
	guildID, err := strconv.ParseInt(m.GuildID, 10, 64)
	if err != nil {
		return nil
	}

	cfg, err := s.repo.GetConfig(ctx, guildID)
	if err == repository.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if !cfg.Enabled || !cfg.AutoEnabled {
		return nil
	}

	channelID, err := strconv.ParseInt(m.ChannelID, 10, 64)
	if err != nil {
		return nil
	}

	channels, err := s.repo.ListChannels(ctx, guildID)
	if err != nil {
		return err
	}
	var override *string
	found := false
	for _, ch := range channels {
		if ch.ChannelID == channelID {
			override = ch.TargetLangOverride
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	target := s.resolveTarget(ctx, guildID, override, cfg)
	return s.translateAndReply(ctx, dg, cfg, m.ChannelID, m.ID, m.GuildID, m.Content, target, model.ReactionDeliveryChannel, "")
}

// TranslateReaction runs the reaction-trigger path for a reaction add event.
func (s *TranslationService) TranslateReaction(ctx context.Context, dg *discordgo.Session, r *discordgo.MessageReactionAdd) error {
	guildID, err := strconv.ParseInt(r.GuildID, 10, 64)
	if err != nil {
		return nil
	}

	cfg, err := s.repo.GetConfig(ctx, guildID)
	if err == repository.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if !cfg.Enabled || !cfg.ReactionEnabled {
		return nil
	}

	reactionKey := emojiKey(r.Emoji)
	emojis, err := s.repo.ListEmojis(ctx, guildID)
	if err != nil {
		return err
	}
	var override *string
	found := false
	for _, e := range emojis {
		if e.Emoji == reactionKey {
			override = e.TargetLangOverride
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	msg, err := dg.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		return fmt.Errorf("translation: fetch reacted message: %w", err)
	}

	target := s.resolveTarget(ctx, guildID, override, cfg)
	delivery := cfg.ReactionDelivery
	if delivery == "" {
		delivery = model.ReactionDeliveryChannel
	}
	return s.translateAndReply(ctx, dg, cfg, r.ChannelID, r.MessageID, r.GuildID, msg.Content, target, delivery, r.UserID)
}

// translateAndReply applies content guards, translates, and delivers the
// result either as a channel reply or (delivery == model.ReactionDeliveryDM)
// a DM to userID, falling back to the channel reply if the DM can't be sent
// (e.g. the user has DMs closed). It is a no-op (nil error) when the message
// should be skipped.
func (s *TranslationService) translateAndReply(ctx context.Context, dg *discordgo.Session, cfg *model.TranslationConfig, channelID, messageID, guildID, content, target, delivery, userID string) error {
	content = strings.TrimSpace(content)
	if !isTranslatable(content) {
		return nil
	}
	if len([]rune(content)) > maxTranslateChars {
		return nil
	}

	apiKey, err := s.decryptKey(cfg)
	if err != nil {
		return err
	}

	endpoint := ""
	if cfg.EndpointURL != nil {
		endpoint = *cfg.EndpointURL
	}

	translator, err := s.newTranslator(cfg.Provider, apiKey, endpoint)
	if err != nil {
		return err
	}

	res, err := translator.Translate(ctx, content, target)
	if err != nil {
		return err
	}

	// Skip when the message is already in the target language, or the engine
	// returned an identical string.
	if res.DetectedSourceLang != "" && res.DetectedSourceLang == normalizeLang(target) {
		return nil
	}
	if strings.TrimSpace(res.TranslatedText) == "" ||
		strings.EqualFold(strings.TrimSpace(res.TranslatedText), content) {
		return nil
	}

	footer := fmt.Sprintf("Auto-translated → %s", strings.ToUpper(target))
	if res.DetectedSourceLang != "" {
		footer = fmt.Sprintf("Auto-translated %s → %s", strings.ToUpper(res.DetectedSourceLang), strings.ToUpper(target))
	}

	embed := &discordgo.MessageEmbed{
		Description: res.TranslatedText,
		Color:       embedColor,
		Footer:      &discordgo.MessageEmbedFooter{Text: footer},
	}

	if delivery == model.ReactionDeliveryDM && userID != "" {
		if dmErr := s.sendDM(dg, userID, channelID, embed); dmErr == nil {
			return nil
		}
		// DMs closed or another failure — fall back to the channel reply below.
	}

	_, err = dg.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embed: embed,
		Reference: &discordgo.MessageReference{
			MessageID: messageID,
			ChannelID: channelID,
			GuildID:   guildID,
		},
	})
	if err != nil {
		return fmt.Errorf("translation: send translated reply: %w", err)
	}
	return nil
}

// sendDM delivers the translation embed privately to userID, tagging it with
// the source channel since that context is otherwise lost outside the guild.
func (s *TranslationService) sendDM(dg *discordgo.Session, userID, sourceChannelID string, embed *discordgo.MessageEmbed) error {
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:  "Source",
		Value: fmt.Sprintf("<#%s>", sourceChannelID),
	})

	dmChannel, err := dg.UserChannelCreate(userID)
	if err != nil {
		return fmt.Errorf("translation: open DM channel: %w", err)
	}
	if _, err := dg.ChannelMessageSendComplex(dmChannel.ID, &discordgo.MessageSend{Embed: embed}); err != nil {
		return fmt.Errorf("translation: send DM: %w", err)
	}
	return nil
}

func (s *TranslationService) decryptKey(cfg *model.TranslationConfig) (string, error) {
	if len(cfg.APIKeyEnc) == 0 {
		return "", nil
	}
	plain, err := crypto.Decrypt(cfg.APIKeyEnc, s.tokenEncKey)
	if err != nil {
		return "", fmt.Errorf("translation: decrypt api key: %w", err)
	}
	return string(plain), nil
}

// resolveTarget picks the target language: per-item override, else config
// default, else the guild language, else "en".
func (s *TranslationService) resolveTarget(ctx context.Context, guildID int64, override *string, cfg *model.TranslationConfig) string {
	if override != nil && *override != "" {
		return *override
	}
	if cfg.TargetLang != nil && *cfg.TargetLang != "" {
		return *cfg.TargetLang
	}
	if s.guildProvider != nil {
		if g, err := s.guildProvider.GetConfig(ctx, guildID); err == nil && g != nil && g.Language != "" {
			return g.Language
		}
	}
	return defaultTargetLang
}

// --- helpers ---

var (
	reCustomEmoji = regexp.MustCompile(`<a?:\w+:\d+>`)
	reMention     = regexp.MustCompile(`<[@#][!&]?\d+>`)
	reURL         = regexp.MustCompile(`https?://\S+`)
)

// isTranslatable strips mentions, custom emojis, and URLs and checks that at
// least a couple of letters remain — so pure links/emoji/mentions are skipped.
func isTranslatable(text string) bool {
	stripped := reCustomEmoji.ReplaceAllString(text, "")
	stripped = reMention.ReplaceAllString(stripped, "")
	stripped = reURL.ReplaceAllString(stripped, "")

	letters := 0
	for _, r := range stripped {
		if unicode.IsLetter(r) {
			letters++
			if letters >= minTranslateChars {
				return true
			}
		}
	}
	return false
}

// emojiKey returns the stored identifier for a reaction emoji: the unicode
// character for standard emojis, or "name:id" for custom guild emojis.
func emojiKey(e discordgo.Emoji) string {
	if e.ID != "" {
		return e.APIName()
	}
	return e.Name
}

// normalizeLang lowercases a code and strips any region suffix ("EN-GB" -> "en").
func normalizeLang(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if i := strings.IndexAny(code, "-_"); i > 0 {
		code = code[:i]
	}
	return code
}

// normalizeStrPtr trims a string pointer and returns nil for empty values.
func normalizeStrPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		return nil
	}
	return &v
}
