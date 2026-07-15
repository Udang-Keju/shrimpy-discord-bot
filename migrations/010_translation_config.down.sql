-- migrations/010_translation_config.down.sql
-- Shrimpy Discord Bot — Rollback message auto-translation feature

DROP INDEX IF EXISTS idx_translation_reaction_emojis_guild_id;
DROP TABLE IF EXISTS translation_reaction_emojis;

DROP INDEX IF EXISTS idx_translation_channels_guild_id;
DROP TABLE IF EXISTS translation_channels;

DROP TRIGGER IF EXISTS set_updated_at_translation_config ON translation_config;
DROP TABLE IF EXISTS translation_config;
