-- +goose Up
-- Add default NOW() for created_at and updated_at
ALTER TABLE chatbots 
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

-- +goose Down
-- Remove defaults
ALTER TABLE chatbots 
    ALTER COLUMN created_at DROP DEFAULT,
    ALTER COLUMN updated_at DROP DEFAULT; 