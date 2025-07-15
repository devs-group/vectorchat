package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// AuthService handles user authentication and user-related operations
type AuthService struct {
	*CommonService
	userRepo   db.UserRepository
	apiKeyRepo db.APIKeyRepository
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo db.UserRepository, apiKeyRepo db.APIKeyRepository) *AuthService {
	return &AuthService{
		CommonService: NewCommonService(),
		userRepo:      userRepo,
		apiKeyRepo:    apiKeyRepo,
	}
}

// GitHubUser represents a GitHub user response
type GitHubUser struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GitHubEmail represents a GitHub email response
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// OAuthState represents OAuth state information
type OAuthState struct {
	State    string
	StateKey string
}

// FindUserByID finds a user by ID
func (s *AuthService) FindUserByID(ctx context.Context, id string) (*db.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// FindUserByEmail finds a user by email
func (s *AuthService) FindUserByEmail(ctx context.Context, email string) (*db.User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

// CreateUser creates a new user
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

	err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// FindAPIKeyByPlaintext finds an API key by comparing against stored hashes
func (s *AuthService) FindAPIKeyByPlaintext(ctx context.Context, plainTextKey string, compareFunc func(hashedKey string) (bool, error)) (*db.APIKey, error) {
	return s.apiKeyRepo.FindByHashComparison(ctx, compareFunc)
}

// GenerateOAuthState generates a random OAuth state for security
func (s *AuthService) GenerateOAuthState() (*OAuthState, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, apperrors.Wrap(err, "failed to generate state")
	}
	state := base64.URLEncoding.EncodeToString(b)
	stateKey := fmt.Sprintf("oauth_state_%s", uuid.New().String())

	return &OAuthState{
		State:    state,
		StateKey: stateKey,
	}, nil
}

// ProcessGitHubCallback processes the GitHub OAuth callback
func (s *AuthService) ProcessGitHubCallback(ctx context.Context, client *http.Client) (*db.User, error) {
	// Get user info from GitHub
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get user info")
	}
	defer resp.Body.Close()

	var githubUser GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, apperrors.Wrap(err, "failed to parse user info")
	}

	// Get email if not present
	if githubUser.Email == "" {
		emails, err := s.getGitHubEmails(client)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get email")
		}
		for _, email := range emails {
			if email.Primary && email.Verified {
				githubUser.Email = email.Email
				break
			}
		}
	}

	// Validate user data
	if len(githubUser.Email) > 255 || len(githubUser.Name) > 100 {
		return nil, apperrors.Wrap(apperrors.ErrInvalidUserData, "invalid user data")
	}

	// Find or create user
	user, err := s.FindUserByEmail(ctx, githubUser.Email)
	if err != nil && !apperrors.Is(err, apperrors.ErrUserNotFound) {
		return nil, apperrors.Wrap(err, "failed to find user")
	}

	if user == nil {
		user, err = s.CreateUser(ctx, uuid.New().String(), githubUser.Name, githubUser.Email, "github")
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to create user")
		}
	}

	return user, nil
}

// getGitHubEmails fetches emails from GitHub API
func (s *AuthService) getGitHubEmails(client *http.Client) ([]GitHubEmail, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, err
	}
	return emails, nil
}
