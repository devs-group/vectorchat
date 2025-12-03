package services

import (
	"context"
	"strings"

	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/llm"
	"github.com/yourusername/vectorchat/pkg/models"
)

// LLMService wraps llm client operations and applies simple plan-based filtering.
type LLMService struct {
	client      llm.Client
	fallbackIDs []string
}

// NewLLMService constructs an LLMService.
func NewLLMService(client llm.Client, fallbackIDs []string) *LLMService {
	dedup := make(map[string]struct{})
	unique := make([]string, 0, len(fallbackIDs))
	for _, id := range fallbackIDs {
		if id == "" {
			continue
		}
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, exists := dedup[id]; exists {
			continue
		}
		dedup[id] = struct{}{}
		unique = append(unique, id)
	}

	return &LLMService{client: client, fallbackIDs: unique}
}

// ListModels fetches model metadata from the provider, falling back to the configured list on failure.
func (s *LLMService) ListModels(ctx context.Context) ([]models.LLMModel, error) {
	infos, err := s.client.ListModels(ctx)
	if err != nil {
		if len(s.fallbackIDs) > 0 {
			return s.buildFallbackModels(), apperrors.Wrap(err, "llm: list models failed, returned fallback list")
		}
		return nil, apperrors.Wrap(err, "llm: list models failed")
	}

	modelsResp := make([]models.LLMModel, 0, len(infos))
	for _, info := range infos {
		if !isAllowedProvider(info.Provider) {
			continue
		}
		modelsResp = append(modelsResp, models.LLMModel{
			ID:       info.ID,
			Label:    friendlyModelLabel(info.ID),
			Provider: info.Provider,
			Advanced: llm.IsAdvancedModel(info),
		})
	}
	return modelsResp, nil
}

// FilterByPlan removes advanced models when the current plan forbids them.
func (s *LLMService) FilterByPlan(modelsList []models.LLMModel, allowAdvanced bool) []models.LLMModel {
	if allowAdvanced {
		return modelsList
	}

	filtered := make([]models.LLMModel, 0, len(modelsList))
	for _, m := range modelsList {
		if m.Advanced {
			continue
		}
		filtered = append(filtered, m)
	}
	return filtered
}

func (s *LLMService) buildFallbackModels() []models.LLMModel {
	fallback := make([]models.LLMModel, 0, len(s.fallbackIDs))
	for _, id := range s.fallbackIDs {
		provider := llm.ProviderFromModelID(id)
		if !isAllowedProvider(provider) {
			continue
		}
		fallback = append(fallback, models.LLMModel{
			ID:       id,
			Label:    friendlyModelLabel(id),
			Provider: provider,
			Advanced: llm.IsAdvancedModel(llm.ModelInfo{ID: id, Provider: provider}),
		})
	}
	return fallback
}

func isAllowedProvider(provider string) bool {
	switch strings.ToLower(provider) {
	case "openai", "google":
		return true
	default:
		return false
	}
}

func friendlyModelLabel(id string) string {
	if id == "" {
		return "Unknown"
	}

	switch id {
	case "chat-default":
		return "Chat Default (GPT-4o Mini)"
	case "prompt-helper":
		return "Prompt Helper (GPT-4o Mini)"
	case "gemini-default":
		return "Gemini 1.5 Flash"
	case "gpt-4o-mini":
		return "GPT-4o Mini"
	}

	// Basic prettifier: replace dashes with spaces and capitalize provider-leading names.
	cleaned := strings.ReplaceAll(id, "-", " ")
	cleaned = strings.ReplaceAll(cleaned, "_", " ")

	words := strings.Fields(cleaned)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	if len(words) == 0 {
		return id
	}
	return strings.Join(words, " ")
}
