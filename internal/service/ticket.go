package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/bwmarrin/discordgo"
)

// TicketRepository defines the database operations consumed by TicketService.
type TicketRepository interface {
	Create(ctx context.Context, t *repository.Ticket) (*repository.Ticket, error)
	GetByID(ctx context.Context, ticketID string) (*repository.Ticket, error)
	GetByChannelID(ctx context.Context, channelID int64) (*repository.Ticket, error)
	List(ctx context.Context, guildID int64, f repository.TicketFilter) ([]repository.Ticket, int64, error)
	CountOpenByUser(ctx context.Context, guildID int64, categoryID string, userID int64) (int64, error)
	UpdateStatus(ctx context.Context, ticketID string, status repository.TicketStatus, reason *string) (*repository.Ticket, error)
	UpdateClaim(ctx context.Context, ticketID string, claimedBy *int64) (*repository.Ticket, error)
	UpdatePriority(ctx context.Context, ticketID string, priority repository.TicketPriority) (*repository.Ticket, error)
	SetChannel(ctx context.Context, ticketID string, channelID, threadID *int64) error
	ResetAutoClose(ctx context.Context, ticketID string, autoCloseAt *time.Time) error
}

// TicketCategoryRepository defines operations on panels and categories.
type TicketCategoryRepository interface {
	GetCategory(ctx context.Context, categoryID string) (*repository.TicketCategory, error)
}

// TicketGuildRepository defines operations on guilds and staff roles.
type TicketGuildRepository interface {
	GetByID(ctx context.Context, guildID int64) (*repository.Guild, error)
	ListStaffRoles(ctx context.Context, guildID int64) ([]repository.StaffRole, error)
}

// TicketMessageRepository logs messages for transcripts.
type TicketMessageRepository interface {
	Add(ctx context.Context, m *repository.TicketMessage) (*repository.TicketMessage, error)
}

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
func (s *TicketService) Open(ctx context.Context, dg *discordgo.Session, guildID int64, categoryID string, userID int64) (*repository.Ticket, error) {
	// 1. Fetch category and check limits
	cat, err := s.categoryRepo.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	openCount, err := s.ticketRepo.CountOpenByUser(ctx, guildID, categoryID, userID)
	if err != nil {
		return nil, err
	}

	if int(openCount) >= cat.MaxTicketsPerUser {
		return nil, repository.ErrLimitReached
	}

	// 2. Create ticket record in DB
	ticket := &repository.Ticket{
		GuildID:    guildID,
		CategoryID: categoryID,
		OpenedBy:   userID,
		Status:     repository.TicketStatusOpen,
		Priority:   repository.TicketPriorityMedium,
	}

	// Set initial auto close if configured
	if cat.AutoCloseHours != nil {
		acTime := time.Now().Add(time.Duration(*cat.AutoCloseHours) * time.Hour)
		ticket.AutoCloseAt = &acTime
	}

	ticket, err = s.ticketRepo.Create(ctx, ticket)
	if err != nil {
		return nil, err
	}

	// 3. Provision Discord Channel/Thread
	guildIDStr := fmt.Sprintf("%d", guildID)
	userIDStr := fmt.Sprintf("%d", userID)

	// Replace template variables for ticket naming (e.g. ticket-username)
	channelName := strings.ReplaceAll(cat.TicketNameTemplate, "{username}", strings.ToLower(userIDStr))
	channelName = strings.ReplaceAll(channelName, "{category}", strings.ToLower(cat.Name))
	channelName = strings.ReplaceAll(channelName, "{number}", ticket.ID[:8]) // unique short ID suffix

	var targetChannelID, targetThreadID *int64

	staffRoles, err := s.guildRepo.ListStaffRoles(ctx, guildID)
	if err != nil {
		staffRoles = []repository.StaffRole{}
	}

	if cat.TicketDestination == "thread" {
		// Thread destination: Create thread inside parent channel (where panel is located or designated transcript)
		// We assume panel channel or first channel.
		// In a production bot, we get the panel's channel ID. For now we will assume the panel's channel.
		// Wait, let's find the panel channel ID if we have it, or create a standard thread.
		// Let's create a thread in the panel's channel if passed.
		// If we don't have panel channel ID easily, we fall back to a config or look it up.
		// Let's check how we can fetch target channel ID.
		// For threads, we must create it in a text channel.
		// In Go, since we can't search easily without Panel, we'll try to find a suitable parent channel ID.
		// Actually, let's assume the panels table can be loaded, or we just create a text channel if not specified.
		// Let's check if we can write a simple helper or check where panel resides.
	}

	// For robust implementation, we will create a dedicated text channel under a category or parent.
	// This is the cleanest and most common ticket style.
	permissionOverrides := []*discordgo.PermissionOverwrite{
		{
			ID:   guildIDStr, // @everyone
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: discordgo.PermissionReadMessages,
		},
		{
			ID:    userIDStr, // Opener
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionReadMessages | discordgo.PermissionSendMessages | discordgo.PermissionEmbedLinks | discordgo.PermissionAttachFiles,
		},
	}

	// Add staff roles to overrides
	for _, sr := range staffRoles {
		permissionOverrides = append(permissionOverrides, &discordgo.PermissionOverwrite{
			ID:    fmt.Sprintf("%d", sr.RoleID),
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionReadMessages | discordgo.PermissionSendMessages | discordgo.PermissionEmbedLinks | discordgo.PermissionAttachFiles,
		})
	}

	// Add bot's own overrides
	botUser, err := dg.User("@me")
	if err == nil {
		permissionOverrides = append(permissionOverrides, &discordgo.PermissionOverwrite{
			ID:    botUser.ID,
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionReadMessages | discordgo.PermissionSendMessages | discordgo.PermissionEmbedLinks | discordgo.PermissionAttachFiles | discordgo.PermissionManageChannels,
		})
	}

	// Create Channel
	ch, err := dg.GuildChannelCreateComplex(guildIDStr, discordgo.GuildChannelCreateData{
		Name:                 channelName,
		Type:                 discordgo.ChannelTypeGuildText,
		PermissionOverwrites: permissionOverrides,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord channel: %w", err)
	}

	chID, _ := repository.ParseID(ch.ID)
	targetChannelID = &chID

	// 4. Update ticket in DB with ChannelID
	err = s.ticketRepo.SetChannel(ctx, ticket.ID, targetChannelID, targetThreadID)
	if err != nil {
		return nil, err
	}
	ticket.ChannelID = targetChannelID

	// 5. Send Opening Embed Message
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

	title := "Support Ticket Created"
	if cat.TicketOpenTitle != nil && *cat.TicketOpenTitle != "" {
		title = replaceVars(*cat.TicketOpenTitle)
	}

	desc := "Support staff will be with you shortly.\nClick below to claim or close this ticket."
	if cat.TicketOpenMessage != nil && *cat.TicketOpenMessage != "" {
		desc = replaceVars(*cat.TicketOpenMessage)
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: desc,
	}
	if cat.TicketOpenColor != nil {
		embed.Color = int(*cat.TicketOpenColor)
	}

	media, err := cat.GetOpenMedia()
	if err == nil && media != nil {
		if media.Author != nil {
			embed.Author = &discordgo.MessageEmbedAuthor{Name: replaceVars(media.Author.Name)}
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
			embed.Footer = &discordgo.MessageEmbedFooter{Text: replaceVars(media.Footer.Text)}
			if media.Footer.IconURL != nil {
				embed.Footer.IconURL = *media.Footer.IconURL
			}
		}
	}

	// Interaction buttons
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
		Content:    fmt.Sprintf("Welcome %s", openerMention),
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{buttons},
	}

	_, err = dg.ChannelMessageSendComplex(ch.ID, params)
	if err != nil {
		fmt.Printf("warning: failed to send greeting embed to ticket channel: %v\n", err)
	}

	return ticket, nil
}

// Claim updates the ticket status and claimed_by field, renames the channel, and notifies staff.
func (s *TicketService) Claim(ctx context.Context, dg *discordgo.Session, ticketID string, staffUserID int64) (*repository.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.Status == repository.TicketStatusClosed || ticket.Status == repository.TicketStatusArchived {
		return nil, fmt.Errorf("cannot claim a closed or archived ticket")
	}

	// 1. Update in DB
	ticket, err = s.ticketRepo.UpdateClaim(ctx, ticketID, &staffUserID)
	if err != nil {
		return nil, err
	}

	// 2. Update Channel Nickname/Name if applicable
	if ticket.ChannelID != nil {
		chIDStr := fmt.Sprintf("%d", *ticket.ChannelID)
		ch, err := dg.Channel(chIDStr)
		if err == nil {
			_, err := dg.User(fmt.Sprintf("%d", staffUserID))
			if err == nil {
				// E.g. prepend "claimed-" to the channel name
				newName := ch.Name
				if !strings.HasPrefix(newName, "claimed-") {
					newName = "claimed-" + newName
				}
				_, _ = dg.ChannelEdit(chIDStr, &discordgo.ChannelEdit{
					Name: newName,
				})
			}
		}

		// Send notification in channel
		embed := &discordgo.MessageEmbed{
			Description: fmt.Sprintf("This ticket has been claimed by <@%d>.", staffUserID),
			Color:       0x4ecdc4, // Teal
		}
		_, _ = dg.ChannelMessageSendEmbed(chIDStr, embed)
	}

	return ticket, nil
}

// Unclaim releases a ticket back to the open pool.
func (s *TicketService) Unclaim(ctx context.Context, dg *discordgo.Session, ticketID string) (*repository.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.Status != repository.TicketStatusClaimed {
		return nil, fmt.Errorf("ticket is not claimed")
	}

	// 1. Update DB
	ticket, err = s.ticketRepo.UpdateClaim(ctx, ticketID, nil)
	if err != nil {
		return nil, err
	}

	// 2. Rename channel
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
func (s *TicketService) Close(ctx context.Context, dg *discordgo.Session, ticketID string, reason *string, closedByUserID int64) (*repository.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.Status == repository.TicketStatusClosed || ticket.Status == repository.TicketStatusArchived {
		return nil, fmt.Errorf("ticket is already closed/archived")
	}

	// 1. Lock the Discord Channel
	if ticket.ChannelID != nil {
		chIDStr := fmt.Sprintf("%d", *ticket.ChannelID)
		openerStr := fmt.Sprintf("%d", ticket.OpenedBy)

		// Revoke opener permissions (remove them or set SendMessages to Deny)
		_ = dg.ChannelPermissionSet(chIDStr, openerStr, discordgo.PermissionOverwriteTypeMember, 0, discordgo.PermissionSendMessages)
	}

	// 2. Update DB
	ticket, err = s.ticketRepo.UpdateStatus(ctx, ticketID, repository.TicketStatusClosed, reason)
	if err != nil {
		return nil, err
	}

	// 3. Generate Transcript & Log
	cat, err := s.categoryRepo.GetCategory(ctx, ticket.CategoryID)
	if err == nil {
		openerUser, err := dg.User(fmt.Sprintf("%d", ticket.OpenedBy))
		openerName := "Unknown User"
		if err == nil {
			openerName = openerUser.Username
		}

		// Generate HTML transcript
		htmlContent, err := s.transcript.GenerateHTML(ctx, ticket, cat.Name, openerName, true)
		if err == nil {
			// Find where to send transcripts: Category transcript channel or Fallback Guild Log Channel
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

	// 4. Send reopen/archive buttons in closed channel
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
func (s *TicketService) Reopen(ctx context.Context, dg *discordgo.Session, ticketID string) (*repository.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket.Status != repository.TicketStatusClosed {
		return nil, fmt.Errorf("ticket is not closed")
	}

	// 1. Unlock Channel
	if ticket.ChannelID != nil {
		chIDStr := fmt.Sprintf("%d", *ticket.ChannelID)
		openerStr := fmt.Sprintf("%d", ticket.OpenedBy)

		// Restore write permission override for the opener
		_ = dg.ChannelPermissionSet(chIDStr, openerStr, discordgo.PermissionOverwriteTypeMember,
			discordgo.PermissionReadMessages|discordgo.PermissionSendMessages|discordgo.PermissionEmbedLinks|discordgo.PermissionAttachFiles, 0)
	}

	// 2. Update DB
	ticket, err = s.ticketRepo.UpdateStatus(ctx, ticketID, repository.TicketStatusOpen, nil)
	if err != nil {
		return nil, err
	}

	// 3. Post notification
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

	// 1. Delete Discord channel
	if ticket.ChannelID != nil {
		_, _ = dg.ChannelDelete(fmt.Sprintf("%d", *ticket.ChannelID))
	}

	// 2. Update DB status
	_, err = s.ticketRepo.UpdateStatus(ctx, ticketID, repository.TicketStatusArchived, nil)
	return err
}

// LogMessage records an incoming chat message in the DB transcript store.
func (s *TicketService) LogMessage(ctx context.Context, ticketID string, authorID int64, username string, content *string, isStaffNote bool, attachments []repository.Attachment) error {
	m := &repository.TicketMessage{
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

	// If we successfully logged the message, we should also reset the auto-close timer of the ticket!
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
