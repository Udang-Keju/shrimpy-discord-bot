-- migrations/001_init.down.sql
-- Shrimpy Discord Bot — Rollback Initial Schema

DROP TRIGGER IF EXISTS set_updated_at_reaction_role_messages ON reaction_role_messages;
DROP TRIGGER IF EXISTS set_updated_at_tickets ON tickets;
DROP TRIGGER IF EXISTS set_updated_at_ticket_categories ON ticket_categories;
DROP TRIGGER IF EXISTS set_updated_at_ticket_panels ON ticket_panels;
DROP TRIGGER IF EXISTS set_updated_at_welcome_config ON welcome_config;
DROP TRIGGER IF EXISTS set_updated_at_guilds ON guilds;
DROP FUNCTION IF EXISTS trigger_set_updated_at();

DROP TABLE IF EXISTS reaction_role_emojis;
DROP TABLE IF EXISTS reaction_role_messages;
DROP TABLE IF EXISTS ticket_messages;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS ticket_categories;
DROP TABLE IF EXISTS ticket_panels;
DROP TABLE IF EXISTS welcome_config;
DROP TABLE IF EXISTS auto_roles;
DROP TABLE IF EXISTS staff_roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS guilds;

DROP TYPE IF EXISTS ticket_priority;
DROP TYPE IF EXISTS ticket_status;
