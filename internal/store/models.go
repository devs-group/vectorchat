package store

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIKey represents an API key in the system
type APIKey struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Key       string     `json:"key"` // Stored as hashed value
	Name      *string    `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// Document represents a document with its vector embedding
type Document struct {
	ID        string
	Content   []byte
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
