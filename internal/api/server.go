package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	api_middleware "github.com/Udang-Keju/shrimpy-discord-bot/internal/api/middleware"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/api/handlers"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/repository"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/service"
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

	// Repositories needed directly by some handlers
	categoryRepo *repository.CategoryRepo
	userRepo     *repository.UserRepo

	// Services
	guildSvc        *service.GuildService
	welcomeSvc      *service.WelcomeService
	autoRoleSvc     *service.AutoRoleService
	reactionRoleSvc *service.ReactionRoleService
	ticketSvc       *service.TicketService
	transcriptSvc   *service.TranscriptService

	// Discord Session
	dg *discordgo.Session
}

// NewServer constructs a new REST API Server.
func NewServer(
	port string,
	jwtSecret []byte,
	tokenEncKey []byte,
	categoryRepo *repository.CategoryRepo,
	userRepo *repository.UserRepo,
	guildSvc *service.GuildService,
	welcomeSvc *service.WelcomeService,
	autoRoleSvc *service.AutoRoleService,
	reactionRoleSvc *service.ReactionRoleService,
	ticketSvc *service.TicketService,
	transcriptSvc *service.TranscriptService,
	dg *discordgo.Session,
) *Server {
	return &Server{
		router:          chi.NewRouter(),
		port:            port,
		jwtSecret:       jwtSecret,
		tokenEncKey:     tokenEncKey,
		categoryRepo:    categoryRepo,
		userRepo:        userRepo,
		guildSvc:        guildSvc,
		welcomeSvc:      welcomeSvc,
		autoRoleSvc:     autoRoleSvc,
		reactionRoleSvc: reactionRoleSvc,
		ticketSvc:       ticketSvc,
		transcriptSvc:   transcriptSvc,
		dg:              dg,
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

	// 2. Instantiate Handlers
	authHandler := handlers.NewAuthHandler(s.userRepo, s.jwtSecret, s.tokenEncKey)
	guildHandler := handlers.NewGuildHandler(s.guildSvc, s.dg)
	welcomeHandler := handlers.NewWelcomeHandler(s.welcomeSvc)
	categoryHandler := handlers.NewCategoryHandler(s.categoryRepo, s.dg)
	ticketHandler := handlers.NewTicketHandler(s.ticketSvc, s.categoryRepo, s.transcriptSvc, s.dg)
	autoRolesHandler := handlers.NewAutoRolesHandler(s.guildSvc, s.reactionRoleSvc, s.dg)
	statsHandler := handlers.NewStatsHandler(s.ticketSvc, s.categoryRepo, s.dg)

	// 3. Register Routes
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	// V1 API Group
	s.router.Route("/api/v1", func(r chi.Router) {
		// Public Auth Callbacks
		r.Post("/auth/callback", authHandler.Callback)

		// Authenticated Routes
		r.Group(func(r chi.Router) {
			r.Use(api_middleware.AuthMiddleware(s.jwtSecret))

			r.Get("/auth/me", authHandler.Me)
			r.Delete("/auth/logout", authHandler.Logout)

			r.Get("/guilds", guildHandler.List)

			// Guild Permission Gateways (for server-specific actions)
			r.Route("/guilds/{guildId}", func(r chi.Router) {
				r.Use(api_middleware.GuildPermissionMiddleware(s.guildSvc, s.dg))

				// Guild Settings & Pickers
				r.Get("/", guildHandler.GetConfig)
				r.Patch("/", guildHandler.UpdateConfig)
				r.Patch("/nickname", guildHandler.UpdateNickname)
				r.Get("/discord/channels", guildHandler.GetDiscordChannels)
				r.Get("/discord/roles", guildHandler.GetDiscordRoles)

				// Welcome onboarding config
				r.Get("/welcome", welcomeHandler.Get)
				r.Put("/welcome", welcomeHandler.Save)
				r.Delete("/welcome", welcomeHandler.Delete)

				// Auto-roles & Staff list config
				r.Get("/auto-roles", autoRolesHandler.ListAutoRoles)
				r.Post("/auto-roles", autoRolesHandler.AddAutoRole)
				r.Delete("/auto-roles/{roleId}", autoRolesHandler.RemoveAutoRole)

				r.Get("/staff-roles", autoRolesHandler.ListStaffRoles)
				r.Post("/staff-roles", autoRolesHandler.AddStaffRole)
				r.Delete("/staff-roles/{roleId}", autoRolesHandler.RemoveStaffRole)

				// Stats
				r.Get("/stats", statsHandler.GetStats)

				// Reaction Role messages
				r.Get("/reaction-roles", autoRolesHandler.ListReactionRoles)
				r.Post("/reaction-roles", autoRolesHandler.CreateReactionRole)
				r.Get("/reaction-roles/{msgId}", autoRolesHandler.GetReactionRole)
				r.Delete("/reaction-roles/{msgId}", autoRolesHandler.DeleteReactionRole)
				r.Post("/reaction-roles/{msgId}/emojis", autoRolesHandler.AddEmojiMapping)
				r.Delete("/reaction-roles/{msgId}/emojis", autoRolesHandler.RemoveEmojiMapping)

				// Ticket Panels & Categories CRUD
				r.Get("/panels", categoryHandler.ListPanels)
				r.Post("/panels", categoryHandler.CreatePanel)
				r.Patch("/panels/{panelId}", categoryHandler.UpdatePanel)
				r.Delete("/panels/{panelId}", categoryHandler.DeletePanel)

				r.Get("/panels/{panelId}/categories", categoryHandler.ListCategories)
				r.Post("/panels/{panelId}/categories", categoryHandler.CreateCategory)
				r.Patch("/panels/{panelId}/categories/{catId}", categoryHandler.UpdateCategory)
				r.Delete("/panels/{panelId}/categories/{catId}", categoryHandler.DeleteCategory)

				// Ticket Management
				r.Get("/tickets", ticketHandler.List)
				r.Get("/tickets/{ticketId}", ticketHandler.Get)
				r.Patch("/tickets/{ticketId}", ticketHandler.Update)
				r.Post("/tickets/{ticketId}/close", ticketHandler.Close)
				r.Post("/tickets/{ticketId}/reopen", ticketHandler.Reopen)
				r.Post("/tickets/{ticketId}/archive", ticketHandler.Archive)
				r.Get("/tickets/{ticketId}/transcript", ticketHandler.DownloadTranscript)
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
