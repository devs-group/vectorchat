package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
func NewDocumentStore(pool *pgxpool.Pool) *DocumentStore {
	return &DocumentStore{
		pool: pool,
	}
}

// StoreDocument stores a document with its vector embedding
func (db *DocumentStore) StoreDocument(ctx context.Context, doc Document) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO documents (id, content, embedding, chatbot_id, file_id, chunk_index)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE
		SET content = $2, embedding = $3, chatbot_id = $4, file_id = $5, chunk_index = $6
	`, doc.ID, doc.Content, pgvector.NewVector(doc.Embedding), doc.ChatbotID, doc.FileID, doc.ChunkIndex)

	if err != nil {
		return apperrors.Wrapf(apperrors.ErrDatabaseOperation, "failed to store document: %v", err)
	}

	return nil
}

// FindSimilarDocuments finds documents similar to the given embedding
func (db *DocumentStore) FindSimilarDocuments(ctx context.Context, embedding []float32, limit int) ([]Document, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, content, embedding, chatbot_id, file_id, chunk_index
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

		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec, &doc.ChatbotID, &doc.FileID, &doc.ChunkIndex); err != nil {
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
		SELECT id, content, embedding, chatbot_id, file_id, chunk_index
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

		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec, &doc.ChatbotID, &doc.FileID, &doc.ChunkIndex); err != nil {
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

// GetDocumentsByChatbotID retrieves documents for a given chatbot_id
func (db *DocumentStore) GetDocumentsByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]Document, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, content, embedding, chatbot_id, file_id, chunk_index
		FROM documents
		WHERE chatbot_id = $1
	`, chatbotID)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents by chatbot_id: %v", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		var pgvec pgvector.Vector

		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec, &doc.ChatbotID, &doc.FileID, &doc.ChunkIndex); err != nil {
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
		SELECT id, content, embedding, chatbot_id, file_id, chunk_index
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

		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec, &doc.ChatbotID, &doc.FileID, &doc.ChunkIndex); err != nil {
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
		SELECT id, content, embedding, chatbot_id, file_id, chunk_index
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

		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec, &doc.ChatbotID, &doc.FileID, &doc.ChunkIndex); err != nil {
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

// InsertFile inserts a new file record into the files table
func (db *DocumentStore) InsertFile(ctx context.Context, fileID uuid.UUID, chatbotID uuid.UUID, filename string) error {
	_, err := db.pool.Exec(ctx, `INSERT INTO files (id, chatbot_id, filename) VALUES ($1, $2, $3)`, fileID, chatbotID, filename)
	if err != nil {
		return apperrors.Wrap(err, "failed to insert file metadata")
	}
	return nil
}

// GetFilesByChatbotID retrieves all files for a given chatbot_id
func (db *DocumentStore) GetFilesByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]File, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, chatbot_id, filename, uploaded_at
		FROM files
		WHERE chatbot_id = $1
		ORDER BY uploaded_at DESC
	`, chatbotID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to query files by chatbot_id")
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var f File
		if err := rows.Scan(&f.ID, &f.ChatbotID, &f.Filename, &f.UploadedAt); err != nil {
			return nil, apperrors.Wrap(err, "failed to scan file row")
		}
		files = append(files, f)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, "error iterating file rows")
	}
	return files, nil
}
