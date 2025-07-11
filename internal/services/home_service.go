package services

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// HomeService handles home-related operations
type HomeService struct {
	pool *pgxpool.Pool
}

// NewHomeService creates a new home service
func NewHomeService(pool *pgxpool.Pool) *HomeService {
	return &HomeService{
		pool: pool,
	}
}

// FindUserByID finds a user by ID (for session validation)
func (s *HomeService) FindUserByID(ctx context.Context, id string) (*User, error) {
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
