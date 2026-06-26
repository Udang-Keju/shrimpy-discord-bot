# User Journey & UX Flow Specification
## Project: **Shrimpy** 🦐 — Frontend Experience, Information Architecture & Journey Design

> **Version**: 1.0.0-draft
> **Status**: In Review
> **Last Updated**: 2026-06-26
> **Applies To**: Next.js Web Dashboard (`dashboard/`)
> **Companion docs**: [PRD](./PRD.md) · [Design System](./DESIGN_SYSTEM.md) · [Technical Spec](./TECHNICAL_SPEC.md)

---

## Table of Contents

1. [Purpose & Scope](#1-purpose--scope)
2. [Relationship to Existing Docs](#2-relationship-to-existing-docs)
3. [Personas & Surface Matrix](#3-personas--surface-matrix)
4. [Experience Principles](#4-experience-principles)
5. [Information Architecture — Current vs Proposed](#5-information-architecture--current-vs-proposed)
6. [Master Journey Map](#6-master-journey-map)
7. [Admin Journey (Primary)](#7-admin-journey-primary)
8. [Support Staff Journey](#8-support-staff-journey)
9. [Member Journey (Discord-side, driven by the dashboard)](#9-member-journey-discord-side-driven-by-the-dashboard)
10. [Gap Analysis — Current Build vs Intended Journey](#10-gap-analysis--current-build-vs-intended-journey)
11. [Improvement Ideas (Prioritized Backlog)](#11-improvement-ideas-prioritized-backlog)
12. [Visual & Interaction Consistency Standards](#12-visual--interaction-consistency-standards)
13. [Phased Implementation Roadmap](#13-phased-implementation-roadmap)
14. [Decisions Log](#14-decisions-log)
- [Appendix A — Annotated Wireframes](#appendix-a--annotated-wireframes)
- [Appendix B — Data Contract & Endpoint Coverage](#appendix-b--data-contract--endpoint-coverage)

---

## 1. Purpose & Scope

The PRD defines **what** Shrimpy does and the [Design System](./DESIGN_SYSTEM.md) defines **how it looks**. Neither defines **how a user moves through the product** — the path from "I just found this bot" to "my server runs on it." That missing layer is why the current dashboard feels disjointed: the screens exist, but the *journey between them* was never designed.

This document defines:

- The **information architecture** (what screens exist and how they relate).
- The **end-to-end journey** for each persona, screen by screen, with **first-run / empty / configured / error** states.
- A **gap analysis** of the current build against that journey.
- A **prioritized backlog** of UX/UI improvements.
- **Consistency standards** so every screen feels like one product.

> **North-star sentence:** *A non-technical server admin should go from login to a working, live ticket panel in under 5 minutes, without ever reading documentation or typing a Discord command.*

---

## 2. Relationship to Existing Docs

| Doc | Owns | This doc references it for |
|-----|------|----------------------------|
| [PRD](./PRD.md) §3, §5 | Personas & user stories | Who travels each journey and which stories each stage satisfies |
| [Design System](./DESIGN_SYSTEM.md) | Color, type, spacing tokens | The single visual language every screen must use |
| [Technical Spec](./TECHNICAL_SPEC.md) §4, §7 | REST endpoints & two-level auth | What data each screen can load/save and who can see it |

User-story IDs referenced below (e.g. `A-02`, `S-01`, `M-01`) map to [PRD §5](./PRD.md#5-user-stories).

---

## 3. Personas & Surface Matrix

There are three personas ([PRD §3](./PRD.md#3-target-users--personas)). Critically, **they do not all use the same surface** — and the current frontend treats the dashboard as if everyone lives there.

| Persona | Primary surface | Secondary surface | Dashboard access level |
|---------|-----------------|-------------------|------------------------|
| **Server Admin** | Web Dashboard | Discord slash commands | Level 1 (Administrator / Manage Server) — full config |
| **Support Staff** | Discord ticket channels | Web Dashboard (tickets only) | Level 2 (Dashboard Access role) — operate, not configure |
| **Member / Inquirer** | Discord (buttons, reactions) | — | None — never logs into the dashboard |

> **Implication for the journey:** The dashboard is an **Admin configuration console + Staff operations console**. The Member never sees it — but **everything the Member experiences in Discord is authored in the dashboard**. So the dashboard must let admins *design and preview the Member's Discord experience* (welcome card, ticket panel, reaction message). The Member journey in §9 is therefore the **downstream output** the dashboard must make tangible.

Access levels come straight from [Technical Spec §7.3](./TECHNICAL_SPEC.md#73-two-level-access-control) — the UI must respect them (a Level-2 staff user should not see the panel builder or settings).

---

## 4. Experience Principles

These are the rules every screen and flow should be judged against.

1. **Guide, don't dump.** Never drop a user onto an empty data table. Every entry point either shows progress (a setup checklist) or a meaningful empty state with one obvious next action.
2. **One product, one skin.** Every pixel uses the [Design System](./DESIGN_SYSTEM.md) tokens (coral + teal + navy). No screen invents its own palette. *(This is the single biggest current violation — see §12.)*
3. **Show the outcome.** Config screens render a **live Discord-accurate preview** of what the Member will see, so the admin edits with confidence.
4. **Progressive disclosure.** Defaults that work out of the box; advanced options tucked behind "Advanced" toggles. A low-technical admin should never be confronted with `supportRoles[]` or "gateway constraints."
5. **Feedback is immediate and human.** Saves confirm with toasts (not `alert()`), errors are surfaced inline with a recovery action, and destructive actions confirm before firing.
6. **Respect the role.** Staff see only what they can act on. Admins see everything. Owners get the multi-bot admin area.
7. **Real ≠ demo.** The sandbox/demo experience is clearly labeled and visually separated from a real authenticated session.

---

## 5. Information Architecture — Current vs Proposed

### 5.1 Current sitemap (as built)

```
/                                  Landing (marketing)            ✅ on-brand (coral/teal)
/login                             Login + "Sandbox Demo" link    ✅ on-brand
/dashboard                         Server selection  ← doubles as picker AND dashboard root
                                                       ❌ OFF-brand (indigo/near-black, inline styles)
/dashboard/[guildId]/tickets       Tickets table   ← default landing after picking a server
/dashboard/[guildId]/panels        Ticket panels + categories
/dashboard/[guildId]/welcome       Welcome config
/dashboard/[guildId]/roles         Reaction roles
/dashboard/[guildId]/settings      Bot params + staff roles + auto-roles
```

**Problems with the current IA**

- **Server selection is conflated with the dashboard.** `/dashboard` is *both* the server picker ([dashboard/page.tsx](../../dashboard/app/dashboard/page.tsx)) and the root of the per-server console. These are two fundamentally different jobs — *"which server?"* vs *"manage this server"* — and collapsing them into one route is a core reason the experience feels muddled. **Server selection must be its own dedicated page, separate from the dashboard.**
- **No server "home."** Selecting a server dumps the admin directly into the **Tickets table** ([dashboard/page.tsx](../../dashboard/app/dashboard/page.tsx) → `/tickets`). On a brand-new server this table is empty and meaningless — the worst possible first impression.
- **No onboarding path.** Nothing tells a first-time admin *what to do first*. Setup order (staff roles → panel → welcome) is implicit knowledge.
- **Settings is a junk drawer.** Bot params, staff/dashboard-access roles, and auto-roles are crammed into one page with engineer-y copy ("Level 2 credentials").
- **Missing screens that the PRD/Spec promise:** Statistics ([Spec §4.8](./TECHNICAL_SPEC.md#48-statistics) exists as an endpoint), Transcripts ([PRD A-11/S-07/M-05]), Multi-bot admin ([Spec §4.9](./TECHNICAL_SPEC.md#49-admin--discord-bot-applications) exists as endpoints, no UI), and a ticket **detail** view.

### 5.2 Proposed sitemap

```
PUBLIC
 /                                 Landing (marketing)
 /login                            Login (Discord OAuth) + clearly-labeled Demo entry

AUTHENTICATED — server selection  (DEDICATED page, separate from the dashboard)
 /servers                          ★ Server selection  ← MOVED off /dashboard + REBRAND
 /dashboard                        (bare) → redirects to /servers

AUTHENTICATED — per-server dashboard  (/dashboard/[guildId])
 ├─ /                              ★ Overview / Home   ← NEW: setup checklist + live stats + health
 │
 │  OPERATE  (Admin + Staff)
 ├─ /tickets                       Tickets inbox (filter, search, priority, bulk)
 ├─ /tickets/[ticketId]            ★ Ticket detail   ← NEW: claim, priority, internal notes (read-only conversation, v1)
 ├─ /transcripts                   ★ Transcripts archive   ← NEW: search + view + export
 │
 │  SERVER MANAGEMENT  (Admin only)
 ├─ /panels                        Ticket panels & categories (multi-button / select menu)
 ├─ /welcome                       Welcome & auto-roles on join
 ├─ /roles                         Reaction roles
 │
 │  SETTINGS  (Admin only)
 ├─ /settings                      General (nickname, prefix, language, log channel, auto-close, ticket limit)
 └─ /settings/access               ★ Staff & dashboard-access roles (split out of General)

OWNER ONLY
 /admin/apps                       ★ Multi-bot application manager   ← NEW UI for existing endpoints
```

### 5.3 Proposed sidebar (grouped, role-aware)

```
┌─────────────────────────┐
│ 🦐 Shrimpy              │
│ [ Server switcher ▾ ]   │   ← rich switcher w/ avatar + status, not a bare <select>
├─────────────────────────┤
│ ▸ Overview              │   (everyone)
│                         │
│ OPERATE                 │
│ ▸ Tickets         (3)   │   ← live open-count badge   (Admin + Staff)
│ ▸ Transcripts           │
│                         │
│ SERVER MANAGEMENT       │   (Admin only — fully hidden for Staff, §14.4)
│ ▸ Ticket Panels         │
│ ▸ Welcome               │
│ ▸ Reaction Roles        │
│                         │
│ SETTINGS                │   (Admin only — fully hidden for Staff, §14.4)
│ ▸ General               │
│ ▸ Staff & Access        │
├─────────────────────────┤
│ 👤 user      ☾ theme    │
└─────────────────────────┘
```

> Group labels ("OPERATE" / "SERVER MANAGEMENT" / "SETTINGS") turn a flat 5-item list into a mental model: *things I do daily* (feature configuration) vs *things I set up once* (bot/server-level plumbing). Staff (Level 2) see only OPERATE — both other groups are absent from their sidebar, not just visually muted (§14.4 decided: fully hidden, not read-only).
>
> Bot-wide status/error logs across all servers it's in (distinct from per-guild health, §7.5) is **out of scope for now** — noted as a future idea, not specced.

---

## 6. Master Journey Map

```
                          ┌────────────────────────────────────────────────────────────┐
                          │                        DISCORD SIDE                          │
   MEMBER  ───────────────│  joins server → gets welcome → clicks ticket button →        │
                          │  chats in private thread → reacts for roles → ticket closed  │
                          └───────────────▲──────────────────────────────▲──────────────┘
                                          │ configured by                │ operated by
                                          │                              │
   ADMIN   ─ Discover ─ Login ─ Pick ─ Invite ─ ★Setup ─ Configure ──────┘                │
            (landing)         server   bot    checklist  (panels/welcome/roles/settings)  │
                                                   │                                       │
                                                   └────────────► Operate (tickets) ◄──────┘
                                                                        ▲
   STAFF   ─ Login ─ Pick server ─────────────────────────────────────-┘
            (Level-2 access: tickets + transcripts only)
```

The three lanes meet at two seams:
- **Configure → Discord:** what the admin builds becomes what the member sees.
- **Discord → Operate:** what members do (open tickets) becomes the staff/admin work queue.

Designing the journey = making both seams **visible, previewable, and fast**.

---

## 7. Admin Journey (Primary)

The admin journey has **7 stages**. For each: the *goal*, the *screen*, the *states it must handle*, what's *built today*, and the *refined target*.

### 7.1 Discover — Landing page (`/`)

- **Goal:** Understand value in 10 seconds; click "Add to Discord" or try the demo.
- **Built today:** [Hero](../../dashboard/components/Hero.tsx), [Features](../../dashboard/components/Features.tsx), [InteractiveDemo](../../dashboard/components/InteractiveDemo.tsx), [Footer](../../dashboard/components/Footer.tsx) — visually on-brand and solid. ✅
- **Refinements:**
  - The Hero "Add to Discord" button is currently inert (no `href`) — wire it to the OAuth bot-invite URL.
  - Feature copy leaks jargon ("under a single GORM relational backend in pgxpool" — [Features.tsx:65](../../dashboard/components/Features.tsx#L65)). Rewrite for the *admin* audience, not engineers.
  - Footer links are all `#` placeholders — point Docs/Commands/Status somewhere real or hide them.

### 7.2 Authenticate — Login (`/login`)

- **Goal:** One-click Discord OAuth.
- **Built today:** [login/page.tsx](../../dashboard/app/(auth)/login/page.tsx) — on-brand coral card. ✅ But:
  - **The "Enter Sandbox Demo Preview" button links to a hardcoded guild id** (`/dashboard/123456789012345678/tickets`). Combined with the API's mock fallback (§10), a logged-out user lands in a fake-but-real-looking dashboard with no "you're in demo mode" signal.
- **Refinements:**
  - **Decided (§14.1):** gate the mock-data fallback behind an explicit **`/demo`** route. Remove the mock-fallback from [lib/api.ts](../../dashboard/lib/api.ts) for real sessions — a real, authenticated user must never silently render mock data because an API call failed; they hit the error state (§12.5/A.13) instead. `/demo` renders the same screens against static mock data with a persistent **"Demo mode" banner**, no login required.
  - On successful auth, redirect to the dedicated **`/servers`** selection page, **not** straight to a guild or to `/dashboard`.
  - Handle the `401` path from `/api/v1/auth/me` ([Spec §7.5](./TECHNICAL_SPEC.md#75-session-verification-nextjs)) → bounce to `/login` with a friendly "session expired" message.

### 7.3 Select a server — dedicated page (`/servers`)

This is its **own page, separate from the dashboard.** `/dashboard` (bare) redirects here. Its single job is *"which server?"* — never per-server management.

- **Goal:** Let the user pick a server to manage, and clearly separate **servers the bot is already in** from **servers they can add the bot to**.
- **Content — two explicit groups:**

```
┌─ Your servers ───────────────────────────────── (bot is active here) ┐
│  [🦐 Shrimpy Sandbox]      [🎮 Gamer Guild]      [⚓ Ocean Crew]       │
│   ● Active                  ● Active              ● Active             │
│   [ Manage → ]              [ Manage → ]          [ Manage → ]         │
└───────────────────────────────────────────────────────────────────────┘
┌─ Add Shrimpy to a server ──────── (you manage these; bot not yet in) ┐
│  [🌊 Reef Talk]            [🏝 Island Hub]        [ + Another server ] │
│   ○ Not added               ○ Not added           invite to any guild │
│   [ Invite Shrimpy ]        [ Invite Shrimpy ]                        │
└───────────────────────────────────────────────────────────────────────┘
```

- **Data:** `GET /api/v1/guilds` already returns each managed guild with a `bot_joined` boolean ([api.ts](../../dashboard/lib/api.ts)). Split the list on that flag: `bot_joined === true` → **"Your servers"** (action: Manage → `/dashboard/[guildId]`); `bot_joined === false` → **"Add Shrimpy to a server"** (action: Invite). Keep the permanent **"+ Another server"** invite affordance for guilds not in the list.
- **Built today:** [dashboard/page.tsx](../../dashboard/app/dashboard/page.tsx) — functional (guild grid, `bot_joined` status, invite card) but with two problems: (1) it lives **on `/dashboard`**, conflated with the console root — it must move to its own `/servers` route; and (2) it's **the single most off-brand screen in the app** — hardcoded indigo/purple on near-black (`#06070a`, `#4f46e5`, `#818cf8→#c084fc`), 100% inline styles, no theming. It looks like a different product than the landing page the user just came from. ❌
- **Refinements (high priority):**
  - **Move to a dedicated `/servers` route**; make `/dashboard` redirect there. The per-server console stays under `/dashboard/[guildId]/…`.
  - **Two clearly-labeled sections** ("Your servers" vs "Add Shrimpy to a server") instead of one mixed grid where joined and not-joined cards sit side by side.
  - **Rebuild against design tokens** (coral primary, teal accent, navy surfaces) using CSS modules — see §12.
  - **Bot-join detection:** after the admin clicks "Invite," the card stays in the "Add" group until a manual refresh. Poll `/api/v1/guilds` (or re-check on window-focus) and move the card up into "Your servers" automatically, then nudge them into Setup.
  - Show a count + a search box once a user has many guilds; distinguish "you're an admin here" (Level 1) vs "you're staff here" (Level 2) on the card.

### 7.4 Invite the bot

- **Goal:** Get Shrimpy into the chosen server with the right permissions.
- **Built today:** invite links use `permissions=8` (Administrator) — works but is a red flag for cautious admins.
- **Refinements:**
  - **Decided (§14.6):** use a **scoped permission integer** instead of blanket Admin (`permissions=8`) — View Channels, Manage Channels, Manage Roles, Manage Threads, Send Messages, Embed Links, Read Message History, Add Reactions, Use External Emojis, Manage Messages (per [PRD §8 assumptions](./PRD.md#8-assumptions--constraints) plus what `/roles` reaction-handling needs).
  - After invite, return the user to a **post-invite interstitial** that confirms the bot joined and links straight to **Setup** (§7.5), closing the loop instead of leaving them on a stale card.

### 7.5 ★ Guided setup — Server Overview / first-run (`/dashboard/[guildId]`) — **NEW**

This is the **missing keystone** of the whole journey. Today there is no home screen; the admin is dumped into an empty Tickets table.

- **Goal:** On a freshly configured server, give the admin an obvious, ordered path to first value. On an established server, give them an at-a-glance operational home.
- **Two states:**

**First-run (nothing configured yet) → Setup Checklist**
```
┌──────────────────────────────────────────────────────────┐
│  Welcome to Shrimpy on  [Server Name] 🦐                  │
│  Let's get you set up. ~5 minutes.                        │
│                                                          │
│  ① Designate staff roles            [ Set up ]   ○        │
│  ② Create your first ticket panel   [ Build  ]   ○        │
│  ③ Set a welcome message            [ Set up ]   ○        │
│  ④ (optional) Add reaction roles    [ Add    ]   ○        │
│                                                          │
│  ▓▓▓░░░░░░░  1 of 4 complete                              │
└──────────────────────────────────────────────────────────┘
```
Each step deep-links to the relevant config screen and returns to the checklist with the step checked. Order matters: **staff roles first** (so the panel's categories have someone to route to), **then** the panel.

**Configured → Overview dashboard**
```
┌───────────── Tickets ─────────────┐  ┌──── Server health ────┐
│  Open 3 · Claimed 2 · Closed 41   │  │ Bot: ✅ connected     │
│  Avg. resolution: 4h 12m          │  │ Perms: ⚠ missing      │
└───────────────────────────────────┘  │       Manage Threads  │
┌──────── Recent activity ──────────┐  └───────────────────────┘
│  • #ticket-0042 opened (Billing)  │  ┌──── Quick actions ────┐
│  • OceanMan claimed #0041         │  │ + New panel           │
│  • CoralReef closed #0039         │  │ ✎ Edit welcome        │
└───────────────────────────────────┘  │ ⤓ Export transcripts  │
                                        └───────────────────────┘
```
- **Decided (§14.5):** counts only for v1 (open/claimed/closed, avg. resolution) — no chart/sparkline. Defer any trend visualization to "Dashboard v2" per [PRD §9](./PRD.md#9-out-of-scope-items); avoids pulling in a charting dependency before the core journey ships.
- **Data source:** `GET /api/v1/guilds/:guildId/stats` ([Spec §4.8](./TECHNICAL_SPEC.md#48-statistics)) — already specced, not yet surfaced. Satisfies PRD `A-09`.
- **Health check** is a high-value add: detect whether the bot's role is above target roles / has needed permissions (the reaction-roles page already warns about this manually — [roles/page.tsx:285-294](../../dashboard/app/dashboard/[guildId]/roles/page.tsx#L285-L294)).

### 7.6 Configure features (Admin only)

The config screens largely exist; the journey work is **ordering, previewing, and depth**. Recommended in-product order mirrors the checklist. (b)–(d) below are the **Server Management** sidebar group (§5.3); (a) and (e) are **Settings**.

**(a) Staff & Access — `/settings/access`** (split out of today's Settings)
- Satisfies `A-05`. Two distinct concepts that today's copy conflates:
  - **Dashboard-access roles** (Level 2 — who can log into this console) — [settings/page.tsx:181-228](../../dashboard/app/dashboard/[guildId]/settings/page.tsx#L181-L228).
  - **Per-category support roles** (who can *see/handle* a given ticket category) — set on the panel screen.
- Rename "Level 2 credentials" → plain language ("People who can manage tickets here").

**(b) Ticket Panels — `/panels`** ([panels/page.tsx](../../dashboard/app/dashboard/[guildId]/panels/page.tsx))
- Satisfies `A-02`, `A-06`. Has a solid two-column **form + live Discord preview** pattern — keep and replicate this everywhere.
- **Depth gaps vs PRD:** UI supports only **one button per panel**; PRD allows up to 3 buttons or a 25-option select menu. No per-category opening embed, no thread-vs-channel choice (`A-06`), categories accept only **one** support role (`supportRoles: [newCatRoleId]` — [panels/page.tsx:115](../../dashboard/app/dashboard/[guildId]/panels/page.tsx#L115)).

**(c) Welcome — `/welcome`** ([welcome/page.tsx](../../dashboard/app/dashboard/[guildId]/welcome/page.tsx))
- Satisfies `A-03`, `A-04` (fold auto-roles-on-join in here — they're conceptually part of "what happens when someone joins," currently stranded in Settings).
- **Depth gaps:** no **template-variable picker** for `{user}` `{server}` `{membercount}` (a PRD `A-03` must-have); no "send test to me" button; the preview hardcodes "Shrimpy Sandbox" / "#99318".

**(d) Reaction Roles — `/roles`** ([roles/page.tsx](../../dashboard/app/dashboard/[guildId]/roles/page.tsx))
- Satisfies `A-04b`. **Depth gaps:** emoji limited to a hardcoded list of 6; no live Discord-style preview of the posted message; the "Gateway Requirements" warning is raw engineering ([roles/page.tsx:285-294](../../dashboard/app/dashboard/[guildId]/roles/page.tsx#L285-L294)) — replace with the automated health check from the Overview.

**(e) General Settings — `/settings`** ([settings/page.tsx](../../dashboard/app/dashboard/[guildId]/settings/page.tsx))
- Bot nickname, prefix, log channel, per-user ticket limit (`A-07` ✅ present). **Missing:** auto-close inactivity duration (`A-08`), language selection.

### 7.7 Operate — Tickets (shared with Staff, see §8)

Once panels are live and members start opening tickets, the admin's daily surface becomes the Tickets inbox. Covered in §8 since it's the Staff persona's primary screen.

### 7.8 Admin journey — stage summary

| # | Stage | Screen | Built? | Key gap |
|---|-------|--------|--------|---------|
| 1 | Discover | `/` | ✅ | Dead "Add to Discord" button; jargon copy |
| 2 | Authenticate | `/login` | ✅ | Demo not labeled; redirects to a guild not the `/servers` selection page |
| 3 | Select server | `/servers` (was `/dashboard`) | ⚠️ | **Conflated w/ dashboard**; off-brand; no join-detection |
| 4 | Invite bot | (Discord) | ⚠️ | Over-broad perms; no return loop |
| 5 | **Setup** | `/dashboard/[id]` | ❌ | **Screen doesn't exist** |
| 6 | Configure | panels/welcome/roles/settings | ⚠️ | Depth gaps vs PRD; reorg needed |
| 7 | Operate | tickets | ⚠️ | No detail view, priority, notes |

---

## 8. Support Staff Journey

Staff (Level 2) are **operators, not configurers**. Their journey is narrow and should be ruthlessly focused.

```
Login ─ Pick server ─ Tickets inbox ─ Open a ticket ─ Claim ─ Set priority ─ Reply / internal note ─ Close ─ (Transcript saved)
```

- **Access:** SERVER MANAGEMENT and SETTINGS groups are **fully hidden** (§14.4 decided — not read-only); the sidebar shows only Overview, Tickets, Transcripts. Enforced server-side per [Spec §7.3](./TECHNICAL_SPEC.md#73-two-level-access-control) and mirrored in the UI.
- **Built today:** [tickets/page.tsx](../../dashboard/app/dashboard/[guildId]/tickets/page.tsx) is a flat table with claim/close/reopen/archive/transcript actions. It satisfies `S-01`, `S-08`, `S-09` at a basic level.
- **Gaps against Staff stories:**
  - `S-02` **Priority** (Low/Med/High/Urgent) — not in the UI at all, though the [Design System §8](./DESIGN_SYSTEM.md#8-component-tokens) already defines priority badge colors.
  - `S-03` **Close with resolution note** — close is a bare action; no note captured.
  - `S-04` **Internal staff notes** — absent.
  - `S-07` **Generate/view transcript** — only an `alert()` stub ([tickets/page.tsx:81-83](../../dashboard/app/dashboard/[guildId]/tickets/page.tsx#L81-L83)).
  - There is **no ticket detail view** — staff can't read the conversation in the dashboard, only act on a row.

> `S-05` (add/remove participants) is **deferred** (§14.8 decided) — per-category support roles already auto-grant every role-holder access to a ticket's thread, covering the real workflow without a per-ticket participants table.

### 8.1 ★ Proposed: Ticket detail (`/tickets/[ticketId]`)

```
┌─ #ticket-0042 · Billing ──────────────────────  [Open ▾] [Priority: High ▾] ┐
│ Creator: ShrimpLover42      Claimed by: —        Opened: 2h ago             │
├──────────────────────────────────────────────────────────────────────────┤
│  conversation (read-only mirror of the Discord thread — §14.7 decided)     │
│   🦐 Shrimpy: Welcome to your support thread…                              │
│   ShrimpLover42: my invoice double-charged                                 │
│                                                                            │
│  ── internal notes (staff-only, never shown to member) ──                  │
│   OceanMan: refunded via Stripe, awaiting confirmation                     │
├──────────────────────────────────────────────────────────────────────────┤
│ [ Claim ]  [ Internal note ]  [ Close w/ note ▾ ]                          │
│ [ ⤓ Export transcript ]                                                    │
└──────────────────────────────────────────────────────────────────────────┘
```
This single screen lights up `S-02`, `S-03`, `S-04`, `S-07` and gives the inbox table a destination to click into. Reply happens in the Discord thread, not here (v1).

### 8.2 Inbox refinements

- Add a **priority column + badge** and a priority filter.
- Add **search** (creator, ticket id, category) and **pagination** for high-volume servers.
- Replace the row's `actionBtnActive || actionBtn` className fallback ([tickets/page.tsx:105](../../dashboard/app/dashboard/[guildId]/tickets/page.tsx#L105)) with a real active style.

---

## 9. Member Journey (Discord-side, driven by the dashboard)

The member never logs into the dashboard — but their entire experience is *authored* there. Surfacing this dependency is what makes the config screens meaningful.

| Member moment (Discord) | PRD story | Authored on dashboard screen | Dashboard must let admin… |
|-------------------------|-----------|------------------------------|---------------------------|
| Joins server, gets welcome DM/card | `M-04` | Welcome | …preview the exact card + test-send it |
| Auto-assigned a role on join | `A-04` | Welcome (proposed) / Settings | …pick roles and confirm bot can assign them |
| Sees support panel, clicks button | `M-01` | Ticket Panels | …preview the embed + button styling live |
| Lands in a private thread/channel | `M-02` | Panels → category | …choose thread vs channel + visibility roles |
| Self-assigns roles via reaction | (A-04b) | Reaction Roles | …preview the message + pick any emoji |
| Closes own ticket, gets transcript | `M-03`, `M-05` | Tickets / Transcripts | …configure transcript delivery |

**Design takeaway:** every CONFIGURE screen must answer *"what will the member actually see?"* with a faithful, live preview. The Panels screen already nails this ([panels/page.tsx:233-267](../../dashboard/app/dashboard/[guildId]/panels/page.tsx#L233-L267) renders a real Discord-style embed); Welcome partially does; Reaction Roles does not. Bring all three to parity.

---

## 10. Gap Analysis — Current Build vs Intended Journey

Mapped to the four problem areas you identified, plus the design issue.

### 10.1 Visual identity mismatch
| Where | Problem | Fix |
|-------|---------|-----|
| [dashboard/page.tsx](../../dashboard/app/dashboard/page.tsx) | Hardcoded indigo/purple (`#06070a`, `#4f46e5`, purple gradients); ignores every design token; dark-locked | Rebuild with `--primary`/`--accent`/`--bg-*` tokens + CSS module (§12) |
| Loading state ([dashboard/page.tsx:60-75](../../dashboard/app/dashboard/page.tsx#L60-L75)) | Hardcoded `#06070a` background | Use `--bg-base`; build a reusable `<PageLoader/>` |
| Discord preview blocks (panels/welcome) | Use Discord's real colors `#5865F2`/`#36393f` | **Keep** — these intentionally emulate Discord; just isolate them in a `<DiscordPreview/>` component |

### 10.2 Unclear / missing flow
| Gap | Impact | Fix |
|-----|--------|-----|
| Server selection conflated with dashboard | `/dashboard` is both picker and console root | Move selection to dedicated `/servers`; `/dashboard` → redirect (§5.2, §7.3) |
| No server Overview/home | Admin lands in an empty table | Build §7.5 Overview + Setup checklist |
| No onboarding order | Admin doesn't know to set staff roles before panels | Setup checklist enforces order |
| Invite → no return loop | Stale "Invite Needed" cards | Join-detection + post-invite interstitial (§7.3–7.4) |
| Login → guild, not selection | Skips server selection | Redirect to `/servers` |

### 10.3 Thin / placeholder screens
| Screen | Missing vs PRD/Spec |
|--------|---------------------|
| Tickets | Detail view, priority (`S-02`), internal notes (`S-04`), real transcript (`S-07`) — `S-05` participants deferred (§14.8) |
| Panels | Multi-button/select-menu, per-category embed, thread-vs-channel (`A-06`), multiple support roles |
| Welcome | Template-variable picker (`A-03`), test-send |
| Settings | Auto-close duration (`A-08`), language |
| (none) | Statistics page (`A-09`), Transcripts archive (`A-11`), Multi-bot admin UI |

### 10.4 Inconsistent UI patterns
| Pattern | Current state | Standard to adopt |
|---------|---------------|-------------------|
| Styling | `dashboard/page.tsx` = inline; feature pages = CSS modules + heavy inline | One approach: CSS modules + token utilities (§12) |
| Feedback | `alert()` for saves/errors ([welcome:44](../../dashboard/app/dashboard/[guildId]/welcome/page.tsx#L44), [settings:57](../../dashboard/app/dashboard/[guildId]/settings/page.tsx#L57), [roles:100](../../dashboard/app/dashboard/[guildId]/roles/page.tsx#L100)) | Toast system + inline field errors |
| Loading | Spinner on Tickets; `return null` elsewhere ([welcome:57](../../dashboard/app/dashboard/[guildId]/welcome/page.tsx#L57), [settings:101](../../dashboard/app/dashboard/[guildId]/settings/page.tsx#L101)) | Shared skeletons per layout |
| Errors | `console.error` only; user sees nothing | Error boundary + retry affordance |
| Save state | No dirty indicator | Sticky "unsaved changes" save bar |
| Guild switcher | Bare `<select>` with raw emoji ([layout.tsx:113-123](../../dashboard/app/dashboard/layout.tsx#L113-L123)) | Custom dropdown w/ avatar + status |
| Copy/voice | Engineer-y ("Banner Image Knobs", "Spawn Thread Channel", "gateway constraints") | Plain, friendly admin language |

---

## 11. Improvement Ideas (Prioritized Backlog)

Beyond closing gaps — ideas to make the experience genuinely good. Grouped by priority.

### Must (fix the journey)
1. **Dedicated `/servers` selection page** (§7.3), split into "Your servers" vs "Add Shrimpy to a server"; `/dashboard` redirects there.
2. **Server Overview + Setup checklist** (§7.5) — the highest-leverage single addition.
3. **Rebrand the server-selection page** to the design system (§12).
4. **Ticket detail view** with priority + internal notes + close-with-note (§8.1).
5. **Toast + inline-error system**; kill all `alert()` calls.
6. **Reorganize IA** per §5 (grouped sidebar, split Settings, fold auto-roles into Welcome).

### Should (depth + polish)
7. **Statistics on the Overview** (wire `/stats`).
8. **Transcripts archive page** with view + export.
9. **Template-variable picker** + live render in Welcome; **test-send** button.
10. **Multi-button / select-menu panels** + per-category thread-vs-channel.
11. **Bot-permission health check** (reused on Overview + Reaction Roles + Panels).
12. **Skeleton loaders + error boundaries** for every data screen.
13. **Sticky "unsaved changes" save bar** on all config forms.
14. **Demo-mode banner** + clean separation of mock vs real session.

### Could (delight / power-user)
15. **Command palette (⌘K / Ctrl-K)** to jump servers/screens/tickets.
16. **Rich server switcher** with search + recent servers.
17. **Responsive/mobile**: collapsible sidebar, card-stacked tables (PRD treats responsive as in-scope).
18. **Accessibility pass**: visible focus rings, ARIA on custom controls, contrast audit (tokens already support it).
19. **Empty-state illustrations** with one-click CTAs on every list.
20. **Multi-bot application manager UI** (`/admin/apps`) for the owner persona.
21. **Onboarding "first ticket" celebration** + inline tips that retire once used.

---

## 12. Visual & Interaction Consistency Standards

The rule: **every screen renders from [Design System](./DESIGN_SYSTEM.md) tokens. No exceptions outside intentional Discord previews.**

### 12.1 Color
- Replace all hardcoded hex in [dashboard/page.tsx](../../dashboard/app/dashboard/page.tsx) with semantic tokens:
  - `#06070a` → `var(--bg-base)`; card `rgba(17,18,25,…)` → `var(--bg-surface)`; borders → `var(--border-subtle)`.
  - `#4f46e5` / `#6366f1` / `#818cf8` (indigo) → `var(--primary)` (coral); accent highlights → `var(--accent)` (teal).
  - status dot green → `var(--success)`; muted text → `var(--text-muted)`.
- **Exception:** `<DiscordPreview/>` may use literal Discord colors (`#5865F2`, `#36393f`, `#2f3136`) — they emulate Discord intentionally. Isolate them in that one component so they never leak.

### 12.2 Styling approach
- **Standardize on CSS Modules + token-driven utility classes** (the feature pages and `globals.css` already establish this). Migrate `dashboard/page.tsx` off inline styles. Acceptable inline styles: truly dynamic, data-derived values (e.g. a role's hex color), not static layout/color.

### 12.3 Shared components to extract
| Component | Replaces repeated code in |
|-----------|---------------------------|
| `<DiscordPreview/>` | panels, welcome, interactive-demo |
| `<PageLoader/>` / `<Skeleton/>` | every data page |
| `<EmptyState/>` | panels, roles, tickets, server-select |
| `<Toast/>` + `useToast()` | all `alert()` sites |
| `<StatusBadge/>` / `<PriorityBadge/>` | tickets (tokens defined in [Design System §8](./DESIGN_SYSTEM.md#8-component-tokens)) |
| `<ServerSwitcher/>` | dashboard layout |
| `<SaveBar/>` | welcome, settings, panels |

### 12.4 Voice & copy
- Write for a **low-technical admin** ([PRD §3.1](./PRD.md#3-target-users--personas)). Concrete renames:
  - "Banner Image Knobs" → "Welcome Card"
  - "Spawn Thread Channel" → "Where tickets open"
  - "Level 2 credentials" / "Dashboard Access Roles (Level 2)" → "Who can manage tickets"
  - "Gateway Requirements / gateway constraint error" → an automated check: "⚠ Move Shrimpy's role above these roles so it can assign them. [Fix]"

### 12.5 Interaction defaults
- Saves → toast confirmation; destructive (delete panel/role, archive ticket) → confirm dialog.
- Every async screen has explicit **loading / empty / error** states.
- Forms track dirty state and warn on navigate-away with unsaved changes.

### 12.6 Heads-up: this is a customized Next.js
`dashboard/AGENTS.md` warns that the project's Next.js has **breaking changes vs upstream** — *"Read the relevant guide in `node_modules/next/dist/docs/` before writing any code."* Any implementation work from this spec must consult those in-repo docs first (routing, server/client components, metadata APIs may differ from defaults).

### 12.7 Responsive behavior
The PRD treats responsive as in-scope; the wireframes above are drawn desktop-first, so define how each pattern degrades. Use the [Design System](./DESIGN_SYSTEM.md) breakpoints; the rules below are the contract every screen follows.

| Breakpoint | Layout rule |
|------------|-------------|
| **≥ 1024px (desktop)** | As drawn: 260px sidebar pinned; two-column form + preview; full data tables. |
| **640–1023px (tablet)** | Sidebar collapses to an icon rail or a hamburger drawer; form + preview **stack** (form first, preview below); tables keep key columns, overflow scrolls horizontally. |
| **< 640px (mobile)** | Sidebar becomes an off-canvas drawer (hamburger in the top bar); `<ServerSwitcher/>` moves into the drawer; **tables become stacked cards** (`#`, creator, two badges per card); the sticky `<SaveBar/>` spans full width at the bottom. |

- **Live previews** (`<DiscordPreview/>`, A.5–A.7) drop below the form on stack; never hide them — the "show the outcome" principle still holds on mobile.
- **Ticket detail (A.3)** keeps the conversation full-width; the action bar becomes a sticky bottom toolbar.
- **The `/servers` grid (A.1)** is already card-based — it reflows from 3-up → 2-up → 1-up.
- Touch targets ≥ 44px; the command palette (A.13) is desktop/keyboard-only and may be hidden on touch.

---

## 13. Phased Implementation Roadmap

Sequenced so each phase ships a coherent, testable improvement.

### Phase 0 — Consistency foundation (unblocks everything)
- Extract `<DiscordPreview/>`, `<PageLoader/>`, `<EmptyState/>`, `<Toast/>`, `<StatusBadge/>`/`<PriorityBadge/>`.
- **Move server selection to a dedicated `/servers` route** (out of `/dashboard`), rebrand it to tokens, and split it into "Your servers" / "Add Shrimpy to a server"; make `/dashboard` redirect to `/servers`.
- Replace all `alert()` with toasts.

### Phase 1 — Fix the journey skeleton
- Build **Server Overview** at `/dashboard/[guildId]` with first-run **Setup checklist** + configured-state cards.
- Regroup the **sidebar** (Operate / Server Management / Settings) and make it **role-aware** (fully hide Server Management + Settings for Staff, §14.4).
- Redirect login → `/servers`; add **demo-mode banner**.
- Invite **join-detection** + post-invite return loop.

### Phase 2 — Operations depth (Staff value)
- **Ticket detail view** (`/tickets/[id]`): read-only conversation mirror (§14.7), claim, **priority**, **internal notes**, **close-with-note**.
- Inbox: priority column/filter, search, pagination.
- **Transcripts archive** page + export (replace the stub).

### Phase 3 — Configuration depth (Admin value)
- Panels: multi-button/select-menu, per-category embed + thread-vs-channel + multiple support roles.
- Welcome: template-variable picker, test-send, fold in auto-roles-on-join.
- Settings: auto-close duration, language; split out **Staff & Access**.
- Wire **statistics** into Overview.

### Phase 4 — Delight & scale
- Command palette, rich server switcher, responsive/mobile, accessibility pass.
- Multi-bot **`/admin/apps`** owner UI.

---

## 14. Decisions Log

All eight items below were open questions; each is now **Decided** and reflected throughout this doc, [TECHNICAL_SPEC.md](./TECHNICAL_SPEC.md), and [CHANGELOG.md](./CHANGELOG.md).

1. **Demo/sandbox strategy — Decided: gate behind `/demo`.** The mock-fallback in [lib/api.ts](../../dashboard/lib/api.ts) is removed for real sessions; `/demo` is an explicit, unauthenticated route rendering the same screens against static mock data with a persistent "Demo mode" banner. A real session never silently falls back to mocks on API failure — it hits the error state (§12.5/A.13) instead. See §7.2.
2. **Ticket detail: route vs drawer — Decided: full route.** `/tickets/[id]`, shareable/deep-linkable, inbox preserved behind it. Staff share ticket links, need back/forward and refresh-survival mid-triage — a drawer loses all of that. See §8.1.
3. **Auto-roles home — Decided: yes, inside `/welcome`,** and the sidebar group that contains Panels/Welcome/Roles is renamed **CONFIGURE → "Server Management."** Settings (General, Staff & Access) remains a separate group — it's bot/server-level plumbing, not feature configuration. Bot-wide status/error-logs across every server it's in was raised but is **out of scope for now** (a future `/admin/apps` extension, not specced here). See §5.2, §5.3.
4. **Staff dashboard scope — Decided: fully hidden, not read-only.** Staff (Level 2) is an operational role — their job is the ticket queue. A read-only Settings/Panels view adds sidebar clutter and a permission surface (visible-vs-editable) for no workflow gain; category-routing context belongs on the ticket itself (category badge), not a parallel Settings page. See §8.
5. **Statistics depth for v1 — Decided: counts only, no chart.** Overview ships open/claimed/closed counts + avg. resolution; sparkline/trend charts deferred to "Dashboard v2" per [PRD §9](./PRD.md#9-out-of-scope-items) — avoids a charting dependency before the core journey ships. See §7.5.
6. **Invite permissions — Decided: scoped, not Administrator.** Request View Channels, Manage Channels, Manage Roles, Manage Threads, Send Messages, Embed Links, Read Message History, Add Reactions, Use External Emojis, Manage Messages — not `permissions=8`. See §7.4.
7. **Dashboard→thread reply scope — Decided: read-only in v1.** No composer, no `POST …/tickets/:id/messages` endpoint. Staff already live in the Discord thread to reply; ticket-detail shows claim/priority/notes/transcript only. Revisit if dashboard-first support is requested. See §8.1, [TECHNICAL_SPEC §4.4](./TECHNICAL_SPEC.md#44-tickets).
8. **Participants (`S-05`) — Decided: defer, no `ticket_participants` table.** Per-category **support roles** ([§A.6](#a6--panels-ticket-panels--categories), `ticket_categories`, multi-role) already auto-grant every role-holder access to a ticket's thread the moment it opens — the real workflow (route to the right team) is already covered. Participants would only add value for one-off individual escalation outside the support-role list, which isn't needed now. Revisit only if that specific case is requested. See §8, [TECHNICAL_SPEC §3.2](./TECHNICAL_SPEC.md#32-postgresql-ddl).

---

## Appendix A — Annotated Wireframes

Low-fidelity, **token-annotated** wireframes for **every primary screen and cross-cutting state** in the journey. They are the visual companion to the screen specs above:

| Wireframe | Screen | Spec section |
|-----------|--------|--------------|
| A.1 | `/servers` selection | [§7.3](#73-select-a-server--dedicated-page-servers) |
| A.2 | `/dashboard/[guildId]` Overview + Setup | [§7.5](#75--guided-setup--server-overview--first-run-dashboardguildid--new) |
| A.3 | `/tickets/[ticketId]` detail | [§8.1](#81--proposed-ticket-detail-ticketsticketid) |
| A.4 | Staff (Level 2) sidebar | [§3](#3-personas--surface-matrix) / [§8](#8-support-staff-journey) |
| A.5 | `/welcome` | [§7.6c](#76-configure-features-admin-only) |
| A.6 | `/panels` | [§7.6b](#76-configure-features-admin-only) |
| A.7 | `/roles` | [§7.6d](#76-configure-features-admin-only) |
| A.8 | `/settings` + `/settings/access` | [§7.6a](#76-configure-features-admin-only) / [§7.6e](#76-configure-features-admin-only) |
| A.9 | Shared component patterns | [§12.3](#123-shared-components-to-extract) |
| A.10 | `/tickets` inbox | [§8.2](#82-inbox-refinements) |
| A.11 | `/transcripts` archive | [§5.2](#52-proposed-sitemap) (A-11/S-07) |
| A.12 | `/admin/apps` owner UI | [§5.2](#52-proposed-sitemap) (owner) |
| A.13 | Cross-cutting states & flows | [§7.4](#74-invite-the-bot) / [§12.5](#125-interaction-defaults) |

**[Appendix B](#appendix-b--data-contract--endpoint-coverage)** then maps every one of these screens to its backing [Technical Spec §4](./TECHNICAL_SPEC.md#4-rest-api-design) endpoint(s), so nothing in the journey is left without a data source.

> These are **layout + behaviour intent, not pixel specs.** Every label in the right margin names the [Design System](./DESIGN_SYSTEM.md) token an implementer must use — no hardcoded hex (§12.1). The customized-Next.js caveat in §12.6 still applies before writing any code.

**Legend**

```
●  filled status dot (--success)        ◯  hollow/pending (--text-muted)
▸  nav item            ▓░  progress fill (--primary) on track (--bg-surface-elevated)
[  Primary  ]  --primary bg / --text-on-primary    [ Secondary ]  --bg-surface + --primary border
↗  external (opens Discord OAuth)       →  internal navigation
↑text  = margin annotation pointing at the element on its left
```

### A.1 — `/servers` (Server Selection)

Dedicated pre-dashboard page; **no per-server sidebar** (nothing selected yet). Top bar + centered content column (`--content-max: 1200px`). Replaces the off-brand inline-styled `dashboard/page.tsx`.

**Populated state**

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  🦐 Shrimpy                                              ☾ theme    👤 Salman ▾ ║  top bar: --bg-surface, h 64px
╠══════════════════════════════════════════════════════════════════════════════╣
║                                                                                ║  page bg: --bg-base
║      Choose a server                                                           ║  --text-3xl, --font-display
║      Manage Shrimpy on a server you own, or add it to a new one.               ║  --text-muted
║                                                            ┌──────────────────┐║
║                                                            │ 🔎 Search servers│║  appears when >8 guilds
║                                                            └──────────────────┘║  --bg-surface, --border-subtle
║                                                                                ║
║   YOUR SERVERS  · 3                                            (bot is active) ║  label: --text-secondary, --text-xs, uppercase
║   ┌────────────────────┐ ┌────────────────────┐ ┌────────────────────┐        ║
║   │  ╭────╮             │ │  ╭────╮             │ │  ╭────╮             │        ║  card: --bg-surface, --radius-lg,
║   │  │ 🦐 │  ● Active   │ │  │ 🎮 │  ● Active   │ │  │ ⚓ │  ● Active   │        ║  border --border-subtle;
║   │  ╰────╯   ↑--success│ │  ╰────╯             │ │  ╰────╯             │        ║  hover: --shadow-md + border --primary
║   │  Shrimpy Sandbox    │ │  Gamer Guild        │ │  Ocean Crew         │        ║  --text-xl title / --text-sm muted
║   │  Admin · 248 members│ │  Admin · 1.2k       │ │  Staff · 89         │        ║
║   │                     │ │                     │ │   ↑ Level-2 badge   │        ║
║   │  [   Manage  →    ] │ │  [   Manage  →    ] │ │  [   Manage  →    ] │        ║  --primary btn
║   └────────────────────┘ └────────────────────┘ └────────────────────┘        ║
║                                                                                ║
║   ADD SHRIMPY TO A SERVER  · 2                  (you manage these; not added)  ║
║   ┌────────────────────┐ ┌────────────────────┐ ┌────────────────────┐        ║
║   │  ╭────╮             │ │  ╭────╮             │ │        ＋           │        ║  invitable: dimmed avatar,
║   │  │ 🌊 │  ○ Not added│ │  │ 🏝 │  ○ Not added│ │                     │        ║  hollow dot --text-muted
║   │  ╰────╯             │ │  ╰────╯             │ │   Another server    │        ║
║   │  Reef Talk          │ │  Island Hub         │ │   Invite Shrimpy to │        ║
║   │  Admin · 512        │ │  Admin · 67         │ │   any server you    │        ║
║   │                     │ │                     │ │   manage            │        ║
║   │ [ Invite Shrimpy ↗ ]│ │ [ Invite Shrimpy ↗ ]│ │ [  Choose server ↗ ]│        ║  secondary btn;
║   └────────────────────┘ └────────────────────┘ └────────────────────┘        ║  dashed border on "+" card
╚══════════════════════════════════════════════════════════════════════════════╝
```

- **Data:** `GET /api/v1/guilds`. Split on `bot_joined`: `true` → *Your servers* (`Manage` → `/dashboard/[guildId]`); `false` → *Add Shrimpy* (`Invite` → scoped-permission OAuth URL, §7.4).
- **Level badge:** Level 1 → "Admin", Level 2 → "Staff" (from permission/role data, Spec §7.3).
- **Join-detection (§7.3):** after *Invite*, poll `/api/v1/guilds` / re-check on window-focus; when `bot_joined` flips, animate the card into *Your servers* and surface a "Set up →" nudge.
- **`+ Another server`** is permanent (covers guilds not in the returned list).

**Empty & loading states**

```
  No managed servers                                Loading
  ┌──────────────────────────────────┐             <PageLoader/> — 6 card-shaped
  │            🦐                     │             skeletons on --bg-base
  │   You don't manage any servers    │             (NOT the hardcoded #06070a
  │   where you can add Shrimpy yet.  │              the current page uses).
  │   [ Add Shrimpy to a server ↗ ]   │
  └──────────────────────────────────┘             If user has Your-servers but no
  <EmptyState/>, --bg-surface, centered            invitable guilds, hide the ADD group.
```

### A.2 — `/dashboard/[guildId]` (Server Overview / Home)

The missing keystone. Shown inside the full app shell (grouped, role-aware sidebar from §5.3). Two states driven by config completeness.

**App shell (frame for every per-server screen)**

```
╔════════════════════════╦═════════════════════════════════════════════════════╗
║ 🦐 Shrimpy             ║  Shrimpy Sandbox ▾        🔔   ☾ theme    👤 Salman ▾ ║  server switcher = <ServerSwitcher/>
║ ┌────────────────────┐ ║                                                       ║  (avatar+name+status, NOT a bare <select>)
║ │ 🦐 Shrimpy Sandbox▾│ ║                                                       ║
║ └────────────────────┘ ║              ↓ content, max 1200px ↓                  ║
║                        ║                                                       ║
║  ▸ Overview        ●   ║                                                       ║  active: --primary-muted bg,
║                        ║                                                       ║  left rail + text --primary
║  OPERATE               ║                                                       ║  group label: --text-muted, --text-xs
║  ▸ Tickets       (3)   ║                                                       ║  count badge: --accent-muted / --accent
║  ▸ Transcripts         ║                                                       ║
║                        ║                                                       ║
║  SERVER MANAGEMENT     ║                                                       ║  ┄ SERVER MGMT + SETTINGS groups
║  ▸ Ticket Panels       ║                                                       ║    HIDDEN for Level-2 Staff (§A.4)
║  ▸ Welcome             ║                                                       ║
║  ▸ Reaction Roles      ║                                                       ║
║                        ║                                                       ║
║  SETTINGS              ║                                                       ║
║  ▸ General             ║                                                       ║
║  ▸ Staff & Access      ║                                                       ║
║ ────────────────────── ║                                                       ║
║  👤 Salman    ☾        ║                                                       ║
╚════════════════════════╩═════════════════════════════════════════════════════╝
  sidebar: 260px, --bg-surface, border-right --border-subtle
```

**First-run — Setup Checklist (nothing configured yet)**

```
┌───────────────────────────────────────────────────────────┐
│  Welcome to Shrimpy on Shrimpy Sandbox  🦐                 │  hero card: --bg-surface, --radius-lg,
│  Let's get you set up — about 5 minutes.                   │  subtle --primary gradient wash
│                                                            │
│   ▓▓▓▓▓░░░░░░░░░░░░░░░  1 of 4 complete                    │  progress: --primary on --bg-surface-elevated
├───────────────────────────────────────────────────────────┤
│  ✅  ①  Designate staff roles                              │  done: --success check, muted text,
│         People who can manage tickets here.   [ Edit  ]    │  secondary btn
│ ─────────────────────────────────────────────────────────  │
│  ◯  ②  Create your first ticket panel        [ Build → ]   │  NEXT step emphasized: --primary btn
│         The button members click to get help. ▲ start here │  + --primary left border
│ ─────────────────────────────────────────────────────────  │
│  ◯  ③  Set a welcome message                 [ Set up  ]   │  pending: hollow circle, --text-muted
│         Greet new members automatically.                   │
│ ─────────────────────────────────────────────────────────  │
│  ◯  ④  Add reaction roles      (optional)    [ Add     ]   │  optional tag: --accent-muted
└───────────────────────────────────────────────────────────┘
┌──── Bot health ────────────────────────────────────────────┐
│  ✅ Connected    ⚠ Missing permission: Manage Threads  [Fix]│  ✅ --success / ⚠ --warning
└─────────────────────────────────────────────────────────────┘
```

- **Order enforced:** staff roles → panel → welcome → reaction roles. Each row deep-links to its config screen and returns with the step checked; progress recomputes on return.
- **Exactly one** step is promoted (primary button + accent rail) — never ambiguous what's next.
- **Completion source:** derive each check from existing config (staff_roles present? ≥1 panel? welcome enabled? ≥1 reaction message?).
- **Health strip** reuses the bot-permission check (§7.5, Backlog #11) — same component later embedded on Reaction Roles & Panels.

**Configured — Overview dashboard**

```
Overview                                                          --text-3xl
┌─ Tickets ──────────────────────┐ ┌─ Bot health ────────────┐   2-col grid, gap --space-6
│  ◍ 3      ◍ 2       ◍ 41        │ │  ● Connected             │   stat: --text-3xl;
│  Open     Claimed   Closed      │ │    Shrimpy#4023          │   label --text-muted --text-sm
│                                 │ │                          │   Open→--success, Claimed→--accent,
│  Avg. resolution  4h 12m        │ │  ⚠ Missing: Manage       │   Closed→--text-muted
│                                 │ │     Threads      [ Fix → ]│
└─────────────────────────────────┘ └──────────────────────────┘
┌─ Recent activity ──────────────┐ ┌─ Quick actions ─────────┐
│  • #0042 opened · Billing  2m   │ │  ＋ New ticket panel     │   activity: --text-sm,
│  • OceanMan claimed #0041  18m  │ │  ✎ Edit welcome message  │   timestamps --text-muted
│  • CoralReef closed #0039  1h   │ │  ⤓ Export transcripts    │   quick actions: ghost btns,
│  • #0040 opened · Bug      2h   │ │  ⚙ Server settings       │   hover --bg-surface-hover
│              [ View all → ]     │ │                          │
└─────────────────────────────────┘ └──────────────────────────┘
```

- **Data:** `GET /api/v1/guilds/:guildId/stats` (Spec §4.8). Satisfies PRD `A-09`.
- **State switch:** render Setup vs Overview on `setup_complete`; a dismissible "resume setup" pill may persist on the configured view until 100%.
- **No chart in v1** (§14.5 decided) — counts + avg. resolution only; trend visualization deferred to "Dashboard v2."

### A.3 — `/tickets/[ticketId]` (Ticket Detail)

The destination the inbox table clicks into — the single screen that lights up `S-02`, `S-03`, `S-04`, `S-07`. Shown in the app shell; visible to **Admin + Staff**. Full route (§14.2 decided), with the inbox preserved behind it.

```
‹ Back to Tickets                                                                  --text-muted link → /tickets
┌─ #ticket-0042 · Billing ───────────────────[ Open ▾ ]  [ ⚑ High ▾ ]  [ Claim ]─┐  header bar: --bg-surface;
│  Creator  ShrimpLover42      Claimed by  —          Opened  2h ago              │  status pill --color (status badge §8),
│  ↑avatar+name                ↑--text-muted          ↑--text-muted               │  priority pill ⚑ --warning (High)
├─────────────────────────────────────────────────────────────────────────────────┤
│  CONVERSATION                              (read-only Discord mirror — §14.7)     │  group label --text-muted --text-xs
│                                                                                   │
│   ╭──╮ Shrimpy  · bot                                                    2h ago   │  message rows; avatar --radius-full
│   ╰──╯ Welcome to your support thread. A staff member will be with you shortly.   │  body --text-base, ts --text-muted
│                                                                                   │
│   ╭──╮ ShrimpLover42                                                     2h ago   │
│   ╰──╯ My invoice was double-charged this month — order #99318.                   │
│                                                                                   │
│  ┄┄┄┄┄ INTERNAL NOTES ┄┄┄┄┄  staff-only · never shown to the member ┄┄┄┄┄┄┄┄┄┄┄  │  divider tinted --warning-muted;
│   ╭──╮ OceanMan                                                          14m ago  │  note rows on --bg-surface-elevated
│   ╰──╯ Refunded via Stripe, awaiting confirmation. Don't close yet.               │  to read as "off the record"
│                                                                                   │
├─────────────────────────────────────────────────────────────────────────────────┤
│  [ ✎ Internal note ]   [ ⤓ Export transcript ]              [ Close w/ note ▾ ]   │  secondary/ghost btns;
│  Reply in the Discord thread — this view is read-only (§14.7)                     │  Close = --danger, opens
└─────────────────────────────────────────────────────────────────────────────────┘    note-capture popover
```

- **Data:** ticket detail (messages + notes) endpoint (Spec §4 tickets); priority/claim/close are mutations. Conversation is a **read-only mirror** of the Discord thread (we don't re-implement chat, and there is no reply-from-dashboard composer in v1 — §14.7).
- **Internal notes (`S-04`)** are visually fenced (tinted divider + elevated surface) so staff never confuse them with member-visible replies. **Close-with-note (`S-03`)** captures a resolution note in a popover before the destructive close (confirm per §12.5).
- **Priority (`S-02`)** dropdown uses the [Design System §8](./DESIGN_SYSTEM.md#8-component-tokens) priority badge tokens (Low→--success, Med→--accent, High→--warning, Urgent→--danger).
- **Inbox link-through:** rows in `/tickets` navigate here; add a priority column + filter and search to the inbox (§8.2).

### A.4 — Role-aware sidebar: Staff (Level 2) variant

Same shell as §A.2, but the **SERVER MANAGEMENT and SETTINGS groups are absent** — not greyed out, *not rendered* (§14.4 decided: fully hidden, enforced server-side per Spec §7.3, not just hidden in the UI). Staff land on Overview and live in Tickets/Transcripts.

```
╔════════════════════════╗
║ 🦐 Shrimpy             ║
║ ┌────────────────────┐ ║
║ │ 🎮 Gamer Guild    ▾│ ║   switcher lists only guilds where this user has access
║ └────────────────────┘ ║
║                        ║
║  ▸ Overview            ║   Staff Overview = ticket stats + recent activity only
║                        ║   (no Setup checklist — that's an Admin concern)
║  OPERATE               ║
║  ▸ Tickets       (3)   ║
║  ▸ Transcripts         ║
║                        ║
║   ┄ no SERVER MGMT     ║   ← these groups simply don't exist for Level 2
║   ┄ no SETTINGS        ║
║ ────────────────────── ║
║  👤 Maya (Staff)  ☾    ║   role surfaced next to the user so it's obvious why
╚════════════════════════╝
```

- **Source of truth:** the same `/api/v1/guilds` access level that tagged the card "Staff" in §A.1 drives which nav groups render. A Level-2 user who deep-links to `/dashboard/[id]/settings` gets a 403 from the API and a friendly "You don't have access to this" screen (§12.5 error handling), not a broken page.
- **Admin = §A.2 full sidebar; Staff = this.** One component, role-filtered group list — don't fork the layout.

### A.5 — `/welcome` (Welcome & Auto-roles)

Admin-only. The reference implementation of **Principle 3 ("Show the outcome")**: a two-column **form + live Discord card preview**. Closes the `A-03` depth gaps — template-variable picker, test-send, and a preview that reflects *real* server data instead of the hardcoded "Shrimpy Sandbox / #99318".

```
Welcome                                                            --text-3xl
┌─ Settings ─────────────────────────────┐ ┌─ Live preview ──────────────────────┐  2-col, gap --space-6
│  Enabled            ●━━○                │ │  what members see in Discord         │  <DiscordPreview/> (real Discord
│  Send to    [ #welcome ▾ ] [ DM ☐ ]     │ │ ┌──────────────────────────────────┐ │  colors, isolated — §12.1)
│                                         │ │ │ ╭──╮ Welcome! 🦐                  │ │
│  Message                                │ │ │ ╰──╯ Hey ShrimpLover42, welcome   │ │  preview renders with the
│  ┌─────────────────────────────────┐    │ │ │      to Ocean Crew! You're our    │ │  CURRENT guild's name +
│  │ Hey {user}, welcome to {server}! │    │ │ │      248th member 🌊              │ │  live member count, not stubs
│  │ You're our {membercount}th member│    │ │ │ ┌──────────────────────────────┐ │ │
│  └─────────────────────────────────┘    │ │ │ │  [welcome card image]        │ │ │
│  Insert:  [{user}] [{server}]           │ │ │ └──────────────────────────────┘ │ │  variable chips: --accent-muted
│           [{membercount}] [{user.tag}]  │ │ └──────────────────────────────────┘ │  bg, --accent text, click =
│           ↑click to insert at cursor    │ │                                      │  insert token at cursor
│                                         │ │  [ ✉ Send test to me ]               │  test-send: secondary btn →
│  Card image (optional)                  │ │   ↑ posts the card to your DM so     │  DMs the requesting admin
│  [ Upload ] or [ paste URL          ]   │ │     you see it exactly as a member   │
│  ─────────────────────────────────────  │ └──────────────────────────────────────┘
│  AUTO-ROLES ON JOIN          (was in Settings — folded in here, §7.6c)            │  grouped under Welcome because
│  Assign on join:  @Member  @Unverified  [ + role ]                                │  it's a "what happens on join"
│   ⚠ Move Shrimpy above @Member so it can assign it.  [ Fix ]                      │  behaviour; ⚠ = health check
└───────────────────────────────────────────────────────────────────────────────────┘
▒ Unsaved changes                                              [ Discard ]  [ Save ]   sticky <SaveBar/> (§A.9)
```

- **Template variables (`A-03`):** chips insert `{user}` `{server}` `{membercount}` `{user.tag}` at the cursor; the preview resolves them against the live guild so the admin never guesses.
- **Test-send:** posts the rendered card to the admin's DM — confidence before going live.
- **Auto-roles folded in (§7.6c, §14.3 decided):** "assign on join" lives here, not in Settings, and shares the **role-height health check** from the Overview (§7.5). Confirms the bot *can* assign each role.
- **Reuse:** same `<DiscordPreview/>` + `<SaveBar/>` as Panels — see §A.9.

### A.6 — `/panels` (Ticket Panels & Categories)

Admin-only. The screen that already has the best form+preview pattern today — the work is **depth** (`A-02`, `A-06`): up to 3 buttons *or* a 25-option select menu (not one button), thread-vs-channel choice, **multiple** support roles per category, and — most importantly — a **fully configurable opening message** (the embed the bot posts into the ticket the instant a member clicks a button / picks a select option), each with its own live preview.

There are **two** things to preview here, so the screen has **two** editors: the **public panel** (what members click) and, behind each ⚙, the **category editor** including the **opening message** (what the bot posts when the ticket opens).

```
Ticket Panels                                            [ + New panel ]   --text-3xl

PANEL — the public message members click
┌─ Panel: "Get Support" ─────────────────┐ ┌─ Preview · public panel ────────────┐
│  Embed title  [ Need a hand?         ]  │ │ ┌──────────────────────────────────┐ │  <DiscordPreview/>
│  Description  [ Pick a topic below…  ]  │ │ │  Need a hand?                    │ │
│  Accent color [ ▦ #FF7B6B ]             │ │ │  Pick a topic below to open a    │ │  embed accent bar uses the
│  Open style   (•) Buttons  ( ) Select   │ │ │  private ticket.                 │ │  chosen color (dynamic inline
│   ┌─ up to 3 buttons ──────────────┐    │ │ │                                  │ │  style is OK here, §12.2)
│   │ 💳 Billing      ⚙   ✕          │    │ │ │  [ 💳 Billing ] [ 🐛 Bug ]        │ │  buttons mirror category list
│   │ 🐛 Bug report   ⚙   ✕          │    │ │ │  [ ❓ Other ]                     │ │
│   │ ❓ Other        ⚙   ✕          │    │ │ └──────────────────────────────────┘ │
│   │ [ + add button ] (3/3)         │    │ │                                      │
│   └────────────────────────────────┘    │ │  ⚙ on a category → opens the         │  per-category editor ↓ (next block)
│  ( ) Select menu  → up to 25 options    │ │     Category editor below            │
│  [ Post panel to ▾ #support ]           │ └──────────────────────────────────────┘
└─────────────────────────────────────────┘

CATEGORY — the ⚙ editor; the OPENING MESSAGE is what the bot posts when the ticket opens
┌─ Category: Billing ─────────────────────┐ ┌─ Preview · opening message ─────────┐
│  Button label  [ 💳 Billing          ]  │ │ first message the member sees inside │  <DiscordPreview/>
│  Opens as   (•) Private thread          │ │ their brand-new ticket               │
│             ( ) New channel             │ │ ┌──────────────────────────────────┐ │
│  Who can see  @Support @Billing [+role] │ │ │ 🦐 Shrimpy                       │ │  MULTIPLE support roles
│  ── Opening message (embed, A-06) ───── │ │ │ Thanks for reaching out, @alice! │ │  ← a FULL embed, not plain text
│  Title  [ Thanks for reaching out!   ]  │ │ │ ────────────────────────────     │ │
│  Body                                   │ │ │ A billing specialist will be     │ │  title + body + accent + media,
│  ┌───────────────────────────────────┐  │ │ │ with you shortly, @alice. Tell   │ │  configurable per category
│  │ A billing specialist will be with │  │ │ │ us your order number.            │ │
│  │ you shortly, {mention}. Tell us   │  │ │ │                                  │ │  body resolves {mention} /
│  │ your order number.                │  │ │ │ [ 🖼 thumbnail ]                 │ │  {category} / {number} live, like A.5
│  └───────────────────────────────────┘  │ │ └──────────────────────────────────┘ │
│  chips [{mention}][{category}][{number}] │ └──────────────────────────────────────┘  reuses A.5 variable picker
│  Accent [ ▦ #FF7B6B ]   Media [ + image ]│
└──────────────────────────────────────────┘
▒ Unsaved changes                                              [ Discard ]  [ Save & post ]
```

- **Multi-button / select-menu (`A-02`):** toggle between ≤3 buttons and a ≤25-option select menu; the panel preview re-renders the component type live.
- **Opening message is a configurable embed (`A-06`):** the message the bot posts into the freshly-opened ticket is a **full embed** — its own **title, body, accent color, and optional media** — edited per category with a **dedicated live preview** of exactly what the member sees on open (not the public panel). The body supports template variables — member tokens shared with Welcome (`{mention}`, `{user}`) plus the **ticket-context** tokens `{category}` and `{number}` ([Command Reference §9](COMMAND_REFERENCE.md#9-template-variables-reference)) — resolved against the live guild (P3 "show the outcome"). **Already backed by schema** — `ticket_categories.ticket_open_{title,message,color,media}` ([Technical Spec §3.2](TECHNICAL_SPEC.md#32-postgresql-ddl)) — so this is frontend + the panels API exposing those fields, **no migration**.
- **Per-category depth (`A-06`):** beyond the opening message, each category gets a **thread vs channel** choice and a **list** of support roles (the current `supportRoles: [oneRole]` becomes a real multi-select).
- **Destructive guards:** removing a button/category or re-posting a panel confirms first (§12.5); "Save & post" updates the live Discord message.

### A.7 — `/roles` (Reaction Roles)

Admin-only. Brings this screen to preview parity with Panels/Welcome (`A-04b`). Replaces the hardcoded 6-emoji list with a **full picker**, adds a **live message preview**, and swaps the raw "Gateway Requirements" wall of text for the same **automated health check** used elsewhere.

```
Reaction Roles                                           [ + New message ]   --text-3xl
┌─ Message ───────────────────────────────┐ ┌─ Live preview ──────────────────────┐
│  Post to   [ #roles ▾ ]                  │ │ ┌──────────────────────────────────┐ │  <DiscordPreview/>
│  Text                                    │ │ │ 🦐 Shrimpy                       │ │
│  ┌─────────────────────────────────┐     │ │ │ React below to pick your roles:  │ │
│  │ React below to pick your roles:  │     │ │ │   🎮  Gamer                      │ │  each mapping renders as a
│  └─────────────────────────────────┘     │ │ │   🎨  Artist                     │ │  reaction row exactly as posted
│                                          │ │ │   🌊  Ocean lover                │ │
│  Emoji → Role                            │ │ │  ┌──┐┌──┐┌──┐                     │ │  reaction pills below message
│   🎮  →  @Gamer        ✕                 │ │ │  │🎮││🎨││🌊│  ← members click    │ │
│   🎨  →  @Artist       ✕                 │ │ │  └──┘└──┘└──┘                     │ │
│   🌊  →  @Ocean lover  ✕                 │ │ └──────────────────────────────────┘ │
│   [ + add mapping ]                      │ │                                      │
│    └ 😀 opens FULL emoji picker          │ │  Mode  (•) Toggle  ( ) Verify-only   │  picker = any unicode/custom
│      (any unicode + server custom),      │ │                    ( ) Single-choice │  emoji, not the hardcoded 6
│      not a fixed list of 6               │ │                                      │
│  ──────────────────────────────────────  │ └──────────────────────────────────────┘
│  ✅ Shrimpy's role is above all 3 target roles — reactions will assign correctly. │  automated health check replaces
│     (was: raw "Gateway Requirements" engineering copy)                            │  the raw warning; ⚠+[Fix] if not
└────────────────────────────────────────────────────────────────────────────────────┘
▒ Unsaved changes                                              [ Discard ]  [ Save ]
```

- **Emoji picker:** full unicode + server custom emoji, replacing the hardcoded six.
- **Live preview:** renders the posted message *and* the reaction pills, matching Panels/Welcome (§9 "bring all three to parity").
- **Health check:** the role-height check is computed and stated plainly with a one-click **[Fix]**, not left as engineer-facing "gateway constraint" text (§12.4).

### A.8 — `/settings` & `/settings/access` (General + Staff & Access)

Admin-only. Splits today's junk-drawer Settings into two focused tabs and adds the missing `A-08` controls. Plain-language copy replaces "Level 2 credentials" (§12.4).

```
Settings   [ General ]  [ Staff & Access ]                  ← sub-nav tabs (--primary underline on active)

┌─ General ──────────────────────────────────────────────────┐
│  Bot nickname        [ Shrimpy                    ]         │  --bg-surface form rows,
│  Command prefix      [ !  ]                                 │  label --text-secondary
│  Language            [ English ▾ ]            ← NEW          │  (A-08 additions marked NEW)
│  Log channel         [ #shrimpy-logs ▾ ]                    │
│  Tickets per user    [ 3  ]  max open at once               │
│  Auto-close after    [ 48 ] hours of inactivity  ← NEW      │  auto-close duration (A-08)
└─────────────────────────────────────────────────────────────┘
▒ Unsaved changes                              [ Discard ]  [ Save ]

┌─ Staff & Access ───────────────────────────────────────────┐
│  WHO CAN MANAGE TICKETS HERE                                │  was "Dashboard Access Roles
│  Roles you grant dashboard access (operate tickets only):   │  (Level 2)" — now plain language
│   @Support   @Mods            [ + add role ]                │
│   ↑ these users see Overview + Tickets + Transcripts only   │  mirrors the §A.4 staff sidebar
│  ──────────────────────────────────────────────────────────│
│  ℹ Admins (Manage Server / Administrator) always have full  │  --info note; explains the
│    access — you don't need to list them here.               │  two-level model without jargon
│  ──────────────────────────────────────────────────────────│
│  Per-category support roles (who handles which ticket type) │  cross-link, not duplicated —
│  are set on each panel category →  [ Go to Panels ]         │  lives on §A.6
└─────────────────────────────────────────────────────────────┘
```

- **General:** adds **language** and **auto-close duration** (`A-08`) alongside the existing nickname/prefix/log-channel/ticket-limit.
- **Staff & Access (`A-05`):** separates the two concepts today's copy conflates — *dashboard-access roles* (Level 2, here) vs *per-category support roles* (on Panels, linked not duplicated). Copy is plain; the two-level model is explained in an info note, not assumed.

### A.9 — Shared component patterns (§12.3)

The cross-cutting primitives every screen above composes from. Building these first (Phase 0) is what makes the rest consistent.

```
<Toast/> + useToast()                      <SaveBar/>  (sticky, appears when form dirty)
 ╭─ ✓ Welcome saved ──────────╮            ┌──────────────────────────────────────────┐
 ╰────────────────────────────╯  --success │ ▒ You have unsaved changes                 │
 ╭─ ✕ Couldn't save · Retry ──╮             │              [ Discard ]   [ Save changes ]│
 ╰────────────────────────────╯  --danger  └──────────────────────────────────────────┘
 top-right stack, auto-dismiss;            --bg-surface-elevated, --shadow-lg, full-width
 replaces every alert() (§10.4)            footer; Save = --primary, Discard = ghost

<EmptyState/>                              <StatusBadge/> / <PriorityBadge/>  (§8 tokens)
 ┌──────────────────────────┐               Open ●  Claimed ●  Closed ●  Archived ●
 │         🦐               │                └ --success  --accent  --text-muted  --text-disabled
 │  No panels yet           │               Low ⚑  Med ⚑  High ⚑  Urgent ⚑
 │  [ Create your first → ]  │                └ --success --accent --warning --danger
 └──────────────────────────┘               pill: tinted bg (token-muted) + token text
 icon + line + ONE CTA;                     <PageLoader/> / <Skeleton/>
 used on panels/roles/tickets/servers        shimmer blocks shaped like the page's content,
                                             on --bg-surface; one per layout (list / form / cards)

<ServerSwitcher/>   (replaces the bare <select> in the layout)
 ┌─────────────────────────────┐
 │ 🦐 Shrimpy Sandbox        ▾ │  trigger: avatar + name + status dot
 ├─────────────────────────────┤
 │ 🔎 Search servers…          │  open: --bg-surface-elevated, --shadow-lg
 │ ● Shrimpy Sandbox   (current)│  current marked; recents on top
 │ ● Gamer Guild               │
 │ ● Ocean Crew                │
 │ ─────────────────────────── │
 │ + Add a server →  /servers  │  always-present escape hatch back to selection
 └─────────────────────────────┘
```

- **One source per primitive.** These map 1:1 to the §12.3 extraction table; every feature screen consumes them rather than re-implementing (e.g. `alert()` → `useToast()` everywhere; `return null` loading → `<PageLoader/>`).
- **`<DiscordPreview/>`** (used in A.5/A.6/A.7) is the only component allowed literal Discord colors (`#5865F2`, `#36393f`) — isolate them there so they never leak into the token system (§12.1).

### A.10 — `/tickets` (Tickets Inbox)

The daily operating surface for **Admin + Staff** and the list that A.3 clicks into. Shown in the app shell. Closes the §8.2 gaps: a priority column + filter, search, pagination, and a real active style on row actions. Backed by `GET …/tickets` with its existing query params (`status`, `priority`, `categoryId`, `openedBy`, `page`, `limit`).

```
Tickets                                                                         --text-3xl
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ [ All ][ Open 3 ][ Claimed 2 ][ Closed ][ Archived ]    ⚑ Priority ▾  🏷 Category ▾  🔎 │  status tabs → ?status=
│  ↑ active tab: --primary underline; counts from /stats                          search │  search → ?openedBy / id / text
├──────────────────────────────────────────────────────────────────────────────────────┤
│  #      Opened by         Category    Priority    Status      Claimed    Age      ⋯     │  header: --text-muted --text-xs
│ ────────────────────────────────────────────────────────────────────────────────────  │
│  0042   ShrimpLover42      Billing     ⚑ High      ● Open      —          2h    [ Claim ]│  whole row → /tickets/0042 (A.3)
│  0041   GuppyFan           Bug         ⚑ Med       ● Claimed   OceanMan    5h    [ Open ]│  ⚑ = <PriorityBadge/>, ● = <StatusBadge/>
│  0039   CoralReef          Other       ⚑ Low       ● Closed    OceanMan    1d    [ View ]│  (§8 tokens); row hover --bg-surface-hover
│  0038   ReefRanger         Billing     ⚑ Urgent    ● Open      —          3d    [ Claim ]│  Urgent priority pill --danger
├──────────────────────────────────────────────────────────────────────────────────────┤
│                                    ‹  1  2  3  …  ›                    25 / page ▾       │  pagination → ?page= &limit=
└──────────────────────────────────────────────────────────────────────────────────────┘
```

```
  Empty (no tickets yet)                           Empty (filter returns nothing)
  ┌──────────────────────────────────┐             ┌──────────────────────────────────┐
  │            🦐                     │             │   No tickets match these filters.  │
  │   No tickets yet.                 │             │   [ Clear filters ]                │
  │   They'll appear here when members │             └──────────────────────────────────┘
  │   open one from your panel.        │             <EmptyState/>, secondary CTA
  │   [ Set up a panel → ]             │
  └──────────────────────────────────┘             keep the filter bar visible above the
  <EmptyState/> → /panels (A.6)                     empty body so the user can adjust.
```

- **Row → detail:** clicking any row opens `/tickets/[ticketId]` (§A.3); the inline `[ Claim ]` is a quick `PATCH …/tickets/:id` without leaving the list.
- **Filters are URL state** (`?status=&priority=&categoryId=&page=`) so a filtered inbox is shareable/bookmarkable and survives refresh.
- **Priority column (`S-02`)** and the priority filter are the headline addition; the badge tokens come straight from [Design System §8](./DESIGN_SYSTEM.md#8-component-tokens) via `<PriorityBadge/>` (§A.9).
- **Staff vs Admin:** identical screen for both roles — it's an OPERATE surface (§A.4); only the surrounding nav differs.

### A.11 — `/transcripts` (Transcripts Archive)

Search, view, and export of closed/archived ticket conversations — for **Admin + Staff**. Replaces the `alert()` stub on the current inbox ([tickets/page.tsx:81-83](../../dashboard/app/dashboard/[guildId]/tickets/page.tsx#L81-L83)). **No new list endpoint required:** it is the tickets list filtered to terminal states (`GET …/tickets?status=closed,archived`) with each row linking to the existing per-ticket transcript (`GET …/tickets/:id/transcript`, JSON or HTML). Satisfies `A-11` / `S-07` / `M-05`.

```
Transcripts                                                                     --text-3xl
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ 🔎 Search creator, ticket #, or keyword           📅 Date range ▾   🏷 Category ▾        │  --bg-surface filter bar
├──────────────────────────────────────────────────────────────────────────────────────┤
│  #0039   Other      CoralReef     closed 1d ago     [ View ]   [ ⤓ HTML ]   [ ⤓ JSON ]  │  View → read-only viewer (below)
│  #0037   Billing    GuppyFan      closed 3d ago     [ View ]   [ ⤓ HTML ]   [ ⤓ JSON ]  │  ⤓ = direct transcript download
│  #0031   Bug        ReefRanger    archived 1w ago   [ View ]   [ ⤓ HTML ]   [ ⤓ JSON ]  │  age/state → --text-muted
└──────────────────────────────────────────────────────────────────────────────────────┘

  Viewer (route or drawer) — read-only, reuses the §A.3 CONVERSATION block WITHOUT the action bar:
  ┌─ Transcript · #0039 · Other ──────────────────────────────  closed by OceanMan · 1d ago ┐
  │  🦐 Shrimpy   Welcome to your support thread…                                    2d ago  │  identical message rendering
  │  CoralReef    how do I change my username?                                       2d ago  │  to A.3 (one shared component)
  │  OceanMan     here's how… (resolved)                                             1d ago  │  internal notes INCLUDED for staff,
  │  ── resolution note ── Refunded + explained. ──                                          │  fenced --warning-muted as in A.3
  │                                                       [ ⤓ Export HTML ]  [ ⤓ Export JSON ]│
  └──────────────────────────────────────────────────────────────────────────────────────┘
```

- **Reuses, doesn't rebuild:** the viewer is the same read-only conversation component as §A.3 minus the composer/action bar — build once, render in both places.
- **Export** maps directly to the `?format=html|json` variants of the transcript endpoint.
- **Internal notes** appear in the staff-facing viewer (fenced) but are stripped from any member-facing/`M-05` delivery — the backend, not the UI, enforces that boundary.

### A.12 — `/admin/apps` (Owner — Multi-Bot Application Manager)

**OWNER-ONLY**, and the only screen that lives **outside** `/dashboard/[guildId]` — it manages bot *applications*, not a single server. New UI over endpoints that already exist ([Spec §4.9](./TECHNICAL_SPEC.md#49-admin--discord-bot-applications)); gated by `AdminMiddleware` / `OWNER_DISCORD_ID`. This is the most sensitive screen in the product, so secrets are **always masked** and every mutation confirms.

```
╔══════════════════════════════════════════════════════════════════════════════╗
║  🦐 Shrimpy · Admin                          ‹ Back to servers   👤 Salman ▾   ║  separate top bar (no per-guild sidebar)
╠══════════════════════════════════════════════════════════════════════════════╣
║   Discord Applications                                        [ + Add app ]    ║  --text-3xl; owner-only badge near title
║                                                                                ║
║   ┌────────────────────────────────────────────────────────────────────────┐ ║
║   │  ● Connected   Production Bot                          Shrimpy#4023      │ ║  ● --success / ○ --danger (disconnected)
║   │    Client ID  123456789012345678                                        │ ║  client id shown (public);
║   │    Token      ••••••••••••••••••••  [ Reveal once ]    [ ⟳ Reconnect ]   │ ║  token/secret masked → •••• (--text-muted)
║   │    Redirect   https://…/api/v1/auth/callback           [ Edit ] [ 🗑 ]   │ ║  🗑 = destructive: confirm + stops session
║   └────────────────────────────────────────────────────────────────────────┘ ║
║   ┌────────────────────────────────────────────────────────────────────────┐ ║
║   │  ○ Disconnected   Dev Bot                              last seen 2d ago  │ ║  disconnected card: --danger dot,
║   │    …                                          [ ⟳ Reconnect ] [ Edit ]   │ ║  --warning "last seen" line
║   └────────────────────────────────────────────────────────────────────────┘ ║
╚════════════════════════════════════════════════════════════════════════════════╝

  Add / Edit app  (modal — <Toast/> on success, inline field errors on 4xx)
  ┌─────────────────────────────────────────────────────────┐
  │  Name              [ Production Bot                 ]    │  --bg-surface-elevated modal
  │  Bot token         [ ••••• paste to replace        ]    │  edit leaves blank = keep current
  │  Client ID         [ 123456789012345678            ]    │  (PUT fields all optional, §4.9)
  │  Client secret     [ ••••• paste to replace        ]    │
  │  OAuth redirect URI[ https://…/api/v1/auth/callback ]   │
  │                                  [ Cancel ]  [ Save app ]│  Save = --primary; POST/PUT → live
  └─────────────────────────────────────────────────────────┘  start/reconnect of the gateway session
```

- **Access:** non-owners never see this in the nav; a direct hit returns 403 → the access-denied screen (§A.13). Surfaced from the user menu / a footer "Admin" link, not the per-server sidebar.
- **Mask by default:** GET returns `"***"` for token + secret ([Spec §4.9](./TECHNICAL_SPEC.md#49-admin--discord-bot-applications)); the form treats an unchanged masked field as "keep existing" so a save never overwrites a secret with dots.
- **Live effect:** `POST`/`PUT` (token) starts or reconnects the session in the background; `DELETE` stops it — reflect the resulting state with the connection dot after the action resolves (poll/refresh).
- **Reconnect** maps to `POST …/apps/:id/reconnect`; show a transient "reconnecting…" state on the card.

### A.13 — Cross-cutting states & flows

The states that don't belong to one screen but make the product feel finished — they realize Principles 5 ("feedback is immediate and human") and 7 ("real ≠ demo") and the §12.5 interaction defaults.

**Post-invite interstitial (closes the §7.4 loop)** — where the OAuth bot-invite returns to, instead of a stale `/servers` card.

```
┌───────────────────────────────────────────────┐
│              🦐  ✅                             │  centered, --bg-surface card on --bg-base
│   Shrimpy is now in  Ocean Crew!                │  --text-2xl --font-display
│   Let's get it set up — about 5 minutes.        │  --text-muted
│        [ Start setup → ]   [ Not now ]          │  primary → /dashboard/[id] (Setup, §A.2);
└───────────────────────────────────────────────┘  ghost "Not now" → /servers
   Reached after the invite returns + a /guilds re-check confirms bot_joined flipped (§A.1 join-detection).
   If the re-check hasn't landed yet: show a brief "Finishing up…" <Skeleton/> then resolve.
```

**Demo-mode banner (Principle 7; §7.2, Backlog #14)** — persistent while the session is unauthenticated / mock-backed (the `lib/api.ts` fallback). Never shown in a real session.

```
╔══════════════════════════════════════════════════════════════════════════════╗
║ 👁  Demo mode — sample data, changes won't be saved.   [ Log in with Discord ↗ ]║  full-width, --warning-muted bg,
╚══════════════════════════════════════════════════════════════════════════════╝  --warning text/border; sits ABOVE top bar
   Pinned across every screen in demo; the only signal that separates a sandbox tour from a real console.
```

**Access denied (403) — Level-2 deep-link or non-owner /admin hit (§A.4, §A.12)**

```
┌──────────────────────────────────┐
│            🔒                     │  <EmptyState/> variant, --bg-surface
│   You don't have access to this.  │  --text-primary
│   This area is for server admins.  │  --text-muted; plain language, no codes
│   [ Back to Tickets ]  [ Servers ]│  primary → a surface they CAN reach
└──────────────────────────────────┘
   Shown when the API returns 403 (server-side enforcement per Spec §7.3) — the UI hides nav, but never relies on hiding alone.
```

**Generic error + retry (§10.4 "user sees nothing" → fix)** — every data screen's error state.

```
┌──────────────────────────────────┐
│            ⚠                      │  --danger icon
│   Couldn't load this.             │  --text-primary
│   [ Try again ]                   │  primary; re-fires the request
└──────────────────────────────────┘
   Error boundary per data region (not a white screen); transient save errors use a <Toast/> with "Retry" instead (§A.9).
```

**Session expired (§7.2)** — `GET /auth/me` → 401 mid-session.

```
 <Toast/>:  ⚠ Your session expired — [ Log in again ]      → /login?reason=expired
   Bounce to /login with a friendly reason param, not a silent redirect or a broken page.
```

**Command palette (⌘K / Ctrl-K — Backlog #15, "Could")** — power-user jump across servers / screens / tickets.

```
┌─────────────────────────────────────────────┐
│ 🔎 Type a command or search…                 │  --bg-surface-elevated, --shadow-lg, centered overlay
├─────────────────────────────────────────────┤
│  Go to · Overview                            │  grouped results; ↑↓ to move, ↵ to run
│  Go to · Tickets                             │  --primary highlight on active row
│  Switch server · Gamer Guild                 │
│  Ticket · #0042 Billing                      │
└─────────────────────────────────────────────┘
   Optional v1; if shipped, it composes the SAME nav + <ServerSwitcher/> + ticket data already loaded — no new endpoints.
```

---

## Appendix B — Data Contract & Endpoint Coverage

Every screen in this journey, mapped to the [Technical Spec §4](./TECHNICAL_SPEC.md#4-rest-api-design) endpoint(s) that feed it — so implementation is never blocked guessing what to call, and any backend work the journey *adds* is explicit. **All endpoints below now exist in the spec** (the ticket sub-resources, `welcome/test`, and `health` were added alongside this appendix); the right-hand column flags anything that is new or computed. Per §14.7–§14.8, `messages` (staff-reply relay) and `participants` are **not** part of v1 scope — see B.2.

### B.1 Screen → endpoint matrix

| Screen (§) | Reads | Writes / actions | Notes |
|------------|-------|------------------|-------|
| `/servers` (A.1) | `GET /guilds` | — | Needs `bot_joined`, `access_level`, `member_count`, `icon` per guild (Spec §4.2 note) |
| Overview (A.2) | `GET /guilds/:id` (`setup` object), `GET …/stats`, `GET …/health` | deep-links to config screens | `setup` + `health` are the new computed reads; counts only, no chart (§14.5) |
| Ticket detail (A.3) | `GET …/tickets/:id` (messages + notes, read-only conversation) | `PATCH` (priority/claim), `…/close`, `…/reopen`, `…/notes` | read-only in v1 (§14.7) — no `messages`/`participants` write endpoints |
| Welcome (A.5) | `GET …/welcome`, `…/discord/channels`, `…/discord/roles`, `…/auto-roles` | `PUT/PATCH …/welcome`, `POST …/welcome/test`, `POST/DELETE …/auto-roles` | `welcome/test` **new** |
| Panels (A.6) | `GET …/panels` (+ categories), `…/discord/channels`, `…/discord/roles` | panel + category CRUD (Spec §4.3) | multi-button/select, multi-role, **and the configurable opening-message embed** (`ticketOpen*` fields) are payload depth, not new routes |
| Reaction Roles (A.7) | `GET …/reaction-roles`, `…/discord/emojis`, `…/discord/roles`, `…/health` | reaction-role CRUD (Spec §4.4) | full emoji picker uses existing `…/discord/emojis` |
| Settings (A.8) | `GET /guilds/:id` | `PATCH /guilds/:id` (prefix/language/log/auto-close) | language + auto-close are existing columns (`guilds.language`, `ticket_categories.auto_close_hours`) |
| Staff & Access (A.8) | `GET …/staff-roles`, `…/discord/roles` | `POST/DELETE …/staff-roles` | — |
| Tickets inbox (A.10) | `GET …/tickets?status&priority&categoryId&page` | `PATCH …/tickets/:id` (quick claim) | all existing |
| Transcripts (A.11) | `GET …/tickets?status=closed,archived`, `…/tickets/:id/transcript` | — | reuses tickets list; **no new list endpoint** |
| Admin apps (A.12) | `GET /admin/apps` | `POST/PUT/DELETE /admin/apps`, `…/reconnect` | owner-gated (Spec §4.9) |
| Post-invite (A.13) | `GET /guilds` (re-check) | — | join-detection poll |
| Demo banner / 403 / error / session (A.13) | `GET /auth/me` (401→expired) | — | client-side state from auth + HTTP status; demo data lives behind `/demo` (§14.1), never a silent fallback |

### B.2 Endpoints added for this journey (now in Spec §4)

| Endpoint | Powers | PRD story |
|----------|--------|-----------|
| `POST …/tickets/:id/notes`, `DELETE …/notes/:noteId` | Internal staff notes (A.3) — rows in `ticket_messages` with `is_staff_note=TRUE` | S-04 |
| `POST …/welcome/test` | "Send test to me" DM (A.5) | A-03 |
| `GET …/guilds/:id/health` | Bot health strip + role-height check (A.2, A.5, A.7) | A-09 / A-04 |
| `setup` object on `GET /guilds/:id` | First-run Setup checklist completion (A.2) | A-09 |

**Deferred/dropped per §14 decisions** — not built in v1: `POST …/tickets/:id/messages` (§14.7, read-only conversation) and `GET/POST/DELETE …/tickets/:id/participants` + the `ticket_participants` table (§14.8, covered by existing per-category support roles).

### B.3 Already-covered (no change needed — common misconception)

- **Priority (`S-02`)** → `tickets.priority` via `PATCH …/tickets/:id`.
- **Close with resolution note (`S-03`)** → `close_reason` via `POST …/tickets/:id/close`.
- **Claim (`S-08`)** → `tickets.claimed_by` via `PATCH …/tickets/:id`.
- **Internal-note storage** → already modeled (`ticket_messages.is_staff_note`); only the *write* endpoint was missing.
- **Per-category support roles, thread-vs-channel, panel style** → already columns on `ticket_categories` (`panel_style`, `ticket_destination`, …); Panels depth (A.6) is richer payloads, not new routes. These same support roles are also why participants (`S-05`) is deferred — §14.8.

> **Implementation takeaway:** the journey is overwhelmingly buildable on the existing API. The only genuinely new backend surface is **4 endpoints + 0 new tables** (B.2) — `messages`/`participants`/`ticket_participants` were scoped out by the §14 decisions, not deferred for lack of clarity.

---

*End of User Journey & UX Flow Specification — Shrimpy v1.0.0-draft*
