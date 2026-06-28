-- migrations/004_panel_handler_roles.down.sql
-- Shrimpy Discord Bot — Rollback per-panel ticket handler roles

DROP INDEX IF EXISTS idx_panel_handler_roles_panel_id;
DROP TABLE IF EXISTS panel_handler_roles;
