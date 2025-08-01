-- Rollback migration: create_refresh_tokens_table
-- Created at: 08/01/2025 15:23:00

-- Add your DOWN migration SQL here

DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_hash;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP TABLE IF EXISTS refresh_tokens;
