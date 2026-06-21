package service

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
)

// TranscriptRepository defines the database operations consumed by TranscriptService.
type TranscriptRepository interface {
	ListByTicket(ctx context.Context, ticketID string) ([]repository.TicketMessage, error)
	ListNonNotesByTicket(ctx context.Context, ticketID string) ([]repository.TicketMessage, error)
}

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
	var messages []repository.TicketMessage
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

		// Handle staff note prefix
		prefix := ""
		if msg.IsStaffNote {
			prefix = "[STAFF NOTE] "
		}

		sb.WriteString(fmt.Sprintf("[%s] %s%s: %s\n", timeStr, prefix, msg.AuthorUsername, content))

		// Render attachments
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
func (s *TranscriptService) GenerateHTML(ctx context.Context, ticket *repository.Ticket, categoryName string, openerUsername string, includeStaffNotes bool) (string, error) {
	var messages []repository.TicketMessage
	var err error

	if includeStaffNotes {
		messages, err = s.repo.ListByTicket(ctx, ticket.ID)
	} else {
		messages, err = s.repo.ListNonNotesByTicket(ctx, ticket.ID)
	}
	if err != nil {
		return "", err
	}

	// Basic metadata strings
	closedBy := "N/A"
	if ticket.ClaimedBy != nil {
		closedBy = fmt.Sprintf("ID: %d", *ticket.ClaimedBy)
	}
	closeReason := "None"
	if ticket.CloseReason != nil {
		closeReason = *ticket.CloseReason
	}

	var sb strings.Builder

	// HTML template header
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
            <h1>🦐 Shrimpy Support Transcript</h1>
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

		// Badges
		if msg.IsStaffNote {
			sb.WriteString(` <span class="badge badge-staff-note">Staff Note</span>`)
		}

		sb.WriteString(fmt.Sprintf(`
                        <span class="timestamp">%s UTC</span>
                    </div>
                    <div class="message-bubble">%s</div>`, timeStr, content))

		// Attachments rendering
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

	// HTML footer
	sb.WriteString(`
        </div>
    </div>
</body>
</html>`)

	return sb.String(), nil
}
