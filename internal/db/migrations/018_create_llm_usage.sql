-- +goose Up
CREATE TABLE llm_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    org_id TEXT,
    trace_id TEXT,
    model_alias TEXT NOT NULL,
    provider TEXT,
    prompt_tokens INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_llm_usage_user_created_at ON llm_usage (user_id, created_at DESC);
CREATE INDEX idx_llm_usage_model_created_at ON llm_usage (model_alias, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS llm_usage;
