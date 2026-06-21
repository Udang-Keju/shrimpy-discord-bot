package bot

import (
	"context"
	"fmt"
	"sync"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/bot/handlers"
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

// Registry manages the lifecycle and routing of multiple concurrent Discord bot sessions.
type Registry struct {
	db               *gorm.DB
	registerHandlers func(s *discordgo.Session)
	devGuildID       string

	mu       sync.RWMutex
	sessions map[string]*discordgo.Session // Key: app UUID string
}

// NewRegistry constructs a new session Registry.
func NewRegistry(db *gorm.DB, devGuildID string, registerHandlers func(s *discordgo.Session)) *Registry {
	return &Registry{
		db:               db,
		registerHandlers: registerHandlers,
		devGuildID:       devGuildID,
		sessions:         make(map[string]*discordgo.Session),
	}
}

// SetHandlers sets the registration callback post-instantiation.
func (r *Registry) SetHandlers(registerHandlers func(s *discordgo.Session)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registerHandlers = registerHandlers
}

// GetSession retrieves an active session by its database app ID (UUID).
func (r *Registry) GetSession(appID string) (*discordgo.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, exists := r.sessions[appID]
	if !exists {
		return nil, fmt.Errorf("registry: no active bot session found for app ID %s", appID)
	}
	return s, nil
}

// GetSessionForGuild retrieves the session managing the specified guild.
func (r *Registry) GetSessionForGuild(ctx context.Context, guildID int64) (*discordgo.Session, error) {
	var appID string
	err := r.db.WithContext(ctx).Table("guilds").
		Where("guild_id = ?", guildID).
		Pluck("discord_app_id", &appID).Error

	if err != nil {
		return nil, fmt.Errorf("registry: failed to look up app for guild %d: %w", guildID, err)
	}

	if appID == "" {
		return nil, fmt.Errorf("registry: guild %d is not associated with any Discord application", guildID)
	}

	return r.GetSession(appID)
}

// IsBotInGuild returns true if any active bot session is currently in the specified guild.
func (r *Registry) IsBotInGuild(guildID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, s := range r.sessions {
		if _, err := s.State.Guild(guildID); err == nil {
			return true
		}
	}
	return false
}

// StartSession initializes and opens a new gateway connection.
func (r *Registry) StartSession(
	appID string,
	token string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Stop existing session if active
	if oldS, exists := r.sessions[appID]; exists {
		_ = oldS.Close()
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("registry: failed to construct session: %w", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentMessageContent

	if r.registerHandlers != nil {
		r.registerHandlers(dg)
	}

	fmt.Printf("Registry: Connecting application %s to Discord Gateway...\n", appID)
	if err := dg.Open(); err != nil {
		return fmt.Errorf("registry: failed to open connection: %w", err)
	}

	// Register Application Commands (Slash Commands)
	commands := handlers.GetSlashCommands()
	if r.devGuildID != "" {
		fmt.Printf("Registry: Registering %d application commands in dev guild: %s for app %s\n", len(commands), r.devGuildID, appID)
		_, err = dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, r.devGuildID, commands)
	} else {
		fmt.Printf("Registry: Registering %d application commands globally for app %s...\n", len(commands), appID)
		_, err = dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, "", commands)
	}
	if err != nil {
		fmt.Printf("Warning: failed to register application commands for app %s: %v\n", appID, err)
	}

	r.sessions[appID] = dg
	return nil
}

// StopSession closes the session gateway connection and removes it from the registry.
func (r *Registry) StopSession(appID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, exists := r.sessions[appID]
	if !exists {
		return nil
	}

	fmt.Printf("Registry: Closing connection for application %s...\n", appID)
	err := s.Close()
	delete(r.sessions, appID)
	return err
}

// Clear closes all active bot connections.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, s := range r.sessions {
		fmt.Printf("Registry: Shutting down connection %s...\n", id)
		_ = s.Close()
	}
	r.sessions = make(map[string]*discordgo.Session)
}
