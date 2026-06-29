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

## Async mutations & toast feedback

Every create/update/delete that calls `ShrimpyAPI` (`lib/api.ts`) must give the user a toast via `useToast()` (`hooks/useToast.ts` → `components/Toast/ToastProvider.tsx`):

- Always show a success toast on completion and an error toast on failure — don't rely on `console.error` alone.
- For requests with a noticeable round-trip (anything that causes the bot to post/edit a Discord channel message, e.g. creating a ticket panel or category), show a `"loading"` toast immediately via `showToast(message, "loading")`, then morph it into the result with `updateToast(id, message, "success" | "error")` instead of stacking a second toast. Fast mutations (role add/remove, simple field updates) don't need the loading step — a single success/error toast on completion is enough.

Forms that create a new resource (panel, category, etc.) must reset to their default values **synchronously at submit time**, not after the request resolves, so the user can immediately start the next one instead of waiting on the network. Guard against a fast double-click re-submitting the same (now-reset) values by checking-and-setting a `useRef` boolean right at the top of the handler, clearing it again immediately after the snapshot+reset (not after the awaited request) — see `handleCreatePanel`/`handleCreateCategory` in `app/dashboard/[guildId]/panels/page.tsx`.

Because the form resets immediately, any state update inside the `.then`/success branch that targets a list scoped to a "currently selected" parent (e.g. categories under a panel) must re-check that the parent is still selected before writing — otherwise a slow request can splice data into the wrong parent's view if the user has since navigated elsewhere. Track the live selection in a `useRef` kept in sync via a `useEffect`, and compare against it at completion time rather than trusting values captured in the closure at request-start (see `selectedPanelIdRef` in the same file).
