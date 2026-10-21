# Go style guide

How to write backend Go code in this repo (`internal/` and `cmd/`). Architecture (vertical feature modules, `Build()` wiring, the multi-bot registry) is in the root [CLAUDE.md](../CLAUDE.md); this file covers the code itself.

## Layout and layering

- One feature per `internal/app/<feature>/` with `model/`, `repository/`, `service/`, `handler/`, optional `bot/`, and `<feature>.go` exposing `Build()` and `Models()`.
- Dependencies point downward only: handler/bot → service → repository. A handler never touches GORM; a repository never calls Discord.
- Cross-module dependencies go through small interfaces declared by the **consumer** (e.g. `service.GuildConfigProvider`) or through `internal/pkg/`, never by importing another feature's internals.
- Business rules (validation, defaults, Discord calls) live in the service. Handlers only decode, call, and map errors to HTTP status.

## Formatting and naming

- `gofmt` clean. Run `gofmt -l .` before committing.
- Exported identifiers have a doc comment starting with the identifier name.
- Comments explain *why* (rate limits, Discord quirks, DST), not what the code already says. Match the surrounding comment density.
- Constructors are `NewXxx`. Name a type after its layer and feature (`TranslationRepo`, `TicketService`).
- Discord snowflakes are `int64` in Go, models, and the DB (`BIGINT`), and **strings in JSON**. Convert with `discordutil.ParseID` / `FormatID`.

## Errors

- Wrap with context: `fmt.Errorf("load timezone %q: %w", tz, err)`. Lowercase, no trailing period.
- Each repository package defines `ErrNotFound`; translate `gorm.ErrRecordNotFound` to it. Check with `errors.Is`.
- Services return sentinel or typed errors for caller mistakes (invalid input, missing channel) so handlers can answer 4xx. Anything else is a 500 with a generic message; never leak internals to the client.
- Background workers must not crash or block on one bad row: log with `fmt.Printf`, continue with the next item.

## Context

- `ctx context.Context` is the first parameter of anything that does I/O, and is passed through to GORM (`db.WithContext(ctx)`) and Discord calls.
- Long-running loops take the shutdown context and exit on `ctx.Done()` (see the ticket `Scheduler`).
- Calls that can block on Discord's rate limiter get a bounded timeout context.

## Database

- GORM for ordinary CRUD, raw SQL for complex queries. Always scope by `guild_id` in queries that take a client-supplied ID.
- Schema is changed only through SQL files in `migrations/` (`NNN_name.up.sql` + `.down.sql`). GORM does **not** auto-migrate, so keep model tags and the DDL in sync by hand.
- Follow existing DDL conventions: `guild_id BIGINT REFERENCES guilds(guild_id) ON DELETE CASCADE`, UUID primary keys for child rows, `created_at`/`updated_at TIMESTAMPTZ` with the `trigger_set_updated_at()` trigger.

## HTTP handlers

- Respond only through `apiutil.WriteJSON` and `apiutil.WriteError(w, status, "CODE", "message")`.
- Use DTO structs for responses; never serialize secrets or ciphertext.
- The guild is read from `chi.URLParam(r, "guildId")`. Permission checks are middleware's job, but re-validate anything sensitive server-side.

## Discord

- Always get a session via `provider.GetSessionForGuild(ctx, guildID)`. Never hold a global session; guilds can belong to different bot apps.
- Respect Discord rate limits. Channel renames are limited to about 2 per 10 minutes per channel: skip no-op edits and don't poll faster than needed.
- Use `discordutil.EmbedMedia` for embed media.

## Testing

- **Write table-driven tests** whenever there is more than one case: a `tests := []struct{ name string; ... }` slice, looped with `t.Run(tt.name, ...)`. Put expected errors in the table (`wantErr bool`) rather than splitting into separate test functions.
- Separate test functions only when setup or assertions genuinely differ between cases.
- Tests are colocated (`service_test.go` next to `service.go`). Use `testify/assert`, and `testify/mock` for mocks.
- Prefer pure functions for logic that needs many cases (formatting, validation) so it can be tested without mocks.
- Make time-dependent code testable by passing `now time.Time` in instead of calling `time.Now()` inside.
- Run `go test ./...` (or `make test`) and `go build ./...` before committing.

## Commits

Conventional Commits with a scope naming the area, small and ordered foundations first, each leaving the build green (see the global git instructions).
