package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Provider  string    `json:"provider" db:"provider"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// APIKey represents an API key in the system
type APIKey struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Key       string     `json:"key" db:"key"` // Stored as hashed value
	Name      *string    `json:"name" db:"name"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// APIKeyResponse represents the response when returning API keys to the client
type APIKeyResponse struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Name      *string    `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	// Key is intentionally omitted for security
}

// ToResponse converts APIKey to APIKeyResponse (without the actual key)
func (a *APIKey) ToResponse() *APIKeyResponse {
	return &APIKeyResponse{
		ID:        a.ID,
		UserID:    a.UserID,
		Name:      a.Name,
		CreatedAt: a.CreatedAt,
		ExpiresAt: a.ExpiresAt,
		RevokedAt: a.RevokedAt,
	}
}

// Chatbot represents a configurable AI assistant
type Chatbot struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	UserID             string    `json:"user_id" db:"user_id"`
	Name               string    `json:"name" db:"name"`
	Description        string    `json:"description" db:"description"`
	SystemInstructions string    `json:"system_instructions" db:"system_instructions"`
	ModelName          string    `json:"model_name" db:"model_name"`
	TemperatureParam   float64   `json:"temperature_param" db:"temperature_param"`
	MaxTokens          int       `json:"max_tokens" db:"max_tokens"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// Document represents a document with its vector embedding
type Document struct {
	ID         string          `json:"id" db:"id"`
	Content    []byte          `json:"content" db:"content"`
	Embedding  pgvector.Vector `json:"embedding" db:"embedding"`
	ChatbotID  uuid.UUID       `json:"chatbot_id" db:"chatbot_id"`
	FileID     *uuid.UUID      `json:"file_id,omitempty" db:"file_id"`
	ChunkIndex *int            `json:"chunk_index,omitempty" db:"chunk_index"`
}

// DocumentWithEmbedding represents a document with embedding as float32 slice for easier handling
type DocumentWithEmbedding struct {
	ID         string     `json:"id"`
	Content    []byte     `json:"content"`
	Embedding  []float32  `json:"embedding"`
	ChatbotID  uuid.UUID  `json:"chatbot_id"`
	FileID     *uuid.UUID `json:"file_id,omitempty"`
	ChunkIndex *int       `json:"chunk_index,omitempty"`
}

// ToDocumentWithEmbedding converts Document to DocumentWithEmbedding
func (d *Document) ToDocumentWithEmbedding() *DocumentWithEmbedding {
	return &DocumentWithEmbedding{
		ID:         d.ID,
		Content:    d.Content,
		Embedding:  d.Embedding.Slice(),
		ChatbotID:  d.ChatbotID,
		FileID:     d.FileID,
		ChunkIndex: d.ChunkIndex,
	}
}

// File represents a file uploaded to a chatbot
type File struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ChatbotID  uuid.UUID `json:"chatbot_id" db:"chatbot_id"`
	Filename   string    `json:"filename" db:"filename"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}
