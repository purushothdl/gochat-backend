-- Rollback migration: create_user_blocks_table
-- Created at: 2025-08-03T19:36:01+05:30
-- Description: Rollback changes from create_user_blocks_table migration

-- Add your DOWN migration SQL here
DROP INDEX IF EXISTS idx_user_blocks_blocked_id;
DROP TABLE IF EXISTS user_blocks;