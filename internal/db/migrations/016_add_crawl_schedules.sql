-- +goose Up
CREATE TABLE crawl_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID REFERENCES chatbots(id) ON DELETE CASCADE,
    shared_knowledge_base_id UUID REFERENCES shared_knowledge_bases(id) ON DELETE CASCADE,
    root_url TEXT NOT NULL,
    cron_expr TEXT NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    last_status TEXT,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT crawl_schedules_scope_ck CHECK (
        (chatbot_id IS NOT NULL AND shared_knowledge_base_id IS NULL) OR
        (chatbot_id IS NULL AND shared_knowledge_base_id IS NOT NULL)
    )
);

CREATE INDEX idx_crawl_schedules_chatbot ON crawl_schedules(chatbot_id) WHERE chatbot_id IS NOT NULL;
CREATE INDEX idx_crawl_schedules_shared ON crawl_schedules(shared_knowledge_base_id) WHERE shared_knowledge_base_id IS NOT NULL;
CREATE UNIQUE INDEX idx_crawl_schedules_chatbot_url ON crawl_schedules(chatbot_id, root_url) WHERE chatbot_id IS NOT NULL;
CREATE UNIQUE INDEX idx_crawl_schedules_shared_url ON crawl_schedules(shared_knowledge_base_id, root_url) WHERE shared_knowledge_base_id IS NOT NULL;
CREATE INDEX idx_crawl_schedules_enabled ON crawl_schedules(enabled);

-- +goose Down
DROP INDEX IF EXISTS idx_crawl_schedules_enabled;
DROP INDEX IF EXISTS idx_crawl_schedules_shared_url;
DROP INDEX IF EXISTS idx_crawl_schedules_chatbot_url;
DROP INDEX IF EXISTS idx_crawl_schedules_shared;
DROP INDEX IF EXISTS idx_crawl_schedules_chatbot;
DROP TABLE IF EXISTS crawl_schedules;
