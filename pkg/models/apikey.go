package models

import "time"

type APIKeyCreateRequest struct {
	Name      string  `json:"name" binding:"required" example:"My API Key"`
	ExpiresAt *string `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
}

type APIKeyResponse struct {
	ID        string     `json:"id" example:"api_key_123"`
	UserID    string     `json:"user_id" example:"user_123"`
	Name      *string    `json:"name" example:"My API Key"`
	CreatedAt time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" example:"2023-06-01T00:00:00Z"`
	// Key is intentionally omitted for security
}

type APIKeyCreateResponse struct {
	APIKey   APIKeyResponse `json:"api_key"`
	PlainKey string         `json:"plain_key" example:"vc_abcd1234efgh5678ijkl9012mnop3456"`
	Message  string         `json:"message" example:"API key created successfully. Save this key as it won't be shown again."`
}

type APIKeysListResponse struct {
	APIKeys    []*APIKeyResponse   `json:"api_keys"`
	Pagination *PaginationMetadata `json:"pagination"`
}

type APIKeysResponse struct {
	APIKeys    []APIKeyResponse   `json:"api_keys"`
	Pagination PaginationMetadata `json:"pagination"`
}
