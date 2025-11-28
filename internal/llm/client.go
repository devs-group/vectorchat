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
}

// Client provides a minimal interface for chat-capable LLM providers.
type Client interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
