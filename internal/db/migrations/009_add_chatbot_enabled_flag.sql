-- +goose Up
-- +goose StatementBegin

-- Add is_enabled column to chatbots table
-- Default to true so existing chatbots remain active
ALTER TABLE chatbots
ADD COLUMN is_enabled BOOLEAN NOT NULL DEFAULT true;

-- Add index for faster filtering of enabled chatbots
CREATE INDEX idx_chatbots_is_enabled ON chatbots(is_enabled);

-- Add composite index for common query pattern (user_id + is_enabled)
CREATE INDEX idx_chatbots_user_id_enabled ON chatbots(user_id, is_enabled);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes first
DROP INDEX IF EXISTS idx_chatbots_user_id_enabled;
DROP INDEX IF EXISTS idx_chatbots_is_enabled;

-- Remove the is_enabled column
ALTER TABLE chatbots
DROP COLUMN IF EXISTS is_enabled;

-- +goose StatementEnd
