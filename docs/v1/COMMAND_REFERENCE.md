# Bot Command Reference
## Project: **Shrimpy** 🦐 — Discord Bot Command Reference

> **Version**: 1.0.0-draft
> **Last Updated**: 2026-06-21
> **Prefix**: `!` (configurable per server)
> **Slash Commands**: All slash commands work server-wide without prefix configuration.

---

## Table of Contents

1. [Permission Levels](#permission-levels)
2. [General Commands](#1-general-commands)
3. [Setup & Configuration Commands](#2-setup--configuration-commands)
4. [Ticket Panel Commands](#3-ticket-panel-commands)
5. [Ticket Management Commands](#4-ticket-management-commands)
6. [Staff Commands](#5-staff-commands)
7. [Admin & Debug Commands](#6-admin--debug-commands)
8. [Button Interactions](#7-button-interactions)
9. [Context Menu Commands](#8-context-menu-commands)
10. [Template Variables Reference](#9-template-variables-reference)

---

## Permission Levels

> [!IMPORTANT]
> All permission checks are enforced by the bot. Discord's native permission system applies in addition to bot-level checks.

| Level | Symbol | Who qualifies |
|-------|--------|--------------|
| **Everyone** | 🌐 | Any server member |
| **Ticket Creator** | 🎫 | The user who opened that specific ticket |
| **Staff** | 🛡️ | Any member with a role designated as "staff" via `/staff add` |
| **Administrator** | ⚙️ | Members with Discord's `Administrator` or `Manage Server` permission |
| **Bot Owner** | 🤖 | The bot's owner (configured in environment) |

---

## 1. General Commands

### `/help`

| Field | Detail |
|-------|--------|
| **Syntax** | `/help [category]` |
| **Description** | Displays the help menu. Optionally filter by command category. |
| **Permission** | 🌐 Everyone |
| **Response** | Embed with category buttons; clicking a category shows commands in that group. |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `category` | String (choice) | Optional | One of: `general`, `setup`, `tickets`, `staff`, `admin` |

**Example:**
```
/help
/help category:tickets
```

---

### `/info`

| Field | Detail |
|-------|--------|
| **Syntax** | `/info` |
| **Description** | Displays information about the bot — version, uptime, guild count, and links. |
| **Permission** | 🌐 Everyone |
| **Response** | Embed with bot stats. |

**Example:**
```
/info
```

---

### `/ping`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ping` |
| **Description** | Check the bot's responsiveness and latency to the Discord gateway. |
| **Permission** | 🌐 Everyone |
| **Response** | `🏓 Pong! Gateway latency: **42ms** \| API latency: **78ms**` |

**Example:**
```
/ping
```

---

## 2. Setup & Configuration Commands

> [!NOTE]
> Setup commands require `Manage Server` or `Administrator` permission. They can also be run using the prefix (e.g., `!setup`) during initial bot configuration when slash commands may not yet be registered.

### `/setup`

| Field | Detail |
|-------|--------|
| **Syntax** | `/setup` |
| **Description** | Runs the interactive server setup wizard. Guides the admin through initial configuration: prefix, log channel, staff roles, welcome message, and auto-roles. |
| **Permission** | ⚙️ Administrator |
| **Response** | Step-by-step embed with button navigation. |
| **Prefix Alias** | `!setup` |

**Example:**
```
/setup
!setup
```

---

### `/setup welcome`

| Field | Detail |
|-------|--------|
| **Syntax** | `/setup welcome` |
| **Description** | Opens the welcome message configuration wizard. Set DM template, channel template, and target channel. |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `enabled` | Boolean | Required | Enable or disable welcome messages |
| `dm-message` | String | Optional | Template for DM welcome message. Supports template variables. |
| `channel` | Channel | Optional | Channel to post public welcome message in |
| `channel-message` | String | Optional | Template for channel welcome message. Supports template variables. |
| `embed-color` | String | Optional | Hex color for the welcome embed's color bar (e.g. `#FF7B6B`) |
| `embed-author-name` | String | Optional | Author name shown at the top of the welcome embed |
| `embed-author-icon-url` | String | Optional | `https://` URL for the author icon image |
| `embed-thumbnail-url` | String | Optional | `https://` URL for the thumbnail image (top-right of embed) |
| `embed-image-url` | String | Optional | `https://` URL for the large image at the bottom of the embed |
| `embed-footer-text` | String | Optional | Footer text shown at the bottom of the welcome embed |
| `embed-footer-icon-url` | String | Optional | `https://` URL for the footer icon image |

> [!NOTE]
> All `*-url` parameters must use **`https://`**. Omitted image fields are simply not rendered in the embed.

**Example:**
```
/setup welcome enabled:true dm-message:"Welcome to {server}, {user}! You are member #{membercount}." channel:#welcome channel-message:"Everyone welcome {mention} to the server! 🎉" embed-color:#FF7B6B embed-author-name:"Welcome to {server}!" embed-thumbnail-url:https://example.com/server-logo.png embed-footer-text:"Joined {date}"
```

---

### `/setup autorole`

| Field | Detail |
|-------|--------|
| **Syntax** | `/setup autorole <action> <role>` |
| **Description** | Add or remove a role from the auto-role list. Members are assigned all auto-roles upon joining. |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | String (choice) | Required | `add` or `remove` |
| `role` | Role | Required | The role to add or remove from the auto-role list |

**Example:**
```
/setup autorole action:add role:@Members
/setup autorole action:remove role:@Members
```

---

### `/set prefix`

| Field | Detail |
|-------|--------|
| **Syntax** | `/set prefix <prefix>` |
| **Description** | Change the bot's command prefix for this server. |
| **Permission** | ⚙️ Administrator |
| **Prefix Alias** | `!set prefix <new_prefix>` |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `prefix` | String | Required | The new prefix (1–5 characters). Cannot contain spaces. |

**Example:**
```
/set prefix prefix:?
!set prefix ?
```

---

### `/set language`

| Field | Detail |
|-------|--------|
| **Syntax** | `/set language <language>` |
| **Description** | Set the bot's response language for this server. |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `language` | String (choice) | Required | Language code: `en` (English), `es` (Spanish), `fr` (French), `de` (German) |

**Example:**
```
/set language language:en
```

---

### `/set logchannel`

| Field | Detail |
|-------|--------|
| **Syntax** | `/set logchannel <channel>` |
| **Description** | Set the server-wide default channel where ticket transcripts and bot audit logs are posted. |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel` | Channel | Required | The text channel for logs |

**Example:**
```
/set logchannel channel:#bot-logs
```

---

### `/set nickname`

| Field | Detail |
|-------|--------|
| **Syntax** | `/set nickname <name>` |
| **Description** | Sets the bot's display name (nickname) in this server. Use `/set nickname reset` to revert to the global name **"Shrimpy"**. |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | Required | The nickname to display in this server. Max 32 characters. Pass `reset` to clear the custom nickname and revert to "Shrimpy". |

**Example:**
```
/set nickname ModBot
```
> Bot will now appear as **ModBot** in this server. Reverting: `/set nickname reset`

> [!NOTE]
> The nickname is stored in the database and **automatically reapplied** if the bot leaves and rejoins the server. The bot's global application name ("Shrimpy") is unchanged on Discord's side.

---

## 3. Ticket Panel & Category Commands

### `/ticket panel create`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket panel create` |
| **Description** | Create a new ticket panel (the embed message) in the specified channel. |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | Required | Internal panel name (e.g., "Main Support Panel") |
| `channel` | Channel | Required | Channel where the panel embed is posted |
| `style` | String (choice) | Optional | `buttons` (max 3 categories) or `select_menu` (max 25 categories). Default: `buttons` |
| `embed-title` | String | Optional | Title of the panel embed. Default: "Open a Ticket" |
| `embed-description` | String | Optional | Description text in the panel embed |
| `embed-color` | String | Optional | Hex color for the embed's left color bar (e.g. `#FF7B6B`). Converted to decimal internally. |
| `embed-author-name` | String | Optional | Text shown in the embed author line (top of embed) |
| `embed-author-icon-url` | String | Optional | `https://` URL for the small icon left of the author name |
| `embed-author-url` | String | Optional | Hyperlink URL for the author name (makes it clickable) |
| `embed-thumbnail-url` | String | Optional | `https://` URL for the small image shown top-right of the embed |
| `embed-image-url` | String | Optional | `https://` URL for the large image shown at the bottom of the embed |
| `embed-footer-text` | String | Optional | Text shown in the footer line (bottom of embed) |
| `embed-footer-icon-url` | String | Optional | `https://` URL for the small icon left of the footer text |

> [!NOTE]
> All `*-url` parameters must use **`https://`** and point to a publicly accessible image (jpg, png, gif, webp). Discord proxies all images through its own CDN — no self-hosting needed.

**Example:**
```
/ticket panel create name:"Main Support Panel" channel:#support style:buttons embed-title:"Need Help?" embed-description:"Click below to open a support ticket." embed-color:#FF7B6B
```

---

### `/ticket panel edit`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket panel edit <panel>` |
| **Description** | Edit an existing ticket panel's configuration (embed details or style). |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `panel` | String (autocomplete) | Required | The panel to edit (shows existing panels) |
| `style` | String (choice) | Optional | Switch between `buttons` and `select_menu` |
| `embed-title` | String | Optional | New embed title |
| `embed-description` | String | Optional | New embed description |
| `embed-color` | String | Optional | New hex color for the embed's color bar |
| `embed-author-name` | String | Optional | New author name text |
| `embed-author-icon-url` | String | Optional | New `https://` URL for the author icon |
| `embed-thumbnail-url` | String | Optional | New `https://` URL for the thumbnail image |
| `embed-image-url` | String | Optional | New `https://` URL for the bottom image |
| `embed-footer-text` | String | Optional | New footer text |

---

### `/ticket panel delete`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket panel delete <panel>` |
| **Description** | Delete a ticket panel and all its categories. The embed message is removed from Discord. |
| **Permission** | ⚙️ Administrator |

---

### `/ticket category add`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket category add <panel>` |
| **Description** | Add a new category to an existing panel. Enforces panel limits (3 for buttons, 25 for select menu). |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `panel` | String (autocomplete) | Required | The panel to add this category to |
| `name` | String | Required | Internal category name (e.g., "Tech Support") |
| `label` | String | Required | Label shown on the button or select menu (max 80 chars) |
| `destination` | String (choice) | Required | `thread` or `channel` — how tickets are created |
| `emoji` | String | Optional | Emoji to prefix the label (e.g., `🛠️`) |
| `button-style` | String (choice) | Optional | `primary`, `secondary`, `success`, `danger`. (Only visible on button panels) |
| `description` | String | Optional | Description text. (Only visible on select menu panels) |
| `name-template` | String | Optional | Channel/thread name format. Default: `ticket-{number}`. Variables: `{category}`, `{username}`, etc. |
| `open-title` | String | Optional | Title of the embed shown when ticket opens |
| `open-message` | String | Optional | Message shown when ticket opens. Variables: `{mention}`, etc. |
| `max-tickets` | Integer | Optional | Max open tickets per user for this category (default: 1) |
| `auto-close-hours` | Integer | Optional | Hours of inactivity before auto-close (0 = disabled) |

**Example:**
```
/ticket category add panel:"Main Support Panel" name:"Bug Report" label:"Bug Report" emoji:🐛 destination:thread name-template:"bug-{number}" open-title:"New Bug Report" open-message:"Hey {mention}, please describe the bug."
```

---

### `/ticket category edit`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket category edit <category>` |
| **Description** | Edit an existing category's properties. Same parameters as `/ticket category add`. |
| **Permission** | ⚙️ Administrator |

---

### `/ticket category remove`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket category remove <category>` |
| **Description** | Remove a category from a panel. Does not affect existing open tickets. |
| **Permission** | ⚙️ Administrator |

---

### `/ticket panel list`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket panel list` |
| **Description** | List all configured ticket categories and their settings for this server. |
| **Permission** | ⚙️ Administrator |

**Example:**
```
/ticket panel list
```

---

## 4. Reaction Role Commands

### `/reactionrole create`

| Field | Detail |
|-------|--------|
| **Syntax** | `/reactionrole create` |
| **Description** | Create a new reaction role message (embed) in a specific channel. |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel` | Channel | Required | The channel where the message will be posted |
| `embed-title` | String | Optional | Title of the embed |
| `embed-description` | String | Optional | Description text of the embed |
| `embed-color` | String | Optional | Hex color code |

---

### `/reactionrole edit`

| Field | Detail |
|-------|--------|
| **Syntax** | `/reactionrole edit <message>` |
| **Description** | Edit the embed properties of an existing reaction role message. |
| **Permission** | ⚙️ Administrator |

---

### `/reactionrole delete`

| Field | Detail |
|-------|--------|
| **Syntax** | `/reactionrole delete <message>` |
| **Description** | Delete a reaction role message. Does NOT remove roles that were already assigned to users. |
| **Permission** | ⚙️ Administrator |

---

### `/reactionrole add-role`

| Field | Detail |
|-------|--------|
| **Syntax** | `/reactionrole add-role <message> <emoji> <role>` |
| **Description** | Add a new emoji-to-role mapping to a reaction role message. The bot will automatically react to the message with the emoji to make it easy for users to click. |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `message` | String (autocomplete)| Required | The reaction role message to add this to |
| `emoji` | String | Required | The emoji users need to click |
| `role` | Role | Required | The role to assign |

---

### `/reactionrole remove-role`

| Field | Detail |
|-------|--------|
| **Syntax** | `/reactionrole remove-role <message> <emoji>` |
| **Description** | Remove an emoji-to-role mapping from a reaction role message. The bot will remove its reaction from the message. Does NOT remove the role from users who already have it. |
| **Permission** | ⚙️ Administrator |

---

## 5. Ticket Management Commands

> [!NOTE]
> Ticket management commands are typically run **inside** an active ticket channel or thread. The bot auto-detects the current ticket from the channel.

### `/ticket close`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket close [reason]` |
| **Description** | Close the current ticket. The channel/thread is locked, a transcript is generated, and a closing embed is posted. |
| **Permission** | 🎫 Ticket Creator OR 🛡️ Staff |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `reason` | String | Optional | Reason for closing the ticket (included in transcript and log) |

**Example:**
```
/ticket close
/ticket close reason:"Issue resolved — user confirmed fix works."
```

---

### `/ticket resolve`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket resolve` |
| **Description** | Mark the current ticket as resolved. Unlike `/ticket close`, the channel/thread stays fully open and no transcript is generated — it's a lightweight "handled" marker. An auto-close timer starts (using the category's auto-close window) so the ticket closes automatically if nobody responds further. |
| **Permission** | 🛡️ Staff |

**Example:**
```
/ticket resolve
```

---

### `/ticket claim`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket claim` |
| **Description** | Claim the current ticket, assigning yourself as the responsible staff member. The ticket embed is updated to show the claim. |
| **Permission** | 🛡️ Staff |

**Example:**
```
/ticket claim
```

---

### `/ticket unclaim`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket unclaim` |
| **Description** | Release your claim on the current ticket, returning it to an unclaimed state. |
| **Permission** | 🛡️ Staff (only the current claimant) |

**Example:**
```
/ticket unclaim
```

---

### `/ticket priority`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket priority <level>` |
| **Description** | Set the priority level of the current ticket. Updates the ticket embed with a priority badge. |
| **Permission** | 🛡️ Staff |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `level` | String (choice) | Required | `low`, `medium`, `high`, or `urgent` |

**Example:**
```
/ticket priority level:urgent
/ticket priority level:low
```

---

### `/ticket add-user`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket add-user <user>` |
| **Description** | Add a user to the current ticket, granting them visibility and the ability to participate. |
| **Permission** | 🛡️ Staff |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `user` | User | Required | The Discord user to add to the ticket |

**Example:**
```
/ticket add-user user:@JaneSmith
```

---

### `/ticket remove-user`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket remove-user <user>` |
| **Description** | Remove a user from the current ticket, revoking their visibility. Cannot remove the original ticket creator. |
| **Permission** | 🛡️ Staff |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `user` | User | Required | The Discord user to remove from the ticket |

**Example:**
```
/ticket remove-user user:@JaneSmith
```

---

### `/ticket rename`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket rename <name>` |
| **Description** | Rename the current ticket's channel or thread to better describe the issue. |
| **Permission** | 🛡️ Staff |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | Required | New channel/thread name (max 100 chars; spaces replaced with hyphens) |

**Example:**
```
/ticket rename name:"login-page-500-error"
```

---

### `/ticket transcript`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket transcript [format]` |
| **Description** | Generate and send a transcript of the current ticket immediately (without closing). Transcript is posted as a file attachment or sent to the log channel. |
| **Permission** | 🛡️ Staff |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `format` | String (choice) | Optional | `txt` (plain text) or `html` (formatted HTML). Default: `html` |

**Example:**
```
/ticket transcript
/ticket transcript format:txt
```

---

### `/ticket note`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket note <message>` |
| **Description** | Post an internal staff-only note in the ticket. The note is displayed as a distinct embed and is **not visible to the ticket creator** (filtered from member-facing transcripts). |
| **Permission** | 🛡️ Staff |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `message` | String | Required | The internal note content |

**Example:**
```
/ticket note message:"Spoke to billing team — refund approved. Waiting on confirmation email."
```

---

### `/ticket reopen`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket reopen` |
| **Description** | Reopen a closed ticket, unlocking the channel/thread and resetting the auto-close timer. Must be run in the closed ticket's channel. |
| **Permission** | 🛡️ Staff |

**Example:**
```
/ticket reopen
```

---

### `/ticket info`

| Field | Detail |
|-------|--------|
| **Syntax** | `/ticket info` |
| **Description** | Display a summary embed for the current ticket: ID, category, status, priority, opened by, claimed by, creation time. |
| **Permission** | 🌐 Everyone (in the ticket channel) |

**Example:**
```
/ticket info
```

---

## 5. Staff Commands

### `/staff add`

| Field | Detail |
|-------|--------|
| **Syntax** | `/staff add <role>` |
| **Description** | Designate a Discord role as a "staff" role for ticket management purposes. Members with this role gain access to all staff-level ticket commands. |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `role` | Role | Required | The role to add as a staff role |

**Example:**
```
/staff add role:@Support Team
/staff add role:@Moderators
```

---

### `/staff remove`

| Field | Detail |
|-------|--------|
| **Syntax** | `/staff remove <role>` |
| **Description** | Remove a role from the staff role list. Members with this role will lose staff-level ticket access. |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `role` | Role | Required | The role to remove from staff |

**Example:**
```
/staff remove role:@Support Team
```

---

### `/staff list`

| Field | Detail |
|-------|--------|
| **Syntax** | `/staff list` |
| **Description** | List all roles currently designated as staff for this server. |
| **Permission** | ⚙️ Administrator |

**Example:**
```
/staff list
```

---

## 6. Admin & Debug Commands

### `/botinfo`

| Field | Detail |
|-------|--------|
| **Syntax** | `/botinfo` |
| **Description** | Display detailed technical information about the bot: version, Go runtime version, uptime, memory usage, guild count, total tickets handled, database connection status. |
| **Permission** | ⚙️ Administrator |
| **Global Name** | The bot's global (application) name is **Shrimpy** 🦐. |
| **Per-Server Name** | Configurable via `/set nickname`. Defaults to **"Shrimpy"** if no custom nickname is set for this server. |

**Example:**
```
/botinfo
```

---

### `/reset config`

| Field | Detail |
|-------|--------|
| **Syntax** | `/reset config <target>` |
| **Description** | Reset a specific configuration section to its default values. A confirmation prompt is shown before executing. |
| **Permission** | ⚙️ Administrator |

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `target` | String (choice) | Required | What to reset: `welcome`, `autoroles`, `staffroles`, `prefix`, `all` |

> [!CAUTION]
> `reset config target:all` will delete ALL bot configuration for this server. This cannot be undone.

**Example:**
```
/reset config target:welcome
/reset config target:all
```

---

### `/diagnostics`

| Field | Detail |
|-------|--------|
| **Syntax** | `/diagnostics` |
| **Description** | Run a configuration health check. Validates that the bot has the necessary permissions in all configured channels (ticket panel channels, log channels, etc.) and reports any issues. |
| **Permission** | ⚙️ Administrator |

**Example:**
```
/diagnostics
```

---

## 7. Button Interactions

Button interactions are triggered by clicking buttons on bot-generated embed messages. They do not require typed commands.

### 7.1 Ticket Panel Buttons

These buttons appear on the panel embed posted by `/ticket panel create`.

| Button Label | Custom ID Pattern | Action | Visible To |
|--------------|------------------|--------|------------|
| `[category button label]` | `ticket:open:{categoryId}` | Opens a new ticket in the configured destination (thread or channel) | 🌐 Everyone |

**Behavior on click:**
1. Bot checks if the user has reached the per-category ticket limit.
2. If limit reached, bot sends an ephemeral error message to the user.
3. Otherwise, creates a private thread or channel with proper permission overrides.
4. Sends an opening embed in the new ticket with ticket details and action buttons.
5. Pings all staff roles (if configured).

---

### 7.2 Ticket Action Buttons

These buttons appear on the **opening embed** inside every new ticket channel/thread.

| Button Label | Custom ID Pattern | Action | Visible To |
|--------------|------------------|--------|------------|
| `✅ Close Ticket` | `ticket:close:{ticketId}` | Closes the ticket (prompts for confirmation) | 🎫 Creator + 🛡️ Staff |
| `🙋 Claim` | `ticket:claim:{ticketId}` | Claims the ticket for the clicking staff member | 🛡️ Staff only |
| `⚡ Set Priority` | `ticket:priority:{ticketId}` | Opens a select menu to set priority | 🛡️ Staff only |
| `📋 Transcript` | `ticket:transcript:{ticketId}` | Generates and sends a transcript immediately | 🛡️ Staff only |

---

### 7.3 Ticket Close Confirmation

When a user clicks the **Close Ticket** button, a confirmation embed appears with:

| Button | Custom ID Pattern | Action |
|--------|------------------|--------|
| `✅ Confirm Close` | `ticket:close:confirm:{ticketId}` | Finalizes ticket close |
| `❌ Cancel` | `ticket:close:cancel:{ticketId}` | Dismisses the confirmation |

---

### 7.4 Priority Select Menu

When the **Set Priority** button is clicked, a select menu appears with:

| Option | Value | Emoji |
|--------|-------|-------|
| Low | `low` | 🟢 |
| Medium | `medium` | 🟡 |
| High | `high` | 🟠 |
| Urgent | `urgent` | 🔴 |

---

## 8. Context Menu Commands

Context menu commands appear when a user right-clicks on a **message** or **user** within Discord. Access via: `Apps` → `[Command Name]`.

### 8.1 Message Context Menu Commands

#### "📩 Create Ticket from Message"

| Field | Detail |
|-------|--------|
| **Type** | Message Context Menu |
| **Description** | Right-click any message to open a ticket referencing that message. The referenced message content is quoted in the ticket's opening embed. |
| **Permission** | 🌐 Everyone |
| **Use case** | A member wants to escalate a message they saw (e.g., a bug report posted in a general channel). |

**Behavior:**
1. A category selection ephemeral menu appears (if multiple categories exist).
2. User selects a category.
3. Ticket is created with the original message's content and a link quoted in the opening embed.

---

### 8.2 User Context Menu Commands

#### "🎫 View User's Tickets"

| Field | Detail |
|-------|--------|
| **Type** | User Context Menu |
| **Description** | Right-click a server member to see a list of their current open and recently closed tickets. |
| **Permission** | 🛡️ Staff |
| **Response** | Ephemeral embed listing up to 10 most recent tickets for that user, with status, category, and links. |

---

#### "👤 User Ticket Stats"

| Field | Detail |
|-------|--------|
| **Type** | User Context Menu |
| **Description** | Right-click a server member to see their ticket history statistics: total tickets opened, closed, and average resolution time. |
| **Permission** | ⚙️ Administrator |
| **Response** | Ephemeral embed with per-user statistics. |

---

## 9. Template Variables Reference

Welcome messages and other configurable text fields support the following variables. Variables are replaced with their actual values when the message is sent.

| Variable | Replaced With | Example Output |
|----------|--------------|----------------|
| `{user}` | The member's username (without discriminator) | `SalmaB` |
| `{mention}` | A Discord mention of the member | `@SalmaB` |
| `{server}` | The server's name | `My Awesome Server` |
| `{membercount}` | The server's current total member count | `1,542` |
| `{date}` | Today's date (YYYY-MM-DD) | `2026-06-21` |
| `{time}` | Current time in UTC (HH:MM) | `10:57 UTC` |
| `{id}` | The member's Discord user ID (snowflake) | `123456789012345678` |

**Example welcome message using variables:**
```
Welcome to {server}, {mention}! 🎉
You are our **{membercount}th** member.
Head to #rules to get started, and feel free to open a ticket if you need help!
```

### 9.1 Ticket-context variables

Ticket **opening messages** and **name templates** (the embed/text the bot posts when a member clicks a panel button or picks a select option — dashboard editor: USER_JOURNEY §A.6) support all of the member/server variables above **plus** the following ticket-context tokens:

| Variable | Replaced With | Example Output |
|----------|--------------|----------------|
| `{category}` | The category the ticket was opened under | `Billing` |
| `{number}` | The ticket's sequential number within the guild | `42` |

**Example opening message:** `A billing specialist will be with you shortly, {mention}. Tell us your order number.` → resolves the mention against the member who opened the ticket, with `{category}`/`{number}` available for the title, body, and name template.

---

## Quick Reference Summary

### By Permission Level

| 🌐 Everyone | 🎫 Creator + 🛡️ Staff | 🛡️ Staff Only | ⚙️ Admin Only |
|-------------|----------------------|--------------|--------------|
| `/help` | `/ticket close` | `/ticket claim` | `/setup` |
| `/info` | Close Button | `/ticket unclaim` | `/setup welcome` |
| `/ping` | | `/ticket priority` | `/setup autorole` |
| `/ticket info` | | `/ticket add-user` | `/set prefix` |
| Panel Buttons (open) | | `/ticket remove-user` | `/set language` |
| | | `/ticket rename` | `/set nickname` |
| | | `/ticket transcript` | `/ticket panel create` |
| | | `/ticket note` | `/ticket panel edit` |
| | | `/ticket reopen` | `/ticket panel delete` |
| | | Claim / Priority Buttons | `/ticket category add` |
| | | "View User's Tickets" | `/ticket category edit` |
| | | | `/ticket category remove` |
| | | | `/reactionrole create` |
| | | | `/reactionrole edit` |
| | | | `/reactionrole delete` |
| | | | `/reactionrole add-role` |
| | | | `/reactionrole remove-role` |
| | | | `/staff add` |
| | | | `/staff remove` |
| | | | `/staff list` |
| | | | `/botinfo` |
| | | | `/reset config` |
| | | | `/diagnostics` |
| | | | "User Ticket Stats" |

---

*End of Bot Command Reference — Shrimpy v1.0.0-draft*
