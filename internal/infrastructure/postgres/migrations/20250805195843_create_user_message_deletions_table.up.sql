-- Migration: create_user_message_deletions_table
-- Created at: 2025-08-05T19:58:44+05:30

-- Add your UP migration SQL here

-- Tracks messages that a user has deleted for themselves.
CREATE TABLE user_message_deletions (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (message_id, user_id)
);