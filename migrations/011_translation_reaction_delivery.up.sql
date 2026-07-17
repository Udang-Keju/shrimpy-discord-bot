-- migrations/011_translation_reaction_delivery.up.sql
-- Shrimpy Discord Bot — reaction-translate delivery mode (channel reply vs DM)

ALTER TABLE translation_config
    ADD COLUMN reaction_delivery VARCHAR(10) NOT NULL DEFAULT 'channel'
        CHECK (reaction_delivery IN ('channel', 'dm'));
