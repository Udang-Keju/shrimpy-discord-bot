-- migrations/013_clock_refresh_interval.up.sql
-- Per-clock refresh interval, chosen from a fixed set in the dashboard.

ALTER TABLE clock_channels
    ADD COLUMN refresh_minutes INT NOT NULL DEFAULT 5
    CONSTRAINT clock_channels_refresh_minutes_check CHECK (refresh_minutes IN (5, 10, 15, 30, 60));
