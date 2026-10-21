-- migrations/012_clock_channels.up.sql
-- Shrimpy Discord Bot — Clock channels (channels renamed to show date/time)

CREATE TABLE clock_channels (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    guild_id    BIGINT       NOT NULL REFERENCES guilds(guild_id) ON DELETE CASCADE,
    channel_id  BIGINT       NOT NULL,
    kind        VARCHAR(8)   NOT NULL,                  -- 'date' | 'time'
    format      VARCHAR(32)  NOT NULL,                  -- preset key, e.g. 'date_long', 'time_12h'
    timezone    VARCHAR(64)  NOT NULL DEFAULT 'UTC',    -- IANA zone name
    emoji       VARCHAR(32),                            -- NULL = default emoji for the kind
    managed     BOOLEAN      NOT NULL DEFAULT FALSE,    -- TRUE if the bot created the channel
    last_name   VARCHAR(100) NOT NULL DEFAULT '',       -- last name written, to skip no-op renames
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(guild_id, channel_id)
);

CREATE INDEX idx_clock_channels_guild_id ON clock_channels(guild_id);

CREATE TRIGGER set_updated_at_clock_channels
    BEFORE UPDATE ON clock_channels
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
