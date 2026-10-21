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

	cmdMu        sync.Mutex
	commandsDone map[string]bool // appID -> slash commands registered since this process started

	ownerMu sync.Mutex
	owners  map[string]ownerEntry // appID -> owners discovered from Discord (see owner.go)
}

// NewRegistry constructs a new session Registry.
func NewRegistry(db *gorm.DB, devGuildID string, registerHandlers func(s *discordgo.Session)) *Registry {
	return &Registry{
		db:               db,
		registerHandlers: registerHandlers,
		devGuildID:       devGuildID,
		sessions:         make(map[string]*discordgo.Session),
		commandsDone:     make(map[string]bool),
		owners:           make(map[string]ownerEntry),
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

// IsConnected reports whether the app's gateway session is open and healthy.
func (r *Registry) IsConnected(appID string) bool {
	r.mu.RLock()
	s, exists := r.sessions[appID]
	r.mu.RUnlock()
	return exists && sessionHealthy(s)
}

// snapshot returns a copy of the app ID -> session map, safe to iterate without the lock.
func (r *Registry) snapshot() map[string]*discordgo.Session {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string]*discordgo.Session, len(r.sessions))
	for id, s := range r.sessions {
		out[id] = s
	}
	return out
}

// StartSession connects an application's bot to the Discord gateway.
//
// If the app already has a session with the same token, that session is restarted in place
// rather than replaced. discordgo runs its own reconnect loop that can reopen a session after
// Close(), so building a second session for the same token could leave two live connections
// (every event would be processed twice).
func (r *Registry) StartSession(
	appID string,
	token string,
) error {
	r.mu.RLock()
	existing, exists := r.sessions[appID]
	r.mu.RUnlock()
	if exists && existing.Token == "Bot "+token {
		fmt.Printf("Registry: Restarting gateway for application %s...\n", appID)
		return r.restartSession(appID, existing)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// The token changed (or this is a new app): retire the old session for good.
	if oldS, exists := r.sessions[appID]; exists {
		oldS.ShouldReconnectOnError = false
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

	// Register before opening so a failed first connect is still retried by the watchdog.
	r.sessions[appID] = dg
	r.cmdMu.Lock()
	delete(r.commandsDone, appID)
	r.cmdMu.Unlock()

	fmt.Printf("Registry: Connecting application %s to Discord Gateway...\n", appID)
	if err := dg.Open(); err != nil {
		return fmt.Errorf("registry: failed to open connection: %w", err)
	}

	r.ensureCommands(appID, dg)
	return nil
}

// restartSession re-opens an existing session's gateway connection in place and makes sure
// its slash commands are registered. Handlers are preserved.
func (r *Registry) restartSession(appID string, s *discordgo.Session) error {
	if err := reopenGateway(s); err != nil {
		return fmt.Errorf("registry: failed to reopen connection: %w", err)
	}
	r.ensureCommands(appID, s)
	return nil
}

// ensureCommands registers the slash commands once per process per app. A failure is logged and
// retried on the next (re)connect.
func (r *Registry) ensureCommands(appID string, dg *discordgo.Session) {
	r.cmdMu.Lock()
	if r.commandsDone[appID] {
		r.cmdMu.Unlock()
		return
	}
	r.commandsDone[appID] = true
	r.cmdMu.Unlock()

	commands := handlers.GetSlashCommands()
	var err error
	if r.devGuildID != "" {
		fmt.Printf("Registry: Registering %d application commands in dev guild: %s for app %s\n", len(commands), r.devGuildID, appID)
		_, err = dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, r.devGuildID, commands)
	} else {
		fmt.Printf("Registry: Registering %d application commands globally for app %s...\n", len(commands), appID)
		_, err = dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, "", commands)
	}
	if err != nil {
		fmt.Printf("Warning: failed to register application commands for app %s: %v\n", appID, err)
		r.cmdMu.Lock()
		delete(r.commandsDone, appID)
		r.cmdMu.Unlock()
	}
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
	s.ShouldReconnectOnError = false
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
		s.ShouldReconnectOnError = false
		_ = s.Close()
	}
	r.sessions = make(map[string]*discordgo.Session)
}
