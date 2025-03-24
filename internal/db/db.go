package db

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)


func New(connStr string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, apperrors.Wrap(err, "unable to parse connection string")
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, apperrors.Wrap(err, "unable to connect to database")
	}
	return pool, nil
}