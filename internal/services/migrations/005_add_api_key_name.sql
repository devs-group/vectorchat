-- +goose Up
ALTER TABLE api_keys 
ADD COLUMN name VARCHAR(255);

-- +goose Down
-- Remove defaults
ALTER TABLE api_keys
DROP COLUMN name;