-- Rollback migration: create_rooms_table
-- Created at: 2025-08-03T19:35:35+05:30
-- Description: Rollback changes from create_rooms_table migration

-- Add your DOWN migration SQL here
DROP TRIGGER IF EXISTS update_rooms_updated_at ON rooms;
DROP TABLE IF EXISTS rooms;
DROP TYPE IF EXISTS room_type;
