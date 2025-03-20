package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	apperrors "github.com/yourusername/vectorchat/pkg/errors"
)

// UserStore implements auth.UserStore
type UserStore struct {
	pool *pgxpool.Pool
}

// User represents an authenticated user
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIKey represents an API key for a user
type APIKey struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// NewUserStore creates a new user store
func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{
		pool: pool,
	}
}

// FindUserByID finds a user by ID
func (s *UserStore) FindUserByID(ctx context.Context, id string) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, email, provider, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Name, &user.Email, &user.Provider, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find user by ID")
	}
	
	return &user, nil
}

// FindUserByEmail finds a user by email
func (s *UserStore) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, email, provider, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.Provider, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find user by email")
	}
	
	return &user, nil
}

// FindUserByAPIKey finds a user by API key
func (s *UserStore) FindUserByAPIKey(ctx context.Context, key string) (*User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u.provider, u.created_at, u.updated_at
		FROM users u
		JOIN api_keys a ON u.id = a.user_id
		WHERE a.key = $1 AND a.revoked_at IS NULL
	`, key).Scan(&user.ID, &user.Name, &user.Email, &user.Provider, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find user by API key")
	}
	
	return &user, nil
}

// CreateUser creates a new user
func (s *UserStore) CreateUser(ctx context.Context, user *User) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (id, name, email, provider, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.Name, user.Email, user.Provider, user.CreatedAt, user.UpdatedAt)
	
	if err != nil {
		return apperrors.Wrap(err, "failed to create user")
	}
	
	return nil
}

// CreateAPIKey creates a new API key
func (s *UserStore) CreateAPIKey(ctx context.Context, apiKey *APIKey) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO api_keys (id, user_id, key, created_at)
		VALUES ($1, $2, $3, $4)
	`, apiKey.ID, apiKey.UserID, apiKey.Key, apiKey.CreatedAt)
	
	if err != nil {
		return apperrors.Wrap(err, "failed to create API key")
	}
	
	return nil
}

// GetAPIKeys gets all API keys for a user
func (s *UserStore) GetAPIKeys(ctx context.Context, userID string) ([]APIKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, key, created_at, revoked_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get API keys")
	}
	defer rows.Close()
	
	var apiKeys []APIKey
	for rows.Next() {
		var apiKey APIKey
		err := rows.Scan(&apiKey.ID, &apiKey.UserID, &apiKey.Key, &apiKey.CreatedAt, &apiKey.RevokedAt)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to scan API key")
		}
		apiKeys = append(apiKeys, apiKey)
	}
	
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, "error iterating API key rows")
	}
	
	return apiKeys, nil
}

// RevokeAPIKey revokes an API key
func (s *UserStore) RevokeAPIKey(ctx context.Context, id string, userID string) error {
	now := time.Now()
	_, err := s.pool.Exec(ctx, `
		UPDATE api_keys
		SET revoked_at = $1
		WHERE id = $2 AND user_id = $3
	`, now, id, userID)
	
	if err != nil {
		return apperrors.Wrap(err, "failed to revoke API key")
	}
	
	return nil
} 