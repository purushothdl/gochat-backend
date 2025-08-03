-- Migration: add_broadcast_only_to_rooms
-- Created at: 2025-08-03T23:30:05+05:30

-- Add your UP migration SQL here
ALTER TABLE rooms
ADD COLUMN is_broadcast_only BOOLEAN NOT NULL DEFAULT FALSE;
