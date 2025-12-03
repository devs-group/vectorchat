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
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             string     `json:"user_id" db:"user_id"`
	OrganizationID     *uuid.UUID `json:"organization_id,omitempty" db:"organization_id"`
	Name               string     `json:"name" db:"name"`
	Description        string     `json:"description" db:"description"`
	SystemInstructions string     `json:"system_instructions" db:"system_instructions"`
	ModelName          string     `json:"model_name" db:"model_name"`
	TemperatureParam   float64    `json:"temperature_param" db:"temperature_param"`
	MaxTokens          int        `json:"max_tokens" db:"max_tokens"`
	UseMaxTokens       bool       `json:"use_max_tokens" db:"use_max_tokens"`
	SaveMessages       bool       `json:"save_messages" db:"save_messages"`
	IsEnabled          bool       `json:"is_enabled" db:"is_enabled"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

type Document struct {
	ID                    string          `json:"id" db:"id"`
	Content               []byte          `json:"content" db:"content"`
	Embedding             pgvector.Vector `json:"embedding" db:"embedding"`
	ChatbotID             *uuid.UUID      `json:"chatbot_id,omitempty" db:"chatbot_id"`
	FileID                *uuid.UUID      `json:"file_id,omitempty" db:"file_id"`
	ChunkIndex            *int            `json:"chunk_index,omitempty" db:"chunk_index"`
	SharedKnowledgeBaseID *uuid.UUID      `json:"shared_knowledge_base_id,omitempty" db:"shared_knowledge_base_id"`
}

type DocumentWithEmbedding struct {
	ID                    string     `json:"id"`
	Content               []byte     `json:"content"`
	Embedding             []float32  `json:"embedding"`
	ChatbotID             *uuid.UUID `json:"chatbot_id,omitempty"`
	FileID                *uuid.UUID `json:"file_id,omitempty"`
	ChunkIndex            *int       `json:"chunk_index,omitempty"`
	SharedKnowledgeBaseID *uuid.UUID `json:"shared_knowledge_base_id,omitempty"`
}

func (d *Document) ToDocumentWithEmbedding() *DocumentWithEmbedding {
	return &DocumentWithEmbedding{
		ID:                    d.ID,
		Content:               d.Content,
		Embedding:             d.Embedding.Slice(),
		ChatbotID:             d.ChatbotID,
		FileID:                d.FileID,
		ChunkIndex:            d.ChunkIndex,
		SharedKnowledgeBaseID: d.SharedKnowledgeBaseID,
	}
}

type File struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	ChatbotID             *uuid.UUID `json:"chatbot_id,omitempty" db:"chatbot_id"`
	Filename              string     `json:"filename" db:"filename"`
	SizeBytes             int64      `json:"size_bytes" db:"size_bytes"`
	UploadedAt            time.Time  `json:"uploaded_at" db:"uploaded_at"`
	SharedKnowledgeBaseID *uuid.UUID `json:"shared_knowledge_base_id,omitempty" db:"shared_knowledge_base_id"`
}

type SharedKnowledgeBase struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OwnerID        string     `json:"owner_id" db:"owner_id"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" db:"organization_id"`
	Name           string     `json:"name" db:"name"`
	Description    *string    `json:"description" db:"description"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

type ChatbotSharedKnowledgeBase struct {
	ChatbotID             uuid.UUID `json:"chatbot_id" db:"chatbot_id"`
	SharedKnowledgeBaseID uuid.UUID `json:"shared_knowledge_base_id" db:"shared_knowledge_base_id"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
}

type ChatMessage struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ChatbotID uuid.UUID `json:"chatbot_id" db:"chatbot_id"`
	SessionID uuid.UUID `json:"session_id" db:"session_id"`
	Role      string    `json:"role" db:"role"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type LLMUsage struct {
	ID               uuid.UUID `json:"id" db:"id"`
	UserID           string    `json:"user_id" db:"user_id"`
	OrgID            *string   `json:"org_id,omitempty" db:"org_id"`
	TraceID          *string   `json:"trace_id,omitempty" db:"trace_id"`
	ModelAlias       string    `json:"model_alias" db:"model_alias"`
	Provider         *string   `json:"provider,omitempty" db:"provider"`
	PromptTokens     int       `json:"prompt_tokens" db:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens" db:"completion_tokens"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
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
	SessionID           uuid.UUID `json:"session_id" db:"session_id"`
	LastMessageAt       time.Time `json:"last_message_at" db:"last_message_at"`
	FirstMessageAt      time.Time `json:"first_message_at" db:"first_message_at"`
	FirstMessageContent string    `json:"first_message_content" db:"first_message_content"`
}

type Organization struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Slug         string    `json:"slug" db:"slug"`
	Description  *string   `json:"description" db:"description"`
	BillingEmail *string   `json:"billing_email" db:"billing_email"`
	PlanTier     string    `json:"plan_tier" db:"plan_tier"`
	CreatedBy    string    `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type OrganizationMember struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	UserID         string     `json:"user_id" db:"user_id"`
	Role           string     `json:"role" db:"role"`
	InvitedBy      *string    `json:"invited_by" db:"invited_by"`
	InvitedAt      *time.Time `json:"invited_at" db:"invited_at"`
	JoinedAt       time.Time  `json:"joined_at" db:"joined_at"`
	LastActiveAt   *time.Time `json:"last_active_at" db:"last_active_at"`
}

type OrganizationInvite struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	Email          string     `json:"email" db:"email"`
	Role           string     `json:"role" db:"role"`
	TokenHash      string     `json:"token_hash" db:"token_hash"`
	InvitedBy      *string    `json:"invited_by" db:"invited_by"`
	Message        *string    `json:"message" db:"message"`
	ExpiresAt      time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	AcceptedAt     *time.Time `json:"accepted_at" db:"accepted_at"`
}
