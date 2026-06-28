-- migrations/005_category_handler_roles.up.sql
-- Shrimpy Discord Bot — Per-category ticket handler roles
--
-- Roles invited into a ticket's created channel/thread to handle it, scoped to
-- a specific category within a panel. Additive to panel_handler_roles.

CREATE TABLE category_handler_roles (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID        NOT NULL REFERENCES ticket_categories(id) ON DELETE CASCADE,
    role_id     BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(category_id, role_id)
);

CREATE INDEX idx_category_handler_roles_category_id ON category_handler_roles(category_id);
