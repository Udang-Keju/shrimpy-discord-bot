-- migrations/009_ticket_auto_close_index_resolved.up.sql
-- The scheduler's auto-close scan (ListDueForAutoClose) now also picks up
-- Resolved tickets (added in 008_ticket_resolved_status), but the original
-- partial index only covered Open/Claimed. Recreate it to include Resolved so
-- the scan can keep using an index instead of falling back to a full scan.
-- This must be its own migration: the 'resolved' enum value added in 008
-- cannot be referenced by name until that migration's transaction commits.

DROP INDEX IF EXISTS idx_tickets_auto_close;
CREATE INDEX idx_tickets_auto_close ON tickets(auto_close_at) WHERE status IN ('open', 'claimed', 'resolved');
