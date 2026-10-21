-- migrations/013_clock_refresh_interval.down.sql

ALTER TABLE clock_channels DROP COLUMN refresh_minutes;
