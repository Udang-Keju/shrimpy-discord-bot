-- migrations/001_init.up.sql
-- Shrimpy Discord Bot — Initial Schema

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── guilds ───────────────────────────────────────────────────────────────────
CREATE TABLE guilds (
    guild_id        BIGINT      PRIMARY KEY,
    prefix          VARCHAR(10) NOT NULL DEFAULT '!',
    language        VARCHAR(10) NOT NULL DEFAULT 'en',
    bot_nickname    VARCHAR(32),
    log_channel_id  BIGINT,
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── users ────────────────────────────────────────────────────────────────────
CREATE TABLE users (
    user_id                     BIGINT      PRIMARY KEY,
    username                    VARCHAR(64) NOT NULL,
    discriminator               VARCHAR(4),
    avatar_hash                 VARCHAR(64),
    discord_access_token_enc    BYTEA,          -- AES-256-GCM encrypted
    discord_refresh_token_enc   BYTEA,          -- AES-256-GCM encrypted
    token_expires_at            TIMESTAMPTZ,
    last_seen                   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── staff_roles ──────────────────────────────────────────────────────────────
CREATE TABLE staff_roles (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id    BIGINT      NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    role_id     BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(guild_id, role_id)
);

-- ─── auto_roles ───────────────────────────────────────────────────────────────
CREATE TABLE auto_roles (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id    BIGINT      NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    role_id     BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(guild_id, role_id)
);

-- ─── welcome_config ───────────────────────────────────────────────────────────
CREATE TABLE welcome_config (
    guild_id        BIGINT      PRIMARY KEY REFERENCES guilds(guild_id) ON DELETE CASCADE,
    enabled         BOOLEAN     NOT NULL DEFAULT TRUE,
    -- DM welcome
    dm_message      TEXT,
    -- Channel welcome
    channel_id      BIGINT,
    channel_message TEXT,
    -- Embed customisation (color + media JSON)
    embed_color     INT,
    embed_media     JSONB,      -- { author, thumbnail, image, footer }
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── ticket_panels ────────────────────────────────────────────────────────────
CREATE TABLE ticket_panels (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id            BIGINT      NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    name                VARCHAR(64) NOT NULL,
    channel_id          BIGINT      NOT NULL,
    message_id          BIGINT,                 -- ID of the posted panel message
    panel_style         VARCHAR(16) NOT NULL DEFAULT 'buttons', -- 'buttons' | 'select_menu'
    embed_title         VARCHAR(256),
    embed_description   TEXT,
    embed_color         INT,
    embed_media         JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── ticket_categories ────────────────────────────────────────────────────────
CREATE TABLE ticket_categories (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    panel_id                UUID        NOT NULL REFERENCES ticket_panels(id) ON DELETE CASCADE,
    name                    VARCHAR(64) NOT NULL,
    emoji                   VARCHAR(32),
    button_label            VARCHAR(64) NOT NULL,
    button_style            VARCHAR(16) NOT NULL DEFAULT 'primary', -- 'primary'|'secondary'|'success'|'danger'
    button_description      VARCHAR(100),        -- used for select_menu style only
    button_order            SMALLINT    NOT NULL DEFAULT 0,
    ticket_destination      VARCHAR(16) NOT NULL DEFAULT 'thread',  -- 'thread' | 'channel'
    ticket_name_template    VARCHAR(64) NOT NULL DEFAULT '{category}-{number}',
    ticket_open_title       VARCHAR(256),
    ticket_open_message     TEXT,
    ticket_open_color       INT,
    ticket_open_media       JSONB,
    max_tickets_per_user    INT         NOT NULL DEFAULT 1,
    auto_close_hours        INT,                 -- NULL = never auto-close
    transcript_channel_id   BIGINT,
    allow_user_close        BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── tickets ──────────────────────────────────────────────────────────────────
CREATE TYPE ticket_status   AS ENUM ('open', 'claimed', 'closed', 'archived');
CREATE TYPE ticket_priority AS ENUM ('low', 'medium', 'high', 'urgent');

CREATE TABLE tickets (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id        BIGINT          NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    category_id     UUID            NOT NULL REFERENCES ticket_categories(id),
    channel_id      BIGINT,                     -- NULL if thread
    thread_id       BIGINT,                     -- NULL if channel
    opened_by       BIGINT          NOT NULL,
    claimed_by      BIGINT,
    status          ticket_status   NOT NULL DEFAULT 'open',
    priority        ticket_priority NOT NULL DEFAULT 'medium',
    close_reason    TEXT,
    auto_close_at   TIMESTAMPTZ,
    closed_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tickets_guild_status ON tickets(guild_id, status);
CREATE INDEX idx_tickets_category     ON tickets(category_id);
CREATE INDEX idx_tickets_auto_close   ON tickets(auto_close_at) WHERE status IN ('open', 'claimed');

-- ─── ticket_messages ──────────────────────────────────────────────────────────
CREATE TABLE ticket_messages (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID        NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    author_id       BIGINT      NOT NULL,
    author_username VARCHAR(64) NOT NULL,
    content         TEXT,
    is_staff_note   BOOLEAN     NOT NULL DEFAULT FALSE,
    attachments     JSONB,      -- array of { filename, url, size }
    sent_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ticket_messages_ticket ON ticket_messages(ticket_id, sent_at);

-- ─── reaction_role_messages ───────────────────────────────────────────────────
CREATE TABLE reaction_role_messages (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id            BIGINT      NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    channel_id          BIGINT      NOT NULL,
    message_id          BIGINT,
    embed_title         VARCHAR(256),
    embed_description   TEXT,
    embed_color         INT,
    embed_media         JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── reaction_role_emojis ─────────────────────────────────────────────────────
CREATE TABLE reaction_role_emojis (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id  UUID        NOT NULL REFERENCES reaction_role_messages(id) ON DELETE CASCADE,
    emoji       VARCHAR(64) NOT NULL,
    role_id     BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(message_id, emoji)
);

-- ─── Triggers: auto-update updated_at ─────────────────────────────────────────
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at_guilds
    BEFORE UPDATE ON guilds
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_welcome_config
    BEFORE UPDATE ON welcome_config
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_ticket_panels
    BEFORE UPDATE ON ticket_panels
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_ticket_categories
    BEFORE UPDATE ON ticket_categories
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_tickets
    BEFORE UPDATE ON tickets
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_reaction_role_messages
    BEFORE UPDATE ON reaction_role_messages
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
