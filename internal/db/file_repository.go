package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

type FileRepository struct {
	db *Database
}

func NewFileRepository(db *Database) *FileRepository {
	return &FileRepository{db: db}
}

// Create creates a new file
func (r *FileRepository) Create(ctx context.Context, file *File) error {
	if file.ID == uuid.Nil {
		file.ID = uuid.New()
	}
	if file.UploadedAt.IsZero() {
		file.UploadedAt = time.Now()
	}

	query := `
		INSERT INTO files (id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at)
		VALUES (:id, :chatbot_id, :shared_knowledge_base_id, :filename, :size_bytes, :uploaded_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, file)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrFileAlreadyExists
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			switch pqErr.Constraint {
			case "files_chatbot_id_fkey":
				return apperrors.ErrChatbotNotFound
			case "files_shared_knowledge_base_id_fkey":
				return apperrors.ErrSharedKnowledgeBaseNotFound
			}
		}
		return apperrors.Wrap(err, "failed to create file")
	}

	return nil
}

// CreateTx creates a new file within a transaction
func (r *FileRepository) CreateTx(ctx context.Context, tx *Transaction, file *File) error {
	if file.ID == uuid.Nil {
		file.ID = uuid.New()
	}
	if file.UploadedAt.IsZero() {
		file.UploadedAt = time.Now()
	}

	query := `
		INSERT INTO files (id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at)
		VALUES (:id, :chatbot_id, :shared_knowledge_base_id, :filename, :size_bytes, :uploaded_at)
	`

	_, err := tx.NamedExecContext(ctx, query, file)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrFileAlreadyExists
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			switch pqErr.Constraint {
			case "files_chatbot_id_fkey":
				return apperrors.ErrChatbotNotFound
			case "files_shared_knowledge_base_id_fkey":
				return apperrors.ErrSharedKnowledgeBaseNotFound
			}
		}
		return apperrors.Wrap(err, "failed to create file")
	}

	return nil
}

// FindByID finds a file by ID
func (r *FileRepository) FindByID(ctx context.Context, id uuid.UUID) (*File, error) {
	var file File
	query := `
		SELECT id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at
		FROM files
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &file, query, id)
	if err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrFileNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find file by ID")
	}

	return &file, nil
}

// FindByChatbotID finds all files for a chatbot
func (r *FileRepository) FindByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]*File, error) {
	var files []*File
	query := `
        SELECT id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at
        FROM files
        WHERE chatbot_id = $1
        ORDER BY uploaded_at DESC
    `

	err := r.db.SelectContext(ctx, &files, query, chatbotID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find files by chatbot ID")
	}

	return files, nil
}

// FindNonTextByChatbotID finds all non-text files for a chatbot
func (r *FileRepository) FindNonTextByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]*File, error) {
	var files []*File
	query := `
        SELECT id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at
        FROM files
        WHERE chatbot_id = $1 AND filename NOT LIKE 'text-%'
        ORDER BY uploaded_at DESC
    `

	err := r.db.SelectContext(ctx, &files, query, chatbotID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find non-text files by chatbot ID")
	}

	return files, nil
}

// FindByChatbotIDAndFilename finds a single file by chatbot and filename
func (r *FileRepository) FindByChatbotIDAndFilename(ctx context.Context, chatbotID uuid.UUID, filename string) (*File, error) {
	var file File
	query := `
        SELECT id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at
        FROM files
        WHERE chatbot_id = $1 AND filename = $2
        LIMIT 1
    `

	err := r.db.GetContext(ctx, &file, query, chatbotID, filename)
	if err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrFileNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find file by chatbot and filename")
	}

	return &file, nil
}

// FindBySharedKnowledgeBaseID returns files attached to a shared knowledge base
func (r *FileRepository) FindBySharedKnowledgeBaseID(ctx context.Context, kbID uuid.UUID) ([]*File, error) {
	var files []*File
	query := `
		SELECT id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at
		FROM files
		WHERE shared_knowledge_base_id = $1
		ORDER BY uploaded_at DESC
	`

	err := r.db.SelectContext(ctx, &files, query, kbID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find files by shared knowledge base ID")
	}
	return files, nil
}

// FindNonTextBySharedKnowledgeBaseID finds non-text files for a shared knowledge base
func (r *FileRepository) FindNonTextBySharedKnowledgeBaseID(ctx context.Context, kbID uuid.UUID) ([]*File, error) {
	var files []*File
	query := `
		SELECT id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at
		FROM files
		WHERE shared_knowledge_base_id = $1 AND filename NOT LIKE 'text-%'
		ORDER BY uploaded_at DESC
	`

	err := r.db.SelectContext(ctx, &files, query, kbID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find files by shared knowledge base ID")
	}

	return files, nil
}

// FindBySharedKnowledgeBaseIDAndFilename finds a file within a shared KB by filename
func (r *FileRepository) FindBySharedKnowledgeBaseIDAndFilename(ctx context.Context, kbID uuid.UUID, filename string) (*File, error) {
	var file File
	query := `
		SELECT id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at
		FROM files
		WHERE shared_knowledge_base_id = $1 AND filename = $2
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &file, query, kbID, filename)
	if err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrFileNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find file by shared knowledge base and filename")
	}

	return &file, nil
}

// FindBySharedKnowledgeBaseIDs returns files for many shared KBs
func (r *FileRepository) FindBySharedKnowledgeBaseIDs(ctx context.Context, kbIDs []uuid.UUID) ([]*File, error) {
	if len(kbIDs) == 0 {
		return []*File{}, nil
	}

	var files []*File
	query := `
		SELECT id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at
		FROM files
		WHERE shared_knowledge_base_id = ANY($1)
		ORDER BY uploaded_at DESC
	`

	err := r.db.SelectContext(ctx, &files, query, pq.Array(kbIDs))
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find files by shared knowledge base IDs")
	}

	return files, nil
}

// FindTextBySharedKnowledgeBaseID returns text sources stored in a shared knowledge base
func (r *FileRepository) FindTextBySharedKnowledgeBaseID(ctx context.Context, kbID uuid.UUID) ([]*File, error) {
	var files []*File
	query := `
		SELECT id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at
		FROM files
		WHERE shared_knowledge_base_id = $1 AND filename LIKE 'text-%'
		ORDER BY uploaded_at DESC
	`

	err := r.db.SelectContext(ctx, &files, query, kbID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find text sources by shared knowledge base ID")
	}

	return files, nil
}

// FindByChatbotIDWithPagination finds files for a chatbot with pagination
func (r *FileRepository) FindByChatbotIDWithPagination(ctx context.Context, chatbotID uuid.UUID, offset, limit int) ([]*File, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM files WHERE chatbot_id = $1`
	err := r.db.GetContext(ctx, &total, countQuery, chatbotID)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to get total files count")
	}

	// Get paginated results
	var files []*File
	query := `
		SELECT id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at
		FROM files
		WHERE chatbot_id = $1
		ORDER BY uploaded_at DESC
		LIMIT $2 OFFSET $3
	`

	err = r.db.SelectContext(ctx, &files, query, chatbotID, limit, offset)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to find files with pagination")
	}

	return files, total, nil
}

// FindTextByChatbotID finds text sources (files with filename starting with 'text-')
func (r *FileRepository) FindTextByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]*File, error) {
	var files []*File
	query := `
        SELECT id, chatbot_id, shared_knowledge_base_id, filename, size_bytes, uploaded_at
        FROM files
        WHERE chatbot_id = $1 AND filename LIKE 'text-%'
        ORDER BY uploaded_at DESC
    `

	err := r.db.SelectContext(ctx, &files, query, chatbotID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find text sources by chatbot ID")
	}

	return files, nil
}

// Update updates a file
func (r *FileRepository) Update(ctx context.Context, file *File) error {
	query := `
		UPDATE files
		SET filename = :filename, size_bytes = :size_bytes
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, file)
	if err != nil {
		return apperrors.Wrap(err, "failed to update file")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrFileNotFound
	}

	return nil
}

// UpdateTx updates a file within a transaction
func (r *FileRepository) UpdateTx(ctx context.Context, tx *Transaction, file *File) error {
	query := `
		UPDATE files
		SET filename = :filename, size_bytes = :size_bytes
		WHERE id = :id
	`

	result, err := tx.NamedExecContext(ctx, query, file)
	if err != nil {
		return apperrors.Wrap(err, "failed to update file")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrFileNotFound
	}

	return nil
}

// Delete deletes a file by ID
func (r *FileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM files WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete file")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrFileNotFound
	}

	return nil
}

// DeleteTx deletes a file by ID within a transaction
func (r *FileRepository) DeleteTx(ctx context.Context, tx *Transaction, id uuid.UUID) error {
	query := `DELETE FROM files WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete file")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrFileNotFound
	}

	return nil
}

// DeleteByChatbotID deletes all files for a chatbot
func (r *FileRepository) DeleteByChatbotID(ctx context.Context, chatbotID uuid.UUID) error {
	query := `DELETE FROM files WHERE chatbot_id = $1`

	_, err := r.db.ExecContext(ctx, query, chatbotID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete files by chatbot ID")
	}

	return nil
}

// DeleteByChatbotIDTx deletes all files for a chatbot within a transaction
func (r *FileRepository) DeleteByChatbotIDTx(ctx context.Context, tx *Transaction, chatbotID uuid.UUID) error {
	query := `DELETE FROM files WHERE chatbot_id = $1`

	_, err := tx.ExecContext(ctx, query, chatbotID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete files by chatbot ID")
	}

	return nil
}

// DeleteBySharedKnowledgeBaseID removes files bound to a shared knowledge base
func (r *FileRepository) DeleteBySharedKnowledgeBaseID(ctx context.Context, kbID uuid.UUID) error {
	query := `DELETE FROM files WHERE shared_knowledge_base_id = $1`

	_, err := r.db.ExecContext(ctx, query, kbID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete files by shared knowledge base ID")
	}

	return nil
}

// DeleteBySharedKnowledgeBaseIDTx removes files bound to a shared knowledge base inside a transaction
func (r *FileRepository) DeleteBySharedKnowledgeBaseIDTx(ctx context.Context, tx *Transaction, kbID uuid.UUID) error {
	query := `DELETE FROM files WHERE shared_knowledge_base_id = $1`

	_, err := tx.ExecContext(ctx, query, kbID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete files by shared knowledge base ID")
	}

	return nil
}

// Count returns the total number of files
func (r *FileRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM files`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to count files")
	}

	return count, nil
}

// CountByChatbotID returns the number of files for a chatbot
func (r *FileRepository) CountByChatbotID(ctx context.Context, chatbotID uuid.UUID) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM files WHERE chatbot_id = $1`

	err := r.db.GetContext(ctx, &count, query, chatbotID)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to count files by chatbot ID")
	}

	return count, nil
}

// CountBySharedKnowledgeBaseID returns the number of files linked to a shared KB
func (r *FileRepository) CountBySharedKnowledgeBaseID(ctx context.Context, kbID uuid.UUID) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM files WHERE shared_knowledge_base_id = $1`

	err := r.db.GetContext(ctx, &count, query, kbID)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to count files by shared knowledge base ID")
	}

	return count, nil
}
