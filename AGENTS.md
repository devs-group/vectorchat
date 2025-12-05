# VectorChat Agent Guide (`AGENTS.md`)

This document is the **authoritative guide** for AI agents working on the VectorChat repository. It defines the coding standards, architectural patterns, and workflows required to maintain the high quality of this codebase.

**Mission:** Build a robust, scalable, and clean AI-powered chat system. We value simplicity, readability, and maintainability over cleverness.

---

## 1. Tech Stack & Key Technologies

### Backend (Go)
- **Language:** Go 1.24+
- **Web Framework:** [Fiber v2](https://github.com/gofiber/fiber) (Fast, Express-like)
- **Database:** PostgreSQL 14+ with `pgvector` extension.
- **ORM/Data Access:** [sqlx](https://github.com/jmoiron/sqlx) (Raw SQL with struct mapping). **No heavy ORMs.**
- **Migrations:** [Goose](https://github.com/pressly/goose).
- **LLM Integration:** [LangChainGo](https://github.com/tmc/langchaingo) & direct OpenAI API usage.
- **Auth:** [Ory Stack](https://www.ory.sh/) (Kratos for identity, Hydra for OAuth2, Oathkeeper for zero-trust proxy).
- **Async/Queue:** NATS JetStream.
- **Documentation:** Swagger/OpenAPI (via `swaggo/swag`).

### Frontend (Nuxt/Vue)
- **Framework:** Nuxt 3 (Vue 3 with Composition API).
- **Language:** TypeScript (Strict mode).
- **Styling:** Tailwind CSS v4.
- **UI Component Library:** [Shadcn UI](https://www.shadcn-vue.com/) (Vue port).
- **Icons:** Lucide Vue Next.
- **State/Data:** Nuxt `useFetch`, `useState`.

### Infrastructure
- **Containerization:** Docker & Docker Compose (`docker-compose.yml`).
- **Services:** PostgreSQL, LiteLLM (LLM Proxy), MarkItDown (Doc processing), Crawl4AI (Web scraping).

---

## 2. Architecture Overview

### Directory Structure
- **`cmd/`**: Entry points. `cmd/main.go` is the API server. `cmd/crawl-scheduler` is the worker.
- **`internal/`**: Private application code.
  - **`api/`**: HTTP Handlers (Controllers). Parse requests, call services, return JSON.
  - **`services/`**: Business logic. The "brain" of the app. Orchestrates DB and other clients.
  - **`db/`**: Data Access Layer (Repositories). SQL queries live here.
  - **`models/`**: (Deprecated/Moved) Prefer `pkg/models` for shared types.
- **`pkg/`**: Public/Shared code.
  - **`models/`**: Domain structs, API request/response types.
  - **`config/`**: App configuration.
- **`frontend/`**: Main Nuxt 3 web application.
- **`deploy/`**: Kubernetes/Docker deployment manifests.
- **`ory/`**: Configuration for Kratos, Hydra, Oathkeeper.

### "Golden Path" Architecture
The application follows a standard **Layered Architecture**:

`Handler (internal/api)` -> `Service (internal/services)` -> `Repository (internal/db)` -> `Database`

1.  **Handlers**: Thin layer. Validate input (using `pkg/models`), call Service, handle errors, return response.
2.  **Services**: Contains ALL business logic. Transaction management happens here (if needed).
3.  **Repositories**: Execute SQL queries. Return domain models.

---

## 3. Coding Standards & patterns

### Go (Backend)
1.  **Dependency Injection**: Use manual DI in `main.go`. Pass dependencies (Repositories, Services) via struct fields.
2.  **Error Handling**:
    -   **ALWAYS** wrap errors using `fmt.Errorf("context: %w", err)`.
    -   Return early (guard clauses).
    -   Handle specific errors (e.g., `sql.ErrNoRows`) in the Repository layer and return domain errors (e.g., `apperrors.ErrNotFound`).
3.  **Database**:
    -   Write raw SQL in `internal/db`.
    -   Use `sqlx` tags (`db:"column_name"`) on structs.
    -   **Context**: All DB methods MUST accept `context.Context` as the first argument.
    -   **Transactions**: Use `*db.Transaction` for multi-step operations.
4.  **Configuration**: Use `pkg/config`. Do not hardcode values.
5.  **Logging**: Use `log/slog` (structured logging).

### TypeScript/Vue (Frontend)
1.  **Composition API**: Always use `<script setup lang="ts">`.
2.  **Typing**: strict TypeScript usage. Avoid `any`. Define interfaces for props and API responses.
3.  **Components**:
    -   Use **Shadcn UI** components from `@/components/ui` for base elements (Buttons, Inputs, etc.).
    -   Keep components small and focused.
4.  **Icons**: Import from `lucide-vue-next`.
    -   Example: `import { Loader2 } from 'lucide-vue-next'`
5.  **API Calls**: Use `useFetch` or custom composables. Handle loading and error states in the UI.

---

## 4. Common Workflows (The "How-To")

### Adding a New API Endpoint
1.  **Model**: Define Request/Response structs in `pkg/models/`.
2.  **Repository**: Add SQL query methods in `internal/db/<entity>_repository.go`.
3.  **Service**: Add business logic method in `internal/services/<entity>_service.go`.
4.  **Handler**: Create handler method in `internal/api/<entity>_handler.go`.
    -   Add Swagger comments (`// @Summary`, `// @Router`, etc.).
5.  **Route**: Register the route in `cmd/main.go` (or `internal/api/routes.go` if it exists).
6.  **Docs**: Run `make swagger` to update API docs.

### Database Changes
1.  **Migration**: Create a new SQL migration file in `internal/db/migrations`.
    -   Format: `YYYYMMDDHHMMSS_description.sql`.
    -   Include both `+goose Up` and `+goose Down` sections.
2.  **Run**: Apply with `make migrate` (runs inside Docker container).
3.  **Verify**: Check `pgadmin` or `psql` to ensure schema is correct.

### Running the Project
-   **Full Stack**: `make run` (starts everything via Docker Compose).
-   **Rebuild**: `make build`.
-   **Stop**: `make stop`.
-   **Clean**: `make clean` (Wipes Data Volumes - CAREFUL).

### Testing
-   **Unit Tests**: Write `_test.go` files next to the code.
-   **Run Tests**: `go test ./...`

---

## 5. Critical Constraints & Rules
-   **Ory Auth**: The app relies on Ory Kratos/Hydra. Do NOT implement custom auth (login/signup) logic. Use the `internal/middleware/auth_middleware.go` which validates Ory sessions/tokens.
-   **Linting**: Ensure code passes standard linters.
-   **No Magic**: Avoid "magical" code or complex reflection. Explicit is better than implicit.

## 6. Definition of Done
1.  Code works and meets requirements.
2.  Code compiles without errors (`go build ./...`).
3.  Linter checks pass.
4.  Swagger documentation is updated (`make swagger`).
5.  New database migrations are included if schema changed.
6.  Frontend builds successfully (`npm run build` in `frontend/`).

## 7. Useful Commands
-   **Go**: `go mod tidy` to clean up dependencies.
-   **Frontend**: `npm install` to install dependencies.
