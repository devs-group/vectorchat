package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkoukk/tiktoken-go"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// OpenAIClient implements Client using an OpenAI-compatible endpoint (LiteLLM, OpenAI, etc.).
type OpenAIClient struct {
	apiKey       string
	baseURL      string
	defaultModel string
	httpClient   *http.Client
}

// NewOpenAIClient creates a new OpenAI-compatible client pointing at the provided baseURL.
func NewOpenAIClient(apiKey, baseURL, defaultModel string, httpClient *http.Client) *OpenAIClient {
	if defaultModel == "" {
		defaultModel = "gpt-4o-mini"
	}
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIClient{
		apiKey:       apiKey,
		baseURL:      strings.TrimRight(baseURL, "/"),
		defaultModel: defaultModel,
		httpClient:   httpClient,
	}
}

// Chat sends a prompt and returns the generated completion along with estimated usage.
func (c *OpenAIClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = c.defaultModel
	}

	options := []openai.Option{
		openai.WithToken(c.apiKey),
		openai.WithModel(model),
		openai.WithBaseURL(c.baseURL),
	}
	if c.httpClient != nil {
		options = append(options, openai.WithHTTPClient(c.httpClient))
	}

	llm, err := openai.New(options...)
	if err != nil {
		return ChatResponse{}, apperrors.Wrap(err, "failed to create OpenAI client")
	}

	completionBuilder := &strings.Builder{}
	promptTokens := estimateTokens(req.Prompt)

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
			text := string(chunk)
			completionBuilder.WriteString(text)
			return req.StreamFn(callCtx, text)
		}))
	}

	response, err := llm.GenerateContent(ctx, []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart(req.Prompt)},
		},
	}, callOptions...)
	if err != nil {
		return ChatResponse{}, apperrors.Wrap(err, "failed to generate completion")
	}

	content := completionBuilder.String()
	if content == "" {
		for _, choice := range response.Choices {
			content += choice.Content
		}
	}

	usage := Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: estimateTokens(content),
	}

	return ChatResponse{Content: content, Usage: usage}, nil
}

// ListModels queries the OpenAI-compatible /models endpoint and returns available ids.
func (c *OpenAIClient) ListModels(ctx context.Context) ([]ModelInfo, error) {
	client := c.httpClient
	if client == nil {
		client = http.DefaultClient
	}

	endpoint := fmt.Sprintf("%s/models", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to build models request")
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to fetch models")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, apperrors.Wrap(fmt.Errorf("models endpoint returned status %d", resp.StatusCode), "unexpected status from llm provider")
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, apperrors.Wrap(err, "failed to decode models response")
	}

	models := make([]ModelInfo, 0, len(payload.Data))
	for _, item := range payload.Data {
		if item.ID == "" {
			continue
		}
		models = append(models, ModelInfo{ID: item.ID, Provider: ProviderFromModelID(item.ID)})
	}

	return models, nil
}

func estimateTokens(text string) int {
	if strings.TrimSpace(text) == "" {
		return 0
	}
	enc, err := tiktoken.EncodingForModel("gpt-4o-mini")
	if err != nil {
		enc, err = tiktoken.GetEncoding("cl100k_base")
	}
	if err != nil {
		return (len([]rune(text)) / 4) + 1
	}

	tokens := enc.Encode(text, nil, nil)
	return len(tokens)
}
