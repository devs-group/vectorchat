package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

// APIKeyService handles API key operations
type APIKeyService struct {
	repo APIKeyRepository
}

// NewAPIKeyService creates a new APIKeyService
func NewAPIKeyService(repo APIKeyRepository) *APIKeyService {
	return &APIKeyService{
		repo: repo,
	}
}

// CreateNewAPIKey generates a new plain text API key and its bcrypt hash.
// It returns the plain text key (show ONCE to the user) and the hash (store in DB).
func (s *APIKeyService) CreateNewAPIKey(name string, expiresAt *time.Time) (plainTextKey string, hashedKey string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		err = errors.Wrap(err, "failed to generate random bytes")
		return
	}

	plainTextKey = "vc_" + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plainTextKey), bcrypt.DefaultCost)
	if err != nil {
		err = errors.Wrap(err, "failed to generate hashed key")
		plainTextKey = ""
		return
	}
	hashedKey = string(hashedBytes)

	return // Return plainTextKey, hashedKey, nil
}

// IsAPIKeyValid compares a provided plain text key against a stored bcrypt hash.
func (s *APIKeyService) IsAPIKeyValid(storedHashedKey, providedPlainTextKey string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(storedHashedKey), []byte(providedPlainTextKey))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	return false, errors.Wrap(err, "failed to compare api key hash")
}

// APIKeyCreateRequest represents the request to create an API key
type APIKeyCreateRequest struct {
	Name      string  `json:"name"`
	ExpiresAt *string `json:"expires_at,omitempty"`
}

// APIKeysListResponse represents the response for listing API keys with pagination
type APIKeysListResponse struct {
	APIKeys    []*APIKeyResponse   `json:"api_keys"`
	Pagination *PaginationMetadata `json:"pagination"`
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

// ParseAPIKeyRequest parses and validates an API key creation request
func (s *APIKeyService) ParseAPIKeyRequest(req *APIKeyCreateRequest) (string, *time.Time, error) {
	name := req.Name
	var expiresAt *time.Time

	// Parse expiration date if provided
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return "", nil, apperrors.Wrap(err, "invalid expiration date format")
		}
		expiresAt = &parsedTime
	}

	return name, expiresAt, nil
}

// CreateAPIKey creates a new API key
func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID, name string, expiresAt *time.Time) (*APIKeyResponse, string, error) {
	// Generate the API key
	plainTextKey, hashedKey, err := s.CreateNewAPIKey(name, expiresAt)
	if err != nil {
		return nil, "", apperrors.Wrap(err, "failed to generate API key")
	}

	// Create the API key record
	apiKey := &APIKey{
		ID:        uuid.New().String(),
		UserID:    userID,
		Key:       hashedKey,
		Name:      &name,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	err = s.repo.Create(ctx, apiKey)
	if err != nil {
		return nil, "", apperrors.Wrap(err, "failed to create API key")
	}

	return apiKey.ToResponse(), plainTextKey, nil
}

// ParsePaginationParams parses and validates pagination parameters
func (s *APIKeyService) ParsePaginationParams(pageStr, limitStr string) (page, limit, offset int) {
	page = 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit = 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset = (page - 1) * limit
	return page, limit, offset
}

// GetAPIKeysWithPagination gets API keys for a user with pagination support
func (s *APIKeyService) GetAPIKeysWithPagination(ctx context.Context, userID string, page, limit, offset int) (*APIKeysListResponse, error) {
	apiKeys, total, err := s.repo.FindByUserIDWithPagination(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]*APIKeyResponse, len(apiKeys))
	for i, apiKey := range apiKeys {
		responses[i] = apiKey.ToResponse()
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	hasNext := page < totalPages
	hasPrev := page > 1

	return &APIKeysListResponse{
		APIKeys: responses,
		Pagination: &PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	}, nil
}

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, id string, userID string) error {
	if id == "" {
		return apperrors.Wrap(apperrors.ErrAPIKeyNotFound, "API key ID is required")
	}
	return s.repo.Revoke(ctx, id, userID)
}
