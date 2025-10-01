# OAuth-Based Authentication Flow

This project now delegates API authentication to the Ory stack. The key players are:

- **Ory Hydra (port 4444/4445)** – Issues OAuth2 client credentials and validates access tokens.
- **Ory Oathkeeper (port 4456)** – Acts as the edge proxy, introspecting access tokens and forwarding trusted headers.
- **Go API (vectorchat-app)** – Trusts the headers from Oathkeeper and never inspects bearer tokens directly.

## Machine-to-Machine Credentials

1. A signed-in user requests new credentials from the dashboard.
2. `POST /auth/apikey` calls Hydra’s Admin API (`/admin/clients`) to create a `client_credentials` OAuth client whose metadata is tagged with the user ID.
3. Hydra returns `client_id` and `client_secret`. The secret is only shown once—store it securely.
4. Users exchange the pair for access tokens via Hydra’s public endpoint:

   ```bash
   curl -u "$CLIENT_ID:$CLIENT_SECRET" \
     -d "grant_type=client_credentials" \
     http://localhost:4444/oauth2/token
   ```

5. The returned access token targets the Oathkeeper proxy (base URL `http://localhost:4456`).

## Request Processing

1. Clients send API calls to Oathkeeper with `Authorization: Bearer <token>`.
2. Oathkeeper’s rule set (`ory/oathkeeper/rules.yaml`) secures non-public paths using the `oauth2_introspection` authenticator against Hydra and, as a fallback, the Kratos session cookie.
3. On success, Oathkeeper injects trusted headers:

   - `X-User-ID`: token subject
   - `X-User-Traits`: optional JSON traits (from a Kratos session)

4. The Go API middleware (`internal/middleware/auth_middleware.go`) looks only at `X-User-ID`, hydrates local user data, and lets the request continue.
5. Missing headers result in a 401, meaning the request bypassed Oathkeeper or the token was rejected upstream.

## Environment Reference

- `HYDRA_ADMIN_URL` – Base URL for admin calls (defaults to `http://localhost:4445` for local binaries; the app container overrides it with `http://hydra:4445`).
- `KRATOS_PUBLIC_URL`, `SESSION_COOKIE_NAME` – Still used for browser-based flows, but token validation is entirely Hydra/Oathkeeper.

## Quick Checklist for Local Runs

1. Start the stack: `docker compose up -d hydra hydra-migrate oathkeeper app`.
2. Create credentials via dashboard or directly:

   ```bash
   curl -X POST http://localhost:4456/auth/apikey \
     -H 'Content-Type: application/json' \
     -H 'X-User-ID: <kratos-user-id>' \
     -d '{"name":"local client"}'
   ```

   (Using the dashboard is preferable because it injects the headers via Oathkeeper.)

3. Exchange client credentials for a token (see step above).
4. Call a protected endpoint through Oathkeeper:

   ```bash
   curl http://localhost:4456/chat/some-endpoint \
     -H "Authorization: Bearer $ACCESS_TOKEN"
   ```

If you receive a 401, inspect Oathkeeper logs first—most failures originate at the gateway before the request hits the Go API.
