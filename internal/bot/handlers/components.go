package handlers

import (
	"github.com/bwmarrin/discordgo"
)

// HandleComponentInteraction routes button clicks and dropdown select menu choices.
func (ctx *HandlerContext) HandleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx.TicketBot.HandleComponentInteraction(s, i)
}
