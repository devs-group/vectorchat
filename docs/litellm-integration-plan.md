# Unified LLM Integration Plan (LiteLLM Proxy + Go Backend)

## 1. Why LiteLLM (and alternatives)
- **LiteLLM proxy** provides a single OpenAI-compatible endpoint that fans out to OpenAI, Anthropic, Gemini, Mistral, etc.
- Handles request-based routing, standardized caching, rate limits, and unified key management.
- Drop-in replacement for existing OpenAI client code (`langchaingo` or `go-openai`), minimizing backend refactors.
- **Alternative:** Building a custom "LLM router" requires writing adapters, auth, and logging for every provider. LiteLLM is the fastest path to value.

## 2. Target Outcomes
- **Single Interface:** All LLM calls (chat, tools, system-prompt gen) go through a unified client abstraction.
- **Configurable Routing:** Switch models (OpenAI vs. Claude vs. Gemini) via configuration without code changes.
- **Deployment:** Docker-first. Local dev uses a container; Prod uses the same container or a managed instance.
- **Resilience:** Multi-provider keys for failover and parallel request handling are first-class citizens.

## 3. Architecture Shape
- **Backend Boundary:** `internal/llm` defines the interface (`Client.Generate`, `Client.ChatStream`, `Client.ListModels`).
- **Implementation:** `OpenAICompatibleClient` points to `LLM_BASE_URL` (LiteLLM). It does **not** read local YAML files; it queries the proxy API for capabilities.
- **Service Usage:** `ChatService` and `PromptService` depend on the interface.
- **Config:** Environment variables drive connection and defaults:
  - `LLM_BASE_URL` (e.g., `http://litellm:4000`)
  - `LLM_API_KEY` (Proxy master key)
  - `LLM_MODEL_CHAT` (Default alias, e.g., `chat-default`)
  - `LLM_MODEL_PROMPT_GEN` (e.g., `prompt-helper`)

## 4. Deployment Plan (Docker-first)
- **Service:** Add `litellm` to `docker-compose.yml`.
  - Image: `ghcr.io/berriai/litellm-proxy:latest`
  - Ports: `4000:4000`
  - Mounts: `./litellm/config.yaml:/app/config.yaml`
  - Command: `--config /app/config.yaml --port 4000`
- **Configuration (`litellm/config.yaml`):**
  - **Providers:** Define API keys via env vars (`${OPENAI_API_KEY}`, etc.).
  - **Models:** Map aliases to real models.
    ```yaml
    model_list:
      - model_name: chat-default
        litellm_params: { model: gpt-4o-mini }
      - model_name: claude-default
        litellm_params: { model: anthropic/claude-3-5-sonnet }
    ```
- **Readiness:** The Go backend container depends on `litellm`. Add a healthcheck or wait-for-it script to ensure the proxy is up before the Go app initializes.

## 5. Backend Tasks

### Phase 1: Client Abstraction & Discovery
1. **Interface Definition (`internal/llm/client.go`):**
   - Methods: `Chat(ctx, req)`, `ChatStream(ctx, req)`, `ListModels(ctx)`.
2. **Implementation (`openai_compatible_client.go`):**
   - Use `langchaingo` (or `sashabaranov/go-openai` for a lighter footprint).
   - Configure `BaseURL` to `LLM_BASE_URL`.
   - **Dynamic Discovery:** Implement `ListModels` by querying `GET /v1/models` on the proxy. This acts as the source of truth for available models, removing the need for the Go app to parse YAML.
3. **Readiness Check:**
   - In `main.go`, perform a startup check: call `ListModels` (or `/health`). If it fails, retry/wait before starting the HTTP server.

### Phase 2: Service Integration
1. **Refactor Services:**
   - Inject `llm.Client` into `ChatService` and `PromptService`.
   - Remove any direct provider dependencies.
2. **Parallel Runner:**
   - Create `llm.ParallelRunner` (worker pool) to handle concurrent tool calls or speculative generation safely within rate limits.

### Phase 3: Usage & Guardrails
1. **Streaming Usage:**
   - Ensure requests send `stream_options: {"include_usage": true}` (supported by OpenAI/LiteLLM).
   - Capture token counts from the final stream chunk.
   - *Fallback:* If provider usage is missing, use a local tokenizer (`tiktoken-go`) to estimate counts.
2. **Usage Recorder:**
   - Async worker that writes to `llm_usage` table.
   - **Store Tokens, Not Just Cost:** Store `prompt_tokens`, `completion_tokens`, and `model_alias`.
   - Calculate financial cost at **reporting time** (or via a Price Registry service) to allow for retroactive pricing updates.

## 6. Frontend Impact & Model Discovery
- **Endpoint:** `GET /llm/models`
- **Logic:**
  1. Backend calls `llmClient.ListModels()` (cached for performance).
  2. Filters list based on user's subscription plan (e.g., "Basic" users don't see Claude Opus).
  3. Returns JSON: `[{ "id": "chat-default", "label": "GPT-4o Mini", "provider": "openai" }, ...]`.
- **UI:** Dropdown populates from this API. If the API fails, fallback to a hardcoded "safe mode" list.

## 7. Config & Secrets
- **.env:**
  ```bash
  LLM_BASE_URL=http://localhost:4000
  LLM_API_KEY=sk-proxy-key...
  LLM_MODEL_CHAT=chat-default
  ```
- **litellm/config.yaml:**
  - Standardized aliases (`chat-default`, `fast-reasoning`, `strong-reasoning`).
  - Router settings: `num_retries: 2`, `timeout: 60s`.

## 8. Usage Accounting Data Model
- **Table:** `llm_usage`
  - `id` (UUID)
  - `user_id` / `org_id`
  - `trace_id` (correlation with logs)
  - `model_alias` (e.g., "chat-default")
  - `provider` (e.g., "openai")
  - `prompt_tokens` (int)
  - `completion_tokens` (int)
  - `created_at` (timestamp)
- **Note:** Cost columns are optional. It is safer to calculate cost in analytics views (`sum(tokens * price_per_token)`) to handle price fluctuations.

## 9. Definition of Done
- [ ] LiteLLM running in Docker with aliases for OpenAI and Anthropic.
- [ ] Go backend waits for LiteLLM readiness on startup.
- [ ] `GET /llm/models` returns the list dynamically from the proxy.
- [ ] Chat flow works via the proxy, respecting the selected model.
- [ ] Token usage is accurately recorded in the DB (including streaming requests).