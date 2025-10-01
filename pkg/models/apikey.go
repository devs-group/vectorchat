package models

import "time"

type APIKeyCreateRequest struct {
	Name      string  `json:"name" binding:"required" example:"My API Key"`
	ExpiresAt *string `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
}

type APIKeyResponse struct {
	ID        string     `json:"id" example:"api_key_123"`
	ClientID  string     `json:"client_id" example:"vc-client-123"`
	UserID    string     `json:"user_id" example:"user_123"`
	Name      *string    `json:"name" example:"My API Key"`
	CreatedAt time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
}

type APIKeyCreateResponse struct {
	ClientID     string     `json:"client_id" example:"vc-client-123"`
	ClientSecret string     `json:"client_secret" example:"client-secret-value"`
	Name         *string    `json:"name,omitempty" example:"My integration"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
	Message      string     `json:"message" example:"OAuth client created successfully. Save the secret as it won't be shown again."`
}

type APIKeysListResponse struct {
	APIKeys    []*APIKeyResponse   `json:"api_keys"`
	Pagination *PaginationMetadata `json:"pagination"`
}

type APIKeysResponse struct {
	APIKeys    []APIKeyResponse   `json:"api_keys"`
	Pagination PaginationMetadata `json:"pagination"`
}
