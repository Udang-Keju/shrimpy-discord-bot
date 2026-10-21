-- migrations/012_clock_channels.down.sql
-- Shrimpy Discord Bot — Rollback clock channels

DROP TRIGGER IF EXISTS set_updated_at_clock_channels ON clock_channels;
DROP INDEX IF EXISTS idx_clock_channels_guild_id;
DROP TABLE IF EXISTS clock_channels;
