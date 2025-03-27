package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// DocumentStore handles interactions with PostgreSQL with pgvector extension
type DocumentStore struct {
	pool *pgxpool.Pool
}

// NewDocumentStore creates a new connection to PostgreSQL with pgvector
func NewDocumentStore(pool *pgxpool.Pool) (*DocumentStore) {
	return &DocumentStore{
		pool: pool,
	}
}

// StoreDocument stores a document with its vector embedding
func (db *DocumentStore) StoreDocument(ctx context.Context, doc Document) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO documents (id, content, embedding, chatbot_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET content = $2, embedding = $3, chatbot_id = $4
	`, doc.ID, doc.Content, pgvector.NewVector(doc.Embedding), doc.ChatbotID)

	if err != nil {
		return apperrors.Wrapf(apperrors.ErrDatabaseOperation, "failed to store document: %v", err)
	}

	return nil
}

// FindSimilarDocuments finds documents similar to the given embedding
func (db *DocumentStore) FindSimilarDocuments(ctx context.Context, embedding []float32, limit int) ([]Document, error) {
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
func (db *DocumentStore) FindSimilarDocumentsByChatID(ctx context.Context, embedding []float32, chatID string, limit int) ([]Document, error) {
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
func (db *DocumentStore) DeleteDocument(ctx context.Context, id string) error {
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
func (db *DocumentStore) GetDocumentsByPrefix(ctx context.Context, prefix string) ([]Document, error) {
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

// FindDocumentsByChatbot retrieves all documents associated with a specific chatbot
func (db *DocumentStore) FindDocumentsByChatbot(ctx context.Context, chatbotID string) ([]Document, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, content, embedding, chatbot_id
		FROM documents
		WHERE chatbot_id = $1
	`, chatbotID)
	
	if err != nil {
		return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, 
			"failed to query documents by chatbot: %v", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		var pgvec pgvector.Vector
		
		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec, &doc.ChatbotID); err != nil {
			return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, 
				"failed to scan document row: %v", err)
		}
		
		doc.Embedding = pgvec.Slice()
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, 
			"error iterating document rows: %v", err)
	}

	return documents, nil
}

// FindSimilarDocumentsByChatbot finds documents similar to the given embedding that belong to a specific chatbot
func (db *DocumentStore) FindSimilarDocumentsByChatbot(ctx context.Context, embedding []float32, chatbotID string, limit int) ([]Document, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, content, embedding, chatbot_id
		FROM documents
		WHERE chatbot_id = $1
		ORDER BY embedding <=> $2
		LIMIT $3
	`, chatbotID, pgvector.NewVector(embedding), limit)
	
	if err != nil {
		return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, 
			"failed to query similar documents by chatbot: %v", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		var pgvec pgvector.Vector
		
		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec, &doc.ChatbotID); err != nil {
			return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, 
				"failed to scan document row: %v", err)
		}
		
		doc.Embedding = pgvec.Slice()
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, 
			"error iterating document rows: %v", err)
	}

	return documents, nil
} 