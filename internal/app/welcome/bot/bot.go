package bot

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/service"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
)

// BotHandler handles Discord bot events and slash commands for onboarding.
type BotHandler struct {
	welcomeSvc *service.WelcomeService
}

// NewBotHandler constructs a new BotHandler.
func NewBotHandler(welcomeSvc *service.WelcomeService) *BotHandler {
	return &BotHandler{
		welcomeSvc: welcomeSvc,
	}
}

// OnGuildMemberAdd handles when a new user joins the server.
func (h *BotHandler) OnGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	guildID, err := strconv.ParseInt(m.GuildID, 10, 64)
	if err != nil {
		return
	}

	go func() {
		err = h.welcomeSvc.SendWelcome(context.Background(), s, guildID, m.Member)
		if err != nil {
			fmt.Printf("Bot Error: failed to run welcome flow for %s: %v\n", m.User.Username, err)
		}
	}()
}

// HandleSetupWelcome configures onboarding and welcome greetings via slash command.
func (h *BotHandler) HandleSetupWelcome(s *discordgo.Session, i *discordgo.InteractionCreate, guildID int64, options []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	var channel *discordgo.Channel
	var message string
	var dmMessage *string
	enabled := true

	for _, opt := range options {
		switch opt.Name {
		case "channel":
			channel = opt.ChannelValue(nil)
		case "message":
			message = opt.StringValue()
		case "dm-message":
			dm := opt.StringValue()
			dmMessage = &dm
		case "enabled":
			enabled = opt.BoolValue()
		}
	}

	if channel == nil {
		return "", fmt.Errorf("invalid channel provided")
	}

	chID, err := discordutil.ParseID(channel.ID)
	if err != nil {
		return "", err
	}

	// Create/Update WelcomeConfig
	cfg := &model.WelcomeConfig{
		GuildID:        guildID,
		Enabled:        enabled,
		ChannelID:      &chID,
		ChannelMessage: &message,
		DMMessage:      dmMessage,
	}

	_, err = h.welcomeSvc.Save(context.Background(), cfg)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("✅ Welcome system configured successfully in <#%d> (Enabled: %v)", chID, enabled), nil
}
