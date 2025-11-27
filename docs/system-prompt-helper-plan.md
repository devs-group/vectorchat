# System Prompt Helper – Implementation Plan

## Goals
- Add a “System Prompt Helper” button above the system prompt textarea on the chat detail form, opening a shadcn-vue modal that gathers a short purpose description.
- Call the backend to generate a tailored system prompt from that purpose, then insert it into the existing system prompt field when the user clicks “Insert”.
- Keep the LLM interaction layer simple and reusable so both chat replies and prompt generation share the same plumbing.

## Current State (what we saw)
- Chat detail UI lives in `frontend/pages/chat/[id]/details.vue`, using `ChatbotForm.vue` for the form and the `Textarea` `v-model="systemInstructions"` for the system prompt.
- API access is centralized in `frontend/composables/useApiService.ts` via `useApiFetch`.
- LLM calls on the backend are in `internal/services/chat_service.go` using langchaingo/openai directly inside `chatWithChatbot`.

## UX Flow
- **Trigger:** “Prompt Helper” button placed just above the system prompt textarea in `ChatbotForm.vue`.
- **Modal:** shadcn-vue dialog with:
  - Textarea for “Describe the assistant’s purpose”.
  - Optional select for tone/style (e.g., default, concise, friendly) – keep simple, allow future extension.
  - Actions: “Generate” (primary), “Cancel”.
  - Loading state on generate; inline error message on failure.
- **Result Stage:** show generated prompt in a readonly preview textarea with “Insert” and “Close” buttons. “Insert” writes into the underlying system prompt field and closes the modal.
- **Validation:** require non-empty purpose; disable Generate while empty or while a request is in flight.

## API Design (backend + frontend contract)
- Endpoint: `POST /chat/system-prompt/generate`
  - Body: `{ purpose: string, tone?: string }`
  - Response: `{ prompt: string }`
- Frontend hook: add `generateSystemPrompt` to `useApiService` returning `{ prompt }`.
- Errors: standardized error payload; surface toast + inline message.

## LLM Abstraction Plan (keep it small & reusable)
- Introduce `internal/llm` package with an interface:
  ```go
  type Client interface {
    Generate(ctx context.Context, req PromptRequest) (PromptResponse, error)
    Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) // optional for future parity
  }
  ```
  - Start with a thin `OpenAIClient` implementation that wraps the existing `langchaingo/openai` usage and accepts model, temperature, max tokens.
- Refactor `chat_service.go` to use the client instead of constructing `openai.New` inline. Keep function signatures stable.
- Add a small helper in `internal/services/prompt_service.go` (new) that depends on the shared LLM client and builds the prompt-engineering template for generating system prompts.
- Wire the client instance through service constructors (likely from app container) to avoid global state.

## Backend Tasks
1) Add `internal/llm` with `client.go` (interface + request/response structs) and `openai_client.go` (OpenAI implementation, temperature/maxTokens options).
2) Add `internal/services/prompt_service.go` with `GenerateSystemPrompt(ctx, purpose, tone)` that crafts the template and calls the LLM client.
3) Add HTTP handler + route `POST /chat/system-prompt/generate` (check existing router in `server` package) that validates input, calls `PromptService`, and returns `{ prompt }`.
4) Refactor `chat_service.go` to consume the shared LLM client (construct call options instead of instantiating `openai.New` locally).
5) Update dependency wiring (likely in `cmd/server` or DI setup) to instantiate one LLM client with API key and pass to both `ChatService` and `PromptService`.

## Frontend Tasks
1) UI placement: In `ChatbotForm.vue`, add the “Prompt Helper” button above the system prompt `Textarea`; manage modal state locally.
2) Modal component: create `frontend/components/chat/SystemPromptHelperDialog.vue` using shadcn-nuxt dialog primitives; props for `open`, `onClose`, `onInsert`.
3) State & calls: within the dialog, use `useApiService().generateSystemPrompt()`; handle loading, validation, errors.
4) Insert flow: when “Insert” clicked, emit the generated prompt to the parent; parent sets `systemInstructions` value and closes modal.
5) Types: extend `frontend/types/api.ts` with request/response for system prompt generation.
6) Styling/UX: keep consistent spacing, keyboard focus traps, ESC to close; ensure mobile width fits.

## Testing & QA
- Backend: add unit test for `PromptService.GenerateSystemPrompt` with a stub LLM client; handler test for validation and happy path.
- Frontend: component test (or Vitest if available) for dialog state transitions; manual QA checklist in dev:
  - Button opens modal; empty purpose blocks Generate.
  - Loading spinner shows during request; error shows on failure.
  - Insert writes prompt into system instructions and closes dialog.
  - Existing chat form submit still works after insertion.

## Rollout Notes
- Keep defaults safe: if the LLM call fails, leave existing system prompt untouched and surface toast + inline error.
- Log LLM generation errors server-side with request ID for traceability; avoid storing generated prompt unless user saves form.
- Keep the prompt template short to reduce token cost; reuse the chatbot’s default model/temperature or a lightweight model configurable via env.

## Definition of Done
- Modal UX works end-to-end in the chat detail page with generated prompt insertion.
- Shared LLM client used by both chat and prompt-generation paths.
- Tests passing; lint/build succeed; documentation of the new endpoint and UI affordance added to README or docs as needed.
