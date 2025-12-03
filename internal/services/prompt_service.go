package services

import (
	"context"
	"fmt"
	"strings"

	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/llm"
)

// PromptService handles generation of tailored system prompts.
type PromptService struct {
	llmClient  llm.Client
	modelAlias string
}

// NewPromptService constructs a PromptService with the provided LLM client.
func NewPromptService(llmClient llm.Client, modelAlias string) *PromptService {
	return &PromptService{
		llmClient:  llmClient,
		modelAlias: modelAlias,
	}
}

// GenerateSystemPrompt crafts a concise system prompt for a chatbot based on purpose and tone.
func (s *PromptService) GenerateSystemPrompt(ctx context.Context, purpose, tone string) (string, error) {
	if strings.TrimSpace(purpose) == "" {
		return "", apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "purpose is required")
	}

	if tone == "" {
		tone = "balanced"
	}

	template := buildPromptTemplate(purpose, tone)
	resp, err := s.llmClient.Chat(ctx, llm.ChatRequest{
		Prompt:      template,
		Temperature: 0.4,
		Model:       s.modelAlias,
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}

func buildPromptTemplate(purpose, tone string) string {
	return fmt.Sprintf(`You are a senior AI engineer and prompt writer. Create a concise, well-structured system prompt for an AI assistant.
Constraints:
- Keep it under 300 words.
- Emphasize clarity, factual accuracy, and the assistant's role.
- Write excellent system prompt.
- Include 3-5 bullet guardrails (tone, safety, citations if relevant).
- Write in Markdown format.
- Keep formatting simple (Markdown list). Do not include any preamble.
- Make sure that the system prompt gives clear constraints on the purpose and should probably avoid answering questions outside of the purpose.

Assistant purpose: %s
Preferred tone/style: %s

Return only the final system prompt.`, strings.TrimSpace(purpose), strings.TrimSpace(tone))
}
