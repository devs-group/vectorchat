package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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

// GetAPIKey finds an API key by ID
func (s *APIKeyService) GetAPIKey(ctx context.Context, id string) (*APIKeyResponse, error) {
	apiKey, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return apiKey.ToResponse(), nil
}

// GetAPIKeysForUser gets all API keys for a user
func (s *APIKeyService) GetAPIKeysForUser(ctx context.Context, userID string) ([]*APIKeyResponse, error) {
	apiKeys, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*APIKeyResponse, len(apiKeys))
	for i, apiKey := range apiKeys {
		responses[i] = apiKey.ToResponse()
	}

	return responses, nil
}

// GetAPIKeysWithPagination gets API keys for a user with pagination support
func (s *APIKeyService) GetAPIKeysWithPagination(ctx context.Context, userID string, offset, limit int) (*PaginatedResponse, error) {
	apiKeys, total, err := s.repo.FindByUserIDWithPagination(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]*APIKeyResponse, len(apiKeys))
	for i, apiKey := range apiKeys {
		responses[i] = apiKey.ToResponse()
	}

	return NewPaginatedResponse(responses, total, offset, limit), nil
}

// FindAPIKeyByPlaintext finds an API key by comparing against stored hashes
func (s *APIKeyService) FindAPIKeyByPlaintext(ctx context.Context, plainTextKey string) (*APIKey, error) {
	compareFunc := func(hashedKey string) (bool, error) {
		return s.IsAPIKeyValid(hashedKey, plainTextKey)
	}

	return s.repo.FindByHashComparison(ctx, compareFunc)
}

// ValidateAPIKey validates an API key and returns the associated user ID
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, plainTextKey string) (string, error) {
	apiKey, err := s.FindAPIKeyByPlaintext(ctx, plainTextKey)
	if err != nil {
		return "", apperrors.Wrap(err, "API key validation failed")
	}

	// Check if the key is revoked
	if apiKey.RevokedAt != nil {
		return "", apperrors.ErrInvalidAPIKey
	}

	// Check if the key is expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return "", apperrors.ErrInvalidAPIKey
	}

	return apiKey.UserID, nil
}

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, id string, userID string) error {
	return s.repo.Revoke(ctx, id, userID)
}

// DeleteAPIKey deletes an API key
func (s *APIKeyService) DeleteAPIKey(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// IsAPIKeyRevoked checks if an API key is revoked
func (s *APIKeyService) IsAPIKeyRevoked(ctx context.Context, id string) (bool, error) {
	return s.repo.IsRevoked(ctx, id)
}

// IsAPIKeyExpired checks if an API key is expired
func (s *APIKeyService) IsAPIKeyExpired(ctx context.Context, id string) (bool, error) {
	return s.repo.IsExpired(ctx, id)
}

// SetAPIKeyExpiration sets an expiration date for an API key
func (s *APIKeyService) SetAPIKeyExpiration(ctx context.Context, id, userID string, expiresAt time.Time) error {
	apiKey, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Check ownership
	if apiKey.UserID != userID {
		return apperrors.ErrAPIKeyNotFound
	}

	apiKey.ExpiresAt = &expiresAt
	return s.repo.Create(ctx, apiKey) // This will update due to ON CONFLICT clause
}

// CleanupExpiredKeys removes expired API keys
func (s *APIKeyService) CleanupExpiredKeys(ctx context.Context) error {
	// This would typically be a background job
	// For now, we'll just mark them as revoked
	// TODO: Implement proper cleanup logic
	return nil
}
