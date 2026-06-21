-- migrations/003_multi_bot_support.up.sql
-- Shrimpy Discord Bot — Multi-bot application credentials & guild link

-- Drop old bot_settings trigger and table
DROP TRIGGER IF EXISTS set_updated_at_bot_settings ON bot_settings;
DROP TABLE IF EXISTS bot_settings;

-- Create discord_apps table
CREATE TABLE discord_apps (
    id                          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name                        VARCHAR(255)NOT NULL,
    discord_token_enc           BYTEA       NOT NULL,   -- AES-256-GCM encrypted bot token
    discord_client_id           VARCHAR(30) NOT NULL UNIQUE,
    discord_client_secret_enc   BYTEA       NOT NULL,   -- AES-256-GCM encrypted
    discord_redirect_uri        TEXT        NOT NULL,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Trigger for auto-updating updated_at
CREATE TRIGGER set_updated_at_discord_apps
    BEFORE UPDATE ON discord_apps
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- Add discord_app_id to guilds
ALTER TABLE guilds ADD COLUMN discord_app_id UUID REFERENCES discord_apps(id) ON DELETE SET NULL;

-- Index for foreign key lookups
CREATE INDEX idx_guilds_discord_app_id ON guilds(discord_app_id);
