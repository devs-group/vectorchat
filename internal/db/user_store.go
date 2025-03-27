package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

// UserStore implements auth.UserStore
type UserStore struct {
	pool *pgxpool.Pool
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

// FindAPIKey finds an API key by its unhashed value
func (s *UserStore) FindAPIKey(ctx context.Context, key string) (*APIKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, key, created_at, expires_at, revoked_at
		FROM api_keys
		WHERE revoked_at IS NULL AND expires_at > NOW()
	`)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to query API keys")
	}
	defer rows.Close()

	for rows.Next() {
		var apiKey APIKey
		err := rows.Scan(&apiKey.ID, &apiKey.UserID, &apiKey.Key, &apiKey.CreatedAt, &apiKey.ExpiresAt, &apiKey.RevokedAt)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to scan API key")
		}

		// Check if the provided key matches the stored hash
		if err := bcrypt.CompareHashAndPassword([]byte(apiKey.Key), []byte(key)); err == nil {
			return &apiKey, nil
		}
	}

	return nil, apperrors.ErrUserNotFound
}

// CreateUser creates a new user with transaction support
func (s *UserStore) CreateUser(ctx context.Context, user *User) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, name, email, provider, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.Name, user.Email, user.Provider, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return apperrors.Wrap(err, "failed to create user")
	}

	if err = tx.Commit(ctx); err != nil {
		return apperrors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// CreateAPIKey creates a new API key
func (s *UserStore) CreateAPIKey(ctx context.Context, apiKey *APIKey) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO api_keys (id, user_id, key, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, apiKey.ID, apiKey.UserID, apiKey.Key, apiKey.CreatedAt, apiKey.ExpiresAt)

	if err != nil {
		return apperrors.Wrap(err, "failed to create API key")
	}

	return nil
}

// GetAPIKeys gets all API keys for a user
func (s *UserStore) GetAPIKeys(ctx context.Context, userID string) ([]APIKey, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, key, created_at, expires_at, revoked_at
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
		err := rows.Scan(&apiKey.ID, &apiKey.UserID, &apiKey.Key, &apiKey.CreatedAt, &apiKey.ExpiresAt, &apiKey.RevokedAt)
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
		WHERE id = $2 AND user_id = $3 AND revoked_at IS NULL
	`, now, id, userID)

	if err != nil {
		return apperrors.Wrap(err, "failed to revoke API key")
	}
	return nil
}
