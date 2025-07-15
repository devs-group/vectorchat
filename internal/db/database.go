package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // PostgreSQL driver
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// Database wraps sqlx.DB with additional functionality
type Database struct {
	*sqlx.DB
}

// NewDatabase creates a new database connection using sqlx
func NewDatabase(connStr string) (*Database, error) {
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to connect to database")
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, apperrors.Wrap(err, "failed to ping database")
	}

	return &Database{DB: db}, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.DB.Close()
}

// Transaction wraps database transaction operations
type Transaction struct {
	*sqlx.Tx
}

// BeginTx starts a new transaction
func (db *Database) BeginTx(ctx context.Context) (*Transaction, error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to begin transaction")
	}
	return &Transaction{Tx: tx}, nil
}

// Commit commits the transaction
func (tx *Transaction) Commit() error {
	if err := tx.Tx.Commit(); err != nil {
		return apperrors.Wrap(err, "failed to commit transaction")
	}
	return nil
}

// Rollback rolls back the transaction
func (tx *Transaction) Rollback() error {
	if err := tx.Tx.Rollback(); err != nil && err != sql.ErrTxDone {
		return apperrors.Wrap(err, "failed to rollback transaction")
	}
	return nil
}

// IsNoRowsError checks if the error is a "no rows" error
func IsNoRowsError(err error) bool {
	return err == sql.ErrNoRows
}

// IsDuplicateKeyError checks if the error is a duplicate key error
func IsDuplicateKeyError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}

// IsForeignKeyViolationError checks if the error is a foreign key violation
func IsForeignKeyViolationError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23503" // foreign_key_violation
	}
	return false
}

// GetErrorCode returns the PostgreSQL error code if it's a pq.Error
func GetErrorCode(err error) string {
	if pqErr, ok := err.(*pq.Error); ok {
		return string(pqErr.Code)
	}
	return ""
}
