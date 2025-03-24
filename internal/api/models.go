package api

import "time"

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
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// UserResponse represents the response for user-related endpoints
type UserResponse struct {
	User User `json:"user"`
}

// APIKeyResponse represents the response for API key endpoints
type APIKeyResponse struct {
	APIKey APIKey `json:"api_key"`
}

// APIKeysResponse represents the response for listing API keys
type APIKeysResponse struct {
	APIKeys []APIKey `json:"api_keys"`
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
type ChatMessage struct {
	ChatID string `json:"chat_id" example:"test-session"`
	Query  string `json:"query" example:"What is this project about?"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Message  string `json:"message"`
	ChatID   string `json:"chat_id"`
	Context  string `json:"context,omitempty"`
}

// FileUploadResponse represents a file upload response
type FileUploadResponse struct {
	Filename string `json:"filename"`
	ChatID   string `json:"chat_id"`
	Size     int64  `json:"size"`
}

// ChatFilesResponse represents the response for listing chat files
type ChatFilesResponse struct {
	Files []ChatFile `json:"files"`
}

// ChatFile represents a chat file
type ChatFile struct {
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	UpdatedAt time.Time `json:"updated_at"`
} 