package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pgvector/pgvector-go"
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

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Provider  string    `json:"provider" db:"provider"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// APIKey represents an API key in the system
type APIKey struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Key       string     `json:"key" db:"key"` // Stored as hashed value
	Name      *string    `json:"name" db:"name"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
}

// APIKeyResponse represents the response when returning API keys to the client
type APIKeyResponse struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Name      *string    `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	// Key is intentionally omitted for security
}

// ToResponse converts APIKey to APIKeyResponse (without the actual key)
func (a *APIKey) ToResponse() *APIKeyResponse {
	return &APIKeyResponse{
		ID:        a.ID,
		UserID:    a.UserID,
		Name:      a.Name,
		CreatedAt: a.CreatedAt,
		ExpiresAt: a.ExpiresAt,
		RevokedAt: a.RevokedAt,
	}
}

// Chatbot represents a configurable AI assistant
type Chatbot struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	UserID             string    `json:"user_id" db:"user_id"`
	Name               string    `json:"name" db:"name"`
	Description        string    `json:"description" db:"description"`
	SystemInstructions string    `json:"system_instructions" db:"system_instructions"`
	ModelName          string    `json:"model_name" db:"model_name"`
	TemperatureParam   float64   `json:"temperature_param" db:"temperature_param"`
	MaxTokens          int       `json:"max_tokens" db:"max_tokens"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// Document represents a document with its vector embedding
type Document struct {
	ID         string          `json:"id" db:"id"`
	Content    []byte          `json:"content" db:"content"`
	Embedding  pgvector.Vector `json:"embedding" db:"embedding"`
	ChatbotID  uuid.UUID       `json:"chatbot_id" db:"chatbot_id"`
	FileID     *uuid.UUID      `json:"file_id,omitempty" db:"file_id"`
	ChunkIndex *int            `json:"chunk_index,omitempty" db:"chunk_index"`
}

// DocumentWithEmbedding represents a document with embedding as float32 slice for easier handling
type DocumentWithEmbedding struct {
	ID         string     `json:"id"`
	Content    []byte     `json:"content"`
	Embedding  []float32  `json:"embedding"`
	ChatbotID  uuid.UUID  `json:"chatbot_id"`
	FileID     *uuid.UUID `json:"file_id,omitempty"`
	ChunkIndex *int       `json:"chunk_index,omitempty"`
}

// ToDocumentWithEmbedding converts Document to DocumentWithEmbedding
func (d *Document) ToDocumentWithEmbedding() *DocumentWithEmbedding {
	return &DocumentWithEmbedding{
		ID:         d.ID,
		Content:    d.Content,
		Embedding:  d.Embedding.Slice(),
		ChatbotID:  d.ChatbotID,
		FileID:     d.FileID,
		ChunkIndex: d.ChunkIndex,
	}
}

// File represents a file uploaded to a chatbot
type File struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ChatbotID  uuid.UUID `json:"chatbot_id" db:"chatbot_id"`
	Filename   string    `json:"filename" db:"filename"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data   interface{} `json:"data"`
	Total  int64       `json:"total"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, total int64, offset, limit int) *PaginatedResponse {
	return &PaginatedResponse{
		Data:   data,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}
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
