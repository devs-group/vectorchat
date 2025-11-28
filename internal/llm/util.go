package llm

import "strings"

// ProviderFromModelID infers a provider name from a model identifier.
func ProviderFromModelID(id string) string {
	if id == "" {
		return ""
	}
	lowered := strings.ToLower(id)
	switch lowered {
	case "chat-default", "prompt-helper":
		return "openai"
	case "claude-default":
		return "anthropic"
	case "gemini-default":
		return "google"
	case "gpt5-default", "gpt5-mini", "gpt5-nano":
		return "openai"
	}
	if strings.Contains(lowered, "/") {
		parts := strings.SplitN(lowered, "/", 2)
		return parts[0]
	}
	switch {
	case strings.HasPrefix(lowered, "gpt"):
		return "openai"
	case strings.HasPrefix(lowered, "claude"):
		return "anthropic"
	case strings.HasPrefix(lowered, "gemini"):
		return "google"
	case strings.HasPrefix(lowered, "mistral"):
		return "mistral"
	default:
		return ""
	}
}

// IsAdvancedModel heuristically marks models that should be hidden on basic plans.
func IsAdvancedModel(info ModelInfo) bool {
	provider := strings.ToLower(info.Provider)
	if provider != "" && provider != "openai" {
		return true
	}

	id := strings.ToLower(info.ID)
	switch {
	case strings.Contains(id, "claude"), strings.Contains(id, "sonnet"), strings.Contains(id, "opus"):
		return true
	case strings.Contains(id, "gpt-4"):
		return true
	case (strings.Contains(id, "gpt-5") || strings.Contains(id, "gpt5")) && !strings.Contains(id, "nano"):
		return true
	}

	return false
}
