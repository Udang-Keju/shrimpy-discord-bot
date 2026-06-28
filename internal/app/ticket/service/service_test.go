package service_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	guild_model "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/service"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
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

// MockTranscriptRepository mocks TranscriptRepository
type MockTranscriptRepository struct {
	ListByTicketFunc         func(ctx context.Context, ticketID string) ([]model.TicketMessage, error)
	ListNonNotesByTicketFunc func(ctx context.Context, ticketID string) ([]model.TicketMessage, error)
}

func (m *MockTranscriptRepository) ListByTicket(ctx context.Context, ticketID string) ([]model.TicketMessage, error) {
	if m.ListByTicketFunc != nil {
		return m.ListByTicketFunc(ctx, ticketID)
	}
	return nil, nil
}

func (m *MockTranscriptRepository) ListNonNotesByTicket(ctx context.Context, ticketID string) ([]model.TicketMessage, error) {
	if m.ListNonNotesByTicketFunc != nil {
		return m.ListNonNotesByTicketFunc(ctx, ticketID)
	}
	return nil, nil
}

// MockTicketRepository mocks TicketRepository
type MockTicketRepository struct {
	CreateFunc          func(ctx context.Context, t *model.Ticket) (*model.Ticket, error)
	GetByIDFunc         func(ctx context.Context, ticketID string) (*model.Ticket, error)
	GetByChannelIDFunc  func(ctx context.Context, channelID int64) (*model.Ticket, error)
	ListFunc            func(ctx context.Context, guildID int64, f model.TicketFilter) ([]model.Ticket, int64, error)
	CountOpenByUserFunc func(ctx context.Context, guildID int64, categoryID string, userID int64) (int64, error)
	UpdateStatusFunc    func(ctx context.Context, ticketID string, status model.TicketStatus, reason *string) (*model.Ticket, error)
	UpdateClaimFunc     func(ctx context.Context, ticketID string, claimedBy *int64) (*model.Ticket, error)
	UpdatePriorityFunc  func(ctx context.Context, ticketID string, priority model.TicketPriority) (*model.Ticket, error)
	SetChannelFunc      func(ctx context.Context, ticketID string, channelID, threadID *int64) error
	ResetAutoCloseFunc  func(ctx context.Context, ticketID string, autoCloseAt *time.Time) error
	GetStatsFunc        func(ctx context.Context, guildID int64) (*model.TicketStats, error)
}

func (m *MockTicketRepository) Create(ctx context.Context, t *model.Ticket) (*model.Ticket, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, t)
	}
	return nil, nil
}

func (m *MockTicketRepository) GetByID(ctx context.Context, ticketID string) (*model.Ticket, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, ticketID)
	}
	return nil, nil
}

func (m *MockTicketRepository) GetByChannelID(ctx context.Context, channelID int64) (*model.Ticket, error) {
	if m.GetByChannelIDFunc != nil {
		return m.GetByChannelIDFunc(ctx, channelID)
	}
	return nil, nil
}

func (m *MockTicketRepository) List(ctx context.Context, guildID int64, f model.TicketFilter) ([]model.Ticket, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, guildID, f)
	}
	return nil, 0, nil
}

func (m *MockTicketRepository) CountOpenByUser(ctx context.Context, guildID int64, categoryID string, userID int64) (int64, error) {
	if m.CountOpenByUserFunc != nil {
		return m.CountOpenByUserFunc(ctx, guildID, categoryID, userID)
	}
	return 0, nil
}

func (m *MockTicketRepository) UpdateStatus(ctx context.Context, ticketID string, status model.TicketStatus, reason *string) (*model.Ticket, error) {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, ticketID, status, reason)
	}
	return nil, nil
}

func (m *MockTicketRepository) UpdateClaim(ctx context.Context, ticketID string, claimedBy *int64) (*model.Ticket, error) {
	if m.UpdateClaimFunc != nil {
		return m.UpdateClaimFunc(ctx, ticketID, claimedBy)
	}
	return nil, nil
}

func (m *MockTicketRepository) UpdatePriority(ctx context.Context, ticketID string, priority model.TicketPriority) (*model.Ticket, error) {
	if m.UpdatePriorityFunc != nil {
		return m.UpdatePriorityFunc(ctx, ticketID, priority)
	}
	return nil, nil
}

func (m *MockTicketRepository) SetChannel(ctx context.Context, ticketID string, channelID, threadID *int64) error {
	if m.SetChannelFunc != nil {
		return m.SetChannelFunc(ctx, ticketID, channelID, threadID)
	}
	return nil
}

func (m *MockTicketRepository) ResetAutoClose(ctx context.Context, ticketID string, autoCloseAt *time.Time) error {
	if m.ResetAutoCloseFunc != nil {
		return m.ResetAutoCloseFunc(ctx, ticketID, autoCloseAt)
	}
	return nil
}

func (m *MockTicketRepository) GetStats(ctx context.Context, guildID int64) (*model.TicketStats, error) {
	if m.GetStatsFunc != nil {
		return m.GetStatsFunc(ctx, guildID)
	}
	return nil, nil
}

// MockTicketCategoryRepository mocks TicketCategoryRepository
type MockTicketCategoryRepository struct {
	GetCategoryFunc              func(ctx context.Context, categoryID string) (*model.TicketCategory, error)
	ListPanelHandlerRolesFunc    func(ctx context.Context, panelID string) ([]model.PanelHandlerRole, error)
	ListCategoryHandlerRolesFunc func(ctx context.Context, categoryID string) ([]model.CategoryHandlerRole, error)
}

func (m *MockTicketCategoryRepository) GetCategory(ctx context.Context, categoryID string) (*model.TicketCategory, error) {
	if m.GetCategoryFunc != nil {
		return m.GetCategoryFunc(ctx, categoryID)
	}
	return nil, nil
}

func (m *MockTicketCategoryRepository) ListPanelHandlerRoles(ctx context.Context, panelID string) ([]model.PanelHandlerRole, error) {
	if m.ListPanelHandlerRolesFunc != nil {
		return m.ListPanelHandlerRolesFunc(ctx, panelID)
	}
	return nil, nil
}

func (m *MockTicketCategoryRepository) ListCategoryHandlerRoles(ctx context.Context, categoryID string) ([]model.CategoryHandlerRole, error) {
	if m.ListCategoryHandlerRolesFunc != nil {
		return m.ListCategoryHandlerRolesFunc(ctx, categoryID)
	}
	return nil, nil
}

// MockTicketGuildRepository mocks TicketGuildRepository
type MockTicketGuildRepository struct {
	GetByIDFunc        func(ctx context.Context, guildID int64) (*guild_model.Guild, error)
	ListStaffRolesFunc func(ctx context.Context, guildID int64) ([]guild_model.StaffRole, error)
}

func (m *MockTicketGuildRepository) GetByID(ctx context.Context, guildID int64) (*guild_model.Guild, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, guildID)
	}
	return nil, nil
}

func (m *MockTicketGuildRepository) ListStaffRoles(ctx context.Context, guildID int64) ([]guild_model.StaffRole, error) {
	if m.ListStaffRolesFunc != nil {
		return m.ListStaffRolesFunc(ctx, guildID)
	}
	return nil, nil
}

// MockTicketMessageRepository mocks TicketMessageRepository
type MockTicketMessageRepository struct {
	AddFunc func(ctx context.Context, m *model.TicketMessage) (*model.TicketMessage, error)
}

func (m *MockTicketMessageRepository) Add(ctx context.Context, msg *model.TicketMessage) (*model.TicketMessage, error) {
	if m.AddFunc != nil {
		return m.AddFunc(ctx, msg)
	}
	return nil, nil
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestTranscriptService_GenerateText(t *testing.T) {
	ctx := context.Background()
	ticketID := "test-ticket-id"
	content1 := "hello there"
	content2 := "staff notes"

	messages := []model.TicketMessage{
		{
			ID:             "msg-1",
			TicketID:       ticketID,
			AuthorID:       111,
			AuthorUsername: "user1",
			Content:        &content1,
			IsStaffNote:    false,
			SentAt:         time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC),
			Attachments:    datatypes.JSON(`[{"filename":"pic.png","url":"https://pic.png","size":100}]`),
		},
		{
			ID:             "msg-2",
			TicketID:       ticketID,
			AuthorID:       222,
			AuthorUsername: "staff1",
			Content:        &content2,
			IsStaffNote:    true,
			SentAt:         time.Date(2026, 6, 21, 10, 5, 0, 0, time.UTC),
		},
	}

	tests := []struct {
		name              string
		includeStaffNotes bool
		repoSetup         func(repo *MockTranscriptRepository)
		expectedStrings   []string
		unexpectedStrings []string
	}{
		{
			name:              "Include Staff Notes",
			includeStaffNotes: true,
			repoSetup: func(repo *MockTranscriptRepository) {
				repo.ListByTicketFunc = func(c context.Context, id string) ([]model.TicketMessage, error) {
					return messages, nil
				}
			},
			expectedStrings: []string{
				"=== Transcript for Ticket: test-ticket-id ===",
				"user1: hello there",
				"[STAFF NOTE] staff1: staff notes",
				"Attachment: pic.png (https://pic.png)",
			},
		},
		{
			name:              "Exclude Staff Notes",
			includeStaffNotes: false,
			repoSetup: func(repo *MockTranscriptRepository) {
				repo.ListNonNotesByTicketFunc = func(c context.Context, id string) ([]model.TicketMessage, error) {
					return []model.TicketMessage{messages[0]}, nil
				}
			},
			expectedStrings: []string{
				"user1: hello there",
				"Attachment: pic.png (https://pic.png)",
			},
			unexpectedStrings: []string{
				"staff notes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockTranscriptRepository{}
			tt.repoSetup(repo)
			svc := service.NewTranscriptService(repo)

			text, err := svc.GenerateText(ctx, ticketID, tt.includeStaffNotes)
			assert.NoError(t, err)
			for _, s := range tt.expectedStrings {
				assert.Contains(t, text, s)
			}
			for _, s := range tt.unexpectedStrings {
				assert.NotContains(t, text, s)
			}
		})
	}
}

func TestTranscriptService_GenerateHTML(t *testing.T) {
	ctx := context.Background()
	ticketID := "test-ticket-id"
	content1 := "hello there"
	content2 := "staff notes"

	messages := []model.TicketMessage{
		{
			ID:             "msg-1",
			TicketID:       ticketID,
			AuthorID:       111,
			AuthorUsername: "user1",
			Content:        &content1,
			IsStaffNote:    false,
			SentAt:         time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC),
			Attachments:    datatypes.JSON(`[{"filename":"pic.png","url":"https://pic.png","size":100}]`),
		},
		{
			ID:             "msg-2",
			TicketID:       ticketID,
			AuthorID:       222,
			AuthorUsername: "staff1",
			Content:        &content2,
			IsStaffNote:    true,
			SentAt:         time.Date(2026, 6, 21, 10, 5, 0, 0, time.UTC),
		},
	}

	repo := &MockTranscriptRepository{
		ListByTicketFunc: func(c context.Context, id string) ([]model.TicketMessage, error) {
			return messages, nil
		},
	}
	svc := service.NewTranscriptService(repo)

	ticket := &model.Ticket{
		ID:        ticketID,
		OpenedBy:  111,
		CreatedAt: time.Date(2026, 6, 21, 9, 0, 0, 0, time.UTC),
	}

	html, err := svc.GenerateHTML(ctx, ticket, "General Support", "user1", true)
	assert.NoError(t, err)
	assert.Contains(t, html, "Shrimpy Support Transcript")
	assert.Contains(t, html, "test-ticket-id")
	assert.Contains(t, html, "General Support")
	assert.Contains(t, html, "user1")
	assert.Contains(t, html, "hello there")
	assert.Contains(t, html, "staff notes")
	assert.Contains(t, html, "pic.png")
}

func TestTicketService_Open(t *testing.T) {
	ctx := context.Background()
	guildID := int64(12345)
	categoryID := "cat-uuid"
	userID := int64(67890)

	tests := []struct {
		name          string
		ticketSetup   func(repo *MockTicketRepository)
		categorySetup func(repo *MockTicketCategoryRepository)
		guildSetup    func(repo *MockTicketGuildRepository)
		mockTransport func(req *http.Request, createdCh *string, msgPayload *string) (*http.Response, error)
		expectError   error
	}{
		{
			name: "Limit Reached",
			ticketSetup: func(repo *MockTicketRepository) {
				repo.CountOpenByUserFunc = func(c context.Context, gID int64, catID string, uID int64) (int64, error) {
					return 3, nil
				}
			},
			categorySetup: func(repo *MockTicketCategoryRepository) {
				repo.GetCategoryFunc = func(c context.Context, id string) (*model.TicketCategory, error) {
					return &model.TicketCategory{ID: categoryID, MaxTicketsPerUser: 2}, nil
				}
			},
			expectError: model.ErrLimitReached,
		},
		{
			name: "Success",
			ticketSetup: func(repo *MockTicketRepository) {
				repo.CountOpenByUserFunc = func(c context.Context, gID int64, catID string, uID int64) (int64, error) {
					return 0, nil
				}
				repo.CreateFunc = func(c context.Context, t *model.Ticket) (*model.Ticket, error) {
					t.ID = "ticket-uuid-1234"
					return t, nil
				}
				repo.SetChannelFunc = func(c context.Context, tID string, channelID, threadID *int64) error {
					assert.Equal(t, "ticket-uuid-1234", tID)
					assert.Equal(t, int64(999888), *channelID)
					return nil
				}
			},
			categorySetup: func(repo *MockTicketCategoryRepository) {
				repo.GetCategoryFunc = func(c context.Context, id string) (*model.TicketCategory, error) {
					return &model.TicketCategory{
						ID:                 categoryID,
						MaxTicketsPerUser:  5,
						TicketNameTemplate: "{category}-{number}",
						Name:               "billing",
						TicketOpenTitle:    stringPtr("Welcome to billing"),
						TicketOpenMessage:  stringPtr("Hello {mention}, ticket ID is {id}"),
					}, nil
				}
			},
			guildSetup: func(repo *MockTicketGuildRepository) {
				repo.ListStaffRolesFunc = func(c context.Context, gID int64) ([]guild_model.StaffRole, error) {
					return []guild_model.StaffRole{{ID: "sr1", GuildID: guildID, RoleID: 888}}, nil
				}
			},
			mockTransport: func(req *http.Request, createdCh *string, msgPayload *string) (*http.Response, error) {
				if req.Method == "GET" && req.URL.Path == "/api/v9/users/@me" {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"id":"777","username":"shrimpy"}`)),
					}, nil
				}
				if req.Method == "POST" && req.URL.Path == "/api/v9/guilds/12345/channels" {
					var data struct {
						Name string `json:"name"`
					}
					_ = json.NewDecoder(req.Body).Decode(&data)
					*createdCh = data.Name
					return &http.Response{
						StatusCode: 201,
						Body:       io.NopCloser(strings.NewReader(`{"id":"999888","name":"billing-ticket","type":0}`)),
					}, nil
				}
				if req.Method == "POST" && req.URL.Path == "/api/v9/channels/999888/messages" {
					b, _ := io.ReadAll(req.Body)
					*msgPayload = string(b)
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"id":"111222"}`)),
					}, nil
				}
				return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader("{}"))}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticketRepo := &MockTicketRepository{}
			if tt.ticketSetup != nil {
				tt.ticketSetup(ticketRepo)
			}
			categoryRepo := &MockTicketCategoryRepository{}
			if tt.categorySetup != nil {
				tt.categorySetup(categoryRepo)
			}
			guildRepo := &MockTicketGuildRepository{}
			if tt.guildSetup != nil {
				tt.guildSetup(guildRepo)
			}

			svc := service.NewTicketService(ticketRepo, categoryRepo, guildRepo, nil, nil)
			dg, _ := discordgo.New("Bot Token")

			var createdCh, msgPayload string
			if tt.mockTransport != nil {
				dg.Client.Transport = &mockTransport{
					roundTrip: func(req *http.Request) (*http.Response, error) {
						return tt.mockTransport(req, &createdCh, &msgPayload)
					},
				}
			}

			ticket, err := svc.Open(ctx, dg, guildID, categoryID, userID)

			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
				assert.Nil(t, ticket)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ticket)
				assert.Equal(t, "ticket-uuid-1234", ticket.ID)
				assert.Equal(t, "billing-ticket-u", createdCh)
				assert.Contains(t, msgPayload, "Welcome to billing")
				assert.Contains(t, msgPayload, "Hello \\u003c@67890\\u003e")
			}
		})
	}
}

func TestTicketService_Claim_Unclaim(t *testing.T) {
	ctx := context.Background()
	ticketID := "ticket-uuid"
	guildID := int64(12345)
	channelID := int64(999888)
	staffUserID := int64(55555)

	t.Run("Claim", func(t *testing.T) {
		ticketRepo := &MockTicketRepository{
			GetByIDFunc: func(c context.Context, id string) (*model.Ticket, error) {
				return &model.Ticket{
					ID:        ticketID,
					GuildID:   guildID,
					ChannelID: &channelID,
					Status:    model.TicketStatusOpen,
				}, nil
			},
			UpdateClaimFunc: func(c context.Context, id string, claimedBy *int64) (*model.Ticket, error) {
				assert.Equal(t, ticketID, id)
				assert.Equal(t, staffUserID, *claimedBy)
				return &model.Ticket{
					ID:        ticketID,
					GuildID:   guildID,
					ChannelID: &channelID,
					Status:    model.TicketStatusClaimed,
					ClaimedBy: claimedBy,
				}, nil
			},
		}

		svc := service.NewTicketService(ticketRepo, nil, nil, nil, nil)
		dg, _ := discordgo.New("Bot Token")

		channelFetched := false
		channelEdited := false
		messageSent := false

		dg.Client.Transport = &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" && req.URL.Path == "/api/v9/channels/999888" {
					channelFetched = true
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"id":"999888","name":"billing-ticket"}`)),
					}, nil
				}
				if req.Method == "PATCH" && req.URL.Path == "/api/v9/channels/999888" {
					channelEdited = true
					b, _ := io.ReadAll(req.Body)
					assert.Contains(t, string(b), "claimed-billing-ticket")
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"id":"999888","name":"claimed-billing-ticket"}`)),
					}, nil
				}
				if req.Method == "POST" && req.URL.Path == "/api/v9/channels/999888/messages" {
					messageSent = true
					b, _ := io.ReadAll(req.Body)
					assert.Contains(t, string(b), "claimed by \\u003c@55555\\u003e")
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"id":"123"}`)),
					}, nil
				}
				return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader("{}"))}, nil
			},
		}

		ticket, err := svc.Claim(ctx, dg, ticketID, staffUserID)
		assert.NoError(t, err)
		assert.NotNil(t, ticket)
		assert.True(t, channelFetched)
		assert.True(t, channelEdited)
		assert.True(t, messageSent)
	})

	t.Run("Unclaim", func(t *testing.T) {
		ticketRepo := &MockTicketRepository{
			GetByIDFunc: func(c context.Context, id string) (*model.Ticket, error) {
				return &model.Ticket{
					ID:        ticketID,
					GuildID:   guildID,
					ChannelID: &channelID,
					Status:    model.TicketStatusClaimed,
					ClaimedBy: &staffUserID,
				}, nil
			},
			UpdateClaimFunc: func(c context.Context, id string, claimedBy *int64) (*model.Ticket, error) {
				assert.Equal(t, ticketID, id)
				assert.Nil(t, claimedBy)
				return &model.Ticket{
					ID:        ticketID,
					GuildID:   guildID,
					ChannelID: &channelID,
					Status:    model.TicketStatusOpen,
					ClaimedBy: nil,
				}, nil
			},
		}

		svc := service.NewTicketService(ticketRepo, nil, nil, nil, nil)
		dg, _ := discordgo.New("Bot Token")

		channelFetched := false
		channelEdited := false
		messageSent := false

		dg.Client.Transport = &mockTransport{
			roundTrip: func(req *http.Request) (*http.Response, error) {
				if req.Method == "GET" && req.URL.Path == "/api/v9/channels/999888" {
					channelFetched = true
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"id":"999888","name":"claimed-billing-ticket"}`)),
					}, nil
				}
				if req.Method == "PATCH" && req.URL.Path == "/api/v9/channels/999888" {
					channelEdited = true
					b, _ := io.ReadAll(req.Body)
					assert.NotContains(t, string(b), "claimed-")
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"id":"999888","name":"billing-ticket"}`)),
					}, nil
				}
				if req.Method == "POST" && req.URL.Path == "/api/v9/channels/999888/messages" {
					messageSent = true
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader(`{"id":"123"}`)),
					}, nil
				}
				return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader("{}"))}, nil
			},
		}

		ticket, err := svc.Unclaim(ctx, dg, ticketID)
		assert.NoError(t, err)
		assert.NotNil(t, ticket)
		assert.True(t, channelFetched)
		assert.True(t, channelEdited)
		assert.True(t, messageSent)
	})
}

func TestTicketService_Close(t *testing.T) {
	ctx := context.Background()
	ticketID := "ticket-uuid"
	guildID := int64(12345)
	channelID := int64(999888)
	closedByUserID := int64(55555)
	reason := "issue resolved"

	ticketRepo := &MockTicketRepository{
		GetByIDFunc: func(c context.Context, id string) (*model.Ticket, error) {
			return &model.Ticket{
				ID:         ticketID,
				GuildID:    guildID,
				ChannelID:  &channelID,
				Status:     model.TicketStatusOpen,
				OpenedBy:   67890,
				CategoryID: "cat-uuid",
			}, nil
		},
		UpdateStatusFunc: func(c context.Context, id string, status model.TicketStatus, r *string) (*model.Ticket, error) {
			assert.Equal(t, ticketID, id)
			assert.Equal(t, model.TicketStatusClosed, status)
			assert.Equal(t, reason, *r)
			return &model.Ticket{
				ID:          ticketID,
				GuildID:     guildID,
				ChannelID:   &channelID,
				Status:      model.TicketStatusClosed,
				OpenedBy:    67890,
				CategoryID:  "cat-uuid",
				CloseReason: r,
			}, nil
		},
	}

	categoryRepo := &MockTicketCategoryRepository{
		GetCategoryFunc: func(c context.Context, id string) (*model.TicketCategory, error) {
			return &model.TicketCategory{
				ID:                  "cat-uuid",
				Name:                "billing",
				TranscriptChannelID: int64Ptr(44444),
			}, nil
		},
	}

	transcriptRepo := &MockTranscriptRepository{
		ListByTicketFunc: func(c context.Context, id string) ([]model.TicketMessage, error) {
			return []model.TicketMessage{
				{
					ID:             "msg-1",
					TicketID:       ticketID,
					AuthorUsername: "user1",
					SentAt:         time.Now(),
				},
			}, nil
		},
	}

	guildRepo := &MockTicketGuildRepository{}

	svc := service.NewTicketService(
		ticketRepo,
		categoryRepo,
		guildRepo,
		nil,
		service.NewTranscriptService(transcriptRepo),
	)

	dg, _ := discordgo.New("Bot Token")

	permissionUpdated := false
	transcriptSent := false
	closeEmbedSent := false

	dg.Client.Transport = &mockTransport{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			if req.Method == "PUT" && req.URL.Path == "/api/v9/channels/999888/permissions/67890" {
				permissionUpdated = true
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("{}"))}, nil
			}
			if req.Method == "GET" && req.URL.Path == "/api/v9/users/67890" {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"id":"67890","username":"user1"}`)),
				}, nil
			}
			if req.Method == "POST" && req.URL.Path == "/api/v9/channels/44444/messages" {
				transcriptSent = true
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":"999"}`))}, nil
			}
			if req.Method == "POST" && req.URL.Path == "/api/v9/channels/999888/messages" {
				closeEmbedSent = true
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":"999"}`))}, nil
			}
			return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader("{}"))}, nil
		},
	}

	ticket, err := svc.Close(ctx, dg, ticketID, &reason, closedByUserID)
	assert.NoError(t, err)
	assert.NotNil(t, ticket)
	assert.True(t, permissionUpdated)
	assert.True(t, transcriptSent)
	assert.True(t, closeEmbedSent)
}
