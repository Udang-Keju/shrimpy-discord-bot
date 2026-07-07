-- migrations/009_ticket_auto_close_index_resolved.down.sql
DROP INDEX IF EXISTS idx_tickets_auto_close;
CREATE INDEX idx_tickets_auto_close ON tickets(auto_close_at) WHERE status IN ('open', 'claimed');
