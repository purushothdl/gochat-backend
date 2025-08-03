-- Migration: create_room_membership_table
-- Created at: 2025-08-03T19:35:48+05:30
-- Description: Add your migration description here

-- Add your UP migration SQL here
CREATE TYPE member_role AS ENUM ('ADMIN', 'MEMBER');

CREATE TABLE room_memberships (
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role member_role NOT NULL DEFAULT 'MEMBER',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (room_id, user_id)
);

CREATE INDEX idx_room_memberships_user_id ON room_memberships(user_id);

CREATE TRIGGER update_room_memberships_updated_at
BEFORE UPDATE ON room_memberships
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();