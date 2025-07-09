package api

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIKey represents an API key
type APIKey struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Key       string     `json:"key"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	RevokedAt *time.Time `json:"revoked_at"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// UserResponse represents the response for user-related endpoints
type UserResponse struct {
	User User `json:"user"`
}

// APIKeyRequest represents the request body to create api key
type APIKeyRequest struct {
	Name      string  `json:"name"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

// APIKeyResponse represents the response for API key endpoints
type APIKeyResponse struct {
	APIKey APIKey `json:"api_key"`
}

// PaginationMetadata represents pagination information
type PaginationMetadata struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// APIKeysResponse represents the response for listing API keys
type APIKeysResponse struct {
	APIKeys    []APIKey           `json:"api_keys"`
	Pagination PaginationMetadata `json:"pagination"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	RedirectURL string `json:"redirect_url"`
}

// SessionResponse represents the current session information
type SessionResponse struct {
	User User `json:"user"`
}

// ChatMessage represents a chat message request
type ChatMessageRequest struct {
	Query string `json:"query" example:"What is this project about?"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Message string `json:"message"`
	ChatID  string `json:"chat_id"`
	Context string `json:"context,omitempty"`
}

// FileUploadResponse represents a file upload response
type FileUploadResponse struct {
	Filename string `json:"filename"`
	ChatID   string `json:"chat_id"`
	Size     int64  `json:"size"`
}

// ChatFilesResponse represents the response for listing chat files
type ChatFilesResponse struct {
	Files []struct {
		Filename  string    `json:"filename"`
		Size      int64     `json:"size"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"files"`
}

// ChatbotCreateRequest represents the request to create a new chatbot
type ChatbotCreateRequest struct {
	Name               string  `json:"name" example:"My AI Assistant"`
	Description        string  `json:"description" example:"A helpful AI assistant for my project"`
	SystemInstructions string  `json:"system_instructions" example:"You are a helpful AI assistant"`
	ModelName          string  `json:"model_name" example:"gpt-4"`
	TemperatureParam   float64 `json:"temperature_param" example:"0.7"`
	MaxTokens          int     `json:"max_tokens" example:"2000"`
}

// ChatbotResponse represents a chatbot in responses
type ChatbotResponse struct {
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
