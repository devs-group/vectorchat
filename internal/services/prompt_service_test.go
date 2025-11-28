package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/yourusername/vectorchat/internal/llm"
)

type stubLLM struct {
	response string
	err      error
	prompt   string
}

func (s *stubLLM) Chat(_ context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	s.prompt = req.Prompt
	if s.err != nil {
		return llm.ChatResponse{}, s.err
	}
	return llm.ChatResponse{Content: s.response}, nil
}

func (s *stubLLM) ListModels(context.Context) ([]llm.ModelInfo, error) {
	return nil, nil
}

func TestGenerateSystemPrompt(t *testing.T) {
	stub := &stubLLM{response: "You are a helpful assistant."}
	service := NewPromptService(stub, "gpt-4o-mini")

	result, err := service.GenerateSystemPrompt(context.Background(), "Help with shipping questions", "concise")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == "" {
		t.Fatalf("expected a prompt, got empty string")
	}

	if !strings.Contains(stub.prompt, "Help with shipping questions") {
		t.Fatalf("prompt template did not include purpose; got %s", stub.prompt)
	}
}

func TestGenerateSystemPromptRequiresPurpose(t *testing.T) {
	stub := &stubLLM{}
	service := NewPromptService(stub, "gpt-4o-mini")

	_, err := service.GenerateSystemPrompt(context.Background(), "   ", "formal")
	if err == nil {
		t.Fatalf("expected error when purpose is empty")
	}
}

func TestGenerateSystemPromptPropagatesError(t *testing.T) {
	stub := &stubLLM{err: errors.New("llm failed")}
	service := NewPromptService(stub, "gpt-4o-mini")

	_, err := service.GenerateSystemPrompt(context.Background(), "Assist", "")
	if err == nil || !strings.Contains(err.Error(), "llm failed") {
		t.Fatalf("expected llm error to propagate, got %v", err)
	}
}
