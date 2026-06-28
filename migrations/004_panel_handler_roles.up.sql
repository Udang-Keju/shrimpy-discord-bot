-- migrations/004_panel_handler_roles.up.sql
-- Shrimpy Discord Bot — Per-panel ticket handler roles
--
-- Roles invited into a ticket's created channel/thread to handle it,
-- distinct from staff_roles (dashboard access).

CREATE TABLE panel_handler_roles (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    panel_id    UUID        NOT NULL REFERENCES ticket_panels(id) ON DELETE CASCADE,
    role_id     BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(panel_id, role_id)
);

CREATE INDEX idx_panel_handler_roles_panel_id ON panel_handler_roles(panel_id);
