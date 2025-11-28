package models

// LLMModel represents a single LLM model/alias available to users.
type LLMModel struct {
	ID       string `json:"id" example:"chat-default"`
	Label    string `json:"label" example:"Chat Default (GPT-4o Mini)"`
	Provider string `json:"provider,omitempty" example:"openai"`
	Advanced bool   `json:"advanced" example:"false"`
}

// LLMModelsResponse wraps the model list for the API.
type LLMModelsResponse struct {
	Models []LLMModel `json:"models"`
}
