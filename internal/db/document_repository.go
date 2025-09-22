package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

type DocumentRepository struct {
	db *Database
}

func NewDocumentRepository(db *Database) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// Store stores a document with its vector embedding
func (r *DocumentRepository) Store(ctx context.Context, doc *Document) error {
	query := `
		INSERT INTO documents (id, content, embedding, chatbot_id, shared_knowledge_base_id, file_id, chunk_index)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
		SET content = $2,
		    embedding = $3,
		    chatbot_id = $4,
		    shared_knowledge_base_id = $5,
		    file_id = $6,
		    chunk_index = $7
	`

	_, err := r.db.ExecContext(ctx, query, doc.ID, doc.Content, doc.Embedding, doc.ChatbotID, doc.SharedKnowledgeBaseID, doc.FileID, doc.ChunkIndex)
	if err != nil {
		return apperrors.Wrap(err, "failed to store document")
	}

	return nil
}

// StoreWithEmbedding stores a document with embedding as float32 slice
func (r *DocumentRepository) StoreWithEmbedding(ctx context.Context, doc *DocumentWithEmbedding) error {
	query := `
		INSERT INTO documents (id, content, embedding, chatbot_id, shared_knowledge_base_id, file_id, chunk_index)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
		SET content = $2,
		    embedding = $3,
		    chatbot_id = $4,
		    shared_knowledge_base_id = $5,
		    file_id = $6,
		    chunk_index = $7
	`

	_, err := r.db.ExecContext(ctx, query, doc.ID, doc.Content, pgvector.NewVector(doc.Embedding), doc.ChatbotID, doc.SharedKnowledgeBaseID, doc.FileID, doc.ChunkIndex)
	if err != nil {
		return apperrors.Wrap(err, "failed to store document with embedding")
	}

	return nil
}

// StoreTx stores a document within a transaction
func (r *DocumentRepository) StoreTx(ctx context.Context, tx *Transaction, doc *Document) error {
	query := `
		INSERT INTO documents (id, content, embedding, chatbot_id, shared_knowledge_base_id, file_id, chunk_index)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
		SET content = $2,
		    embedding = $3,
		    chatbot_id = $4,
		    shared_knowledge_base_id = $5,
		    file_id = $6,
		    chunk_index = $7
	`

	_, err := tx.ExecContext(ctx, query, doc.ID, doc.Content, doc.Embedding, doc.ChatbotID, doc.SharedKnowledgeBaseID, doc.FileID, doc.ChunkIndex)
	if err != nil {
		return apperrors.Wrap(err, "failed to store document")
	}

	return nil
}

// StoreWithEmbeddingTx stores a document with embedding within a transaction
func (r *DocumentRepository) StoreWithEmbeddingTx(ctx context.Context, tx *Transaction, doc *DocumentWithEmbedding) error {
	query := `
		INSERT INTO documents (id, content, embedding, chatbot_id, shared_knowledge_base_id, file_id, chunk_index)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
		SET content = $2,
		    embedding = $3,
		    chatbot_id = $4,
		    shared_knowledge_base_id = $5,
		    file_id = $6,
		    chunk_index = $7
	`

	_, err := tx.ExecContext(ctx, query, doc.ID, doc.Content, pgvector.NewVector(doc.Embedding), doc.ChatbotID, doc.SharedKnowledgeBaseID, doc.FileID, doc.ChunkIndex)
	if err != nil {
		return apperrors.Wrap(err, "failed to store document with embedding")
	}

	return nil
}

// FindByID finds a document by ID
func (r *DocumentRepository) FindByID(ctx context.Context, id string) (*Document, error) {
	var doc Document
	query := `
		SELECT id, content, embedding, chatbot_id, shared_knowledge_base_id, file_id, chunk_index
		FROM documents
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &doc, query, id)
	if err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrDocumentNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find document by ID")
	}

	return &doc, nil
}

// FindSimilar finds documents similar to the given embedding
func (r *DocumentRepository) FindSimilar(ctx context.Context, embedding []float32, limit int) ([]*DocumentWithEmbedding, error) {
	var docs []*Document
	query := `
		SELECT id, content, embedding, chatbot_id, shared_knowledge_base_id, file_id, chunk_index
		FROM documents
		ORDER BY embedding <=> $1
		LIMIT $2
	`

	err := r.db.SelectContext(ctx, &docs, query, pgvector.NewVector(embedding), limit)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find similar documents")
	}

	// Convert to DocumentWithEmbedding
	result := make([]*DocumentWithEmbedding, len(docs))
	for i, doc := range docs {
		result[i] = doc.ToDocumentWithEmbedding()
	}

	return result, nil
}

// FindSimilarByChatbot finds documents similar to the given embedding for a specific chatbot
func (r *DocumentRepository) FindSimilarByChatbot(ctx context.Context, embedding []float32, chatbotID uuid.UUID, limit int) ([]*DocumentWithEmbedding, error) {
	var docs []*Document
	query := `
		SELECT id, content, embedding, chatbot_id, shared_knowledge_base_id, file_id, chunk_index
		FROM documents
		WHERE chatbot_id = $1
		ORDER BY embedding <=> $2
		LIMIT $3
	`

	err := r.db.SelectContext(ctx, &docs, query, chatbotID, pgvector.NewVector(embedding), limit)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find similar documents by chatbot")
	}

	// Convert to DocumentWithEmbedding
	result := make([]*DocumentWithEmbedding, len(docs))
	for i, doc := range docs {
		result[i] = doc.ToDocumentWithEmbedding()
	}

	return result, nil
}

// FindSimilarBySharedKnowledgeBases finds documents similar to the embedding across shared KBs
func (r *DocumentRepository) FindSimilarBySharedKnowledgeBases(ctx context.Context, embedding []float32, kbIDs []uuid.UUID, limit int) ([]*DocumentWithEmbedding, error) {
	if len(kbIDs) == 0 {
		return []*DocumentWithEmbedding{}, nil
	}

	var docs []*Document
	query := `
		SELECT id, content, embedding, chatbot_id, shared_knowledge_base_id, file_id, chunk_index
		FROM documents
		WHERE shared_knowledge_base_id = ANY($1)
		ORDER BY embedding <=> $2
		LIMIT $3
	`

	err := r.db.SelectContext(ctx, &docs, query, pq.Array(kbIDs), pgvector.NewVector(embedding), limit)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find similar documents by shared knowledge bases")
	}

	result := make([]*DocumentWithEmbedding, len(docs))
	for i, doc := range docs {
		result[i] = doc.ToDocumentWithEmbedding()
	}

	return result, nil
}

// FindByChatbotID finds all documents for a chatbot
func (r *DocumentRepository) FindByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]*Document, error) {
	var docs []*Document
	query := `
		SELECT id, content, embedding, chatbot_id, shared_knowledge_base_id, file_id, chunk_index
		FROM documents
		WHERE chatbot_id = $1
		ORDER BY id
	`

	err := r.db.SelectContext(ctx, &docs, query, chatbotID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find documents by chatbot ID")
	}

	return docs, nil
}

// FindBySharedKnowledgeBaseID finds all documents for a shared knowledge base
func (r *DocumentRepository) FindBySharedKnowledgeBaseID(ctx context.Context, kbID uuid.UUID) ([]*Document, error) {
	var docs []*Document
	query := `
		SELECT id, content, embedding, chatbot_id, shared_knowledge_base_id, file_id, chunk_index
		FROM documents
		WHERE shared_knowledge_base_id = $1
		ORDER BY id
	`

	err := r.db.SelectContext(ctx, &docs, query, kbID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find documents by shared knowledge base ID")
	}

	return docs, nil
}

// FindByFileID finds all documents for a file
func (r *DocumentRepository) FindByFileID(ctx context.Context, fileID uuid.UUID) ([]*Document, error) {
	var docs []*Document
	query := `
		SELECT id, content, embedding, chatbot_id, shared_knowledge_base_id, file_id, chunk_index
		FROM documents
		WHERE file_id = $1
		ORDER BY chunk_index NULLS LAST, id
	`

	err := r.db.SelectContext(ctx, &docs, query, fileID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find documents by file ID")
	}

	return docs, nil
}

// Delete deletes a document by ID
func (r *DocumentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM documents WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete document")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrDocumentNotFound
	}

	return nil
}

// DeleteTx deletes a document by ID within a transaction
func (r *DocumentRepository) DeleteTx(ctx context.Context, tx *Transaction, id string) error {
	query := `DELETE FROM documents WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete document")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrDocumentNotFound
	}

	return nil
}

// DeleteByChatbotID deletes all documents for a chatbot
func (r *DocumentRepository) DeleteByChatbotID(ctx context.Context, chatbotID uuid.UUID) error {
	query := `DELETE FROM documents WHERE chatbot_id = $1`

	_, err := r.db.ExecContext(ctx, query, chatbotID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete documents by chatbot ID")
	}

	return nil
}

// DeleteByChatbotIDTx deletes all documents for a chatbot within a transaction
func (r *DocumentRepository) DeleteByChatbotIDTx(ctx context.Context, tx *Transaction, chatbotID uuid.UUID) error {
	query := `DELETE FROM documents WHERE chatbot_id = $1`

	_, err := tx.ExecContext(ctx, query, chatbotID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete documents by chatbot ID")
	}

	return nil
}

// DeleteBySharedKnowledgeBaseID deletes all documents associated with a shared KB
func (r *DocumentRepository) DeleteBySharedKnowledgeBaseID(ctx context.Context, kbID uuid.UUID) error {
	query := `DELETE FROM documents WHERE shared_knowledge_base_id = $1`

	_, err := r.db.ExecContext(ctx, query, kbID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete documents by shared knowledge base ID")
	}

	return nil
}

// DeleteBySharedKnowledgeBaseIDTx deletes all documents associated with a shared KB within a transaction
func (r *DocumentRepository) DeleteBySharedKnowledgeBaseIDTx(ctx context.Context, tx *Transaction, kbID uuid.UUID) error {
	query := `DELETE FROM documents WHERE shared_knowledge_base_id = $1`

	_, err := tx.ExecContext(ctx, query, kbID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete documents by shared knowledge base ID")
	}

	return nil
}

// DeleteByFileID deletes all documents for a file
func (r *DocumentRepository) DeleteByFileID(ctx context.Context, fileID uuid.UUID) error {
	query := `DELETE FROM documents WHERE file_id = $1`

	_, err := r.db.ExecContext(ctx, query, fileID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete documents by file ID")
	}

	return nil
}

// DeleteByFileIDTx deletes all documents for a file within a transaction
func (r *DocumentRepository) DeleteByFileIDTx(ctx context.Context, tx *Transaction, fileID uuid.UUID) error {
	query := `DELETE FROM documents WHERE file_id = $1`

	_, err := tx.ExecContext(ctx, query, fileID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete documents by file ID")
	}

	return nil
}

// Count returns the total number of documents
func (r *DocumentRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM documents`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to count documents")
	}

	return count, nil
}

// CountByChatbotID returns the number of documents for a chatbot
func (r *DocumentRepository) CountByChatbotID(ctx context.Context, chatbotID uuid.UUID) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM documents WHERE chatbot_id = $1`

	err := r.db.GetContext(ctx, &count, query, chatbotID)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to count documents by chatbot ID")
	}

	return count, nil
}
