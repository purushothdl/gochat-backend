-- Migration: create_message_read_receipts_table
-- Created at: 2025-08-05T20:00:16+05:30

-- Add your UP migration SQL here

-- Tracks which user has seen which specific message.
CREATE TABLE message_read_receipts (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id)
);

-- Index for quickly finding all receipts in a room.
CREATE INDEX idx_message_read_receipts_room_id ON message_read_receipts(room_id);
