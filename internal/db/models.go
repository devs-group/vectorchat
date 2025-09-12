package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Provider  string    `json:"provider" db:"provider"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type APIKey struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Key       string     `json:"key" db:"key"` // Stored as hashed value
	Name      *string    `json:"name" db:"name"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

type APIKeyResponse struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Name      *string    `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	// Key is intentionally omitted for security
}

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

type Chatbot struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	UserID             string    `json:"user_id" db:"user_id"`
	Name               string    `json:"name" db:"name"`
	Description        string    `json:"description" db:"description"`
	SystemInstructions string    `json:"system_instructions" db:"system_instructions"`
	ModelName          string    `json:"model_name" db:"model_name"`
	TemperatureParam   float64   `json:"temperature_param" db:"temperature_param"`
	MaxTokens          int       `json:"max_tokens" db:"max_tokens"`
	IsEnabled          bool      `json:"is_enabled" db:"is_enabled"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type Document struct {
	ID         string          `json:"id" db:"id"`
	Content    []byte          `json:"content" db:"content"`
	Embedding  pgvector.Vector `json:"embedding" db:"embedding"`
	ChatbotID  uuid.UUID       `json:"chatbot_id" db:"chatbot_id"`
	FileID     *uuid.UUID      `json:"file_id,omitempty" db:"file_id"`
	ChunkIndex *int            `json:"chunk_index,omitempty" db:"chunk_index"`
}

type DocumentWithEmbedding struct {
	ID         string     `json:"id"`
	Content    []byte     `json:"content"`
	Embedding  []float32  `json:"embedding"`
	ChatbotID  uuid.UUID  `json:"chatbot_id"`
	FileID     *uuid.UUID `json:"file_id,omitempty"`
	ChunkIndex *int       `json:"chunk_index,omitempty"`
}

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

type File struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ChatbotID  uuid.UUID `json:"chatbot_id" db:"chatbot_id"`
	Filename   string    `json:"filename" db:"filename"`
	SizeBytes  int64     `json:"size_bytes" db:"size_bytes"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}
