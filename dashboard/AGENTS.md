<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->

# Specification docs (source of truth)

Before building or changing dashboard UI, read these (paths relative to this `dashboard/` directory):

- **[../docs/v1/USER_JOURNEY.md](../docs/v1/USER_JOURNEY.md)** — the **primary spec for dashboard work**: information architecture, per-persona user journeys, screen specs, routing, and UI/interaction consistency standards.
- **[../docs/v1/DESIGN_SYSTEM.md](../docs/v1/DESIGN_SYSTEM.md)** — color tokens (dark/light), typography, spacing, component tokens, theming. **Every screen must render from these tokens** — no hardcoded hex.
- **[../docs/v1/PRD.md](../docs/v1/PRD.md)** — product scope, personas, and user stories the UI must satisfy.
- **[../docs/v1/TECHNICAL_SPEC.md](../docs/v1/TECHNICAL_SPEC.md)** — REST API endpoints (§4) and the auth/permission model (§7) the dashboard calls.

**Routing reminder** (USER_JOURNEY §5): `/servers` = pick a server · `/dashboard/[guildId]/…` = manage one server · bare `/dashboard` redirects to `/servers`.
