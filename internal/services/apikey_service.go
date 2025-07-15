package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// APIKeyService handles API key operations
type APIKeyService struct {
	*CommonService
	repo db.APIKeyRepository
}

// NewAPIKeyService creates a new APIKeyService
func NewAPIKeyService(repo db.APIKeyRepository) *APIKeyService {
	return &APIKeyService{
		CommonService: NewCommonService(),
		repo:          repo,
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

// ParseAPIKeyRequest parses and validates an API key creation request
func (s *APIKeyService) ParseAPIKeyRequest(req *models.APIKeyCreateRequest) (string, *time.Time, error) {
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
func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID, name string, expiresAt *time.Time) (*db.APIKeyResponse, string, error) {
	// Generate the API key
	plainTextKey, hashedKey, err := s.CreateNewAPIKey(name, expiresAt)
	if err != nil {
		return nil, "", apperrors.Wrap(err, "failed to generate API key")
	}

	// Create the API key record
	apiKey := &db.APIKey{
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
func (s *APIKeyService) GetAPIKeysWithPagination(ctx context.Context, userID string, page, limit, offset int) (*models.APIKeysListResponse, error) {
	apiKeys, total, err := s.repo.FindByUserIDWithPagination(ctx, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.APIKeyResponse, len(apiKeys))
	for i, apiKey := range apiKeys {
		responses[i] = &models.APIKeyResponse{
			ID:        apiKey.ToResponse().ID,
			UserID:    apiKey.ToResponse().UserID,
			Name:      apiKey.ToResponse().Name,
			CreatedAt: apiKey.ToResponse().CreatedAt,
			ExpiresAt: apiKey.ToResponse().ExpiresAt,
			RevokedAt: apiKey.ToResponse().RevokedAt,
		}
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	hasNext := page < totalPages
	hasPrev := page > 1

	return &models.APIKeysListResponse{
		APIKeys: responses,
		Pagination: &models.PaginationMetadata{
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
