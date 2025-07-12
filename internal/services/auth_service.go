package services

import (
	"context"
	"time"

	"github.com/google/uuid"
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

// FindAPIKeyByPlaintext finds an API key by comparing against stored hashes
func (s *AuthService) FindAPIKeyByPlaintext(ctx context.Context, plainTextKey string, compareFunc func(hashedKey string) (bool, error)) (*APIKey, error) {
	return s.apiKeyRepo.FindByHashComparison(ctx, compareFunc)
}
