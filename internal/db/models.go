package db

import (
	"time"

	"github.com/google/uuid"
)

// User represents an authenticated user
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIKey represents an API key for a user
type APIKey struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// Document represents a document with its vector embedding
type Document struct {
	ID        string
	Content   string
	Embedding []float32
	ChatbotID uuid.UUID
}

// Chatbot represents a configurable AI assistant
type Chatbot struct {
	ID                 uuid.UUID `json:"id"`
	UserID             string    `json:"user_id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	SystemInstructions string    `json:"system_instructions"`
	ModelName          string    `json:"model_name"`
	TemperatureParam   float64   `json:"temperature_param"`
	MaxTokens          int       `json:"max_tokens"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
