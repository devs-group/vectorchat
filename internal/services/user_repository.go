package services

import (
	"context"
	"time"

	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// userRepository implements UserRepository interface
type userRepository struct {
	db *Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *Database) UserRepositoryTx {
	return &userRepository{db: db}
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(ctx context.Context, id string) (*User, error) {
	var user User
	query := `
		SELECT id, name, email, provider, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find user by ID")
	}

	return &user, nil
}

// FindByEmail finds a user by email
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	query := `
		SELECT id, name, email, provider, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find user by email")
	}

	return &user, nil
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *User) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO users (id, name, email, provider, created_at, updated_at)
		VALUES (:id, :name, :email, :provider, :created_at, :updated_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrUserAlreadyExists
		}
		return apperrors.Wrap(err, "failed to create user")
	}

	return nil
}

// CreateTx creates a new user within a transaction
func (r *userRepository) CreateTx(ctx context.Context, tx *Transaction, user *User) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO users (id, name, email, provider, created_at, updated_at)
		VALUES (:id, :name, :email, :provider, :created_at, :updated_at)
	`

	_, err := tx.NamedExecContext(ctx, query, user)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrUserAlreadyExists
		}
		return apperrors.Wrap(err, "failed to create user")
	}

	return nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET name = :name, email = :email, provider = :provider, updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrUserAlreadyExists
		}
		return apperrors.Wrap(err, "failed to update user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}

// UpdateTx updates an existing user within a transaction
func (r *userRepository) UpdateTx(ctx context.Context, tx *Transaction, user *User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET name = :name, email = :email, provider = :provider, updated_at = :updated_at
		WHERE id = :id
	`

	result, err := tx.NamedExecContext(ctx, query, user)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrUserAlreadyExists
		}
		return apperrors.Wrap(err, "failed to update user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}

// Delete deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}

// DeleteTx deletes a user by ID within a transaction
func (r *userRepository) DeleteTx(ctx context.Context, tx *Transaction, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete user")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}
