-- +goose Up
-- Create files table
CREATE TABLE files (
  id UUID PRIMARY KEY,
  chatbot_id UUID NOT NULL,
  filename TEXT NOT NULL,
  uploaded_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_files_chatbot_id ON files(chatbot_id);
CREATE INDEX idx_files_uploaded_at ON files(uploaded_at);

-- Update documents table: add file_id and chunk_index
ALTER TABLE documents ADD COLUMN file_id UUID;
ALTER TABLE documents ADD COLUMN chunk_index INTEGER;

CREATE INDEX idx_documents_file_id ON documents(file_id);
CREATE INDEX idx_documents_chunk_index ON documents(chunk_index);

-- +goose Down
-- Remove indexes and columns
DROP INDEX IF EXISTS idx_documents_chunk_index;
DROP INDEX IF EXISTS idx_documents_file_id;
ALTER TABLE documents DROP COLUMN IF EXISTS chunk_index;
ALTER TABLE documents DROP COLUMN IF EXISTS file_id;

-- Drop files table and indexes
DROP INDEX IF EXISTS idx_files_uploaded_at;
DROP INDEX IF EXISTS idx_files_chatbot_id;
DROP TABLE IF EXISTS files; 