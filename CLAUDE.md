# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Shrimpy** is a multi-bot Discord server management and help desk system written in Go with a Next.js web dashboard. A single Go binary serves as both the Discord bot runtime and REST API server. The system supports managing multiple Discord bot applications from a single backend, with credentials stored encrypted in PostgreSQL.

## Development Commands

### Go Backend

```bash
# Build the Go binary
make build

# Run locally (requires PostgreSQL and .env configured)
make run

# Run all tests
make test

# Clean build artifacts
make clean
```

### Database Migrations

Migrations are stored in `migrations/` and should be run against PostgreSQL directly. Use the direct connection (port 5432), not PgBouncer's transaction mode (port 6543), as migrations use `SET` statements.

### Docker Development

```bash
# Start full stack (PostgreSQL, Go backend, Next.js frontend)
make docker-fresh

# Stop containers
make docker-down

# Clean containers and volumes
make docker-clean
```

After running `docker-fresh`, services are available at:
- **Go Backend + Bot**: http://localhost:8080
- **Next.js Dashboard**: http://localhost:3000

### Next.js Dashboard

```bash
cd dashboard
npm install
npm run dev      # Development server
npm run build    # Production build
npm run lint     # ESLint
```

## Architecture Patterns

### Vertical Feature Modules

The codebase uses **vertical slice architecture**. Each feature lives under `internal/app/<feature>/` and is subdivided into layers:

```
internal/app/<feature>/
├── model/model.go           # GORM models (database schema)
├── repository/repository.go # Data access layer
├── service/service.go       # Business logic
├── handler/handler.go       # REST API endpoints
├── bot/bot.go              # Discord bot event handlers
└── <feature>.go            # Module builder (wires dependencies)
```

Example: `internal/app/ticket/`, `internal/app/guild/`, `internal/app/welcome/`

**Key principle**: Each module is self-contained. Cross-feature dependencies go through the module's exported `Module` struct or via interfaces in `internal/pkg/`.

### Module Builder Pattern

Every feature has a `Build()` function that wires all layers together and returns a `Module` struct:

```go
// Example: internal/app/guild/guild.go
func Build(db *gorm.DB, cacheTTL time.Duration, provider discordutil.DiscordSessionProvider) *Module {
    repo := repository.NewGuildRepo(db)
    svc := service.NewGuildService(repo, guildCache)
    h := handler.NewHandler(svc, provider)
    b := bot.NewBotHandler(svc)
    return &Module{Repo: repo, Service: svc, Handler: h, Bot: b}
}
```

All modules are instantiated in `internal/app/app.go` and passed to both the REST API server and bot handlers.

### Multi-Bot Registry Pattern

The `bot.Registry` (internal/bot/registry.go) manages multiple concurrent Discord bot sessions:

- Each bot application is stored in the `discord_apps` table with encrypted credentials
- The registry maps `app_id` → `*discordgo.Session`
- Services retrieve the correct session via `registry.GetSessionForGuild(ctx, guildID)`
- Sessions can be dynamically started/stopped/reconnected without restarting the backend

**Important**: The registry is both a `DiscordSessionProvider` (for services to get sessions) and a `BotSessionController` (for the settings service to start/stop sessions).

### Dependency Injection

Dependencies flow from `cmd/shrimpy/main.go` → `internal/app/app.go` (module builder) → individual feature modules. Circular dependencies are avoided by:

1. Using interfaces defined in `internal/pkg/` (e.g., `discordutil.DiscordSessionProvider`)
2. Late-binding handlers via `registry.SetHandlers()` after all modules are built

### Caching Strategy

- **Guild configs**: Cached in-memory with configurable TTL (default 5 minutes) to avoid DB hits on every Discord event
- **Permission checks**: 5-minute TTL on Discord API permission lookups to avoid rate limits
- Cache implementation: `internal/app/guild/repository/cache.go` (generic, reusable)

### Bot Event Flow

```
Discord Gateway
  ↓
internal/bot/handlers/events.go (OnGuildMemberAdd, OnInteractionCreate, etc.)
  ↓
internal/bot/handlers/components.go (button routing) or commands.go (slash command routing)
  ↓
Feature bot handler (internal/app/<feature>/bot/bot.go)
  ↓
Feature service (internal/app/<feature>/service/service.go)
  ↓
Feature repository (internal/app/<feature>/repository/repository.go)
  ↓
PostgreSQL
```

### REST API Flow

```
Dashboard HTTP Request
  ↓
internal/api/server.go (chi router)
  ↓
internal/api/middleware/*.go (auth, guild permission check, rate limit)
  ↓
Feature handler (internal/app/<feature>/handler/handler.go)
  ↓
Feature service
  ↓
Feature repository
  ↓
PostgreSQL
```

## Key Files to Understand

| File | Purpose |
|------|---------|
| `cmd/shrimpy/main.go` | Application entry point; wires all dependencies |
| `internal/app/app.go` | Central module builder; instantiates all features |
| `internal/bot/registry.go` | Multi-bot session manager |
| `internal/bot/handlers/events.go` | Discord Gateway event dispatcher |
| `internal/bot/handlers/commands.go` | Slash command registration and routing |
| `internal/api/server.go` | REST API router and middleware setup |
| `internal/api/middleware/auth.go` | JWT validation middleware |
| `internal/api/middleware/guild.go` | Guild permission check middleware (uses Discord API + DB roles) |

## Database Schema

- All tables use `BIGINT` for Discord snowflake IDs (guild_id, user_id, channel_id, role_id, message_id)
- UUIDs (`gen_random_uuid()`) are used for internal record IDs (tickets, panels, categories)
- Sensitive credentials (bot tokens, OAuth secrets) are encrypted using AES-256-GCM before storage
- Schema is managed via SQL migrations in `migrations/`

**Important tables**:
- `discord_apps`: Multi-bot application credentials (encrypted)
- `guilds`: Per-server configuration (prefix, language, nickname, log channel)
- `ticket_panels` / `ticket_categories` / `tickets`: Ticketing system schema
- `welcome_config`: Welcome message settings
- `auto_roles` / `staff_roles`: Role assignment configs
- `reaction_role_messages` / `reaction_role_emojis`: Reaction role system

## Authentication & Authorization

### OAuth2 Flow (Dashboard Login)

1. User clicks "Login with Discord" → redirected to Go backend `/api/v1/auth/login`
2. Backend redirects to Discord OAuth2 authorize page
3. Discord redirects to `/api/v1/auth/callback?code=...` with authorization code
4. Backend exchanges code for access token, fetches user + guilds from Discord API
5. Backend signs a JWT with claims: `{sub, managed_guilds[], exp, jti}` and sets HttpOnly cookie
6. Dashboard requests are authenticated via JWT cookie

### Permission Checks

Dashboard access to a guild is granted if **either**:
1. User has Discord `ADMINISTRATOR` or `MANAGE_GUILD` permission on the guild (checked via Discord API)
2. User has a role listed in the guild's `staff_roles` table (checked via DB + Discord API)

Middleware: `internal/api/middleware/guild.go`

## Testing Approach

Tests are colocated with implementation files:
- `internal/app/<feature>/repository/repository_test.go`
- `internal/app/<feature>/service/service_test.go`

Run all tests: `make test` or `go test ./...`

Use `github.com/stretchr/testify/assert` for assertions and `github.com/stretchr/testify/mock` for mocking.

## Common Patterns

### Adding a New Feature Module

1. Create `internal/app/<feature>/` directory
2. Define GORM models in `model/model.go`
3. Implement repository layer (DB queries)
4. Implement service layer (business logic)
5. Implement handler layer (REST API endpoints)
6. Implement bot layer (Discord event handlers)
7. Create module builder in `<feature>.go` with `Build()` and `Models()` functions
8. Wire into `internal/app/app.go`
9. Register REST endpoints in `internal/api/server.go`
10. Register bot handlers in `internal/bot/handlers/events.go` or `commands.go`

### Working with Discord Sessions

Always retrieve the correct session via the provider:

```go
// In a service that needs to send Discord messages
session, err := s.provider.GetSessionForGuild(ctx, guildID)
if err != nil {
    return err
}
_, err = session.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{...})
```

### Encrypted Credentials

Use `internal/pkg/crypto/aes_256_gcm.go` for encrypting/decrypting sensitive data. The encryption key is derived from the `TOKEN_ENCRYPTION_KEY` environment variable (64 hex characters = 32 bytes).

Example: `internal/app/settings/service/service.go` encrypts bot tokens before storing in `discord_apps` table.

## Environment Variables

**First-boot seeding**: If `discord_apps` table is empty on startup, the backend reads `DISCORD_TOKEN`, `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, and `DISCORD_REDIRECT_URI` from env vars and creates a "First Boot App" entry. After first boot, these env vars can be removed and credentials managed via `/api/v1/admin/apps` endpoints.

**Required always**:
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Random 32+ byte string for signing JWTs
- `TOKEN_ENCRYPTION_KEY`: 64-character hex string (32 bytes for AES-256-GCM)

**Optional**:
- `API_PORT` or `PORT`: HTTP server port (default: 8080)
- `DEV_GUILD_ID`: Discord guild ID for instant slash command registration during development
- `OWNER_DISCORD_ID`: Discord user ID granted unconditional admin access to `/admin/apps` endpoints
- `CORS_ALLOWED_ORIGINS`: Comma-separated allowed origins for CORS (default: http://localhost:3000)
- `CACHE_TTL_SECONDS`: Guild config cache TTL (default: 300)
- `ENVIRONMENT`: `development` or `production`

## Deployment

Production deployment uses a three-service split:
- **Railway** (Hobby): Go binary (bot + API) — always-on for Discord Gateway connection
- **Supabase** (Free): PostgreSQL database with PgBouncer connection pooling
- **Vercel** (Free): Next.js dashboard frontend

Use Supabase's **Transaction mode** connection string (port 6543) with `?pgbouncer=true` for runtime. Run migrations against the direct connection (port 5432).

## Code Style Conventions

- Go: Follow standard Go conventions; use `gofmt`
- Error handling: Always wrap errors with context using `fmt.Errorf("...: %w", err)`
- Logging: Use `fmt.Printf` for structured logs (future: migrate to `slog`)
- Database: Use GORM for ORM; raw SQL for complex queries
- REST responses: Use `internal/pkg/apiutil` helpers (`WriteJSON`, `WriteError`)
- Discord embeds: Use `discordutil.EmbedMedia` struct for consistent embed media handling

## Important Notes

- Never run multiple bot instances with the same Discord token (causes duplicate event processing)
- Always re-validate permissions server-side for sensitive operations (don't trust client)
- Use the registry's `GetSessionForGuild()` to route Discord API calls to the correct bot session
- Cache invalidation: When staff roles or guild config changes, invalidate relevant cache keys immediately
- Migrations: Always run against direct Postgres connection (port 5432), not PgBouncer
