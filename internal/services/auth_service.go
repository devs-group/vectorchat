package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// AuthService handles user authentication and user-related operations
type AuthService struct {
	userRepo   UserRepository
	apiKeyRepo APIKeyRepository
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo UserRepository, apiKeyRepo APIKeyRepository) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		apiKeyRepo: apiKeyRepo,
	}
}

// FindUserByID finds a user by ID
func (s *AuthService) FindUserByID(ctx context.Context, id string) (*User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// FindUserByEmail finds a user by email
func (s *AuthService) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

// CreateUser creates a new user
func (s *AuthService) CreateUser(ctx context.Context, id, name, email, provider string) (*User, error) {
	if id == "" {
		id = uuid.New().String()
	}

	user := &User{
		ID:        id,
		Name:      name,
		Email:     email,
		Provider:  provider,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *AuthService) UpdateUser(ctx context.Context, user *User) error {
	return s.userRepo.Update(ctx, user)
}

// DeleteUser deletes a user
func (s *AuthService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepo.Delete(ctx, id)
}

// FindAPIKeyByPlaintext finds an API key by comparing against stored hashes
func (s *AuthService) FindAPIKeyByPlaintext(ctx context.Context, plainTextKey string, compareFunc func(hashedKey string) (bool, error)) (*APIKey, error) {
	return s.apiKeyRepo.FindByHashComparison(ctx, compareFunc)
}

// ValidateAPIKey validates an API key and returns the associated user
func (s *AuthService) ValidateAPIKey(ctx context.Context, plainTextKey string, compareFunc func(hashedKey string) (bool, error)) (*User, error) {
	// Find the API key
	apiKey, err := s.apiKeyRepo.FindByHashComparison(ctx, compareFunc)
	if err != nil {
		return nil, apperrors.Wrap(err, "API key validation failed")
	}

	// Check if the key is revoked
	if apiKey.RevokedAt != nil {
		return nil, apperrors.ErrInvalidAPIKey
	}

	// Check if the key is expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, apperrors.ErrInvalidAPIKey
	}

	// Get the user associated with the API key
	user, err := s.userRepo.FindByID(ctx, apiKey.UserID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find user for API key")
	}

	return user, nil
}

// CreateAPIKey creates a new API key for a user
func (s *AuthService) CreateAPIKey(ctx context.Context, userID, hashedKey string, name *string, expiresAt *time.Time) (*APIKey, error) {
	apiKey := &APIKey{
		ID:        uuid.New().String(),
		UserID:    userID,
		Key:       hashedKey,
		Name:      name,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	err := s.apiKeyRepo.Create(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

// GetAPIKeysForUser gets all API keys for a user
func (s *AuthService) GetAPIKeysForUser(ctx context.Context, userID string) ([]*APIKey, error) {
	return s.apiKeyRepo.FindByUserID(ctx, userID)
}

// GetAPIKeysWithPagination gets API keys for a user with pagination support
func (s *AuthService) GetAPIKeysWithPagination(ctx context.Context, userID string, offset, limit int) ([]*APIKey, int64, error) {
	return s.apiKeyRepo.FindByUserIDWithPagination(ctx, userID, offset, limit)
}

// RevokeAPIKey revokes an API key
func (s *AuthService) RevokeAPIKey(ctx context.Context, id string, userID string) error {
	return s.apiKeyRepo.Revoke(ctx, id, userID)
}

// DeleteAPIKey deletes an API key
func (s *AuthService) DeleteAPIKey(ctx context.Context, id string) error {
	return s.apiKeyRepo.Delete(ctx, id)
}

// IsAPIKeyValid checks if an API key is valid (not revoked and not expired)
func (s *AuthService) IsAPIKeyValid(ctx context.Context, id string) (bool, error) {
	apiKey, err := s.apiKeyRepo.FindByID(ctx, id)
	if err != nil {
		return false, err
	}

	// Check if revoked
	if apiKey.RevokedAt != nil {
		return false, nil
	}

	// Check if expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}

// GetUserStats returns statistics about a user
func (s *AuthService) GetUserStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	apiKeys, err := s.apiKeyRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Count active API keys
	activeAPIKeys := 0
	for _, apiKey := range apiKeys {
		if apiKey.RevokedAt == nil && (apiKey.ExpiresAt == nil || apiKey.ExpiresAt.After(time.Now())) {
			activeAPIKeys++
		}
	}

	stats := map[string]interface{}{
		"user_id":         user.ID,
		"user_name":       user.Name,
		"user_email":      user.Email,
		"provider":        user.Provider,
		"created_at":      user.CreatedAt,
		"updated_at":      user.UpdatedAt,
		"total_api_keys":  len(apiKeys),
		"active_api_keys": activeAPIKeys,
	}

	return stats, nil
}
