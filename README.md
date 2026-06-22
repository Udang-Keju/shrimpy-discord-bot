# Shrimpy 🦐

> A Go-powered Discord server management & help desk bot with a web-based admin dashboard.

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Overview

**Shrimpy** is a general-purpose Discord bot written in Go, designed to serve as the backbone of community server management and structured user support. It combines:

- **Server management** — welcome messages, auto-role assignment, reaction roles
- **Ticket / help-desk system** — powered by Discord buttons, slash commands, and embeds
- **Web admin dashboard** — Next.js UI for admins to manage everything without touching Discord

A single compiled Go binary acts as both the **Discord bot** and the **REST API server** for the dashboard.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Bot runtime** | Go 1.26, [discordgo](https://github.com/bwmarrin/discordgo) |
| **HTTP router** | [go-chi/chi](https://github.com/go-chi/chi) |
| **Database** | PostgreSQL 16 via [pgx/v5](https://github.com/jackc/pgx) |
| **Auth** | Discord OAuth2, JWT ([golang-jwt/jwt](https://github.com/golang-jwt/jwt)) |
| **Dashboard** | Next.js 16, TypeScript, CSS Modules |
| **Deployment** | Railway (bot + API), Supabase (DB), Vercel (dashboard) |

---

## Project Structure

```
shrimpy-discord-bot/
├── cmd/shrimpy/          # Application entrypoint
├── internal/
│   ├── app/             # Vertical business features
│   │   ├── auth/        # Auth feature (model, repository, handler)
│   │   ├── guild/       # Guild configs/roles (model, repository, service, handler, bot)
│   │   ├── welcome/     # Welcome/onboarding (model, repository, service, handler, bot)
│   │   ├── reactionrole/# Reaction roles (model, repository, service, handler, bot)
│   │   └── ticket/      # Ticketing system (model, repository, service, handler, bot, config)
│   ├── pkg/             # Shared utility packages
│   │   ├── apiutil/     # Common HTTP API response & context helpers
│   │   └── discordutil/ # Common Discord types & Snowflake validation helpers
│   ├── bot/             # discordgo session and event dispatching logic
│   │   └── handlers/    # Bot action delegate context (events, commands, buttons, prefix)
│   ├── api/             # REST API server definition and routes
│   │   └── middleware/  # JWT auth, guild permissions, rate limiting
│   └── config/          # Environment variable loader
├── migrations/          # SQL schema migrations
├── dashboard/           # Next.js 16 Web Admin Dashboard
│   ├── app/             # App Router pages and styling
│   ├── lib/             # Theme & API client helpers
│   └── Dockerfile.dev   # Frontend dev container configuration
├── docs/
│   ├── CHANGELOG.md     # Version history
│   └── v1/              # v1.0 specifications
│       ├── PRD.md
│       ├── TECHNICAL_SPEC.md
│       ├── COMMAND_REFERENCE.md
│       └── DESIGN_SYSTEM.md
├── .env.example         # Environment variable template
├── .CLAUDE.md           # AI assistant developer guidelines
├── Dockerfile           # Production multi-stage image
├── docker-compose.yml   # Local dev stack (db, backend, frontend)
└── Makefile             # Build & dev tasks
```

---

## Getting Started

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [PostgreSQL 16+](https://www.postgresql.org/) (or Docker)
- A [Discord Application](https://discord.com/developers/applications) with a bot token

### 1. Clone the repository

```bash
git clone https://github.com/Udang-Keju/shrimpy-discord-bot.git
cd shrimpy-discord-bot
```

### 2. Set up environment variables

```bash
cp .env.example .env
# Edit .env and fill in your values
```

| Variable | Required | Description |
|----------|----------|-------------|
| `DISCORD_TOKEN` | ✅ | Bot token from Discord Developer Portal |
| `DISCORD_CLIENT_ID` | ✅ | Application client ID |
| `DISCORD_CLIENT_SECRET` | ✅ | OAuth2 client secret |
| `DATABASE_URL` | ✅ | PostgreSQL connection string |
| `JWT_SECRET` | ✅ | 32+ byte random string for JWT signing |
| `API_PORT` | ❌ | REST API port (default: `8080`) |
| `ENVIRONMENT` | ❌ | `development` or `production` |

---

## Running Locally

### Option A: Using Docker Compose (Full Stack)
Make sure **Docker Desktop** is running, then run:

```bash
# Clean up and build/start all services (DB, backend, frontend)
make docker-fresh

# Stop the containers
make docker-down
```
- **Go REST API & Bot** runs at `http://localhost:8080`
- **Dashboard Web UI** runs at `http://localhost:3000` (changes are automatically hot-reloaded)

### Option B: Bare Metal Development
If you prefer running services directly on your host machine:

#### Start PostgreSQL database
Ensure PostgreSQL is active on port `5432` with a database named `shrimpy`. Apply migrations:
```bash
make migrate-up
```

#### Run Go Bot & REST API
```bash
go run cmd/shrimpy/main.go
```

#### Run Next.js Dashboard
In a separate terminal, start the Next.js development server:
```bash
cd dashboard
npm install
npm run dev
```

---

## Available Commands

See the full command reference in [docs/v1/COMMAND_REFERENCE.md](docs/v1/COMMAND_REFERENCE.md).

| Category | Commands |
|----------|---------|
| General | `/help`, `/info`, `/ping` |
| Setup | `/setup`, `/setup welcome`, `/setup autorole`, `/set prefix` |
| Ticket Panels | `/ticket panel create/edit/delete`, `/ticket category add/edit/remove` |
| Ticket Actions | `/ticket close`, `/ticket claim`, `/ticket priority`, `/ticket note` |
| Reaction Roles | `/reactionrole create/edit/delete`, `/reactionrole add-role` |
| Staff | `/staff add/remove/list` |
| Admin | `/botinfo`, `/diagnostics`, `/reset config` |

---

## Documentation

All project specifications live under [`docs/`](docs/):

- **[Changelog](docs/CHANGELOG.md)** — Version history across all releases
- **[PRD](docs/v1/PRD.md)** — Product requirements and feature scope
- **[Technical Spec](docs/v1/TECHNICAL_SPEC.md)** — Architecture, DB schema, REST API, and auth design
- **[Command Reference](docs/v1/COMMAND_REFERENCE.md)** — Every command, permission, and parameter
- **[Design System](docs/v1/DESIGN_SYSTEM.md)** — Dashboard colors, typography, and spacing tokens

---

## Deployment

Shrimpy is designed for a three-service split at approximately **$5/month**:

| Service | Platform | Cost |
|---------|----------|------|
| Go Bot + API | [Railway](https://railway.app) Hobby | ~$5/mo |
| PostgreSQL | [Supabase](https://supabase.com) Free | $0 |
| Dashboard | [Vercel](https://vercel.com) Hobby | $0 |

See [TECHNICAL_SPEC.md § Deployment](docs/v1/TECHNICAL_SPEC.md) for full setup instructions.

---

## License

MIT © Udang-Keju
