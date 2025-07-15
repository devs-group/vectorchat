package models

import "time"

type APIResponse struct {
	Error   string `json:"error,omitempty" example:"validation failed"`
	Message string `json:"message,omitempty" example:"Operation completed successfully"`
	Data    any    `json:"data,omitempty"`
}

type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

type PaginationMetadata struct {
	Page       int   `json:"page" example:"1"`
	Limit      int   `json:"limit" example:"10"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int   `json:"total_pages" example:"10"`
	HasNext    bool  `json:"has_next" example:"true"`
	HasPrev    bool  `json:"has_prev" example:"false"`
}

type PaginatedResponse struct {
	Data   interface{} `json:"data"`
	Total  int64       `json:"total" example:"100"`
	Offset int         `json:"offset" example:"0"`
	Limit  int         `json:"limit" example:"10"`
}

type User struct {
	ID        string    `json:"id" example:"user_123"`
	Email     string    `json:"email" example:"user@example.com"`
	Name      string    `json:"name" example:"John Doe"`
	Provider  string    `json:"provider" example:"google"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

type UserResponse struct {
	User User `json:"user"`
}

type SessionResponse struct {
	User User `json:"user"`
}

type LoginResponse struct {
	RedirectURL string `json:"redirect_url" example:"https://app.example.com/dashboard"`
}
