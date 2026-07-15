-- migrations/010_translation_config.up.sql
-- Shrimpy Discord Bot — Message auto-translation feature

-- Per-guild translation configuration.
CREATE TABLE translation_config (
    guild_id         BIGINT      PRIMARY KEY REFERENCES guilds(guild_id) ON DELETE CASCADE,
    enabled          BOOLEAN     NOT NULL DEFAULT FALSE,  -- master feature toggle
    auto_enabled     BOOLEAN     NOT NULL DEFAULT FALSE,  -- auto-translate in configured channels
    reaction_enabled BOOLEAN     NOT NULL DEFAULT FALSE,  -- translate on configured emoji reaction
    provider         VARCHAR(32) NOT NULL DEFAULT 'deepl',
    api_key_enc      BYTEA,                               -- AES-256-GCM encrypted engine API key
    endpoint_url     TEXT,                                -- for self-hosted engines (LibreTranslate)
    target_lang      VARCHAR(10),                         -- NULL = fall back to guilds.language
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER set_updated_at_translation_config
    BEFORE UPDATE ON translation_config
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- Channels where member messages are auto-translated.
CREATE TABLE translation_channels (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id            BIGINT      NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    channel_id          BIGINT      NOT NULL,
    target_lang_override VARCHAR(10),                     -- NULL = use config/guild target
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(guild_id, channel_id)
);

CREATE INDEX idx_translation_channels_guild_id ON translation_channels(guild_id);

-- Emojis that trigger translation when reacted to a message.
CREATE TABLE translation_reaction_emojis (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id            BIGINT      NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    emoji               VARCHAR(128) NOT NULL,            -- unicode char or "name:id" for custom
    target_lang_override VARCHAR(10),                     -- NULL = use config/guild target
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(guild_id, emoji)
);

CREATE INDEX idx_translation_reaction_emojis_guild_id ON translation_reaction_emojis(guild_id);
