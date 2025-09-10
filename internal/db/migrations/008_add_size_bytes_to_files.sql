-- +goose Up
-- Add size_bytes column to files to track ingested bytes per source
ALTER TABLE files ADD COLUMN IF NOT EXISTS size_bytes BIGINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE files DROP COLUMN IF EXISTS size_bytes;

