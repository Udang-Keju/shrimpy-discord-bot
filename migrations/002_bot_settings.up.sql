-- migrations/002_bot_settings.up.sql
-- Shrimpy Discord Bot — Bot application credential settings (singleton)

-- ─── bot_settings ──────────────────────────────────────────────────────────────
-- This is a singleton table (always exactly one row with id = 1).
-- It stores all Discord credentials that can be updated from the dashboard
-- without needing to change Railway environment variables.
-- All sensitive values are encrypted with AES-256-GCM at rest.
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
