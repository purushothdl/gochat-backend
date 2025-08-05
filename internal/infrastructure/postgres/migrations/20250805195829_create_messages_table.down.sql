-- Rollback migration: create_messages_table
-- Created at: 2025-08-05T19:58:29+05:30

-- Add your DOWN migration SQL here

DROP TRIGGER IF EXISTS update_messages_updated_at ON messages;
DROP INDEX IF EXISTS idx_messages_room_id_created_at;
DROP INDEX IF EXISTS idx_messages_user_id;
DROP TABLE IF EXISTS messages;
DROP TYPE IF EXISTS message_type;