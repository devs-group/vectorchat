package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// AuthService handles user authentication and user-related operations backed by Ory Kratos.
type AuthService struct {
	*CommonService
	userRepo   *db.UserRepository
	apiKeyRepo *db.APIKeyRepository
}

// NewAuthService creates a new auth service instance.
func NewAuthService(userRepo *db.UserRepository, apiKeyRepo *db.APIKeyRepository) *AuthService {
	return &AuthService{
		CommonService: NewCommonService(),
		userRepo:      userRepo,
		apiKeyRepo:    apiKeyRepo,
	}
}

// FindUserByID finds a user by ID.
func (s *AuthService) FindUserByID(ctx context.Context, id string) (*db.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// FindUserByEmail finds a user by email address.
func (s *AuthService) FindUserByEmail(ctx context.Context, email string) (*db.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

// CreateUser creates a new user record.
func (s *AuthService) CreateUser(ctx context.Context, id, name, email, provider string) (*db.User, error) {
	if id == "" {
		id = uuid.New().String()
	}

	user := &db.User{
		ID:        id,
		Name:      name,
		Email:     email,
		Provider:  provider,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// FindAPIKeyByPlaintext finds an API key by comparing against stored hashes.
func (s *AuthService) FindAPIKeyByPlaintext(ctx context.Context, plainTextKey string, compareFunc func(hashedKey string) (bool, error)) (*db.APIKey, error) {
	return s.apiKeyRepo.FindByHashComparison(ctx, compareFunc)
}

// SyncIdentity merges an Ory Kratos identity into the local user store.
func (s *AuthService) SyncIdentity(ctx context.Context, identityID string, traitsJSON string) (*db.User, error) {
	traits := make(map[string]any)
	if err := json.Unmarshal([]byte(traitsJSON), &traits); err != nil {
		return nil, apperrors.Wrap(err, "failed to decode identity traits")
	}

	email, _ := traits["email"].(string)
	if email == "" {
		return nil, apperrors.ErrInvalidUserData
	}

	name, _ := traits["name"].(string)
	if name == "" {
		name = email
	}

	provider := "kratos"
	if p, ok := traits["provider"].(string); ok && p != "" {
		provider = p
	}

	user, err := s.userRepo.FindByID(ctx, identityID)
	if err == nil {
		updated := false
		if user.Email != email {
			user.Email = email
			updated = true
		}
		if user.Name != name {
			user.Name = name
			updated = true
		}
		if user.Provider != provider {
			user.Provider = provider
			updated = true
		}
		if updated {
			if updateErr := s.userRepo.Update(ctx, user); updateErr != nil {
				return nil, apperrors.Wrap(updateErr, "failed to update user")
			}
		}
		return user, nil
	}

	if !apperrors.Is(err, apperrors.ErrUserNotFound) {
		return nil, apperrors.Wrap(err, "failed to look up user")
	}

	user = &db.User{
		ID:        identityID,
		Email:     email,
		Name:      name,
		Provider:  provider,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if createErr := s.userRepo.Create(ctx, user); createErr != nil {
		if apperrors.Is(createErr, apperrors.ErrUserAlreadyExists) {
			existing, findErr := s.userRepo.FindByEmail(ctx, email)
			if findErr != nil {
				return nil, apperrors.Wrap(findErr, "failed to load existing user by email")
			}

			needsUpdate := false
			if existing.Name != name {
				existing.Name = name
				needsUpdate = true
			}
			if existing.Provider != provider {
				existing.Provider = provider
				needsUpdate = true
			}
			if needsUpdate {
				if updateErr := s.userRepo.Update(ctx, existing); updateErr != nil {
					return nil, apperrors.Wrap(updateErr, "failed to update existing user")
				}
			}
			return existing, nil
		}
		return nil, apperrors.Wrap(createErr, "failed to create user from identity")
	}

	return user, nil
}
