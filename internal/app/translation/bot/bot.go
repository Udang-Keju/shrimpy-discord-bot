package bot

import (
	"context"
	"fmt"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/translation/service"
	"github.com/bwmarrin/discordgo"
)

// BotHandler handles Discord message and reaction events for translation.
type BotHandler struct {
	translationSvc *service.TranslationService
}

// NewBotHandler constructs a new BotHandler.
func NewBotHandler(translationSvc *service.TranslationService) *BotHandler {
	return &BotHandler{translationSvc: translationSvc}
}

// OnMessageCreate runs the auto-translate path for member messages.
func (h *BotHandler) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID == "" || m.Author == nil || m.Author.Bot {
		return
	}

	go func() {
		if err := h.translationSvc.TranslateMessage(context.Background(), s, m); err != nil {
			fmt.Printf("Bot Error: failed to auto-translate message: %v\n", err)
		}
	}()
}

// OnMessageReactionAdd runs the reaction-trigger translate path.
func (h *BotHandler) OnMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.GuildID == "" || r.UserID == s.State.User.ID {
		return
	}

	go func() {
		if err := h.translationSvc.TranslateReaction(context.Background(), s, r); err != nil {
			fmt.Printf("Bot Error: failed to handle translation reaction: %v\n", err)
		}
	}()
}
