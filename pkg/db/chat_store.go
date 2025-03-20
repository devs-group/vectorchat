package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	apperrors "github.com/yourusername/vectorchat/pkg/errors"
)

// ChatStore handles interactions with PostgreSQL with pgvector extension
type ChatStore struct {
	pool *pgxpool.Pool
}

// Document represents a document with its vector embedding
type Document struct {
	ID        string
	Content   string
	Embedding []float32
}

// NewChatStore creates a new connection to PostgreSQL with pgvector
func NewChatStore(connStr string) (*ChatStore, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, apperrors.Wrap(err, "unable to parse connection string")
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, apperrors.Wrap(err, "unable to connect to database")
	}

	db := &ChatStore{
		pool: pool,
	}

	return db, nil
}

// Close closes the database connection
func (db *ChatStore) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// StoreDocument stores a document with its vector embedding
func (db *ChatStore) StoreDocument(ctx context.Context, doc Document) error {
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
func (db *ChatStore) FindSimilarDocuments(ctx context.Context, embedding []float32, limit int) ([]Document, error) {
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
func (db *ChatStore) FindSimilarDocumentsByChatID(ctx context.Context, embedding []float32, chatID string, limit int) ([]Document, error) {
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
func (db *ChatStore) DeleteDocument(ctx context.Context, id string) error {
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
func (db *ChatStore) GetDocumentsByPrefix(ctx context.Context, prefix string) ([]Document, error) {
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