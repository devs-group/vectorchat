package llm

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// OpenAIClient implements Client using the langchaingo OpenAI driver.
type OpenAIClient struct {
	apiKey       string
	defaultModel string
}

// NewOpenAIClient creates a new OpenAI-backed LLM client.
func NewOpenAIClient(apiKey string, defaultModel string) *OpenAIClient {
	if defaultModel == "" {
		defaultModel = "gpt-4"
	}
	return &OpenAIClient{
		apiKey:       apiKey,
		defaultModel: defaultModel,
	}
}

// Chat sends a single prompt to OpenAI and returns the combined completion text.
func (c *OpenAIClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = c.defaultModel
	}

	llm, err := openai.New(
		openai.WithToken(c.apiKey),
		openai.WithModel(model),
	)
	if err != nil {
		return ChatResponse{}, apperrors.Wrap(err, "failed to create OpenAI client")
	}

	callOptions := []llms.CallOption{
		llms.WithTemperature(req.Temperature),
	}

	if req.MaxTokens != nil {
		callOptions = append(callOptions, llms.WithMaxTokens(*req.MaxTokens))
	}

	if req.StreamFn != nil {
		callOptions = append(callOptions, llms.WithStreamingFunc(func(callCtx context.Context, chunk []byte) error {
			if len(chunk) == 0 {
				return nil
			}
			return req.StreamFn(callCtx, string(chunk))
		}))
	}

	response, err := llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(req.Prompt),
			},
		},
	}, callOptions...)
	if err != nil {
		return ChatResponse{}, apperrors.Wrap(err, "failed to generate completion")
	}

	// If streaming was used we return aggregated content from the streamFn caller.
	if req.StreamFn != nil {
		return ChatResponse{}, nil
	}

	content := ""
	for _, choice := range response.Choices {
		content += choice.Content
	}

	return ChatResponse{Content: content}, nil
}
