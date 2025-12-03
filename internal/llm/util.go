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
	case "gemini-default":
		return "google"
	case "gpt-4o-mini":
		return "openai"
	}
	if strings.Contains(lowered, "/") {
		parts := strings.SplitN(lowered, "/", 2)
		return parts[0]
	}
	switch {
	case strings.HasPrefix(lowered, "gpt"):
		return "openai"
	case strings.HasPrefix(lowered, "gemini"):
		return "google"
	default:
		return ""
	}
}

// IsAdvancedModel heuristically marks models that should be hidden on basic plans.
func IsAdvancedModel(info ModelInfo) bool {
	provider := strings.ToLower(info.Provider)
	if provider != "" && provider != "openai" && provider != "google" {
		return true
	}

	id := strings.ToLower(info.ID)
	switch {
	case strings.HasPrefix(id, "gemini"):
		return true
	case strings.Contains(id, "gpt-4"):
		return true
	case (strings.Contains(id, "gpt-5") || strings.Contains(id, "gpt5")):
		return true
	}

	return false
}
