# Minimal Ory-Based Auth Stack

## Baseline Goals
- Reuse existing `docker-compose` PostgreSQL instance for identities and app data.
- Centralize authentication for `vectorchat-light` (Nuxt + backend) and the main Nuxt + Go backend application.
- Deliver GitHub login with shared SSO today; leave room for additional social providers and email/password later.

## Phase 1 – Decide Ory Components & Topology
- [x] Confirm Kratos as identity provider (password + social strategies) and plan to postpone Hydra unless OAuth2 tokens become necessary.
- [x] Adopt Oathkeeper as a lightweight gateway to validate Kratos sessions and forward identity headers to both backends.
- [x] Define local topology: Postgres, Kratos (public/admin), Oathkeeper, Mailhog (for future flows), frontends/backends, optional reverse proxy.
- [x] Establish shared local domain convention (e.g. `.vectorchat.local`) and document `/etc/hosts` entries.

## Phase 2 – Compose Stack & Configuration
- [x] Extend `docker-compose.yml` with Kratos, Oathkeeper, Mailhog services wired to the shared Postgres.
- [x] Create Kratos config: identity schema (basic profile), GitHub social connector, secure cookie settings, public/admin endpoints.
- [x] Author Oathkeeper rules mapping public endpoints to upstream backends with session validation and identity header injection.
- [x] Provide container images / Dockerfiles for Nuxt apps and Go backend if missing, ensuring they sit behind Oathkeeper or reverse proxy.

## Phase 3 – Frontend Integration
- [x] Update `vectorchat-light` Nuxt app to drive Kratos self-service flows (login, logout) via public API and reuse session cookies.
- [x] Apply identical session integration to the main Nuxt frontend; ensure SSR/CSR logic reads Kratos session.
- [x] Implement shared auth store/composable to detect active session and refresh via Kratos `whoami` endpoint.
- [x] Build minimal UI screens for login/logout success/error states; later extend for registration/recovery.

## Phase 4 – Backend Integration
- [x] Update Nuxt backend and Go backend to sit behind Oathkeeper, trusting forwarded identity headers for user context.
- [x] Add middleware helpers to parse Oathkeeper headers (user ID, traits) and attach them to request context.
- [x] Gate protected routes on presence of validated identity; allow anonymous access where needed.

## Phase 5 – Future-Proofing
- [x] Sketch Kratos password strategy enablement: registration, verification, recovery, settings flows.
- [x] Wire Mailhog SMTP for local email delivery; create templates for verification/recovery emails.
- [x] Plan configuration for additional social providers (GitHub today, others later) via Kratos OAuth2 connectors.

## Testing & Verification
- [x] Manual SSO test: log into `vectorchat-light`, confirm automatic auth on main app, and verify logout symmetry.
- [x] Backend regression checks ensuring Oathkeeper protects sensitive endpoints and identity headers populate correctly.
- [x] Add automated smoke tests or scripts that hit Kratos `whoami` via both apps to confirm session sharing.

## Documentation & Handoff
- [x] Document environment variables (GitHub client ID/secret, cookie domain, secrets) and secure storage guidance (`docs/authentication.md`).
- [x] Write developer onboarding instructions: compose startup, OAuth app setup, login flow walkthrough (`README.md`, `docs/authentication.md`).
- [x] Track follow-up decision on adopting Hydra for service-to-service OAuth if/when required.
