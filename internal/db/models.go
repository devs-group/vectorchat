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
	SaveMessages       bool      `json:"save_messages" db:"save_messages"`
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

type ChatMessage struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ChatbotID uuid.UUID `json:"chatbot_id" db:"chatbot_id"`
	SessionID uuid.UUID `json:"session_id" db:"session_id"`
	Role      string    `json:"role" db:"role"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type AnswerRevision struct {
	ID                uuid.UUID       `json:"id" db:"id"`
	ChatbotID         uuid.UUID       `json:"chatbot_id" db:"chatbot_id"`
	OriginalMessageID *uuid.UUID      `json:"original_message_id" db:"original_message_id"`
	Question          string          `json:"question" db:"question"`
	OriginalAnswer    string          `json:"original_answer" db:"original_answer"`
	RevisedAnswer     string          `json:"revised_answer" db:"revised_answer"`
	QuestionEmbedding pgvector.Vector `json:"question_embedding" db:"question_embedding"`
	RevisionReason    *string         `json:"revision_reason" db:"revision_reason"`
	RevisedBy         string          `json:"revised_by" db:"revised_by"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
	IsActive          bool            `json:"is_active" db:"is_active"`
}

type AnswerRevisionWithEmbedding struct {
	ID                uuid.UUID  `json:"id"`
	ChatbotID         uuid.UUID  `json:"chatbot_id"`
	OriginalMessageID *uuid.UUID `json:"original_message_id"`
	Question          string     `json:"question"`
	OriginalAnswer    string     `json:"original_answer"`
	RevisedAnswer     string     `json:"revised_answer"`
	QuestionEmbedding []float32  `json:"question_embedding"`
	RevisionReason    *string    `json:"revision_reason"`
	RevisedBy         string     `json:"revised_by"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	IsActive          bool       `json:"is_active"`
	Similarity        float64    `json:"similarity,omitempty"` // Used when returning search results
}

func (r *AnswerRevision) ToAnswerRevisionWithEmbedding() *AnswerRevisionWithEmbedding {
	return &AnswerRevisionWithEmbedding{
		ID:                r.ID,
		ChatbotID:         r.ChatbotID,
		OriginalMessageID: r.OriginalMessageID,
		Question:          r.Question,
		OriginalAnswer:    r.OriginalAnswer,
		RevisedAnswer:     r.RevisedAnswer,
		QuestionEmbedding: r.QuestionEmbedding.Slice(),
		RevisionReason:    r.RevisionReason,
		RevisedBy:         r.RevisedBy,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
		IsActive:          r.IsActive,
	}
}

type Conversation struct {
	SessionID      uuid.UUID `json:"session_id" db:"session_id"`
	LastMessageAt  time.Time `json:"last_message_at" db:"last_message_at"`
	FirstMessageAt time.Time `json:"first_message_at" db:"first_message_at"`
}
