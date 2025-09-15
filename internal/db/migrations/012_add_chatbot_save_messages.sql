-- +goose Up
-- +goose StatementBegin

-- Add save_messages column controlling chat history persistence
ALTER TABLE chatbots
ADD COLUMN save_messages BOOLEAN NOT NULL DEFAULT true;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the save_messages column
ALTER TABLE chatbots
DROP COLUMN IF EXISTS save_messages;

-- +goose StatementEnd
