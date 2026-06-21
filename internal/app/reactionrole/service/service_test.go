package service_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/service"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

// mockTransport intercepts HTTP requests made by discordgo.Session
type mockTransport struct {
	roundTrip func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.roundTrip != nil {
		return m.roundTrip(req)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("{}")),
	}, nil
}

// MockReactionRoleRepository mocks ReactionRoleRepository
type MockReactionRoleRepository struct {
	ListByGuildFunc           func(ctx context.Context, guildID int64) ([]model.ReactionRoleMessage, error)
	GetMessageFunc            func(ctx context.Context, messageID string) (*model.ReactionRoleMessage, error)
	GetByDiscordMessageIDFunc func(ctx context.Context, discordMsgID int64) (*model.ReactionRoleMessage, error)
	CreateMessageFunc         func(ctx context.Context, msg *model.ReactionRoleMessage) (*model.ReactionRoleMessage, error)
	UpdateMessageFunc         func(ctx context.Context, msg *model.ReactionRoleMessage) (*model.ReactionRoleMessage, error)
	SetDiscordMessageIDFunc   func(ctx context.Context, id string, discordMsgID int64) error
	DeleteMessageFunc         func(ctx context.Context, messageID string) error

	AddEmojiFunc     func(ctx context.Context, e *model.ReactionRoleEmoji) (*model.ReactionRoleEmoji, error)
	RemoveEmojiFunc  func(ctx context.Context, messageID, emoji string) error
	GetEmojiRoleFunc func(ctx context.Context, discordMsgID int64, emoji string) (*model.ReactionRoleEmoji, error)
}

func (m *MockReactionRoleRepository) ListByGuild(ctx context.Context, guildID int64) ([]model.ReactionRoleMessage, error) {
	if m.ListByGuildFunc != nil {
		return m.ListByGuildFunc(ctx, guildID)
	}
	return nil, nil
}

func (m *MockReactionRoleRepository) GetMessage(ctx context.Context, messageID string) (*model.ReactionRoleMessage, error) {
	if m.GetMessageFunc != nil {
		return m.GetMessageFunc(ctx, messageID)
	}
	return nil, nil
}

func (m *MockReactionRoleRepository) GetByDiscordMessageID(ctx context.Context, discordMsgID int64) (*model.ReactionRoleMessage, error) {
	if m.GetByDiscordMessageIDFunc != nil {
		return m.GetByDiscordMessageIDFunc(ctx, discordMsgID)
	}
	return nil, nil
}

func (m *MockReactionRoleRepository) CreateMessage(ctx context.Context, msg *model.ReactionRoleMessage) (*model.ReactionRoleMessage, error) {
	if m.CreateMessageFunc != nil {
		return m.CreateMessageFunc(ctx, msg)
	}
	return nil, nil
}

func (m *MockReactionRoleRepository) UpdateMessage(ctx context.Context, msg *model.ReactionRoleMessage) (*model.ReactionRoleMessage, error) {
	if m.UpdateMessageFunc != nil {
		return m.UpdateMessageFunc(ctx, msg)
	}
	return nil, nil
}

func (m *MockReactionRoleRepository) SetDiscordMessageID(ctx context.Context, id string, discordMsgID int64) error {
	if m.SetDiscordMessageIDFunc != nil {
		return m.SetDiscordMessageIDFunc(ctx, id, discordMsgID)
	}
	return nil
}

func (m *MockReactionRoleRepository) DeleteMessage(ctx context.Context, messageID string) error {
	if m.DeleteMessageFunc != nil {
		return m.DeleteMessageFunc(ctx, messageID)
	}
	return nil
}

func (m *MockReactionRoleRepository) AddEmoji(ctx context.Context, e *model.ReactionRoleEmoji) (*model.ReactionRoleEmoji, error) {
	if m.AddEmojiFunc != nil {
		return m.AddEmojiFunc(ctx, e)
	}
	return nil, nil
}

func (m *MockReactionRoleRepository) RemoveEmoji(ctx context.Context, messageID, emoji string) error {
	if m.RemoveEmojiFunc != nil {
		return m.RemoveEmojiFunc(ctx, messageID, emoji)
	}
	return nil
}

func (m *MockReactionRoleRepository) GetEmojiRole(ctx context.Context, discordMsgID int64, emoji string) (*model.ReactionRoleEmoji, error) {
	if m.GetEmojiRoleFunc != nil {
		return m.GetEmojiRoleFunc(ctx, discordMsgID, emoji)
	}
	return nil, nil
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestReactionRoleService_ListAndGet(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	msgID := "msg-uuid-1"

	t.Run("List", func(t *testing.T) {
		repo := &MockReactionRoleRepository{
			ListByGuildFunc: func(c context.Context, gID int64) ([]model.ReactionRoleMessage, error) {
				assert.Equal(t, guildID, gID)
				return []model.ReactionRoleMessage{{ID: msgID, GuildID: gID}}, nil
			},
		}
		svc := service.NewReactionRoleService(repo)
		msgs, err := svc.List(ctx, guildID)
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)
		assert.Equal(t, msgID, msgs[0].ID)
	})

	t.Run("Get", func(t *testing.T) {
		repo := &MockReactionRoleRepository{
			GetMessageFunc: func(c context.Context, mID string) (*model.ReactionRoleMessage, error) {
				assert.Equal(t, msgID, mID)
				return &model.ReactionRoleMessage{ID: msgID, GuildID: guildID}, nil
			},
		}
		svc := service.NewReactionRoleService(repo)
		msg, err := svc.Get(ctx, msgID)
		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, msgID, msg.ID)
	})
}

func TestReactionRoleService_Create(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	channelID := int64(999888)
	title := "Get Roles"
	desc := "Click reaction to get roles"

	repo := &MockReactionRoleRepository{
		CreateMessageFunc: func(c context.Context, msg *model.ReactionRoleMessage) (*model.ReactionRoleMessage, error) {
			msg.ID = "msg-uuid"
			return msg, nil
		},
		SetDiscordMessageIDFunc: func(c context.Context, id string, discordMsgID int64) error {
			assert.Equal(t, "msg-uuid", id)
			assert.Equal(t, int64(777), discordMsgID)
			return nil
		},
	}
	svc := service.NewReactionRoleService(repo)

	dg, _ := discordgo.New("Bot Token")
	discordCalled := false
	dg.Client.Transport = &mockTransport{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			discordCalled = true
			assert.Equal(t, "POST", req.Method)
			assert.Equal(t, "/api/v9/channels/999888/messages", req.URL.Path)
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"id":"777"}`)),
			}, nil
		},
	}

	created, err := svc.Create(ctx, dg, guildID, channelID, title, desc, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.True(t, discordCalled)
	assert.Equal(t, int64(777), *created.MessageID)
}

func TestReactionRoleService_AddRemoveDelete(t *testing.T) {
	ctx := context.Background()
	messageID := "msg-uuid"
	discordMsgID := int64(777)
	channelID := int64(999888)
	emoji := "🦀"
	roleID := int64(555)

	t.Run("AddEmoji", func(t *testing.T) {
		repo := &MockReactionRoleRepository{
			GetMessageFunc: func(c context.Context, mID string) (*model.ReactionRoleMessage, error) {
				return &model.ReactionRoleMessage{
					ID:        messageID,
					ChannelID: channelID,
					MessageID: &discordMsgID,
				}, nil
			},
			AddEmojiFunc: func(c context.Context, e *model.ReactionRoleEmoji) (*model.ReactionRoleEmoji, error) {
				assert.Equal(t, messageID, e.MessageID)
				assert.Equal(t, emoji, e.Emoji)
				assert.Equal(t, roleID, e.RoleID)
				return e, nil
			},
		}
		svc := service.NewReactionRoleService(repo)

		dg, _ := discordgo.New("Bot Token")
		discordCalled := false
		dg.Client.Transport = &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				discordCalled = true
				assert.Equal(t, "PUT", req.Method)
				// Path should target adding reaction to message
				assert.Contains(t, req.URL.Path, "/channels/999888/messages/777/reactions/🦀/@me")
				return &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader(""))}, nil
			},
		}

		created, err := svc.AddEmoji(ctx, dg, messageID, emoji, roleID)
		assert.NoError(t, err)
		assert.NotNil(t, created)
		assert.True(t, discordCalled)
	})

	t.Run("RemoveEmoji", func(t *testing.T) {
		repoCalled := false
		repo := &MockReactionRoleRepository{
			GetMessageFunc: func(c context.Context, mID string) (*model.ReactionRoleMessage, error) {
				return &model.ReactionRoleMessage{
					ID:        messageID,
					ChannelID: channelID,
					MessageID: &discordMsgID,
				}, nil
			},
			RemoveEmojiFunc: func(c context.Context, mID, em string) error {
				assert.Equal(t, messageID, mID)
				assert.Equal(t, emoji, em)
				repoCalled = true
				return nil
			},
		}
		svc := service.NewReactionRoleService(repo)

		dg, _ := discordgo.New("Bot Token")
		discordCalled := false
		dg.Client.Transport = &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				discordCalled = true
				assert.Equal(t, "DELETE", req.Method)
				assert.Contains(t, req.URL.Path, "/channels/999888/messages/777/reactions/🦀/@me")
				return &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader(""))}, nil
			},
		}

		err := svc.RemoveEmoji(ctx, dg, messageID, emoji)
		assert.NoError(t, err)
		assert.True(t, repoCalled)
		assert.True(t, discordCalled)
	})

	t.Run("Delete", func(t *testing.T) {
		repoCalled := false
		repo := &MockReactionRoleRepository{
			GetMessageFunc: func(c context.Context, mID string) (*model.ReactionRoleMessage, error) {
				return &model.ReactionRoleMessage{
					ID:        messageID,
					ChannelID: channelID,
					MessageID: &discordMsgID,
				}, nil
			},
			DeleteMessageFunc: func(c context.Context, mID string) error {
				assert.Equal(t, messageID, mID)
				repoCalled = true
				return nil
			},
		}
		svc := service.NewReactionRoleService(repo)

		dg, _ := discordgo.New("Bot Token")
		discordCalled := false
		dg.Client.Transport = &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				discordCalled = true
				assert.Equal(t, "DELETE", req.Method)
				assert.Contains(t, req.URL.Path, "/channels/999888/messages/777")
				return &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader(""))}, nil
			},
		}

		err := svc.Delete(ctx, dg, messageID)
		assert.NoError(t, err)
		assert.True(t, repoCalled)
		assert.True(t, discordCalled)
	})
}

func TestReactionRoleService_HandleReactions(t *testing.T) {
	ctx := context.Background()
	discordMsgID := int64(777)
	guildID := int64(12345)
	userID := int64(67890)
	emoji := "🦀"
	roleID := int64(555)

	t.Run("HandleReactionAdd", func(t *testing.T) {
		repo := &MockReactionRoleRepository{
			GetEmojiRoleFunc: func(c context.Context, dID int64, em string) (*model.ReactionRoleEmoji, error) {
				assert.Equal(t, discordMsgID, dID)
				assert.Equal(t, emoji, em)
				return &model.ReactionRoleEmoji{RoleID: roleID}, nil
			},
		}
		svc := service.NewReactionRoleService(repo)

		dg, _ := discordgo.New("Bot Token")
		discordCalled := false
		dg.Client.Transport = &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				discordCalled = true
				assert.Equal(t, "PUT", req.Method)
				assert.Contains(t, req.URL.Path, "/guilds/12345/members/67890/roles/555")
				return &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader(""))}, nil
			},
		}

		err := svc.HandleReactionAdd(ctx, dg, discordMsgID, guildID, userID, emoji)
		assert.NoError(t, err)
		assert.True(t, discordCalled)
	})

	t.Run("HandleReactionRemove", func(t *testing.T) {
		repo := &MockReactionRoleRepository{
			GetEmojiRoleFunc: func(c context.Context, dID int64, em string) (*model.ReactionRoleEmoji, error) {
				assert.Equal(t, discordMsgID, dID)
				assert.Equal(t, emoji, em)
				return &model.ReactionRoleEmoji{RoleID: roleID}, nil
			},
		}
		svc := service.NewReactionRoleService(repo)

		dg, _ := discordgo.New("Bot Token")
		discordCalled := false
		dg.Client.Transport = &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				discordCalled = true
				assert.Equal(t, "DELETE", req.Method)
				assert.Contains(t, req.URL.Path, "/guilds/12345/members/67890/roles/555")
				return &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader(""))}, nil
			},
		}

		err := svc.HandleReactionRemove(ctx, dg, discordMsgID, guildID, userID, emoji)
		assert.NoError(t, err)
		assert.True(t, discordCalled)
	})
}
