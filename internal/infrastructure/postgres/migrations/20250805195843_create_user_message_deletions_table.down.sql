-- Rollback migration: create_user_message_deletions_table
-- Created at: 2025-08-05T19:58:44+05:30

-- Add your DOWN migration SQL here
DROP TABLE IF EXISTS user_message_deletions;