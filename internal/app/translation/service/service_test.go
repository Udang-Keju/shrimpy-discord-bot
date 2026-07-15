package service

import (
	"context"
	"testing"

	guildmodel "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/crypto"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/translate"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 32-byte key for AES-256-GCM.
var testKey = []byte("0123456789abcdef0123456789abcdef")

// --- fakes ---

type fakeRepo struct {
	cfg      *model.TranslationConfig
	channels []model.TranslationChannel
	emojis   []model.TranslationReactionEmoji
	upserted *model.TranslationConfig
}

func (f *fakeRepo) GetConfig(ctx context.Context, guildID int64) (*model.TranslationConfig, error) {
	if f.cfg == nil {
		return nil, repository.ErrNotFound
	}
	return f.cfg, nil
}
func (f *fakeRepo) UpsertConfig(ctx context.Context, cfg *model.TranslationConfig) (*model.TranslationConfig, error) {
	f.upserted = cfg
	f.cfg = cfg
	return cfg, nil
}
func (f *fakeRepo) SetEnabled(ctx context.Context, guildID int64, enabled bool) error { return nil }
func (f *fakeRepo) ListChannels(ctx context.Context, guildID int64) ([]model.TranslationChannel, error) {
	return f.channels, nil
}
func (f *fakeRepo) AddChannel(ctx context.Context, ch *model.TranslationChannel) (*model.TranslationChannel, error) {
	f.channels = append(f.channels, *ch)
	return ch, nil
}
func (f *fakeRepo) RemoveChannel(ctx context.Context, guildID, channelID int64) error { return nil }
func (f *fakeRepo) ListEmojis(ctx context.Context, guildID int64) ([]model.TranslationReactionEmoji, error) {
	return f.emojis, nil
}
func (f *fakeRepo) AddEmoji(ctx context.Context, e *model.TranslationReactionEmoji) (*model.TranslationReactionEmoji, error) {
	f.emojis = append(f.emojis, *e)
	return e, nil
}
func (f *fakeRepo) RemoveEmoji(ctx context.Context, guildID int64, emoji string) error { return nil }

type fakeGuildProvider struct{ lang string }

func (f *fakeGuildProvider) GetConfig(ctx context.Context, guildID int64) (*guildmodel.Guild, error) {
	return &guildmodel.Guild{GuildID: guildID, Language: f.lang}, nil
}

// fakeTranslator records calls and returns a canned result.
type fakeTranslator struct {
	called bool
	result translate.Result
	err    error
}

func (f *fakeTranslator) Translate(ctx context.Context, text, targetLang string) (translate.Result, error) {
	f.called = true
	return f.result, f.err
}

// --- tests ---

func TestGetConfig_GracefulDefault(t *testing.T) {
	svc := NewTranslationService(&fakeRepo{}, &fakeGuildProvider{lang: "en"}, testKey)
	cfg, err := svc.GetConfig(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, int64(42), cfg.GuildID)
	assert.False(t, cfg.Enabled)
	assert.Equal(t, translate.ProviderDeepL, cfg.Provider)
}

func TestSaveConfig_EncryptsNewKeyAndPreservesOnMask(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewTranslationService(repo, &fakeGuildProvider{lang: "en"}, testKey)

	// 1. New key is encrypted at rest and decryptable.
	_, err := svc.SaveConfig(context.Background(), 1, SaveConfigInput{
		Enabled:  true,
		Provider: "deepl",
		APIKey:   "secret-key",
	})
	require.NoError(t, err)
	require.NotEmpty(t, repo.upserted.APIKeyEnc)
	plain, err := crypto.Decrypt(repo.upserted.APIKeyEnc, testKey)
	require.NoError(t, err)
	assert.Equal(t, "secret-key", string(plain))

	// 2. Saving again with the masked placeholder preserves the stored key.
	_, err = svc.SaveConfig(context.Background(), 1, SaveConfigInput{
		Enabled:  true,
		Provider: "deepl",
		APIKey:   MaskedAPIKey,
	})
	require.NoError(t, err)
	plain2, err := crypto.Decrypt(repo.upserted.APIKeyEnc, testKey)
	require.NoError(t, err)
	assert.Equal(t, "secret-key", string(plain2))

	// 3. Empty key also preserves.
	_, err = svc.SaveConfig(context.Background(), 1, SaveConfigInput{Enabled: false, Provider: "deepl", APIKey: ""})
	require.NoError(t, err)
	plain3, err := crypto.Decrypt(repo.upserted.APIKeyEnc, testKey)
	require.NoError(t, err)
	assert.Equal(t, "secret-key", string(plain3))
}

func TestResolveTarget_Precedence(t *testing.T) {
	svc := NewTranslationService(&fakeRepo{}, &fakeGuildProvider{lang: "de"}, testKey)

	override := "fr"
	cfgTarget := "es"
	cfg := &model.TranslationConfig{TargetLang: &cfgTarget}

	// Override wins.
	assert.Equal(t, "fr", svc.resolveTarget(context.Background(), 1, &override, cfg))
	// Then config default.
	assert.Equal(t, "es", svc.resolveTarget(context.Background(), 1, nil, cfg))
	// Then guild language.
	assert.Equal(t, "de", svc.resolveTarget(context.Background(), 1, nil, &model.TranslationConfig{}))
	// Then hard default when no guild language.
	svcNoLang := NewTranslationService(&fakeRepo{}, &fakeGuildProvider{lang: ""}, testKey)
	assert.Equal(t, defaultTargetLang, svcNoLang.resolveTarget(context.Background(), 1, nil, &model.TranslationConfig{}))
}

func TestIsTranslatable(t *testing.T) {
	assert.True(t, isTranslatable("hola amigo"))
	assert.False(t, isTranslatable(""))
	assert.False(t, isTranslatable("a"))
	assert.False(t, isTranslatable("https://example.com/foo"))
	assert.False(t, isTranslatable("<@1234567890>"))
	assert.False(t, isTranslatable("<:custom:1234567890>"))
	assert.False(t, isTranslatable("😀🎉"))
	assert.True(t, isTranslatable("hi <@1234567890>"))
}

func TestEmojiKey(t *testing.T) {
	assert.Equal(t, "🇫🇷", emojiKey(discordgo.Emoji{Name: "🇫🇷"}))
	assert.Equal(t, "flag:123", emojiKey(discordgo.Emoji{Name: "flag", ID: "123"}))
}

// TranslateMessage should not translate when the feature is disabled.
func TestTranslateMessage_SkipsWhenDisabled(t *testing.T) {
	repo := &fakeRepo{cfg: &model.TranslationConfig{GuildID: 1, Enabled: false, AutoEnabled: true}}
	svc := NewTranslationService(repo, &fakeGuildProvider{lang: "en"}, testKey)
	ft := &fakeTranslator{}
	svc.newTranslator = func(_, _, _ string) (translate.Translator, error) { return ft, nil }

	err := svc.TranslateMessage(context.Background(), nil, &discordgo.MessageCreate{
		Message: &discordgo.Message{GuildID: "1", ChannelID: "10", Content: "hola amigo"},
	})
	require.NoError(t, err)
	assert.False(t, ft.called, "translator should not be called when feature disabled")
}

// TranslateMessage should skip channels that are not configured.
func TestTranslateMessage_SkipsUnconfiguredChannel(t *testing.T) {
	repo := &fakeRepo{
		cfg:      &model.TranslationConfig{GuildID: 1, Enabled: true, AutoEnabled: true},
		channels: []model.TranslationChannel{{GuildID: 1, ChannelID: 99}},
	}
	svc := NewTranslationService(repo, &fakeGuildProvider{lang: "en"}, testKey)
	ft := &fakeTranslator{}
	svc.newTranslator = func(_, _, _ string) (translate.Translator, error) { return ft, nil }

	err := svc.TranslateMessage(context.Background(), nil, &discordgo.MessageCreate{
		Message: &discordgo.Message{GuildID: "1", ChannelID: "10", Content: "hola amigo"},
	})
	require.NoError(t, err)
	assert.False(t, ft.called)
}

// TranslateMessage should not post when the detected source equals the target.
func TestTranslateMessage_SkipsSameLanguage(t *testing.T) {
	target := "es"
	repo := &fakeRepo{
		cfg:      &model.TranslationConfig{GuildID: 1, Enabled: true, AutoEnabled: true, Provider: "deepl", APIKeyEnc: mustEncrypt("k"), TargetLang: &target},
		channels: []model.TranslationChannel{{GuildID: 1, ChannelID: 10}},
	}
	svc := NewTranslationService(repo, &fakeGuildProvider{lang: "en"}, testKey)
	ft := &fakeTranslator{result: translate.Result{TranslatedText: "hola amigo", DetectedSourceLang: "es"}}
	svc.newTranslator = func(_, _, _ string) (translate.Translator, error) { return ft, nil }

	// dg is nil: the same-language guard returns before the send path uses it.
	err := svc.TranslateMessage(context.Background(), nil, &discordgo.MessageCreate{
		Message: &discordgo.Message{GuildID: "1", ChannelID: "10", Content: "hola amigo"},
	})
	require.NoError(t, err)
	assert.True(t, ft.called, "translator is called")
}

func mustEncrypt(s string) []byte {
	enc, err := crypto.Encrypt([]byte(s), testKey)
	if err != nil {
		panic(err)
	}
	return enc
}
