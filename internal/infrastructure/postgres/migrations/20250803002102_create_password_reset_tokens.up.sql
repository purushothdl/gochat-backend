-- Migration: create_password_reset_tokens
-- Created at: 08/03/2025 00:21:02

CREATE TABLE password_reset_tokens (
    token_hash TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL
);

-- Index to efficiently clean up expired tokens.
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
