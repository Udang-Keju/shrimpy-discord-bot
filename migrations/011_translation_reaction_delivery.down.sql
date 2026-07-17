-- migrations/011_translation_reaction_delivery.down.sql

ALTER TABLE translation_config
    DROP COLUMN reaction_delivery;
