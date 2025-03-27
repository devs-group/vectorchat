-- +goose Up
-- First drop the foreign key constraint from documents
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_chatbot_id_fkey;

-- Now we can safely modify the chatbots table
ALTER TABLE chatbots DROP CONSTRAINT chatbots_pkey;
ALTER TABLE chatbots ALTER COLUMN id DROP DEFAULT;
ALTER TABLE chatbots ALTER COLUMN id SET DATA TYPE UUID USING (gen_random_uuid());
ALTER TABLE chatbots ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE chatbots ADD PRIMARY KEY (id);

-- Update documents table to match UUID foreign key
ALTER TABLE documents ALTER COLUMN chatbot_id TYPE UUID USING (
    CASE 
        WHEN chatbot_id IS NULL OR chatbot_id = '' THEN NULL 
        ELSE chatbot_id::uuid 
    END
);

-- Recreate the foreign key constraint
ALTER TABLE documents
    ADD CONSTRAINT documents_chatbot_id_fkey
    FOREIGN KEY (chatbot_id)
    REFERENCES chatbots(id)
    ON DELETE SET NULL;

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
    ON DELETE SET NULL; 