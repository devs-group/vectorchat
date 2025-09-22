-- +goose Up
-- Ensure pgcrypto extension is available for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Shared knowledge base catalog
CREATE TABLE shared_knowledge_bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id VARCHAR(100) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shared_knowledge_bases_owner_id
    ON shared_knowledge_bases(owner_id);

-- Relationship between chatbots and shared knowledge bases
CREATE TABLE chatbot_shared_knowledge_bases (
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    shared_knowledge_base_id UUID NOT NULL REFERENCES shared_knowledge_bases(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chatbot_id, shared_knowledge_base_id)
);

CREATE INDEX idx_chatbot_shared_kbs_shared_id
    ON chatbot_shared_knowledge_bases(shared_knowledge_base_id);

-- Allow files to belong either to a chatbot-specific store or a shared knowledge base
ALTER TABLE files DROP CONSTRAINT IF EXISTS files_chatbot_id_fkey;
ALTER TABLE files ALTER COLUMN chatbot_id DROP NOT NULL;
ALTER TABLE files ADD COLUMN shared_knowledge_base_id UUID;

ALTER TABLE files
    ADD CONSTRAINT files_chatbot_id_fkey
    FOREIGN KEY (chatbot_id)
    REFERENCES chatbots(id)
    ON DELETE CASCADE;

ALTER TABLE files
    ADD CONSTRAINT files_shared_knowledge_base_id_fkey
    FOREIGN KEY (shared_knowledge_base_id)
    REFERENCES shared_knowledge_bases(id)
    ON DELETE CASCADE;

ALTER TABLE files
    ADD CONSTRAINT files_chatbot_or_shared_ck
    CHECK (
        (chatbot_id IS NOT NULL AND shared_knowledge_base_id IS NULL)
        OR (chatbot_id IS NULL AND shared_knowledge_base_id IS NOT NULL)
    );

CREATE INDEX idx_files_shared_knowledge_base_id
    ON files(shared_knowledge_base_id)
    WHERE shared_knowledge_base_id IS NOT NULL;

-- Allow documents to be associated with either a chatbot or a shared knowledge base
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_chatbot_id_fkey;
ALTER TABLE documents ALTER COLUMN chatbot_id DROP NOT NULL;
ALTER TABLE documents ADD COLUMN shared_knowledge_base_id UUID;

ALTER TABLE documents
    ADD CONSTRAINT documents_chatbot_id_fkey
    FOREIGN KEY (chatbot_id)
    REFERENCES chatbots(id)
    ON DELETE CASCADE;

ALTER TABLE documents
    ADD CONSTRAINT documents_shared_knowledge_base_id_fkey
    FOREIGN KEY (shared_knowledge_base_id)
    REFERENCES shared_knowledge_bases(id)
    ON DELETE CASCADE;

ALTER TABLE documents
    ADD CONSTRAINT documents_chatbot_or_shared_ck
    CHECK (
        (chatbot_id IS NOT NULL AND shared_knowledge_base_id IS NULL)
        OR (chatbot_id IS NULL AND shared_knowledge_base_id IS NOT NULL)
    );

CREATE INDEX idx_documents_shared_knowledge_base_id
    ON documents(shared_knowledge_base_id)
    WHERE shared_knowledge_base_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_documents_shared_knowledge_base_id;
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_chatbot_or_shared_ck;
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_shared_knowledge_base_id_fkey;
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_chatbot_id_fkey;
ALTER TABLE documents DROP COLUMN IF EXISTS shared_knowledge_base_id;
ALTER TABLE documents ALTER COLUMN chatbot_id SET NOT NULL;
ALTER TABLE documents
    ADD CONSTRAINT documents_chatbot_id_fkey
    FOREIGN KEY (chatbot_id)
    REFERENCES chatbots(id)
    ON DELETE CASCADE;

DROP INDEX IF EXISTS idx_files_shared_knowledge_base_id;
ALTER TABLE files DROP CONSTRAINT IF EXISTS files_chatbot_or_shared_ck;
ALTER TABLE files DROP CONSTRAINT IF EXISTS files_shared_knowledge_base_id_fkey;
ALTER TABLE files DROP CONSTRAINT IF EXISTS files_chatbot_id_fkey;
ALTER TABLE files DROP COLUMN IF EXISTS shared_knowledge_base_id;
ALTER TABLE files ALTER COLUMN chatbot_id SET NOT NULL;
ALTER TABLE files
    ADD CONSTRAINT files_chatbot_id_fkey
    FOREIGN KEY (chatbot_id)
    REFERENCES chatbots(id)
    ON DELETE CASCADE;

DROP TABLE IF EXISTS chatbot_shared_knowledge_bases;
DROP TABLE IF EXISTS shared_knowledge_bases;
