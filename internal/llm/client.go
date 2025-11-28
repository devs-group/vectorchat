package llm

import "context"

// ChatRequest represents a simple chat completion request.
type ChatRequest struct {
	Prompt      string
	Model       string
	Temperature float64
	MaxTokens   *int
	StreamFn    func(context.Context, string) error
}

// ChatResponse contains the generated message content.
type ChatResponse struct {
	Content string
	Usage   Usage
}

// Client provides a minimal interface for chat-capable LLM providers.
type Client interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
	ListModels(ctx context.Context) ([]ModelInfo, error)
}

// Usage captures estimated token usage for a completion.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
}

// ModelInfo describes an available model alias/id and optional provider hint.
type ModelInfo struct {
	ID       string
	Provider string
}
