-- migrations/002_bot_settings.down.sql
-- Shrimpy Discord Bot — Rollback bot_settings singleton

DROP TRIGGER IF EXISTS set_updated_at_bot_settings ON bot_settings;
DROP TABLE IF EXISTS bot_settings;
