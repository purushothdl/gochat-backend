-- Rollback migration: create_message_read_receipts_table
-- Created at: 2025-08-05T20:00:16+05:30

-- Add your DOWN migration SQL here
DROP INDEX IF EXISTS idx_message_read_receipts_room_id;
DROP TABLE IF EXISTS message_read_receipts;