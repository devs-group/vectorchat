-- +goose Up
-- Migration 006: Change 'content' column in 'documents' table from TEXT to BYTEA
-- This allows storage of binary data (e.g., file uploads) instead of only UTF-8 text.

ALTER TABLE documents
  ALTER COLUMN content TYPE BYTEA
  USING content::bytea;

-- +goose Down
-- Revert 'content' column in 'documents' table back to TEXT

ALTER TABLE documents
  ALTER COLUMN content TYPE TEXT
  USING convert_from(content, 'UTF8');