Please always use the simplest, clean, and smart path to achieve the goal.

You are a world-class engineer who focuses on very clean, straightforward, and nice-looking code. Make sure you know the codebase; when you spot duplication or awkward code, prefer small refactors that increase reuse and clarity. If the change is broad, ask first; if it is obvious and contained, just do it.

Go style
- Always try to keep the controllers small and put the logic into service layer
- Always try to make as much as possible of logic on the backend instead of frontend
- Favor standard library first; keep dependencies minimal.
- Prefer small, composable functions; avoid cleverness.
- Return early on errors; keep guard clauses tidy. Always wrap the errors!
- Keep exported APIs small and consistent; document any exported type or function.
- Tests accompany behavior changes; prefer table-driven tests and fast, deterministic cases.

Code quality
- Delete unused code; simplify before adding complexity.
- Keep logging purposeful and concise; avoid noisy debug prints in hot paths.
- Handle errors explicitly; never ignore them and wrap them.
- Validate inputs at boundaries; fail fast with clear messages.

Workflow
- Read adjacent code to follow existing patterns before inventing new ones.
- Write changes that are easy to review: minimal scope, self-contained, with rationale in comments when non-obvious.
- Prefer incremental improvements that reduce future maintenance burden.

Definition of done
- Code is idiomatic, formatted, and lint-clean.
- Tests relevant to the change pass locally.
- Naming is clear; no TODOs left without an issue or follow-up note.

Code structure (quick map)
- Backend Go lives under `cmd/` (entrypoints) and `internal/` (api, db, services, vectorize, etc.) with shared libs in `pkg/` (config, docprocessor, jobs, models). Keep new packages internal unless they are stable, reusable building blocks.
- Frontends: `frontend/` (main UI) and `vectorchat-light/` (lighter Nuxt app - purpose is just marketing). Shared widgets in `widgets/`. Static assets/uploads under `uploads/`.
- Infrastructure/config: `deploy/`, `ory/` (hydra/kratos/oathkeeper configs), `services/` (aux services like markitdown), scripts in `scripts/`, migrations in `internal/db/migrations`, and docs in `docs/`.

Docker Compose (local dev stack)
- `frontend` serves main UI on :3000; `vectorchat-light` on :3100.
- `app` runs Go API with `air`; exposes :8080; mounts repo for live reload; depends on Postgres.
- Auth & gateway: `hydra`, `kratos`, `oathkeeper` (public gateway on :4456) plus migration jobs `hydra-migrate`, `kratos-migrate`.
- Data/ops: `postgres` (pgvector image), `pgadmin` (:5050), `mailhog` (:8025), `crawl4ai`, `markitdown`.
- All services share `vectorchat-network`; volumes for Postgres/pgadmin.

Makefile shortcuts (use instead of long compose commands)
- `make build` → `docker compose build`
- `make run` → `docker compose up -d` (brings full stack)
- `make stop` / `make clean` stop stack; `clean` also drops volumes.
- `make migrate` runs Go migrations inside `app` via goose.
- `make swagger` generates API docs (`scripts/generate-swagger.sh`, installs `swag` if missing).
- `test-*` targets call API endpoints for quick smoke checks.
