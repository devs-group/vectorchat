# Organizations Implementation Plan (revalidated 2025-12-03)

## Reality Check (current system)
- Backend: Fiber handlers under `internal/api`, services in `internal/services`, ownership middleware checks `X-User-ID` from Oathkeeper; no org concept yet.
- IDs: `users.id` is `TEXT`; core resource IDs (`chatbots.id`, `files.id`, `shared_knowledge_bases.id`) are `UUID` with `owner_id/user_id` as `TEXT` foreign keys.
- Frontend: Nuxt + Pinia; existing UI leans toward Tailwind utility classes. We want a cleaner Notion/Linear-inspired vibe (calm typography, greys, single accent, minimal chrome).

## Principles
1) Stateless org context via optional `X-Organization-ID` header; personal workspace when absent.
2) Keep scope tight: only organizations, memberships, invites, and scoping for chatbots + shared knowledge bases.
3) Use existing patterns: middleware injects context, services hold logic, handlers stay thin, errors wrapped.
4) Minimal surface area: avoid premature settings/billing screens; ship a usable switcher + member mgmt first.

## Data Model (aligned with existing schema)

```sql
-- Organizations
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

-- Memberships
CREATE TABLE organization_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('owner','admin','member','billing')),
    invited_by TEXT REFERENCES users(id),
    invited_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ,
    UNIQUE(organization_id, user_id)
);

-- Invites (token stored as hash)
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
    UNIQUE(organization_id, email)
);

-- Resource scoping (preserve creator columns)
ALTER TABLE chatbots
    ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    ADD COLUMN organization_role_cached VARCHAR(20);

ALTER TABLE shared_knowledge_bases
    ADD COLUMN organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL;

CREATE INDEX idx_chatbots_org ON chatbots(organization_id);
CREATE INDEX idx_shared_kb_org ON shared_knowledge_bases(organization_id);
CREATE INDEX idx_org_members_user ON organization_members(user_id);
```

Rules:
- `user_id/owner_id` stay as creator for audit; filtering uses `organization_id` when header present.
- Personal resources have `organization_id IS NULL`.

## Backend Architecture (minimal)
- Middleware: `OrganizationContextMiddleware` reads `X-Organization-ID` (optional), loads membership via repository, attaches `{OrgID, Role}` to `fiber.Ctx` locals. Rejects if header is present but membership missing.
- Services: `OrganizationService` (CRUD + membership lookup), `InviteService` (hash token, send email, consume token), small helper `PermissionService` for role → permission map.
- Handlers: new `OrgHandler` with routes under `/orgs` to mirror existing Fiber style; keep logic delegated to services.
- Query scoping: repositories accept `OrgID *uuid.UUID` and `UserID string`; enforce `WHERE organization_id = $1` when present else fallback to user-owned rows.

### Endpoints (aligned with current routing conventions)
- `GET /orgs` (list mine) / `POST /orgs` (create)
- `GET /orgs/:id` / `PATCH /orgs/:id` (owner/admin) / `DELETE /orgs/:id` (owner)
- `GET /orgs/:id/members` / `PATCH /orgs/:id/members/:userID` / `DELETE ...` (admin+)
- `POST /orgs/:id/invites` / `GET /orgs/:id/invites`
- `POST /org-invites/accept` with token
- Existing chat/KB endpoints accept `X-Organization-ID`; ownership middleware expands to "member with permission" when org header present.

## Frontend Direction (Notion/Linear feel)
- Palette: warm greys (#0F172A text, #F8FAFC bg) with a single accent (electric blue or jade). Lots of white space, subtle borders, rounded-8 corners.
- Typography: switch to "Inter var" already? If not, use `"SF Pro Text", "Inter", system` stack with 14–16px base.
- Components: lightweight org switcher + member list; no heavy chrome. Translucent cards, soft shadows, micro-interactions (opacity/scale on hover).
- API client: Axios plugin already exists; set/remove `X-Organization-ID` centrally.
- Navigation: show org name pill + personal fallback; keep rest unchanged until org pages ship.

## Execution Checklist (for me)
### Foundation
- [x] Confirm goose migration numbers and add 019 covering org tables + resource columns.
- [x] Add repository interfaces for organizations, members, invites (tests pending).
- [x] Update ownership middleware to accept org context + permission checks.

### Backend features
- [x] Org CRUD handler under `/orgs` (thin) backed by service.
- [x] Invite flow: create, accept endpoint that joins member + invalidates token (email hook TBD).
- [x] Chatbot/KB listing & create: scope by org when header set; preserve personal behavior when absent.

### Frontend
- [x] Org state composable (`currentOrgId`, hydration from `/orgs`).
- [x] Axios/fetch header injection for `X-Organization-ID`.
- [x] Minimal org switcher + organizations page (list + create). Member management UI pending.

### QA & rollout
- [ ] Table-driven unit tests for permission logic and invite token flow.
- [ ] Integration tests for scoped list/create on chatbots & shared KBs.
- [ ] Migration dry-run on local stack; verify existing personal data remains accessible with no header.
- [ ] Update docs/README + swagger after routes are stable.

## Notes on Minimalism
- Defer billing UI/settings until org CRUD + invite + resource scoping are solid.
- Avoid duplicating logic between personal/org paths—use shared repository filters.
- Keep logging concise and wrap all errors with context.
