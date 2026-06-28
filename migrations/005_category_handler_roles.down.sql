-- migrations/005_category_handler_roles.down.sql
-- Shrimpy Discord Bot — Rollback per-category ticket handler roles

DROP INDEX IF EXISTS idx_category_handler_roles_category_id;
DROP TABLE IF EXISTS category_handler_roles;
