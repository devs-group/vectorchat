-- +goose Up
-- Ensure pgcrypto is available for uuid generation functions used elsewhere
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Drop existing constraints if present
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_chatbot_id_fkey;
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_file_id_fkey;
ALTER TABLE files DROP CONSTRAINT IF EXISTS files_chatbot_id_fkey;

-- Remove orphaned records to satisfy upcoming constraints
DELETE FROM documents WHERE chatbot_id IS NULL;
DELETE FROM documents
WHERE file_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM files WHERE files.id = documents.file_id);
DELETE FROM documents
WHERE chatbot_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM chatbots WHERE chatbots.id = documents.chatbot_id);
DELETE FROM files
WHERE NOT EXISTS (SELECT 1 FROM chatbots WHERE chatbots.id = files.chatbot_id);

-- Enforce non-null chatbot references on documents
ALTER TABLE documents ALTER COLUMN chatbot_id SET NOT NULL;

-- Recreate foreign keys with cascading behaviour
ALTER TABLE documents
    ADD CONSTRAINT documents_chatbot_id_fkey
    FOREIGN KEY (chatbot_id)
    REFERENCES chatbots(id)
    ON DELETE CASCADE;

ALTER TABLE files
    ADD CONSTRAINT files_chatbot_id_fkey
    FOREIGN KEY (chatbot_id)
    REFERENCES chatbots(id)
    ON DELETE CASCADE;

ALTER TABLE documents
    ADD CONSTRAINT documents_file_id_fkey
    FOREIGN KEY (file_id)
    REFERENCES files(id)
    ON DELETE CASCADE;

-- +goose Down
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_file_id_fkey;
ALTER TABLE files DROP CONSTRAINT IF EXISTS files_chatbot_id_fkey;
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_chatbot_id_fkey;
ALTER TABLE documents ALTER COLUMN chatbot_id DROP NOT NULL;
ALTER TABLE documents
    ADD CONSTRAINT documents_chatbot_id_fkey
    FOREIGN KEY (chatbot_id)
    REFERENCES chatbots(id)
    ON DELETE CASCADE;
