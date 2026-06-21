package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	api_middleware "github.com/Udang-Keju/shrimpy-discord-bot/internal/api/middleware"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app"
	auth_handler "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth/handler"
	guild_handler "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/handler"
	guild_svc "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/guild/service"
	rr_handler "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/reactionrole/handler"
	ticket_handler "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/ticket/handler"
	welcome_handler "github.com/Udang-Keju/shrimpy-discord-bot/internal/app/welcome/handler"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server coordinates routing, HTTP middlewares, and mounts the dashboard endpoint handlers.
type Server struct {
	router      *chi.Mux
	port        string
	jwtSecret   []byte
	tokenEncKey []byte

	// Handlers
	authHandler         *auth_handler.AuthHandler
	guildHandler        *guild_handler.Handler
	welcomeHandler      *welcome_handler.Handler
	ticketHandler       *ticket_handler.Handler
	reactionRoleHandler *rr_handler.Handler

	// Services/Deps needed by middleware
	guildSvc *guild_svc.GuildService
	dg       *discordgo.Session
}

// NewServer constructs a new REST API Server.
func NewServer(
	port string,
	jwtSecret []byte,
	tokenEncKey []byte,
	modules *app.Modules,
	dg *discordgo.Session,
) *Server {
	return &Server{
		router:              chi.NewRouter(),
		port:                port,
		jwtSecret:           jwtSecret,
		tokenEncKey:         tokenEncKey,
		authHandler:         modules.Auth.Handler,
		guildHandler:        modules.Guild.Handler,
		welcomeHandler:      modules.Welcome.Handler,
		ticketHandler:       modules.Ticket.Handler,
		reactionRoleHandler: modules.ReactionRole.Handler,
		guildSvc:            modules.Guild.Service,
		dg:                  dg,
	}
}


// SetupRoutes registers global middlewares, sets up public/private route groups, and mounts handlers.
func (s *Server) SetupRoutes(allowedOrigins string) {
	// 1. Global Middlewares
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))
	s.router.Use(api_middleware.RateLimitMiddleware)
	s.router.Use(corsMiddleware(allowedOrigins))

	// 2. Register Routes
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	// V1 API Group
	s.router.Route("/api/v1", func(r chi.Router) {
		// Public Auth Callbacks
		r.Post("/auth/callback", s.authHandler.Callback)

		// Authenticated Routes
		r.Group(func(r chi.Router) {
			r.Use(api_middleware.AuthMiddleware(s.jwtSecret))

			r.Get("/auth/me", s.authHandler.Me)
			r.Delete("/auth/logout", s.authHandler.Logout)

			r.Get("/guilds", s.guildHandler.List)

			// Guild Permission Gateways (for server-specific actions)
			r.Route("/guilds/{guildId}", func(r chi.Router) {
				r.Use(api_middleware.GuildPermissionMiddleware(s.guildSvc, s.dg))

				// Guild Settings & Pickers
				r.Get("/", s.guildHandler.GetConfig)
				r.Patch("/", s.guildHandler.UpdateConfig)
				r.Patch("/nickname", s.guildHandler.UpdateNickname)
				r.Get("/discord/channels", s.guildHandler.GetDiscordChannels)
				r.Get("/discord/roles", s.guildHandler.GetDiscordRoles)

				// Welcome onboarding config
				r.Get("/welcome", s.welcomeHandler.Get)
				r.Put("/welcome", s.welcomeHandler.Save)
				r.Delete("/welcome", s.welcomeHandler.Delete)

				// Auto-roles & Staff list config
				r.Get("/auto-roles", s.guildHandler.ListAutoRoles)
				r.Post("/auto-roles", s.guildHandler.AddAutoRole)
				r.Delete("/auto-roles/{roleId}", s.guildHandler.RemoveAutoRole)

				r.Get("/staff-roles", s.guildHandler.ListStaffRoles)
				r.Post("/staff-roles", s.guildHandler.AddStaffRole)
				r.Delete("/staff-roles/{roleId}", s.guildHandler.RemoveStaffRole)

				// Stats
				r.Get("/stats", s.ticketHandler.GetStats)

				// Reaction Role messages
				r.Get("/reaction-roles", s.reactionRoleHandler.ListReactionRoles)
				r.Post("/reaction-roles", s.reactionRoleHandler.CreateReactionRole)
				r.Get("/reaction-roles/{msgId}", s.reactionRoleHandler.GetReactionRole)
				r.Delete("/reaction-roles/{msgId}", s.reactionRoleHandler.DeleteReactionRole)
				r.Post("/reaction-roles/{msgId}/emojis", s.reactionRoleHandler.AddEmojiMapping)
				r.Delete("/reaction-roles/{msgId}/emojis", s.reactionRoleHandler.RemoveEmojiMapping)

				// Ticket Panels & Categories CRUD
				r.Get("/panels", s.ticketHandler.ListPanels)
				r.Post("/panels", s.ticketHandler.CreatePanel)
				r.Patch("/panels/{panelId}", s.ticketHandler.UpdatePanel)
				r.Delete("/panels/{panelId}", s.ticketHandler.DeletePanel)

				r.Get("/panels/{panelId}/categories", s.ticketHandler.ListCategories)
				r.Post("/panels/{panelId}/categories", s.ticketHandler.CreateCategory)
				r.Patch("/panels/{panelId}/categories/{catId}", s.ticketHandler.UpdateCategory)
				r.Delete("/panels/{panelId}/categories/{catId}", s.ticketHandler.DeleteCategory)

				// Ticket Management
				r.Get("/tickets", s.ticketHandler.List)
				r.Get("/tickets/{ticketId}", s.ticketHandler.Get)
				r.Patch("/tickets/{ticketId}", s.ticketHandler.Update)
				r.Post("/tickets/{ticketId}/close", s.ticketHandler.Close)
				r.Post("/tickets/{ticketId}/reopen", s.ticketHandler.Reopen)
				r.Post("/tickets/{ticketId}/archive", s.ticketHandler.Archive)
				r.Get("/tickets/{ticketId}/transcript", s.ticketHandler.DownloadTranscript)
			})
		})
	})
}


// Start launches the HTTP server listening on the configured port.
func (s *Server) Start() error {
	addr := ":" + s.port
	fmt.Printf("API: Starting server on %s\n", addr)
	return http.ListenAndServe(addr, s.router)
}

func corsMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.Split(allowedOrigins, ",")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false
			for _, o := range origins {
				if o == origin || o == "*" {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
