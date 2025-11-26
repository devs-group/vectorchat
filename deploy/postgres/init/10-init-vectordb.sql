\connect vectordb;

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create the documents table if missing (legacy support)
CREATE TABLE IF NOT EXISTS documents (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    embedding vector(1536) NOT NULL
);

CREATE INDEX IF NOT EXISTS documents_embedding_idx
ON documents
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);
