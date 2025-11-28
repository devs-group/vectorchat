# How to add a new LLM model (LiteLLM + Go + Frontend)

Use this checklist whenever you want to expose another model through the proxy and UI.

1) **Add alias in LiteLLM config**
   - Edit `litellm/config.yaml` → `model_list`.
   - Append an entry with your alias and underlying provider model, e.g.
     ```yaml
     - model_name: my-new-model
       litellm_params:
         model: provider/model-id
     ```
   - If the provider needs a new key, add the corresponding `*_API_KEY` env var to `docker-compose.yml` (litellm service).

2) **Restart proxy/app**
   - `docker compose up -d --build litellm app` (rebuild ensures the config mount is refreshed).

3) **Label (optional but recommended)**
   - Map the alias to a friendly label in `internal/services/llm_service.go` (`friendlyModelLabel` switch).
   - If the model should be considered “advanced” for plan filtering, update `internal/llm/util.go` (`IsAdvancedModel` heuristic).

4) **Frontend fallback (optional)**
   - Add the alias to the fallback list in `frontend/pages/chat/components/ChatbotForm.vue` so it appears even if `/llm/models` is temporarily unavailable.
   - Otherwise, the dropdown will still populate from `/llm/models` once the backend responds.

5) **Defaults (if desired)**
   - Set `LLM_MODEL_CHAT` and/or `LLM_MODEL_PROMPT_GEN` in `.env` / compose to make the new alias the default.

6) **Sanity check**
   - `docker compose logs -f litellm` to confirm the proxy loads the config.
   - Call `GET /llm/models` (or use the UI form) and verify the alias appears with the expected provider/advanced flag.

That’s it—the app always talks to LiteLLM via `/v1/models` discovery, so no other code changes are needed for most additions.
