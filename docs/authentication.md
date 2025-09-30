# Authentication Stack

This project now uses an Ory-based stack for identity and access control. All services run locally through `docker-compose` and share the main PostgreSQL instance.

## Services

- **Kratos** (`oryd/kratos`) – identity management, self-service flows, and GitHub social login.
- **Oathkeeper** (`oryd/oathkeeper`) – reverse proxy that validates Kratos sessions and injects identity headers for backend services.
- **Mailhog** (`mailhog/mailhog`) – captures recovery and verification emails for local development.
- **PostgreSQL** (`pgvector/pgvector`) – re-used for Kratos (database `vectorauth`) as well as application data (`vectordb`).

Ports exposed on the host:

| Service        | Port | Purpose                       |
| -------------- | ---- | ----------------------------- |
| Frontend (Nuxt)| 3000 | Main dashboard                |
| VectorChat Light | 3100 | Marketing site               |
| Backend API (via Oathkeeper) | 8080 | Protected API gateway |
| Kratos public  | 4433 | Self-service and session APIs |
| Kratos admin   | 4434 | Administrative API            |
| Mailhog UI     | 8025 | Local email inbox             |

## Required Environment Variables

Populate the following variables in `.env` before running `docker-compose up`:

| Variable | Description |
| -------- | ----------- |
| `OPENAI_API_KEY` | Existing requirement for vectorisation. |
| `GITHUB_CLIENT_ID` | GitHub OAuth app client ID used by Kratos. |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth app client secret used by Kratos. |
| `KRATOS_COOKIE_SECRET` | 32+ byte secret for signing Kratos session cookies. |
| `KRATOS_CIPHER_SECRET` | 32+ byte secret for encrypting identity data. |
| `BASE_URL` | Public host for the Go API (e.g. `localhost:8080`). |
| `FRONTEND_URL` | Host:port for the main Nuxt app (`localhost:3000`). |
| `LIGHT_FRONTEND_URL` | Host:port for VectorChat Light (`localhost:3100`). |

The compose file propagates these variables to the relevant containers. Kratos requires the GitHub credentials and secrets at start-up; update them whenever the OAuth app changes.

## Local Development Flow

1. Ensure PostgreSQL volume is initialised. The script in `deploy/postgres/init/00-create-databases.sql` creates the `vectorauth` database automatically on first boot.
2. Start the stack:

   ```bash
   docker-compose up --build
   ```

3. Add a developers GitHub OAuth app callback to `http://localhost:4433/self-service/methods/oidc/callback/github`. The homepage/allowed logout URLs should point to `http://localhost:3000` (main app) and `http://localhost:3100` (VectorChat Light).
4. Visit `http://localhost:3000/login` to authenticate. Successful login through Kratos/GitHub sets the `vectorauth_session` cookie on `localhost`, which Oathkeeper validates for both Nuxt applications.
5. VectorChat Light forwards "Sign in" to the main login page with a `return_to` query param. After login you are redirected back and both SPAs immediately see the shared session via Kratos’ `whoami` endpoint.
6. Use `POST /auth/logout` (or the app UI) to end the session; the endpoint proxies Kratos’ logout API and clears the local cookie.

## Future Enhancements

- Enable the Kratos password strategy and wire the Nuxt login/registration/recovery UIs.
- Configure additional OAuth providers by extending `ory/kratos/kratos.yml` (add entries under `selfservice.methods.oidc.config.providers`).
- Move Kratos/Oathkeeper secrets to a secure secret manager for non-development environments.
- Introduce Ory Hydra if backend-to-backend OAuth tokens become necessary.

## Troubleshooting

- **Session missing headers** – verify Oathkeeper is running; backend expects `X-User-Id` and `X-User-Traits` headers that the proxy injects.
- **Cookie not shared** – ensure requests originate from `http://localhost` so the host-only cookie is reused across ports.
- **Database errors on first boot** – delete the Postgres volume if `vectorauth` was not created and restart `docker-compose`.
