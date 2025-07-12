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

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, id string, userID string) error {
	return s.repo.Revoke(ctx, id, userID)
}
