package models

import (
	"time"

	"github.com/google/uuid"
)

type ChatbotCreateRequest struct {
	Name                   string      `json:"name" binding:"required" example:"Customer Support Bot"`
	Description            string      `json:"description" binding:"required" example:"AI assistant for customer support"`
	SystemInstructions     string      `json:"system_instructions" binding:"required" example:"You are a helpful customer support assistant"`
	ModelName              string      `json:"model_name" binding:"required" example:"gpt-3.5-turbo"`
	TemperatureParam       float64     `json:"temperature_param" binding:"required,min=0,max=2" example:"0.7"`
	MaxTokens              int         `json:"max_tokens" binding:"required,min=1" example:"1000"`
	SaveMessages           *bool       `json:"save_messages,omitempty" example:"true"`
	UseMaxTokens           *bool       `json:"use_max_tokens,omitempty" example:"true"`
	IsEnabled              *bool       `json:"is_enabled,omitempty" example:"false"`
	SharedKnowledgeBaseIDs []uuid.UUID `json:"shared_knowledge_base_ids,omitempty"`
}

type ChatbotUpdateRequest struct {
	Name                   *string     `json:"name,omitempty" example:"Updated Bot Name"`
	Description            *string     `json:"description,omitempty" example:"Updated description"`
	SystemInstructions     *string     `json:"system_instructions,omitempty" example:"Updated system instructions"`
	ModelName              *string     `json:"model_name,omitempty" example:"gpt-4"`
	TemperatureParam       *float64    `json:"temperature_param,omitempty" example:"0.8"`
	MaxTokens              *int        `json:"max_tokens,omitempty" example:"1500"`
	SaveMessages           *bool       `json:"save_messages,omitempty" example:"true"`
	UseMaxTokens           *bool       `json:"use_max_tokens,omitempty" example:"true"`
	SharedKnowledgeBaseIDs []uuid.UUID `json:"shared_knowledge_base_ids,omitempty"`
}

type ChatbotResponse struct {
	ID                     uuid.UUID   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID                 string      `json:"user_id" example:"user_123"`
	Name                   string      `json:"name" example:"Customer Support Bot"`
	Description            string      `json:"description" example:"AI assistant for customer support"`
	SystemInstructions     string      `json:"system_instructions" example:"You are a helpful customer support assistant"`
	ModelName              string      `json:"model_name" example:"gpt-3.5-turbo"`
	TemperatureParam       float64     `json:"temperature_param" example:"0.7"`
	MaxTokens              int         `json:"max_tokens" example:"1000"`
	SaveMessages           bool        `json:"save_messages" example:"true"`
	UseMaxTokens           bool        `json:"use_max_tokens" example:"true"`
	IsEnabled              bool        `json:"is_enabled" example:"true"`
	CreatedAt              time.Time   `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt              time.Time   `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	AIMessagesAmount       int64       `json:"ai_messages_amount" example:"42"`
	SharedKnowledgeBaseIDs []uuid.UUID `json:"shared_knowledge_base_ids"`
}

type ChatbotsListResponse struct {
	Chatbots []ChatbotResponse `json:"chatbots"`
}

type ChatbotToggleRequest struct {
	IsEnabled bool `json:"is_enabled" binding:"required" example:"false"`
}

type ChatMessageRequest struct {
	Query     string  `json:"query" binding:"required" example:"Hello, how can you help me?"`
	SessionID *string `json:"session_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// CreateRevisionRequest represents a request to create an answer revision
type CreateRevisionRequest struct {
	ChatbotID         uuid.UUID  `json:"chatbot_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	OriginalMessageID *uuid.UUID `json:"original_message_id,omitempty" example:"660e8400-e29b-41d4-a716-446655440001"`
	Question          string     `json:"question" binding:"required" example:"What are your business hours?"`
	OriginalAnswer    string     `json:"original_answer" binding:"required" example:"We are open 24/7"`
	RevisedAnswer     string     `json:"revised_answer" binding:"required" example:"We are open Monday-Friday 9AM-5PM EST"`
	RevisionReason    *string    `json:"revision_reason,omitempty" example:"Incorrect business hours"`
	RevisedBy         string     `json:"revised_by" binding:"required" example:"admin_user_123"`
}

// UpdateRevisionRequest represents a request to update an existing revision
type UpdateRevisionRequest struct {
	Question       *string `json:"question,omitempty" example:"What are your updated business hours?"`
	RevisedAnswer  *string `json:"revised_answer,omitempty" example:"We are open Monday-Friday 8AM-6PM EST"`
	RevisionReason *string `json:"revision_reason,omitempty" example:"Updated hours for summer schedule"`
	IsActive       *bool   `json:"is_active,omitempty" example:"true"`
}

// RevisionResponse represents an answer revision in API responses
type RevisionResponse struct {
	ID                uuid.UUID  `json:"id" example:"770e8400-e29b-41d4-a716-446655440002"`
	ChatbotID         uuid.UUID  `json:"chatbot_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OriginalMessageID *uuid.UUID `json:"original_message_id,omitempty" example:"660e8400-e29b-41d4-a716-446655440001"`
	Question          string     `json:"question" example:"What are your business hours?"`
	OriginalAnswer    string     `json:"original_answer" example:"We are open 24/7"`
	RevisedAnswer     string     `json:"revised_answer" example:"We are open Monday-Friday 9AM-5PM EST"`
	RevisionReason    *string    `json:"revision_reason,omitempty" example:"Incorrect business hours"`
	RevisedBy         string     `json:"revised_by" example:"admin_user_123"`
	CreatedAt         time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt         time.Time  `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	IsActive          bool       `json:"is_active" example:"true"`
	Similarity        float64    `json:"similarity,omitempty" example:"0.95"`
}

type ConversationResponse struct {
	SessionID           uuid.UUID `json:"session_id" example:"880e8400-e29b-41d4-a716-446655440003"`
	FirstMessageContent string    `json:"first_message_content" example:"Hi there"`
	FirstMessageAt      time.Time `json:"first_message_at" example:"2023-01-01T00:00:00Z"`
	LastMessageAt       time.Time `json:"last_message_at" example:"2023-01-01T00:00:00Z"`
}

type ConversationPagination struct {
	Page            int   `json:"page" example:"1"`
	PerPage         int   `json:"per_page" example:"20"`
	TotalItems      int64 `json:"total_items" example:"100"`
	TotalPages      int   `json:"total_pages" example:"5"`
	HasNextPage     bool  `json:"has_next_page" example:"true"`
	HasPrevPage     bool  `json:"has_prev_page" example:"false"`
	Offset          int   `json:"offset" example:"0"`
	RequestedOffset *int  `json:"requested_offset,omitempty" example:"40"`
	NextPage        *int  `json:"next_page,omitempty" example:"2"`
	PrevPage        *int  `json:"prev_page,omitempty" example:"1"`
	NextOffset      *int  `json:"next_offset,omitempty" example:"20"`
	PrevOffset      *int  `json:"prev_offset,omitempty" example:"0"`
}

// ConversationsResponse represents a conversation with pagination metadata
type ConversationsResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	Pagination    ConversationPagination `json:"pagination"`
}

// MessageDetails represents individual message details in a conversation
type MessageDetails struct {
	ID        uuid.UUID `json:"id" example:"990e8400-e29b-41d4-a716-446655440004"`
	ChatbotID uuid.UUID `json:"chatbot_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role      string    `json:"role" example:"user"`
	Content   string    `json:"content" example:"Hello, how can you help me?"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
}

// ConversationsListResponse represents a list of conversations
type ConversationsListResponse struct {
	Conversations []ConversationResponse `json:"conversations"`
	TotalCount    int                    `json:"total_count" example:"100"`
	Limit         int                    `json:"limit" example:"20"`
	Offset        int                    `json:"offset" example:"0"`
}

// RevisionsListResponse represents a list of revisions
type RevisionsListResponse struct {
	Revisions []RevisionResponse `json:"revisions"`
}

type ChatResponse struct {
	Message string `json:"message" example:"Hello! I'm here to help you with any questions you might have."`
	ChatID  string `json:"chat_id" example:"chat_123"`
	Context string `json:"context,omitempty" example:"Previous conversation context"`
}

// TextUploadRequest represents a plain text payload to index for a chatbot
type TextUploadRequest struct {
	Text string `json:"text" binding:"required" example:"Paste your knowledge base text here."`
}

// WebsiteUploadRequest represents a request to index a website starting at a URL
type WebsiteUploadRequest struct {
	URL string `json:"url" binding:"required" example:"https://docs.example.com"`
}

type FileInfo struct {
	Filename   string    `json:"filename" example:"document.pdf"`
	ID         uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Size       int64     `json:"size" example:"1024"`
	UploadedAt time.Time `json:"uploaded_at" example:"2023-01-01T00:00:00Z"`
}

type FileUploadResponse struct {
	Message   string    `json:"message" example:"File uploaded successfully"`
	ChatID    uuid.UUID `json:"chat_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ChatbotID uuid.UUID `json:"chatbot_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	File      string    `json:"file" example:"document.pdf"`
	Filename  string    `json:"filename,omitempty" example:"document.pdf"`
	Size      int64     `json:"size,omitempty" example:"1024"`
}

type ChatFilesResponse struct {
	ChatID uuid.UUID  `json:"chat_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Files  []FileInfo `json:"files"`
}

type ChatFilesListResponse struct {
	Files []struct {
		Filename  string    `json:"filename" example:"document.pdf"`
		Size      int64     `json:"size" example:"1024"`
		UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	} `json:"files"`
}

type TextSourceInfo struct {
	ID         uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title      string    `json:"title" example:"text-20240101-120000.txt"`
	Size       int64     `json:"size" example:"2048"`
	UploadedAt time.Time `json:"uploaded_at" example:"2023-01-01T00:00:00Z"`
}

type TextSourcesResponse struct {
	ChatID  uuid.UUID        `json:"chat_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Sources []TextSourceInfo `json:"sources"`
}
