-- Rollback migration: create_room_membership_table
-- Created at: 2025-08-03T19:35:48+05:30
-- Description: Rollback changes from create_room_membership_table migration

-- Add your DOWN migration SQL here

DROP TRIGGER IF EXISTS update_room_memberships_updated_at ON room_memberships;
DROP TABLE IF EXISTS room_memberships;
DROP TYPE IF EXISTS member_role;