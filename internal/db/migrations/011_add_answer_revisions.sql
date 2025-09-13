-- +goose Up
-- Store revised Q&A pairs with embeddings for RAG-based answer improvements
CREATE TABLE answer_revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    original_message_id UUID REFERENCES chat_messages(id) ON DELETE SET NULL,
    question TEXT NOT NULL,
    original_answer TEXT NOT NULL,
    revised_answer TEXT NOT NULL,
    question_embedding vector(1536), -- OpenAI embedding dimension
    revision_reason TEXT,
    revised_by TEXT NOT NULL, -- admin user ID
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT true
);

-- Indexes for efficient queries
CREATE INDEX idx_answer_revisions_chatbot_id ON answer_revisions(chatbot_id);
CREATE INDEX idx_answer_revisions_active ON answer_revisions(is_active) WHERE is_active = true;
CREATE INDEX idx_answer_revisions_created_at ON answer_revisions(created_at DESC);

-- Vector similarity search index using ivfflat
CREATE INDEX idx_answer_revisions_embedding ON answer_revisions
    USING ivfflat (question_embedding vector_cosine_ops)
    WITH (lists = 100)
    WHERE is_active = true;

-- +goose Down
DROP TABLE IF EXISTS answer_revisions;
