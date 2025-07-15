package db

import (
	"context"
	"time"

	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

type APIKeyRepository struct {
	db *Database
}

func NewAPIKeyRepository(db *Database) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(ctx context.Context, apiKey *APIKey) error {
	if apiKey.CreatedAt.IsZero() {
		apiKey.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO api_keys (id, user_id, key, name, created_at, expires_at)
		VALUES (:id, :user_id, :key, :name, :created_at, :expires_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, apiKey)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrAPIKeyAlreadyExists
		}
		return apperrors.Wrap(err, "failed to create API key")
	}

	return nil
}

// CreateTx creates a new API key within a transaction
func (r *APIKeyRepository) CreateTx(ctx context.Context, tx *Transaction, apiKey *APIKey) error {
	if apiKey.CreatedAt.IsZero() {
		apiKey.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO api_keys (id, user_id, key, name, created_at, expires_at)
		VALUES (:id, :user_id, :key, :name, :created_at, :expires_at)
	`

	_, err := tx.NamedExecContext(ctx, query, apiKey)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrAPIKeyAlreadyExists
		}
		return apperrors.Wrap(err, "failed to create API key")
	}

	return nil
}

// FindByID finds an API key by ID
func (r *APIKeyRepository) FindByID(ctx context.Context, id string) (*APIKey, error) {
	var apiKey APIKey
	query := `
		SELECT id, user_id, key, name, created_at, expires_at, revoked_at
		FROM api_keys
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &apiKey, query, id)
	if err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrAPIKeyNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find API key by ID")
	}

	return &apiKey, nil
}

// FindByUserID finds all API keys for a user
func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID string) ([]*APIKey, error) {
	var apiKeys []*APIKey
	query := `
		SELECT id, user_id, key, name, created_at, expires_at, revoked_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &apiKeys, query, userID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find API keys by user ID")
	}

	return apiKeys, nil
}

// FindByUserIDWithPagination finds API keys for a user with pagination
func (r *APIKeyRepository) FindByUserIDWithPagination(ctx context.Context, userID string, offset, limit int) ([]*APIKey, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM api_keys WHERE user_id = $1`
	err := r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to get total API keys count")
	}

	// Get paginated results
	var apiKeys []*APIKey
	query := `
		SELECT id, user_id, key, name, created_at, expires_at, revoked_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err = r.db.SelectContext(ctx, &apiKeys, query, userID, limit, offset)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to find API keys with pagination")
	}

	return apiKeys, total, nil
}

// FindByHashComparison finds an API key by comparing against stored hashes
func (r *APIKeyRepository) FindByHashComparison(ctx context.Context, compareFunc func(hashedKey string) (bool, error)) (*APIKey, error) {
	var apiKeys []*APIKey
	query := `
		SELECT id, user_id, key, name, created_at, expires_at, revoked_at
		FROM api_keys
		WHERE revoked_at IS NULL
	`

	err := r.db.SelectContext(ctx, &apiKeys, query)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to query API keys for comparison")
	}

	for _, apiKey := range apiKeys {
		// Check if key is expired
		if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
			continue
		}

		isValid, err := compareFunc(apiKey.Key)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to compare API key hash")
		}
		if isValid {
			return apiKey, nil
		}
	}

	return nil, apperrors.ErrAPIKeyNotFound
}

// Revoke revokes an API key
func (r *APIKeyRepository) Revoke(ctx context.Context, id, userID string) error {
	now := time.Now()
	query := `
		UPDATE api_keys
		SET revoked_at = $1
		WHERE id = $2 AND user_id = $3 AND revoked_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, now, id, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to revoke API key")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrAPIKeyNotFound
	}

	return nil
}

// RevokeTx revokes an API key within a transaction
func (r *APIKeyRepository) RevokeTx(ctx context.Context, tx *Transaction, id, userID string) error {
	now := time.Now()
	query := `
		UPDATE api_keys
		SET revoked_at = $1
		WHERE id = $2 AND user_id = $3 AND revoked_at IS NULL
	`

	result, err := tx.ExecContext(ctx, query, now, id, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to revoke API key")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrAPIKeyNotFound
	}

	return nil
}

// Delete deletes an API key by ID
func (r *APIKeyRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM api_keys WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete API key")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrAPIKeyNotFound
	}

	return nil
}

// DeleteTx deletes an API key by ID within a transaction
func (r *APIKeyRepository) DeleteTx(ctx context.Context, tx *Transaction, id string) error {
	query := `DELETE FROM api_keys WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete API key")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrAPIKeyNotFound
	}

	return nil
}

// IsRevoked checks if an API key is revoked
func (r *APIKeyRepository) IsRevoked(ctx context.Context, id string) (bool, error) {
	var revokedAt *time.Time
	query := `SELECT revoked_at FROM api_keys WHERE id = $1`

	err := r.db.GetContext(ctx, &revokedAt, query, id)
	if err != nil {
		if IsNoRowsError(err) {
			return false, apperrors.ErrAPIKeyNotFound
		}
		return false, apperrors.Wrap(err, "failed to check if API key is revoked")
	}

	return revokedAt != nil, nil
}

// IsExpired checks if an API key is expired
func (r *APIKeyRepository) IsExpired(ctx context.Context, id string) (bool, error) {
	var expiresAt *time.Time
	query := `SELECT expires_at FROM api_keys WHERE id = $1`

	err := r.db.GetContext(ctx, &expiresAt, query, id)
	if err != nil {
		if IsNoRowsError(err) {
			return false, apperrors.ErrAPIKeyNotFound
		}
		return false, apperrors.Wrap(err, "failed to check if API key is expired")
	}

	if expiresAt == nil {
		return false, nil // No expiration date means it never expires
	}

	return expiresAt.Before(time.Now()), nil
}
