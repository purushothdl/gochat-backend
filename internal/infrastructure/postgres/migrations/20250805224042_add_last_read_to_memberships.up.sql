-- Migration: add_last_read_to_memberships
-- Created at: 2025-08-05T22:40:42+05:30

-- Add your UP migration SQL here
ALTER TABLE room_memberships
ADD COLUMN last_read_timestamp TIMESTAMPTZ;
