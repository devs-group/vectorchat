-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(120) NOT NULL UNIQUE,
    description TEXT,
    billing_email VARCHAR(255),
    plan_tier VARCHAR(50) NOT NULL DEFAULT 'free' CHECK (plan_tier IN ('free','pro','enterprise')),
    created_by TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE organization_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('owner','admin','member','billing')),
    invited_by TEXT REFERENCES users(id),
    invited_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ,
    UNIQUE (organization_id, user_id)
);

CREATE TABLE organization_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'member' CHECK (role IN ('owner','admin','member','billing')),
    token_hash TEXT NOT NULL,
    invited_by TEXT REFERENCES users(id),
    message TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    accepted_at TIMESTAMPTZ,
    UNIQUE (organization_id, email)
);

ALTER TABLE chatbots
    ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    ADD COLUMN organization_role_cached VARCHAR(20);

ALTER TABLE shared_knowledge_bases
    ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL;

CREATE INDEX idx_chatbots_org ON chatbots (organization_id);
CREATE INDEX idx_shared_kb_org ON shared_knowledge_bases (organization_id);
CREATE INDEX idx_org_members_user ON organization_members (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_org_members_user;
DROP INDEX IF EXISTS idx_shared_kb_org;
DROP INDEX IF EXISTS idx_chatbots_org;

ALTER TABLE shared_knowledge_bases
    DROP COLUMN IF EXISTS organization_id;

ALTER TABLE chatbots
    DROP COLUMN IF EXISTS organization_role_cached,
    DROP COLUMN IF EXISTS organization_id;

DROP TABLE IF EXISTS organization_invites;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
