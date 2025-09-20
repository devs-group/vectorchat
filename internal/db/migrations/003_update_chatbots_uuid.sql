-- +goose Up
-- Ensure pgcrypto is available for gen_random_uuid
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- First drop the foreign key constraint from documents
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_chatbot_id_fkey;

-- Now we can safely modify the chatbots table
ALTER TABLE chatbots DROP CONSTRAINT chatbots_pkey;
ALTER TABLE chatbots ALTER COLUMN id DROP DEFAULT;
ALTER TABLE chatbots ALTER COLUMN id SET DATA TYPE UUID USING (
    CASE
        WHEN id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN id::uuid
        ELSE gen_random_uuid()
    END
);
ALTER TABLE chatbots ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE chatbots ADD PRIMARY KEY (id);

-- Update documents table to match UUID foreign key
ALTER TABLE documents ALTER COLUMN chatbot_id TYPE UUID USING (
    CASE
        WHEN chatbot_id IS NULL OR chatbot_id = '' THEN NULL
        WHEN chatbot_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN chatbot_id::uuid
        ELSE NULL
    END
);

-- Recreate the foreign key constraint with cascade semantics
ALTER TABLE documents
    ADD CONSTRAINT documents_chatbot_id_fkey
    FOREIGN KEY (chatbot_id)
    REFERENCES chatbots(id)
    ON DELETE CASCADE;

-- +goose Down
-- First drop the foreign key constraint
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_chatbot_id_fkey;

-- Revert UUID changes
ALTER TABLE documents ALTER COLUMN chatbot_id TYPE TEXT;
ALTER TABLE chatbots DROP CONSTRAINT chatbots_pkey;
ALTER TABLE chatbots ALTER COLUMN id DROP DEFAULT;
ALTER TABLE chatbots ALTER COLUMN id SET DATA TYPE TEXT;
ALTER TABLE chatbots ADD PRIMARY KEY (id);

-- Recreate the original foreign key constraint
ALTER TABLE documents
    ADD CONSTRAINT documents_chatbot_id_fkey
    FOREIGN KEY (chatbot_id)
    REFERENCES chatbots(id)
    ON DELETE CASCADE;
