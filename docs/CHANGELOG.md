# Changelog

All notable changes to the **Shrimpy ??** Discord Bot project documentation and specifications will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]
### Added
- **User Journey & UX Flow Specification ([USER_JOURNEY.md](file:///d:/Pesronal/Projects/Discord%20Bot/docs/v1/USER_JOURNEY.md))**: Defines the end-to-end frontend journey (login → server select → bot invite → guided setup → configure → operate) for all three personas, a proposed information architecture, a gap analysis of the current dashboard, a prioritized improvement backlog, visual/interaction consistency standards, and a phased implementation roadmap.
- **Annotated wireframes (USER_JOURNEY.md Appendix A)**: Token-annotated low-fidelity layouts for every primary screen — A.1 `/servers` selection (populated/empty/loading), A.2 `/dashboard/[guildId]` Overview (app shell + first-run Setup Checklist + configured dashboard), A.3 `/tickets/[ticketId]` Ticket Detail, A.4 role-aware Staff (Level 2) sidebar, A.5 `/welcome` (template-variable picker, test-send, live card preview, folded-in auto-roles), A.6 `/panels` (multi-button/select-menu, per-category embed, thread-vs-channel, multiple support roles), A.7 `/roles` (full emoji picker, live preview, automated role-height health check), A.8 `/settings` + `/settings/access` split (adds language + auto-close, plain-language access copy), and A.9 shared component patterns (Toast, SaveBar, EmptyState, Status/Priority badges, PageLoader, ServerSwitcher, DiscordPreview). Every element is mapped to a [Design System](file:///d:/Pesronal/Projects/Discord%20Bot/docs/v1/DESIGN_SYSTEM.md) token to keep implementation on-brand.
- **Multi-Bot Support Design**: Drafted database and runtime architecture specs for supporting multiple Discord Applications simultaneously, mapped to Guilds (Option 1).
- **`discord_apps` schema migration**: Planned SQL migration to replace `bot_settings` singleton table with a multi-tenant `discord_apps` table and link `guilds` via `discord_app_id` foreign key.
- **REST API Endpoints for Apps**: Designed CRUD routes at `/api/v1/admin/apps` replacing single-bot `/api/v1/admin/settings`.

### Changed
- **Dashboard information architecture — dedicated server selection**: Server selection is now its own **`/servers`** page, separate from the per-server dashboard (`/dashboard/[guildId]/…`); bare `/dashboard` redirects to `/servers`. The `/servers` page splits guilds into "Your servers" (bot active) vs "Add Shrimpy to a server" (invitable). Propagated to [TECHNICAL_SPEC.md](file:///d:/Pesronal/Projects/Discord%20Bot/docs/v1/TECHNICAL_SPEC.md) (OAuth post-login redirect target, auth flow diagram, Next.js directory structure, `DASHBOARD_URL` description) and the root [CLAUDE.md](file:///d:/Pesronal/Projects/Discord%20Bot/CLAUDE.md) (frontend-IA pointer). Defined in [USER_JOURNEY.md](file:///d:/Pesronal/Projects/Discord%20Bot/docs/v1/USER_JOURNEY.md) §5 & §7.3.
- Renamed bot display name from **Shrimp** to **Shrimpy** across all documentation.

---

## [1.1.0] - 2026-06-21
### Added
- **`bot_settings` singleton table** (`migrations/002_bot_settings.up.sql`): Stores all four Discord credentials (`DISCORD_TOKEN`, `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, `DISCORD_REDIRECT_URI`) encrypted at rest with AES-256-GCM. Eliminates the need to update Railway environment variables when credentials change.
- **`internal/app/settings/` vertical module**: Full model / repository / service / handler stack following the established vertical module pattern.
- **Admin REST API endpoints** (`GET`, `PUT /api/v1/admin/settings`, `POST /api/v1/admin/settings/reconnect`): Dashboard-facing CRUD for bot credentials. Token and client secret are always masked in GET responses.
- **Live bot reconnect** (`bot.Reconnect()`): Updating the token via `PUT /api/v1/admin/settings` automatically closes and reopens the Discord Gateway connection on the same `*discordgo.Session` pointer — no restart required, all handlers preserved.
- **`internal/api/middleware/admin.go`**: Admin-only middleware that checks `managed_guilds` JWT claim or `OWNER_DISCORD_ID` env var override.
- **`OWNER_DISCORD_ID` env var**: Optional; grants a specific Discord user unconditional access to the admin settings endpoints.
- **First-boot seeding**: On startup, if `bot_settings` is empty and `DISCORD_*` env vars are set, credentials are automatically encrypted and persisted to the DB. After first boot, env vars can be safely removed from Railway.

### Changed
- **`internal/config/config.go`**: All four `DISCORD_*` environment variables changed from required (`mustGetEnv`) to optional (`getEnv`). Added `HasDiscordSeed()` helper and `getEnvFallback()` for Railway's `PORT` precedence.
- **`cmd/shrimpy/main.go`**: Startup sequence reordered — `bot_settings` table migrated first, credentials seeded/loaded from DB, Discord session constructed from DB token, then remaining modules built.
- **`internal/app/auth/handler/handler.go`**: Auth callback now fetches OAuth2 credentials from the settings service (30s cached DB lookup) instead of from startup config.
- **`internal/api/server.go`**: Admin route group added under `/api/v1/admin/`.
- **`railway.toml`**: `healthcheckTimeout` increased from 10s to 30s to accommodate cold-start DB connection + migration time.
- **Technical Spec** ([TECHNICAL_SPEC.md](file:///d:/Pesronal/Projects/Discord Bot/docs/v1/TECHNICAL_SPEC.md)): Updated sections 3 (schema), 4 (API), 8 (directory structure), 9 (config), 10.2 (Railway vars), 10.4 (Vercel vars), 10.5 (docker-compose).

### Fixed
- **Railway health check failure**: HTTP server now starts before the Discord Gateway connection so `/health` is reachable during Railway's health check window.
- **Railway `PORT` env var**: Config now reads Railway's injected `PORT` variable first, falling back to `API_PORT`, then `8080`.

---

## [1.0.0] - 2026-06-21
### Added
- **Initial MVP Specifications ([v1](file:///d:/Pesronal/Projects/Discord%20Bot/docs/v1/))**:
  - [PRD](file:///d:/Pesronal/Projects/Discord%20Bot/docs/v1/PRD.md): Detailed product requirements for ticket systems, auto-roles, and dashboards.
  - [Technical Spec](file:///d:/Pesronal/Projects/Discord%20Bot/docs/v1/TECHNICAL_SPEC.md): Database schemas (PostgreSQL), REST API design, OAuth2 flow, and directory structure.
  - [Command Reference](file:///d:/Pesronal/Projects/Discord%20Bot/docs/v1/COMMAND_REFERENCE.md): Detailed parameters and permission levels for all slash commands.
  - [Design System](file:///d:/Pesronal/Projects/Discord%20Bot/docs/v1/DESIGN_SYSTEM.md): Light/Dark UI color tokens, typography, and spacing scales.
