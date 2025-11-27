-- +goose Up
-- Replace partial unique indexes with full unique constraints so UPSERT works.

-- Drop old partial unique indexes if they exist
DROP INDEX IF EXISTS idx_crawl_schedules_chatbot_url;
DROP INDEX IF EXISTS idx_crawl_schedules_shared_url;

-- Add unique constraints (PostgreSQL creates backing indexes)
ALTER TABLE crawl_schedules
    ADD CONSTRAINT crawl_schedules_chatbot_root_uniq UNIQUE (chatbot_id, root_url);

ALTER TABLE crawl_schedules
    ADD CONSTRAINT crawl_schedules_shared_root_uniq UNIQUE (shared_knowledge_base_id, root_url);

-- +goose Down
ALTER TABLE crawl_schedules DROP CONSTRAINT IF EXISTS crawl_schedules_shared_root_uniq;
ALTER TABLE crawl_schedules DROP CONSTRAINT IF EXISTS crawl_schedules_chatbot_root_uniq;

-- Recreate previous partial unique indexes (best-effort)
CREATE UNIQUE INDEX IF NOT EXISTS idx_crawl_schedules_chatbot_url
    ON crawl_schedules(chatbot_id, root_url)
    WHERE chatbot_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_crawl_schedules_shared_url
    ON crawl_schedules(shared_knowledge_base_id, root_url)
    WHERE shared_knowledge_base_id IS NOT NULL;
