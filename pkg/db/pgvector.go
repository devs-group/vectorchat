package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	apperrors "github.com/yourusername/vectorchat/pkg/errors"
)

// PgVectorDB handles interactions with PostgreSQL with pgvector extension
type PgVectorDB struct {
	pool *pgxpool.Pool
}

// Document represents a document with its vector embedding
type Document struct {
	ID        string
	Content   string
	Embedding []float32
}

// VectorDB defines the interface for vector database operations
type VectorDB interface {
	Close()
	StoreDocument(ctx context.Context, doc Document) error
	FindSimilarDocuments(ctx context.Context, embedding []float32, limit int) ([]Document, error)
	FindSimilarDocumentsByChatID(ctx context.Context, embedding []float32, chatID string, limit int) ([]Document, error)
	DeleteDocument(ctx context.Context, id string) error
	GetDocumentsByPrefix(ctx context.Context, prefix string) ([]Document, error)
}

// NewPgVectorDB creates a new connection to PostgreSQL with pgvector
func NewPgVectorDB(connStr string) (*PgVectorDB, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, apperrors.Wrap(err, "unable to parse connection string")
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, apperrors.Wrap(err, "unable to connect to database")
	}

	db := &PgVectorDB{
		pool: pool,
	}

	// Initialize the database schema
	if err := db.initSchema(); err != nil {
		pool.Close()
		return nil, apperrors.Wrap(err, "failed to initialize schema")
	}

	return db, nil
}

// Close closes the database connection
func (db *PgVectorDB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// initSchema creates the necessary tables and extensions if they don't exist
func (db *PgVectorDB) initSchema() error {
	ctx := context.Background()
	
	// Create pgvector extension if it doesn't exist
	_, err := db.pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("failed to create vector extension: %v", err)
	}

	// Create documents table
	_, err = db.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			embedding vector(1536) NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create documents table: %v", err)
	}

	// Create index for vector similarity search
	_, err = db.pool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS documents_embedding_idx 
		ON documents 
		USING ivfflat (embedding vector_cosine_ops) 
		WITH (lists = 100)
	`)
	if err != nil {
		return fmt.Errorf("failed to create vector index: %v", err)
	}

	return nil
}

// StoreDocument stores a document with its vector embedding
func (db *PgVectorDB) StoreDocument(ctx context.Context, doc Document) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO documents (id, content, embedding)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE
		SET content = $2, embedding = $3
	`, doc.ID, doc.Content, pgvector.NewVector(doc.Embedding))

	if err != nil {
		return fmt.Errorf("failed to store document: %v", err)
	}

	return nil
}

// FindSimilarDocuments finds documents similar to the given embedding
func (db *PgVectorDB) FindSimilarDocuments(ctx context.Context, embedding []float32, limit int) ([]Document, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, content, embedding
		FROM documents
		ORDER BY embedding <=> $1
		LIMIT $2
	`, pgvector.NewVector(embedding), limit)
	
	if err != nil {
		return nil, fmt.Errorf("failed to query similar documents: %v", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		var pgvec pgvector.Vector
		
		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec); err != nil {
			return nil, fmt.Errorf("failed to scan document row: %v", err)
		}
		
		doc.Embedding = pgvec.Slice()
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating document rows: %v", err)
	}

	return documents, nil
}

// FindSimilarDocumentsByChatID finds documents similar to the given embedding that belong to a specific chat ID
func (db *PgVectorDB) FindSimilarDocumentsByChatID(ctx context.Context, embedding []float32, chatID string, limit int) ([]Document, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, content, embedding
		FROM documents
		WHERE id LIKE $1 || '%'
		ORDER BY embedding <=> $2
		LIMIT $3
	`, chatID+"-", pgvector.NewVector(embedding), limit)
	
	if err != nil {
		return nil, fmt.Errorf("failed to query similar documents by chat ID: %v", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		var pgvec pgvector.Vector
		
		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec); err != nil {
			return nil, fmt.Errorf("failed to scan document row: %v", err)
		}
		
		doc.Embedding = pgvec.Slice()
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating document rows: %v", err)
	}

	return documents, nil
}

// DeleteDocument removes a document from the database
func (db *PgVectorDB) DeleteDocument(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx, `
		DELETE FROM documents
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete document: %v", err)
	}

	return nil
}

// GetDocumentsByPrefix retrieves documents with IDs starting with the given prefix
func (db *PgVectorDB) GetDocumentsByPrefix(ctx context.Context, prefix string) ([]Document, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, content, embedding
		FROM documents
		WHERE id LIKE $1 || '%'
	`, prefix)
	
	if err != nil {
		return nil, fmt.Errorf("failed to query documents by prefix: %v", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		var pgvec pgvector.Vector
		
		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec); err != nil {
			return nil, fmt.Errorf("failed to scan document row: %v", err)
		}
		
		doc.Embedding = pgvec.Slice()
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating document rows: %v", err)
	}

	return documents, nil
} 