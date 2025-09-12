package models

import (
	"time"

	"github.com/google/uuid"
)

type ChatbotCreateRequest struct {
	Name               string  `json:"name" binding:"required" example:"Customer Support Bot"`
	Description        string  `json:"description" binding:"required" example:"AI assistant for customer support"`
	SystemInstructions string  `json:"system_instructions" binding:"required" example:"You are a helpful customer support assistant"`
	ModelName          string  `json:"model_name" binding:"required" example:"gpt-3.5-turbo"`
	TemperatureParam   float64 `json:"temperature_param" binding:"required,min=0,max=2" example:"0.7"`
	MaxTokens          int     `json:"max_tokens" binding:"required,min=1" example:"1000"`
}

type ChatbotUpdateRequest struct {
	Name               *string  `json:"name,omitempty" example:"Updated Bot Name"`
	Description        *string  `json:"description,omitempty" example:"Updated description"`
	SystemInstructions *string  `json:"system_instructions,omitempty" example:"Updated system instructions"`
	ModelName          *string  `json:"model_name,omitempty" example:"gpt-4"`
	TemperatureParam   *float64 `json:"temperature_param,omitempty" example:"0.8"`
	MaxTokens          *int     `json:"max_tokens,omitempty" example:"1500"`
}

type ChatbotResponse struct {
	ID                 uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID             string    `json:"user_id" example:"user_123"`
	Name               string    `json:"name" example:"Customer Support Bot"`
	Description        string    `json:"description" example:"AI assistant for customer support"`
	SystemInstructions string    `json:"system_instructions" example:"You are a helpful customer support assistant"`
	ModelName          string    `json:"model_name" example:"gpt-3.5-turbo"`
	TemperatureParam   float64   `json:"temperature_param" example:"0.7"`
	MaxTokens          int       `json:"max_tokens" example:"1000"`
	IsEnabled          bool      `json:"is_enabled" example:"true"`
	CreatedAt          time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt          time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
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
