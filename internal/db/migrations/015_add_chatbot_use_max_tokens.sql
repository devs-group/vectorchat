-- +goose Up
-- +goose StatementBegin

ALTER TABLE chatbots
ADD COLUMN IF NOT EXISTS use_max_tokens BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE chatbots
DROP COLUMN IF EXISTS use_max_tokens;

-- +goose StatementEnd
