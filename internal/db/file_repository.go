package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// fileRepository implements FileRepository interface
type fileRepository struct {
	db *Database
}

// NewFileRepository creates a new file repository
func NewFileRepository(db *Database) FileRepositoryTx {
	return &fileRepository{db: db}
}

// Create creates a new file
func (r *fileRepository) Create(ctx context.Context, file *File) error {
	if file.ID == uuid.Nil {
		file.ID = uuid.New()
	}
	if file.UploadedAt.IsZero() {
		file.UploadedAt = time.Now()
	}

	query := `
		INSERT INTO files (id, chatbot_id, filename, uploaded_at)
		VALUES (:id, :chatbot_id, :filename, :uploaded_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, file)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrFileAlreadyExists
		}
		if IsForeignKeyViolationError(err) {
			return apperrors.ErrChatbotNotFound
		}
		return apperrors.Wrap(err, "failed to create file")
	}

	return nil
}

// CreateTx creates a new file within a transaction
func (r *fileRepository) CreateTx(ctx context.Context, tx *Transaction, file *File) error {
	if file.ID == uuid.Nil {
		file.ID = uuid.New()
	}
	if file.UploadedAt.IsZero() {
		file.UploadedAt = time.Now()
	}

	query := `
		INSERT INTO files (id, chatbot_id, filename, uploaded_at)
		VALUES (:id, :chatbot_id, :filename, :uploaded_at)
	`

	_, err := tx.NamedExecContext(ctx, query, file)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrFileAlreadyExists
		}
		if IsForeignKeyViolationError(err) {
			return apperrors.ErrChatbotNotFound
		}
		return apperrors.Wrap(err, "failed to create file")
	}

	return nil
}

// FindByID finds a file by ID
func (r *fileRepository) FindByID(ctx context.Context, id uuid.UUID) (*File, error) {
	var file File
	query := `
		SELECT id, chatbot_id, filename, uploaded_at
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
func (r *fileRepository) FindByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]*File, error) {
	var files []*File
	query := `
		SELECT id, chatbot_id, filename, uploaded_at
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

// FindByChatbotIDWithPagination finds files for a chatbot with pagination
func (r *fileRepository) FindByChatbotIDWithPagination(ctx context.Context, chatbotID uuid.UUID, offset, limit int) ([]*File, int64, error) {
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
		SELECT id, chatbot_id, filename, uploaded_at
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

// Update updates a file
func (r *fileRepository) Update(ctx context.Context, file *File) error {
	query := `
		UPDATE files
		SET filename = :filename
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
func (r *fileRepository) UpdateTx(ctx context.Context, tx *Transaction, file *File) error {
	query := `
		UPDATE files
		SET filename = :filename
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
func (r *fileRepository) Delete(ctx context.Context, id uuid.UUID) error {
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
func (r *fileRepository) DeleteTx(ctx context.Context, tx *Transaction, id uuid.UUID) error {
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
func (r *fileRepository) DeleteByChatbotID(ctx context.Context, chatbotID uuid.UUID) error {
	query := `DELETE FROM files WHERE chatbot_id = $1`

	_, err := r.db.ExecContext(ctx, query, chatbotID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete files by chatbot ID")
	}

	return nil
}

// DeleteByChatbotIDTx deletes all files for a chatbot within a transaction
func (r *fileRepository) DeleteByChatbotIDTx(ctx context.Context, tx *Transaction, chatbotID uuid.UUID) error {
	query := `DELETE FROM files WHERE chatbot_id = $1`

	_, err := tx.ExecContext(ctx, query, chatbotID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete files by chatbot ID")
	}

	return nil
}

// Count returns the total number of files
func (r *fileRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM files`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to count files")
	}

	return count, nil
}

// CountByChatbotID returns the number of files for a chatbot
func (r *fileRepository) CountByChatbotID(ctx context.Context, chatbotID uuid.UUID) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM files WHERE chatbot_id = $1`

	err := r.db.GetContext(ctx, &count, query, chatbotID)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to count files by chatbot ID")
	}

	return count, nil
}
