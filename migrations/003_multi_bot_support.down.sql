-- migrations/003_multi_bot_support.down.sql
-- Shrimpy Discord Bot — Rollback multi-bot application credentials & guild link

-- Drop guilds foreign key index and column
DROP INDEX IF EXISTS idx_guilds_discord_app_id;
ALTER TABLE guilds DROP COLUMN IF EXISTS discord_app_id;

-- Drop discord_apps table and trigger
DROP TRIGGER IF EXISTS set_updated_at_discord_apps ON discord_apps;
DROP TABLE IF EXISTS discord_apps;

-- Re-create bot_settings singleton table
CREATE TABLE bot_settings (
    id                          SMALLINT    PRIMARY KEY DEFAULT 1,
    discord_token_enc           BYTEA       NOT NULL,   -- AES-256-GCM encrypted bot token
    discord_client_id           VARCHAR(30) NOT NULL,
    discord_client_secret_enc   BYTEA       NOT NULL,   -- AES-256-GCM encrypted
    discord_redirect_uri        TEXT        NOT NULL,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT singleton CHECK (id = 1)
);

CREATE TRIGGER set_updated_at_bot_settings
    BEFORE UPDATE ON bot_settings
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
