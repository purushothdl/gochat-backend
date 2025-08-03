-- Migration: create_rooms_table
-- Created at: 2025-08-03T19:35:35+05:30

-- Add your UP migration SQL here

CREATE TYPE room_type AS ENUM ('DIRECT', 'PRIVATE', 'PUBLIC');

CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100), -- Name is optional for DIRECT rooms
    type room_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_rooms_type ON rooms(type) WHERE deleted_at IS NULL;

CREATE TRIGGER update_rooms_updated_at
BEFORE UPDATE ON rooms
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();