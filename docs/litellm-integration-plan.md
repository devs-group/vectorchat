# Unified LLM Integration Plan (LiteLLM Proxy First, Pluggable Later)

## Why LiteLLM (and alternatives)
- **LiteLLM proxy** gives a single OpenAI-compatible endpoint that can fan out to OpenAI, Anthropic, Gemini, Mistral, etc., with request-based routing, caching, rate limits, and key management baked in.
- It is drop-in for our current OpenAI-langchaingo usage (same schema), minimizing backend refactors.
- Alternative: build our own thin “LLM router” service. This keeps dependencies smaller but requires us to implement provider adapters, auth, logging, and cost/rate limiting. Given time-to-value, LiteLLM proxy is the fastest path; we can still wrap it in a small interface so we can swap later.

## Target Outcomes
- All LLM calls (chat and system-prompt generation) go through a single client abstraction that speaks an OpenAI-compatible API.
- We can switch/model-route to providers (OpenAI, Claude, Gemini, Mistral, etc.) via config, without code changes.
- Deployable in Docker compose; local dev uses a LiteLLM container; prod can point at a managed endpoint if needed.

## Architecture Shape
- **Backend boundary:** `internal/llm` defines a narrow interface (`Client.Generate`, `Client.ChatStream`), implemented by an `OpenAICompatibleClient` using `langchaingo/llms/openai` configured to point to the LiteLLM proxy `BASE_URL` and `API_KEY`.
- **Service usage:** `ChatService` and new `PromptService` depend only on the interface; no provider-specific code inside services.
- **Config:** environment variables drive model choice and routing hints:
  - `LLM_BASE_URL` (LiteLLM proxy URL)
  - `LLM_API_KEY`
  - `LLM_MODEL_CHAT` (default chat model, e.g., `anthropic/claude-3-5-sonnet` or `gpt-4o-mini`)
  - `LLM_MODEL_PROMPT_GEN` (lighter model for system prompt generation)
  - `LLM_REQUEST_TIMEOUT_MS`
- **Routing:** LiteLLM can map logical model ids (aliases) to providers; we keep model strings opaque in app config.

## Deployment Plan (Docker-first)
- Add a `litellm` service to `docker-compose.yml`:
  - Image: `ghcr.io/berriai/litellm-proxy:latest`
  - Expose port `4000`
  - Mount `litellm/config.yaml` for provider keys and model aliases.
  - Optional: enable logging, caching (`redis_url`), and rate-limits.
- Provide a sample `litellm/config.yaml.example` with:
  - Provider keys via env placeholders (e.g., `${OPENAI_API_KEY}`, `${ANTHROPIC_API_KEY}`, `${GOOGLE_API_KEY}`).
  - Model alias block:
    ```yaml
    model_list:
      - model_name: chat-default
        litellm_params: { model: gpt-4o-mini }
      - model_name: prompt-helper
        litellm_params: { model: gpt-4o-mini }
      - model_name: claude-fast
        litellm_params: { model: anthropic/claude-3-5-sonnet }
    ```
- Local dev: set `LLM_BASE_URL=http://localhost:4000` and `LLM_API_KEY=<proxy_key>`.
- Prod: same vars but point to deployed proxy (or use direct OpenAI if proxy is absent—client should allow direct mode).

## Backend Tasks
1) **LLM client abstraction**
   - Add `internal/llm/client.go` with interface and request/response structs.
   - Add `openai_compatible_client.go` that wraps `langchaingo` but accepts `BaseURL` + `Model` + `APIKey` from config; handles streaming and non-streaming.
   - Add small options struct to pass temperature, max tokens; defaults from caller or config.
2) **Config wiring**
   - Extend config struct/env parsing to include `LLM_BASE_URL`, `LLM_API_KEY`, `LLM_MODEL_CHAT`, `LLM_MODEL_PROMPT_GEN`, `LLM_TIMEOUT_MS`.
   - Provide safe defaults: if `LLM_BASE_URL` empty, fallback to OpenAI direct.
3) **Service refactor**
   - `ChatService`: replace inline `openai.New` with injected `llm.Client`; pass model/temperature/maxTokens from chatbot settings; streaming supported.
   - `PromptService`: already planned; use `LLM_MODEL_PROMPT_GEN`.
4) **HTTP handlers**
   - No schema change needed; existing endpoints keep their shapes because the client remains OpenAI-compatible.
5) **Tests**
   - Add a fake `llm.Client` for unit tests.
   - Integration test hitting a mock LiteLLM server (or httptest) to ensure BaseURL override works.

## Frontend Impact
- Expose model choices dynamically instead of hardcoded list in `ChatbotForm.vue`.
- Add `/llm/models` endpoint that returns the configured logical model aliases (e.g., `chat-default`, `claude-default`, `gemini-default`) plus a human label and provider hint.
- Update the existing model dropdown to fetch from this endpoint on load; keep a graceful fallback list if the call fails.
- Display provider badge (OpenAI/Anthropic/Google/Mistral) so users understand the selection; prevent users from seeing models that aren’t configured.

## Config & Secrets
- New envs in `.env.example`:
  - `LLM_BASE_URL=`
  - `LLM_API_KEY=`
  - `LLM_MODEL_CHAT=chat-default`
  - `LLM_MODEL_PROMPT_GEN=prompt-helper`
  - `LLM_TIMEOUT_MS=60000`
- Add `litellm/config.yaml.example` and include in repo; real `config.yaml` ignored via `.gitignore`.

## Multi‑Provider Keys & Parallel Support
- LiteLLM proxy supports multiple providers/keys simultaneously; we’ll list each provider block in `litellm/config.yaml`:
  ```yaml
  general_settings:
    api_keys:
      openai: ${OPENAI_API_KEY}
      anthropic: ${ANTHROPIC_API_KEY}
      google: ${GOOGLE_API_KEY}
      mistral: ${MISTRAL_API_KEY}
  model_list:
    - model_name: chat-default
      litellm_params: { model: gpt-4o-mini }
    - model_name: claude-default
      litellm_params: { model: anthropic/claude-3-5-sonnet }
    - model_name: gemini-default
      litellm_params: { model: google/gemini-1.5-flash }
  ```
- Services pick which logical model alias to use; LiteLLM chooses the correct provider/key under the hood. We can add per-model overrides (timeout, num_retries) and fallbacks if a provider is down.

## Model Discovery Endpoint (for the dropdown)
- New handler `GET /llm/models` returns:
  ```json
  [
    { "id": "chat-default", "label": "OpenAI GPT-4o Mini", "provider": "openai" },
    { "id": "claude-default", "label": "Claude 3.5 Sonnet", "provider": "anthropic" },
    { "id": "gemini-default", "label": "Gemini 1.5 Flash", "provider": "google" }
  ]
  ```
- Source of truth: LiteLLM `model_list` in `config.yaml`; server reads it at startup (or from env) and exposes the filtered list.
- Frontend dropdown uses these ids as `model_name` values; no code changes required when adding new models—just update LiteLLM config, restart, and the list updates.

## Per-Plan / Per-User Model Eligibility
- Add a small server-side filter layer:
  - Maintain a mapping of plan → allowed model ids (e.g., Basic: `chat-default`; Pro: `chat-default`, `claude-default`; Enterprise: all).
  - In `/llm/models`, read the authenticated user’s plan/tier (existing subscription service) and return only allowed models.
  - If no subscription is found, fall back to a safe baseline list.
- Pricing changes become config-only:
  - Store the plan→models mapping in config (`LLM_PLAN_MODELS_JSON` env or a YAML block). Reload on startup; hot-reload optional.
  - LiteLLM config remains the superset; server filter enforces per-user visibility without changing LiteLLM.
- Frontend behavior:
  - Dropdown renders only what `/llm/models` returns, so users see only the models their plan allows.
  - Optionally show a “locked” hint for upsell if the API includes `locked: true` entries; otherwise omit.

## Ops/Monitoring
- Enable LiteLLM logging to stdout; optionally add `redis_url` for caching.
- Consider `num_retries`, `timeout`, and `rate_limit_per_minute` in config.
- Document how to add new models: edit `litellm/config.yaml`, restart proxy, update env model alias if needed.

## Definition of Done
- All backend LLM calls route through the new interface and respect `LLM_BASE_URL`.
- Docker compose can spin up LiteLLM proxy locally; app works with it.
- Tests updated/passing; docs added for config and model onboarding.
