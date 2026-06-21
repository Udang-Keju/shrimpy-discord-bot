# Product Requirements Document (PRD)
## Project: **Shrimp** 🦐 — Go-Powered Discord Server Management & Help Desk Bot

> **Version**: 1.0.0-draft
> **Status**: In Review
> **Last Updated**: 2026-06-21
> **Author**: Engineering Team

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Goals & Objectives](#2-goals--objectives)
3. [Target Users & Personas](#3-target-users--personas)
4. [MVP Scope](#4-mvp-scope)
5. [User Stories](#5-user-stories)
6. [Feature List & Prioritization](#6-feature-list--prioritization)
7. [Success Metrics](#7-success-metrics)
8. [Assumptions & Constraints](#8-assumptions--constraints)
9. [Out-of-Scope Items](#9-out-of-scope-items)

---

## 1. Project Overview

**Shrimp** is a general-purpose Discord bot written in Go, designed to serve as the backbone of community server management and structured user support. It combines server lifecycle management (welcoming members, assigning roles) with a full-featured ticket/help-desk system powered by Discord's native UI components — slash commands, interactive buttons, and embedded messages.

Shrimp is designed to be deployed for a single server initially but is **architecturally multi-tenant from day one**, enabling straightforward onboarding of additional guilds without code changes. A companion **web-based admin dashboard** provides a graphical interface for server administrators and support staff to manage configuration, monitor tickets, and export transcripts — without ever needing to use Discord commands.

### Name Rationale

> **Shrimp** 🦐 — a playful, memorable, and unique name that stands out in any server's member list. While unconventional, the name is instantly recognizable, easy to mention (`@Shrimp`), and gives the bot a distinct personality. The shrimp motif carries through the web dashboard's design language — coral pinks, ocean teals, and warm sandy tones — forming a cohesive brand identity.

---

## 2. Goals & Objectives

| # | Goal | Rationale |
|---|------|-----------|
| G1 | Provide a structured, Discord-native ticket/helpdesk system | Replaces ad-hoc DMs and chaotic support channels |
| G2 | Automate new member onboarding (welcome messages, auto-roles) | Reduces manual admin overhead and improves first impressions |
| G3 | Offer a full-featured web dashboard for non-technical admins | Lowers barrier to configuration; no command memorization needed |
| G4 | Maintain multi-server readiness from the initial build | Avoids expensive architectural refactoring when scaling |
| G5 | Ensure all interactions are auditable (transcripts, logs) | Enables dispute resolution and compliance tracking |

---

## 3. Target Users & Personas

### 3.1 Persona: Server Administrator (Admin)

| Attribute | Detail |
|-----------|--------|
| **Role** | Guild owner or administrator with `Manage Server` / `Administrator` permissions |
| **Technical Level** | Low to Medium |
| **Primary Goals** | Configure the bot, set up ticket panels, define staff roles, manage welcome messages |
| **Pain Points** | Spending too much time on manual onboarding; no visibility into support requests |
| **Key Touchpoints** | Web Dashboard (primary), `/setup` slash commands, prefix commands |

### 3.2 Persona: Staff Member (Support Agent)

| Attribute | Detail |
|-----------|--------|
| **Role** | Moderator, support agent, or any user with a designated staff role |
| **Technical Level** | Low |
| **Primary Goals** | Claim tickets, respond to inquiries, set priorities, close resolved tickets, leave internal notes |
| **Pain Points** | No structured way to track who owns which ticket; messages getting lost |
| **Key Touchpoints** | Discord ticket channels/threads (primary), Web Dashboard (secondary) |

### 3.3 Persona: Regular Member / Inquirer

| Attribute | Detail |
|-----------|--------|
| **Role** | Standard server member needing assistance |
| **Technical Level** | Very Low |
| **Primary Goals** | Open a ticket easily, get a response, close the ticket when resolved |
| **Pain Points** | Not knowing where to get help; feeling ignored in busy servers |
| **Key Touchpoints** | Ticket panel buttons in Discord (primary), slash commands (secondary) |

---

## 4. MVP Scope

The MVP delivers the **core value proposition** of Shrimp: structured ticket management and automated member onboarding.

### 4.1 In-Scope for MVP

#### Server Management
- [x] Configurable welcome messages (DM and/or channel, with template variables)
- [x] Auto-role assignment on member join (one or more roles)
- [x] Reaction roles (users self-assign roles by reacting to an admin-configured message)

#### Ticket System
- [x] Ticket panels (customizable embeds using Buttons or Select Menu styles)
- [x] Multiple ticket categories per panel (max 3 for buttons, max 25 for select menu)
- [x] Ticket destination: private thread OR private channel (configurable per category)
- [x] Custom ticket channel/thread name templates (e.g., `{category}-{number}`)
- [x] Custom ticket opening embed messages per category
- [x] Role-restricted visibility (creator + staff roles only)
- [x] Ticket claiming by staff
- [x] Priority levels: Low, Medium, High, Urgent
- [x] Ticket closing by creator or staff
- [x] Transcript generation on close (optional log channel posting)
- [x] Auto-close after configurable inactivity duration
- [x] Per-user open ticket limit
- [x] Internal staff notes (invisible to inquirer)

#### Web Dashboard
- [x] Discord OAuth2 login for admins/staff
- [x] Ticket panel & category setup UI
- [x] Open ticket management (view, assign, close, reopen, archive)
- [x] Welcome message template editor
- [x] Auto-role configuration
- [x] Reaction role configuration
- [x] Basic server statistics
- [x] Per-server bot settings (prefix, language, bot nickname)
- [x] Transcript export

### 4.2 Out-of-Scope for MVP

See [Section 9](#9-out-of-scope-items) for the full list.

---

## 5. User Stories

### 5.1 Server Administrator Stories

| ID | Story | Priority |
|----|-------|----------|
| A-01 | As a **Server Admin**, I want to run a single setup command so that the bot configures itself with sensible defaults for my server. | Must Have |
| A-02 | As a **Server Admin**, I want to create ticket panels with custom embed messages and category options (buttons or select menus) so that members can open categorized support requests. | Must Have |
| A-03 | As a **Server Admin**, I want to configure a welcome message with variables like `{user}`, `{server}`, and `{membercount}` so that new members receive a personalized greeting. | Must Have |
| A-04 | As a **Server Admin**, I want to assign one or more roles to new members automatically so that I don't have to manually grant roles on every join. | Must Have |
| A-04b | As a **Server Admin**, I want to create reaction role messages so that members can self-assign roles by reacting to specific emojis. | Must Have |
| A-05 | As a **Server Admin**, I want to designate which roles count as "staff" for ticket handling so that only authorized users can manage tickets. | Must Have |
| A-06 | As a **Server Admin**, I want to choose whether tickets open as private threads or private channels so that I can match my server's organizational style. | Must Have |
| A-07 | As a **Server Admin**, I want to set a per-user open ticket limit so that one member can't flood the system with duplicate tickets. | Should Have |
| A-08 | As a **Server Admin**, I want to configure an auto-close duration for inactive tickets so that stale tickets don't pile up. | Should Have |
| A-09 | As a **Server Admin**, I want to view a dashboard with server-wide ticket statistics so that I can monitor support performance. | Should Have |
| A-10 | As a **Server Admin**, I want to manage all bot settings through a web UI so that I don't need to memorize Discord commands. | Should Have |
| A-11 | As a **Server Admin**, I want to designate a transcript log channel so that all closed ticket records are archived automatically. | Must Have |

### 5.2 Staff Member Stories

| ID | Story | Priority |
|----|-------|----------|
| S-01 | As a **Staff Member**, I want to claim a ticket so that my colleagues know I am handling it and don't duplicate effort. | Must Have |
| S-02 | As a **Staff Member**, I want to set a priority level on a ticket so that urgent issues are visually distinguishable and addressed first. | Must Have |
| S-03 | As a **Staff Member**, I want to close a ticket with a resolution note so that the inquirer understands the outcome. | Must Have |
| S-04 | As a **Staff Member**, I want to leave an internal note on a ticket that the inquirer cannot see so that I can communicate privately with other staff. | Must Have |
| S-05 | As a **Staff Member**, I want to add or remove users from a ticket so that additional members can assist with or observe the ticket. | Should Have |
| S-06 | As a **Staff Member**, I want to rename a ticket so that its channel/thread name reflects the issue clearly. | Could Have |
| S-07 | As a **Staff Member**, I want to generate a transcript at any time so that I can share the ticket log before closing. | Should Have |
| S-08 | As a **Staff Member**, I want to view all open tickets in the web dashboard so that I can prioritize my work without searching Discord manually. | Should Have |
| S-09 | As a **Staff Member**, I want to reopen a closed ticket so that I can address follow-up issues without making the member start over. | Should Have |

### 5.3 Regular Member / Inquirer Stories

| ID | Story | Priority |
|----|-------|----------|
| M-01 | As a **Member**, I want to open a ticket by clicking a button so that I don't need to know any commands. | Must Have |
| M-02 | As a **Member**, I want my ticket to be private so that only I and support staff can see it. | Must Have |
| M-03 | As a **Member**, I want to close my ticket when my issue is resolved so that I don't have to wait for a staff member to do it. | Must Have |
| M-04 | As a **Member**, I want to receive a welcome message when I join the server so that I know where to find help and what to do next. | Must Have |
| M-05 | As a **Member**, I want to receive a transcript of my closed ticket so that I have a record of the support I received. | Could Have |

---

## 6. Feature List & Prioritization

Using the **MoSCoW** prioritization framework:

### Must Have (MVP Blockers)

| Feature | Description |
|---------|-------------|
| Ticket Panel Setup | Create embed+button panels in any channel |
| Ticket Category Configuration | Multiple categories per panel, each independently configured |
| Private Ticket Creation | Open private thread or channel on button click |
| Role-restricted Visibility | Only creator + staff roles can see tickets |
| Ticket Claiming | Staff can claim ownership of a ticket |
| Ticket Priority Setting | Set Low / Medium / High / Urgent priority |
| Ticket Closing | Closeable by creator or staff; archives channel/thread |
| Transcript on Close | Generates message log; optionally posts to log channel |
| Welcome Messages | Configurable DM and/or channel welcome with template variables |
| Auto-Role Assignment | Assign roles automatically on member join |
| Staff Role Designation | Admin designates which roles are "staff" for ticket purposes |
| Internal Staff Notes | Staff-only notes inside tickets invisible to the inquirer |
| Discord OAuth2 Dashboard Login | Secure dashboard authentication |

### Should Have (High Value, Post-MVP OK)

| Feature | Description |
|---------|-------------|
| Per-user Ticket Limit | Prevent flooding by capping open tickets per user |
| Auto-close on Inactivity | Automatically close tickets with no activity after X hours/days |
| Add/Remove Users from Ticket | Manually expand ticket visibility to other users |
| Web Dashboard Ticket Management | View, manage, close, reopen via web UI |
| Server Statistics Page | Member count, ticket volume, resolution rate on dashboard |
| Transcript Export from Dashboard | Download transcripts as HTML or plaintext |
| Ticket Reopening | Reopen a closed ticket to continue the conversation |
| Bot Settings Page | Prefix, language, and custom bot nickname settings per server via dashboard |

### Could Have (Nice to Have)

| Feature | Description |
|---------|-------------|
| Ticket Renaming | Staff can rename the ticket channel/thread |
| Member Transcript DM | DM the transcript to the member on ticket close |
| Multi-language Support | Localized bot responses |
| Ticket Tags/Labels | Freeform tagging system for categorization |
| Ticket Rating System | Members rate their support experience on close |
| SLA Tracking | Track response time targets per priority level |
| Custom Ticket Embed Colors | Per-category embed color theming |

---

## 7. Success Metrics

> [!NOTE]
> These metrics apply once the bot is deployed to at least one active server for 30+ days.

| Metric | Target | Measurement Method |
|--------|--------|--------------------|
| Ticket Creation Success Rate | ≥ 99% of button clicks result in a ticket | Bot log analytics |
| Average Ticket Resolution Time | Baseline established in first 30 days | Ticket `opened_at` vs `closed_at` timestamps |
| Staff Claim Rate | ≥ 80% of tickets are claimed before closing | `claimed_by` field not null |
| Auto-close Rate | < 15% of tickets closed by auto-close | `close_reason` field in DB |
| Dashboard Adoption | ≥ 50% of admin configuration done via dashboard | Dashboard API request logs |
| Transcript Generation Uptime | ≥ 99.5% | Error rate monitoring |
| Welcome Message Delivery Rate | ≥ 98% of joins trigger a message | Bot log analytics |

---

## 8. Assumptions & Constraints

### Assumptions

- The bot will have the necessary Discord permissions on all guilds it joins (`Manage Channels`, `Manage Threads`, `Manage Roles`, `Send Messages`, `Embed Links`, `Read Message History`, etc.).
- Server admins are responsible for granting the bot proper permissions during onboarding.
- Discord's API rate limits are respected; the bot uses exponential backoff for retries.
- The Go backend (Railway Hobby) and Next.js dashboard (Vercel) are deployed as separate services; the dashboard communicates with the bot API over HTTPS.
- The bot is initially targeted at a single server but must not require code changes to support additional servers.
- PostgreSQL will be the sole persistent data store.

### Constraints

| Constraint | Impact |
|------------|--------|
| Discord API rate limits (50 req/s per bot globally) | Transcript generation must be batched; large channels may take time |
| Discord thread limit (1000 active threads per channel) | Ticket channels should be used when thread volume is high |
| Discord message history API (limited lookups per minute) | Transcript generation is asynchronous |
| Private threads require Discord Nitro boost level on some server tiers | Fallback to private channels must be available |
| Go binary deployment | Single compiled binary simplifies deployment but requires cross-compilation for CI/CD |

---

## 9. Out-of-Scope Items

The following items are explicitly **not** part of the MVP and will be considered for a future version:

| Item | Reason for Exclusion |
|------|---------------------|
| Moderation features (warn, kick, ban, mute) | Separate product concern; adds significant complexity |
| Leveling/XP system | Out of scope for help-desk focus |
| Music playback | Unrelated to server management/tickets |
| Giveaway system | Out of scope for MVP |
| Economy system (coins, shop) | Future consideration |
| Custom command builder | Significant complexity; post-MVP |
| Advanced analytics & reporting (charts, trends) | Dashboard v2 feature |
| Mobile-optimized dashboard | Responsive web design is considered; native app is not |
| Self-hosting documentation | Deferred to post-MVP |
| Stripe/payment integrations | Out of scope |
| AI-powered ticket routing | Post-MVP consideration |

---

*End of Product Requirements Document — Shrimp v1.0.0-draft*
