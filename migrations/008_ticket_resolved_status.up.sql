-- migrations/008_ticket_resolved_status.up.sql
-- Add a "resolved" state to the ticket lifecycle, distinct from "closed": marks that
-- staff consider the issue handled without locking the channel or generating a
-- transcript. Tickets left in this state with no further activity are still picked
-- up by the existing auto-close scheduler.

ALTER TYPE ticket_status ADD VALUE IF NOT EXISTS 'resolved';
