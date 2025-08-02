-- Rollback migration: create_password_reset_tokens
-- Created at: 08/03/2025 00:21:02

DROP INDEX IF EXISTS idx_password_reset_tokens_expires_at;
DROP TABLE IF EXISTS password_reset_tokens;