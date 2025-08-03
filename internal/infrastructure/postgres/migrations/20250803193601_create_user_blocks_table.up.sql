-- Migration: create_user_blocks_table
-- Created at: 2025-08-03T19:36:01+05:30
-- Description: Add your migration description here

-- Add your UP migration SQL here

CREATE TABLE user_blocks (
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (blocker_id, blocked_id)
);

CREATE INDEX idx_user_blocks_blocked_id ON user_blocks(blocked_id);