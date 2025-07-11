package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

type APIKeyService struct {
	pool *pgxpool.Pool
}

func NewAPIKeyService(pool *pgxpool.Pool) *APIKeyService {
	return &APIKeyService{
		pool: pool,
	}
}

// CreateNewAPIKey generates a new plain text API key and its bcrypt hash.
// It returns the plain text key (show ONCE to the user) and the hash (store in DB).
func (s *APIKeyService) CreateNewAPIKey() (plainTextKey string, hashedKey string, err error) {
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
func (s *APIKeyService) CreateAPIKey(ctx context.Context, apiKey *APIKey) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO api_keys (id, user_id, key, name, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, apiKey.ID, apiKey.UserID, apiKey.Key, apiKey.Name, apiKey.CreatedAt, apiKey.ExpiresAt)

	if err != nil {
		return apperrors.Wrap(err, "failed to create API key")
	}

	return nil
}

// GetAPIKeysWithPagination gets API keys for a user with pagination support
func (s *APIKeyService) GetAPIKeysWithPagination(ctx context.Context, userID string, offset, limit int) ([]APIKey, int64, error) {
	// Get total count
	var total int64
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM api_keys WHERE user_id = $1
	`, userID).Scan(&total)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to get total API keys count")
	}

	// Get paginated results
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, key, name, created_at, expires_at, revoked_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)

	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to get API keys")
	}
	defer rows.Close()

	var apiKeys []APIKey
	for rows.Next() {
		var apiKey APIKey
		err := rows.Scan(&apiKey.ID, &apiKey.UserID, &apiKey.Key, &apiKey.Name, &apiKey.CreatedAt, &apiKey.ExpiresAt, &apiKey.RevokedAt)
		if err != nil {
			return nil, 0, apperrors.Wrap(err, "failed to scan API key")
		}
		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.Wrap(err, "error iterating API key rows")
	}

	return apiKeys, total, nil
}

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, id string, userID string) error {
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
