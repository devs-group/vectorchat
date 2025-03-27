-- +goose Up
-- +goose StatementBegin

-- Create chatbots table
CREATE TABLE IF NOT EXISTS chatbots (
    id VARCHAR(100) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    system_instructions TEXT,
    model_name VARCHAR(50) DEFAULT 'gpt-3.5-turbo',
    temperature_param FLOAT DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 2000,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index for faster querying by user
CREATE INDEX idx_chatbots_user_id ON chatbots(user_id);

-- Modify documents table to associate documents with chatbots
ALTER TABLE documents ADD COLUMN chatbot_id VARCHAR(100) REFERENCES chatbots(id) ON DELETE CASCADE;
CREATE INDEX idx_documents_chatbot_id ON documents(chatbot_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove chatbot association from documents
DROP INDEX IF EXISTS idx_documents_chatbot_id;
ALTER TABLE documents DROP COLUMN IF EXISTS chatbot_id;

-- Drop chatbots table and index
DROP INDEX IF EXISTS idx_chatbots_user_id;
DROP TABLE IF EXISTS chatbots;

-- +goose StatementEnd 