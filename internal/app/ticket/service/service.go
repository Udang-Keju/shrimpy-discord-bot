package service

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	guild_model "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
)

// ─── Interfaces ──────────────────────────────────────────────────────────────

// TicketRepository defines the database operations consumed by TicketService.
type TicketRepository interface {
	Create(ctx context.Context, t *model.Ticket) (*model.Ticket, error)
	GetByID(ctx context.Context, ticketID string) (*model.Ticket, error)
	GetByChannelID(ctx context.Context, channelID int64) (*model.Ticket, error)
	List(ctx context.Context, guildID int64, f model.TicketFilter) ([]model.Ticket, int64, error)
	CountOpenByUser(ctx context.Context, guildID int64, categoryID string, userID int64) (int64, error)
	UpdateStatus(ctx context.Context, ticketID string, status model.TicketStatus, reason *string) (*model.Ticket, error)
	UpdateClaim(ctx context.Context, ticketID string, claimedBy *int64) (*model.Ticket, error)
	UpdatePriority(ctx context.Context, ticketID string, priority model.TicketPriority) (*model.Ticket, error)
	SetChannel(ctx context.Context, ticketID string, channelID, threadID *int64) error
	ResetAutoClose(ctx context.Context, ticketID string, autoCloseAt *time.Time) error
	GetStats(ctx context.Context, guildID int64) (*model.TicketStats, error)
}

// TicketCategoryRepository defines operations on panels and categories.
type TicketCategoryRepository interface {
	GetCategory(ctx context.Context, categoryID string) (*model.TicketCategory, error)
	ListPanelHandlerRoles(ctx context.Context, panelID string) ([]model.PanelHandlerRole, error)
	ListCategoryHandlerRoles(ctx context.Context, categoryID string) ([]model.CategoryHandlerRole, error)
	GetPanel(ctx context.Context, panelID string) (*model.TicketPanel, error)
	ListCategoriesByPanel(ctx context.Context, panelID string) ([]model.TicketCategory, error)
	SetPanelMessage(ctx context.Context, panelID string, messageID int64) error
	ClearPanelMessage(ctx context.Context, panelID string) error
}

// TicketGuildRepository defines operations on guilds and staff roles.
type TicketGuildRepository interface {
	GetByID(ctx context.Context, guildID int64) (*guild_model.Guild, error)
	ListStaffRoles(ctx context.Context, guildID int64) ([]guild_model.StaffRole, error)
}

// TicketMessageRepository logs messages for transcripts.
type TicketMessageRepository interface {
	Add(ctx context.Context, m *model.TicketMessage) (*model.TicketMessage, error)
}

// TranscriptRepository defines the database operations consumed by TranscriptService.
type TranscriptRepository interface {
	ListByTicket(ctx context.Context, ticketID string) ([]model.TicketMessage, error)
	ListNonNotesByTicket(ctx context.Context, ticketID string) ([]model.TicketMessage, error)
}

// SchedulerRepository defines the operations needed by the Ticket Scheduler.
type SchedulerRepository interface {
	ListDueForAutoClose(ctx context.Context) ([]model.Ticket, error)
}

// ─── TranscriptService ────────────────────────────────────────────────────────

// TranscriptService generates plain-text and rich HTML transcripts for ticket messages.
type TranscriptService struct {
	repo TranscriptRepository
}

// NewTranscriptService constructs a new TranscriptService.
func NewTranscriptService(repo TranscriptRepository) *TranscriptService {
	return &TranscriptService{repo: repo}
}

// GenerateText creates a plain text transcript of the ticket.
func (s *TranscriptService) GenerateText(ctx context.Context, ticketID string, includeStaffNotes bool) (string, error) {
	var messages []model.TicketMessage
	var err error

	if includeStaffNotes {
		messages, err = s.repo.ListByTicket(ctx, ticketID)
	} else {
		messages, err = s.repo.ListNonNotesByTicket(ctx, ticketID)
	}
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== Transcript for Ticket: %s ===\n", ticketID))
	sb.WriteString(fmt.Sprintf("Generated on: %s\n\n", time.Now().UTC().Format(time.RFC1123)))

	for _, msg := range messages {
		timeStr := msg.SentAt.UTC().Format("2006-01-02 15:04:05 UTC")
		content := ""
		if msg.Content != nil {
			content = *msg.Content
		}

		prefix := ""
		if msg.IsStaffNote {
			prefix = "[STAFF NOTE] "
		}

		sb.WriteString(fmt.Sprintf("[%s] %s%s: %s\n", timeStr, prefix, msg.AuthorUsername, content))

		attachments, err := msg.GetAttachments()
		if err == nil && len(attachments) > 0 {
			for _, att := range attachments {
				sb.WriteString(fmt.Sprintf("   -> Attachment: %s (%s)\n", att.Filename, att.URL))
			}
		}
	}

	return sb.String(), nil
}

// GenerateHTML creates a beautiful self-contained HTML transcript page.
func (s *TranscriptService) GenerateHTML(ctx context.Context, ticket *model.Ticket, categoryName string, openerUsername string, includeStaffNotes bool) (string, error) {
	var messages []model.TicketMessage
	var err error

	if includeStaffNotes {
		messages, err = s.repo.ListByTicket(ctx, ticket.ID)
	} else {
		messages, err = s.repo.ListNonNotesByTicket(ctx, ticket.ID)
	}
	if err != nil {
		return "", err
	}

	closedBy := "N/A"
	if ticket.ClaimedBy != nil {
		closedBy = fmt.Sprintf("ID: %d", *ticket.ClaimedBy)
	}
	closeReason := "None"
	if ticket.CloseReason != nil {
		closeReason = *ticket.CloseReason
	}

	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Ticket Transcript - Shrimpy</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=Outfit:wght@500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: #1a1830;
            --container-bg: #22203f;
            --text-primary: #ffffff;
            --text-secondary: #a0aec0;
            --primary-coral: #ff7b6b;
            --accent-teal: #4ecdc4;
            --staff-note-bg: rgba(251, 191, 36, 0.1);
            --staff-note-border: #fbbf24;
            --msg-bg-member: #2d2a4e;
            --msg-bg-staff: #2b4545;
            --border-color: #312e5c;
        }

        body {
            font-family: 'Inter', sans-serif;
            background-color: var(--bg-color);
            color: var(--text-primary);
            margin: 0;
            padding: 2rem 1rem;
            display: flex;
            justify-content: center;
        }

        .container {
            width: 100%;
            max-width: 900px;
            background-color: var(--container-bg);
            border-radius: 12px;
            border: 1px solid var(--border-color);
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
            overflow: hidden;
            display: flex;
            flex-direction: column;
        }

        .header {
            padding: 2rem;
            background: linear-gradient(135deg, #2b2853 0%, #1c1a36 100%);
            border-bottom: 1px solid var(--border-color);
        }

        .header h1 {
            font-family: 'Outfit', sans-serif;
            font-size: 1.8rem;
            margin: 0 0 1rem 0;
            color: var(--primary-coral);
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .metadata-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            font-size: 0.9rem;
            color: var(--text-secondary);
        }

        .meta-item strong {
            color: var(--text-primary);
        }

        .messages-list {
            padding: 2rem;
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
            background-color: var(--container-bg);
            overflow-y: auto;
        }

        .message-row {
            display: flex;
            gap: 1rem;
            align-items: flex-start;
        }

        .avatar {
            width: 42px;
            height: 42px;
            background-color: var(--primary-coral);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: 600;
            font-family: 'Outfit', sans-serif;
            color: white;
            flex-shrink: 0;
        }

        .message-content-wrapper {
            flex-grow: 1;
            display: flex;
            flex-direction: column;
            gap: 0.3rem;
        }

        .message-info {
            display: flex;
            align-items: center;
            gap: 0.6rem;
            font-size: 0.85rem;
        }

        .username {
            font-weight: 600;
            color: var(--text-primary);
        }

        .timestamp {
            color: var(--text-secondary);
            font-size: 0.75rem;
        }

        .badge {
            font-size: 0.75rem;
            padding: 0.1rem 0.4rem;
            border-radius: 4px;
            font-weight: 500;
        }

        .badge-staff {
            background-color: var(--accent-teal);
            color: #1a1830;
        }

        .badge-staff-note {
            background-color: var(--staff-note-border);
            color: #1a1830;
        }

        .message-bubble {
            background-color: var(--msg-bg-member);
            padding: 0.8rem 1rem;
            border-radius: 0 12px 12px 12px;
            border: 1px solid var(--border-color);
            font-size: 0.95rem;
            line-height: 1.4;
            word-break: break-word;
            white-space: pre-wrap;
        }

        .message-row.is-note .message-bubble {
            background-color: var(--staff-note-bg);
            border-color: var(--staff-note-border);
        }

        .attachments-list {
            margin-top: 0.5rem;
            display: flex;
            flex-direction: column;
            gap: 0.4rem;
        }

        .attachment-item {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            background-color: rgba(0, 0, 0, 0.2);
            padding: 0.4rem 0.8rem;
            border-radius: 6px;
            font-size: 0.85rem;
            text-decoration: none;
            color: var(--accent-teal);
            border: 1px solid rgba(78, 205, 196, 0.2);
            width: fit-content;
        }

        .attachment-item:hover {
            background-color: rgba(78, 205, 196, 0.1);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1><h1>🦐 Shrimpy Support Transcript</h1></h1>
            <div class="metadata-grid">
                <div class="meta-item"><strong>Ticket ID:</strong> ` + ticket.ID + `</div>
                <div class="meta-item"><strong>Category:</strong> ` + html.EscapeString(categoryName) + `</div>
                <div class="meta-item"><strong>Opened By:</strong> ` + html.EscapeString(openerUsername) + `</div>
                <div class="meta-item"><strong>Opened At:</strong> ` + ticket.CreatedAt.UTC().Format("2006-01-02 15:04:05 UTC") + `</div>
                <div class="meta-item"><strong>Closed By:</strong> ` + html.EscapeString(closedBy) + `</div>
                <div class="meta-item"><strong>Close Reason:</strong> ` + html.EscapeString(closeReason) + `</div>
            </div>
        </div>
        <div class="messages-list">`)

	for _, msg := range messages {
		noteClass := ""
		if msg.IsStaffNote {
			noteClass = " is-note"
		}

		authorInitial := "?"
		if len(msg.AuthorUsername) > 0 {
			authorInitial = strings.ToUpper(string(msg.AuthorUsername[0]))
		}

		timeStr := msg.SentAt.UTC().Format("2006-01-02 15:04:05")

		content := ""
		if msg.Content != nil {
			content = html.EscapeString(*msg.Content)
		}

		sb.WriteString(fmt.Sprintf(`
            <div class="message-row%s">
                <div class="avatar">%s</div>
                <div class="message-content-wrapper">
                    <div class="message-info">
                        <span class="username">%s</span>`, noteClass, authorInitial, html.EscapeString(msg.AuthorUsername)))

		if msg.IsStaffNote {
			sb.WriteString(` <span class="badge badge-staff-note">Staff Note</span>`)
		}

		sb.WriteString(fmt.Sprintf(`
                        <span class="timestamp">%s UTC</span>
                    </div>
                    <div class="message-bubble">%s</div>`, timeStr, content))

		attachments, err := msg.GetAttachments()
		if err == nil && len(attachments) > 0 {
			sb.WriteString(`
                    <div class="attachments-list">`)
			for _, att := range attachments {
				sb.WriteString(fmt.Sprintf(`
                        <a href="%s" target="_blank" class="attachment-item" rel="noopener noreferrer">
                            📎 %s (%d bytes)
                        </a>`, html.EscapeString(att.URL), html.EscapeString(att.Filename), att.Size))
			}
			sb.WriteString(`
                    </div>`)
		}

		sb.WriteString(`
                </div>
            </div>`)
	}

	sb.WriteString(`
        </div>
    </div>
</body>
</html>`)

	return sb.String(), nil
}

// ─── TicketService ───────────────────────────────────────────────────────────

// TicketService coordinates opening, claiming, closing, and archiving tickets.
type TicketService struct {
	ticketRepo   TicketRepository
	categoryRepo TicketCategoryRepository
	guildRepo    TicketGuildRepository
	msgRepo      TicketMessageRepository
	transcript   *TranscriptService
}

// NewTicketService constructs a new TicketService.
func NewTicketService(
	ticketRepo TicketRepository,
	categoryRepo TicketCategoryRepository,
	guildRepo TicketGuildRepository,
	msgRepo TicketMessageRepository,
	transcript *TranscriptService,
) *TicketService {
	return &TicketService{
		ticketRepo:   ticketRepo,
		categoryRepo: categoryRepo,
		guildRepo:    guildRepo,
		msgRepo:      msgRepo,
		transcript:   transcript,
	}
}

// Open creates a new ticket in the DB, provisions a channel or thread on Discord, and sends the greeting.
func (s *TicketService) Open(ctx context.Context, dg *discordgo.Session, guildID int64, categoryID string, userID int64) (*model.Ticket, error) {
	cat, err := s.categoryRepo.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	openCount, err := s.ticketRepo.CountOpenByUser(ctx, guildID, categoryID, userID)
	if err != nil {
		return nil, err
	}

	if int(openCount) >= cat.MaxTicketsPerUser {
		return nil, model.ErrLimitReached
	}

	ticket := &model.Ticket{
		GuildID:    guildID,
		CategoryID: categoryID,
		OpenedBy:   userID,
		Status:     model.TicketStatusOpen,
		Priority:   model.TicketPriorityMedium,
	}

	if cat.AutoCloseHours != nil {
		acTime := time.Now().Add(time.Duration(*cat.AutoCloseHours) * time.Hour)
		ticket.AutoCloseAt = &acTime
	}

	ticket, err = s.ticketRepo.Create(ctx, ticket)
	if err != nil {
		return nil, err
	}

	guildIDStr := fmt.Sprintf("%d", guildID)
	userIDStr := fmt.Sprintf("%d", userID)

	channelName := strings.ReplaceAll(cat.TicketNameTemplate, "{username}", strings.ToLower(userIDStr))
	channelName = strings.ReplaceAll(channelName, "{category}", strings.ToLower(cat.Name))
	channelName = strings.ReplaceAll(channelName, "{number}", ticket.ID[:8])

	var targetChannelID, targetThreadID *int64

	staffRoles, err := s.guildRepo.ListStaffRoles(ctx, guildID)
	if err != nil {
		staffRoles = []guild_model.StaffRole{}
	}

	permissionOverrides := []*discordgo.PermissionOverwrite{
		{
			ID:   guildIDStr,
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: discordgo.PermissionReadMessages,
		},
		{
			ID:    userIDStr,
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionReadMessages | discordgo.PermissionSendMessages | discordgo.PermissionEmbedLinks | discordgo.PermissionAttachFiles,
		},
	}

	addedRoleIDs := make(map[int64]bool)
	for _, sr := range staffRoles {
		permissionOverrides = append(permissionOverrides, &discordgo.PermissionOverwrite{
			ID:    fmt.Sprintf("%d", sr.RoleID),
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionReadMessages | discordgo.PermissionSendMessages | discordgo.PermissionEmbedLinks | discordgo.PermissionAttachFiles,
		})
		addedRoleIDs[sr.RoleID] = true
	}

	panelHandlerRoles, err := s.categoryRepo.ListPanelHandlerRoles(ctx, cat.PanelID)
	if err != nil {
		panelHandlerRoles = []model.PanelHandlerRole{}
	}
	for _, hr := range panelHandlerRoles {
		if addedRoleIDs[hr.RoleID] {
			continue
		}
		permissionOverrides = append(permissionOverrides, &discordgo.PermissionOverwrite{
			ID:    fmt.Sprintf("%d", hr.RoleID),
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionReadMessages | discordgo.PermissionSendMessages | discordgo.PermissionEmbedLinks | discordgo.PermissionAttachFiles,
		})
		addedRoleIDs[hr.RoleID] = true
	}

	categoryHandlerRoles, err := s.categoryRepo.ListCategoryHandlerRoles(ctx, cat.ID)
	if err != nil {
		categoryHandlerRoles = []model.CategoryHandlerRole{}
	}
	for _, hr := range categoryHandlerRoles {
		if addedRoleIDs[hr.RoleID] {
			continue
		}
		permissionOverrides = append(permissionOverrides, &discordgo.PermissionOverwrite{
			ID:    fmt.Sprintf("%d", hr.RoleID),
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionReadMessages | discordgo.PermissionSendMessages | discordgo.PermissionEmbedLinks | discordgo.PermissionAttachFiles,
		})
		addedRoleIDs[hr.RoleID] = true
	}

	botUser, err := dg.User("@me")
	if err == nil {
		permissionOverrides = append(permissionOverrides, &discordgo.PermissionOverwrite{
			ID:    botUser.ID,
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionReadMessages | discordgo.PermissionSendMessages | discordgo.PermissionEmbedLinks | discordgo.PermissionAttachFiles | discordgo.PermissionManageChannels,
		})
	}

	ch, err := dg.GuildChannelCreateComplex(guildIDStr, discordgo.GuildChannelCreateData{
		Name:                 channelName,
		Type:                 discordgo.ChannelTypeGuildText,
		PermissionOverwrites: permissionOverrides,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord channel: %w", err)
	}

	chID, _ := discordutil.ParseID(ch.ID)
	targetChannelID = &chID

	err = s.ticketRepo.SetChannel(ctx, ticket.ID, targetChannelID, targetThreadID)
	if err != nil {
		return nil, err
	}
	ticket.ChannelID = targetChannelID

	openerMention := fmt.Sprintf("<@%d>", userID)
	replaceVars := func(text string) string {
		r := strings.NewReplacer(
			"{user}", openerMention,
			"{mention}", openerMention,
			"{category}", cat.Name,
			"{id}", ticket.ID,
		)
		return r.Replace(text)
	}

	media, _ := cat.GetOpenMedia()

	plainText := cat.TicketOpenContent
	embedFields := discordutil.EmbedFields{
		Title:       cat.TicketOpenTitle,
		Description: cat.TicketOpenMessage,
		Color:       cat.TicketOpenColor,
		Media:       media,
	}

	legacyTitle := "Support Ticket Created"
	legacyDesc := "Support staff will be with you shortly.\nClick below to claim or close this ticket."

	// Preserve the legacy hardcoded greeting only when the admin has configured
	// nothing at all (no plain text, no embed fields), so pre-existing/unconfigured
	// categories keep behaving exactly as before after upgrade.
	if (plainText == nil || *plainText == "") && !embedFields.HasContent() {
		embedFields.Title = &legacyTitle
		embedFields.Description = &legacyDesc
	}

	content, embed := discordutil.BuildContentAndEmbed(plainText, embedFields, replaceVars)

	// Keep the bot's own "Welcome {mention}" ping; the configurable plain text,
	// if any, is appended below it rather than replacing it.
	fullContent := fmt.Sprintf("Welcome %s", openerMention)
	if content != "" {
		fullContent = fullContent + "\n" + content
	}

	buttons := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Claim",
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("ticket:claim:%s", ticket.ID),
				Emoji:    &discordgo.ComponentEmoji{Name: "🙋"},
			},
			discordgo.Button{
				Label:    "Close",
				Style:    discordgo.DangerButton,
				CustomID: fmt.Sprintf("ticket:close:%s", ticket.ID),
				Emoji:    &discordgo.ComponentEmoji{Name: "🔒"},
			},
		},
	}

	params := &discordgo.MessageSend{
		Content:    fullContent,
		Components: []discordgo.MessageComponent{buttons},
	}
	if embed != nil {
		params.Embeds = []*discordgo.MessageEmbed{embed}
	}

	_, err = dg.ChannelMessageSendComplex(ch.ID, params)
	if err != nil {
		fmt.Printf("warning: failed to send greeting message to ticket channel: %v\n", err)
	}

	return ticket, nil
}

// SyncPanelMessage rebuilds the panel's message (text, embed, category buttons) from
// current DB state and posts it if not yet posted, or edits the existing message in place.
// If the edit fails (e.g. the message was deleted manually), it falls back to posting fresh.
func (s *TicketService) SyncPanelMessage(ctx context.Context, dg *discordgo.Session, panelID string) error {
	panel, err := s.categoryRepo.GetPanel(ctx, panelID)
	if err != nil {
		return fmt.Errorf("failed to load panel: %w", err)
	}

	cats, err := s.categoryRepo.ListCategoriesByPanel(ctx, panelID)
	if err != nil {
		return fmt.Errorf("failed to load categories: %w", err)
	}

	identity := func(text string) string { return text }

	media, _ := panel.GetMedia()
	embedFields := discordutil.EmbedFields{
		Title:       panel.EmbedTitle,
		Description: panel.EmbedDescription,
		Color:       panel.EmbedColor,
		Media:       media,
	}
	content, embed := discordutil.BuildContentAndEmbed(panel.Content, embedFields, identity)

	components := buildPanelButtons(cats)

	params := &discordgo.MessageSend{Content: content, Components: components}
	if embed != nil {
		params.Embeds = []*discordgo.MessageEmbed{embed}
	}

	channelIDStr := discordutil.FormatID(panel.ChannelID)

	if panel.MessageID == nil {
		return s.postNewPanelMessage(ctx, dg, panel, channelIDStr, params)
	}

	embeds := params.Embeds
	editParams := &discordgo.MessageEdit{
		ID:         discordutil.FormatID(*panel.MessageID),
		Channel:    channelIDStr,
		Content:    &params.Content,
		Components: &params.Components,
		Embeds:     &embeds,
	}

	_, err = dg.ChannelMessageEditComplex(editParams)
	if err != nil {
		return s.postNewPanelMessage(ctx, dg, panel, channelIDStr, params)
	}
	return nil
}

// postNewPanelMessage sends a brand-new panel message and persists the returned Discord message ID.
func (s *TicketService) postNewPanelMessage(ctx context.Context, dg *discordgo.Session, panel *model.TicketPanel, channelIDStr string, params *discordgo.MessageSend) error {
	msg, err := dg.ChannelMessageSendComplex(channelIDStr, params)
	if err != nil {
		return fmt.Errorf("failed to post panel message: %w", err)
	}
	msgID, err := discordutil.ParseID(msg.ID)
	if err != nil {
		return fmt.Errorf("invalid Discord message ID returned: %w", err)
	}
	return s.categoryRepo.SetPanelMessage(ctx, panel.ID, msgID)
}

// DeletePanelMessage best-effort deletes the Discord message backing a panel, if one was ever posted.
func (s *TicketService) DeletePanelMessage(ctx context.Context, dg *discordgo.Session, panel *model.TicketPanel) error {
	if panel.MessageID == nil {
		return nil
	}
	return dg.ChannelMessageDelete(discordutil.FormatID(panel.ChannelID), discordutil.FormatID(*panel.MessageID))
}

// buildPanelButtons renders one button per category (ordered by ButtonOrder via the
// repository query), grouped into ActionsRows of up to 5 buttons (Discord's per-row
// limit), capped at 5 rows (Discord's per-message limit).
func buildPanelButtons(cats []model.TicketCategory) []discordgo.MessageComponent {
	if len(cats) == 0 {
		return nil
	}
	var rows []discordgo.MessageComponent
	var current []discordgo.MessageComponent
	for _, cat := range cats {
		btn := discordgo.Button{
			Label:    cat.ButtonLabel,
			Style:    buttonStyleFromString(cat.ButtonStyle),
			CustomID: fmt.Sprintf("ticket:open:%s", cat.ID),
		}
		if cat.Emoji != nil && *cat.Emoji != "" {
			btn.Emoji = &discordgo.ComponentEmoji{Name: *cat.Emoji}
		}
		current = append(current, btn)
		if len(current) == 5 {
			rows = append(rows, discordgo.ActionsRow{Components: current})
			current = nil
		}
	}
	if len(current) > 0 {
		rows = append(rows, discordgo.ActionsRow{Components: current})
	}
	if len(rows) > 5 {
		rows = rows[:5]
	}
	return rows
}

func buttonStyleFromString(s string) discordgo.ButtonStyle {
	switch s {
	case "secondary":
		return discordgo.SecondaryButton
	case "success":
		return discordgo.SuccessButton
	case "danger":
		return discordgo.DangerButton
	default:
		return discordgo.PrimaryButton
	}
}

// Claim updates the ticket status and claimed_by field, renames the channel, and notifies staff.
func (s *TicketService) Claim(ctx context.Context, dg *discordgo.Session, ticketID string, staffUserID int64) (*model.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.Status == model.TicketStatusClosed || ticket.Status == model.TicketStatusArchived {
		return nil, fmt.Errorf("cannot claim a closed or archived ticket")
	}

	ticket, err = s.ticketRepo.UpdateClaim(ctx, ticketID, &staffUserID)
	if err != nil {
		return nil, err
	}

	if ticket.ChannelID != nil {
		chIDStr := fmt.Sprintf("%d", *ticket.ChannelID)
		ch, err := dg.Channel(chIDStr)
		if err == nil {
			newName := ch.Name
			if !strings.HasPrefix(newName, "claimed-") {
				newName = "claimed-" + newName
			}
			_, _ = dg.ChannelEdit(chIDStr, &discordgo.ChannelEdit{
				Name: newName,
			})
		}

		embed := &discordgo.MessageEmbed{
			Description: fmt.Sprintf("This ticket has been claimed by <@%d>.", staffUserID),
			Color:       0x4ecdc4,
		}
		_, _ = dg.ChannelMessageSendEmbed(chIDStr, embed)
	}

	return ticket, nil
}

// Unclaim releases a ticket back to the open pool.
func (s *TicketService) Unclaim(ctx context.Context, dg *discordgo.Session, ticketID string) (*model.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.Status != model.TicketStatusClaimed {
		return nil, fmt.Errorf("ticket is not claimed")
	}

	ticket, err = s.ticketRepo.UpdateClaim(ctx, ticketID, nil)
	if err != nil {
		return nil, err
	}

	if ticket.ChannelID != nil {
		chIDStr := fmt.Sprintf("%d", *ticket.ChannelID)
		ch, err := dg.Channel(chIDStr)
		if err == nil {
			newName := strings.TrimPrefix(ch.Name, "claimed-")
			_, _ = dg.ChannelEdit(chIDStr, &discordgo.ChannelEdit{
				Name: newName,
			})
		}

		embed := &discordgo.MessageEmbed{
			Description: "This ticket has been unclaimed and is now open for other staff.",
			Color:       0xff7b6b,
		}
		_, _ = dg.ChannelMessageSendEmbed(chIDStr, embed)
	}

	return ticket, nil
}

// Close locks the channel, compiles and posts the transcript, and updates DB status.
func (s *TicketService) Close(ctx context.Context, dg *discordgo.Session, ticketID string, reason *string, closedByUserID int64) (*model.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.Status == model.TicketStatusClosed || ticket.Status == model.TicketStatusArchived {
		return nil, fmt.Errorf("ticket is already closed/archived")
	}

	if ticket.ChannelID != nil {
		chIDStr := fmt.Sprintf("%d", *ticket.ChannelID)
		openerStr := fmt.Sprintf("%d", ticket.OpenedBy)
		_ = dg.ChannelPermissionSet(chIDStr, openerStr, discordgo.PermissionOverwriteTypeMember, 0, discordgo.PermissionSendMessages)
	}

	ticket, err = s.ticketRepo.UpdateStatus(ctx, ticketID, model.TicketStatusClosed, reason)
	if err != nil {
		return nil, err
	}

	cat, err := s.categoryRepo.GetCategory(ctx, ticket.CategoryID)
	if err == nil {
		openerUser, err := dg.User(fmt.Sprintf("%d", ticket.OpenedBy))
		openerName := "Unknown User"
		if err == nil {
			openerName = openerUser.Username
		}

		htmlContent, err := s.transcript.GenerateHTML(ctx, ticket, cat.Name, openerName, true)
		if err == nil {
			var logChannelID *int64
			if cat.TranscriptChannelID != nil {
				logChannelID = cat.TranscriptChannelID
			} else {
				g, err := s.guildRepo.GetByID(ctx, ticket.GuildID)
				if err == nil {
					logChannelID = g.LogChannelID
				}
			}

			if logChannelID != nil {
				logChStr := fmt.Sprintf("%d", *logChannelID)
				fileReader := strings.NewReader(htmlContent)
				fileName := fmt.Sprintf("transcript-%s.html", ticket.ID[:8])

				_, _ = dg.ChannelMessageSendComplex(logChStr, &discordgo.MessageSend{
					Content: fmt.Sprintf("Transcript for closed ticket **%s** (Category: %s)", ticket.ID, cat.Name),
					Files: []*discordgo.File{
						{
							Name:        fileName,
							ContentType: "text/html",
							Reader:      fileReader,
						},
					},
				})
			}
		}
	}

	if ticket.ChannelID != nil {
		chIDStr := fmt.Sprintf("%d", *ticket.ChannelID)
		reasonStr := "No reason provided"
		if reason != nil {
			reasonStr = *reason
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Ticket Closed",
			Description: fmt.Sprintf("Closed by: <@%d>\nReason: %s", closedByUserID, reasonStr),
			Color:       0x312e5c,
		}

		buttons := discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Reopen",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("ticket:reopen:%s", ticket.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "🔓"},
				},
				discordgo.Button{
					Label:    "Archive / Delete",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("ticket:archive:%s", ticket.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "🗑️"},
				},
			},
		}

		_, _ = dg.ChannelMessageSendComplex(chIDStr, &discordgo.MessageSend{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{buttons},
		})
	}

	return ticket, nil
}

// Reopen unlocks the ticket channel and restores the opener's permissions.
func (s *TicketService) Reopen(ctx context.Context, dg *discordgo.Session, ticketID string) (*model.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.Status != model.TicketStatusClosed {
		return nil, fmt.Errorf("ticket is not closed")
	}

	if ticket.ChannelID != nil {
		chIDStr := fmt.Sprintf("%d", *ticket.ChannelID)
		openerStr := fmt.Sprintf("%d", ticket.OpenedBy)
		_ = dg.ChannelPermissionSet(chIDStr, openerStr, discordgo.PermissionOverwriteTypeMember,
			discordgo.PermissionReadMessages|discordgo.PermissionSendMessages|discordgo.PermissionEmbedLinks|discordgo.PermissionAttachFiles, 0)
	}

	ticket, err = s.ticketRepo.UpdateStatus(ctx, ticketID, model.TicketStatusOpen, nil)
	if err != nil {
		return nil, err
	}

	if ticket.ChannelID != nil {
		chIDStr := fmt.Sprintf("%d", *ticket.ChannelID)
		embed := &discordgo.MessageEmbed{
			Description: "This ticket has been reopened.",
			Color:       0x4ecdc4,
		}
		_, _ = dg.ChannelMessageSendEmbed(chIDStr, embed)
	}

	return ticket, nil
}

// Archive deletes the Discord channel/thread, retaining DB records.
func (s *TicketService) Archive(ctx context.Context, dg *discordgo.Session, ticketID string) error {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.ChannelID != nil {
		_, _ = dg.ChannelDelete(fmt.Sprintf("%d", *ticket.ChannelID))
	}

	_, err = s.ticketRepo.UpdateStatus(ctx, ticketID, model.TicketStatusArchived, nil)
	return err
}

// LogMessage records an incoming chat message in the DB transcript store.
func (s *TicketService) LogMessage(ctx context.Context, ticketID string, authorID int64, username string, content *string, isStaffNote bool, attachments []discordutil.Attachment) error {
	m := &model.TicketMessage{
		TicketID:       ticketID,
		AuthorID:       authorID,
		AuthorUsername: username,
		Content:        content,
		IsStaffNote:    isStaffNote,
		SentAt:         time.Now().UTC(),
	}

	if len(attachments) > 0 {
		b, err := json.Marshal(attachments)
		if err == nil {
			m.Attachments = b
		}
	}

	_, err := s.msgRepo.Add(ctx, m)

	if err == nil {
		ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
		if err == nil && ticket.AutoCloseAt != nil {
			cat, err := s.categoryRepo.GetCategory(ctx, ticket.CategoryID)
			if err == nil && cat.AutoCloseHours != nil {
				newCloseTime := time.Now().Add(time.Duration(*cat.AutoCloseHours) * time.Hour)
				_ = s.ticketRepo.ResetAutoClose(ctx, ticketID, &newCloseTime)
			}
		}
	}

	return err
}

func (s *TicketService) GetByChannelID(ctx context.Context, channelID int64) (*model.Ticket, error) {
	return s.ticketRepo.GetByChannelID(ctx, channelID)
}

func (s *TicketService) UpdatePriority(ctx context.Context, ticketID string, priority model.TicketPriority) (*model.Ticket, error) {
	return s.ticketRepo.UpdatePriority(ctx, ticketID, priority)
}

func (s *TicketService) GetByID(ctx context.Context, ticketID string) (*model.Ticket, error) {
	return s.ticketRepo.GetByID(ctx, ticketID)
}

func (s *TicketService) List(ctx context.Context, guildID int64, f model.TicketFilter) ([]model.Ticket, int64, error) {
	return s.ticketRepo.List(ctx, guildID, f)
}

func (s *TicketService) GetStats(ctx context.Context, guildID int64) (*model.TicketStats, error) {
	return s.ticketRepo.GetStats(ctx, guildID)
}

// ─── Scheduler ────────────────────────────────────────────────────────────────

// Scheduler runs checks in the background to close due tickets automatically.
type Scheduler struct {
	ticketRepo SchedulerRepository
	ticketSvc  *TicketService
	interval   time.Duration
}

// NewScheduler constructs a new Scheduler.
func NewScheduler(ticketRepo SchedulerRepository, ticketSvc *TicketService, interval time.Duration) *Scheduler {
	return &Scheduler{
		ticketRepo: ticketRepo,
		ticketSvc:  ticketSvc,
		interval:   interval,
	}
}

// Start initiates the periodic poll. Blocks until the context is cancelled.
func (s *Scheduler) Start(ctx context.Context, provider discordutil.DiscordSessionProvider) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	fmt.Printf("Scheduler: Auto-close background check active (every %s)\n", s.interval)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Scheduler: Auto-close check stopped.")
			return
		case <-ticker.C:
			s.runCheck(ctx, provider)
		}
	}
}

func (s *Scheduler) runCheck(ctx context.Context, provider discordutil.DiscordSessionProvider) {
	tickets, err := s.ticketRepo.ListDueForAutoClose(ctx)
	if err != nil {
		fmt.Printf("Scheduler Error: failed to list due tickets: %v\n", err)
		return
	}

	if len(tickets) == 0 {
		return
	}

	fmt.Printf("Scheduler: Found %d tickets due for auto-closing.\n", len(tickets))

	for _, t := range tickets {
		dg, err := provider.GetSessionForGuild(ctx, t.GuildID)
		if err != nil {
			fmt.Printf("Scheduler Error: failed to resolve bot session for guild %d: %v\n", t.GuildID, err)
			continue
		}

		botUser, err := dg.User("@me")
		var botUserID int64
		if err == nil {
			botUserID, _ = discordutil.ParseID(botUser.ID)
		}

		reason := "Auto-closed due to inactivity."
		_, err = s.ticketSvc.Close(ctx, dg, t.ID, &reason, botUserID)
		if err != nil {
			fmt.Printf("Scheduler Error: failed to close ticket %s: %v\n", t.ID, err)
		} else {
			fmt.Printf("Scheduler: Successfully auto-closed ticket %s\n", t.ID)
		}
	}
}
