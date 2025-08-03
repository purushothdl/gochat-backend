-- Rollback migration: add_broadcast_only_to_rooms
-- Created at: 2025-08-03T23:30:05+05:30

-- Add your DOWN migration SQL here

ALTER TABLE rooms
DROP COLUMN is_broadcast_only;