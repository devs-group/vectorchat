package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

type ChatbotRepository struct {
	db *Database
}

func NewChatbotRepository(db *Database) *ChatbotRepository {
	return &ChatbotRepository{db: db}
}

// Create creates a new chatbot
func (r *ChatbotRepository) Create(ctx context.Context, chatbot *Chatbot) error {
	if chatbot.ID == uuid.Nil {
		chatbot.ID = uuid.New()
	}
	if chatbot.CreatedAt.IsZero() {
		chatbot.CreatedAt = time.Now()
	}
	if chatbot.UpdatedAt.IsZero() {
		chatbot.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO chatbots (
			id, user_id, organization_id, name, description, system_instructions,
			model_name, temperature_param, max_tokens, use_max_tokens, save_messages, is_enabled, created_at, updated_at
		) VALUES (
			:id, :user_id, :organization_id, :name, :description, :system_instructions,
			:model_name, :temperature_param, :max_tokens, :use_max_tokens, :save_messages, :is_enabled, :created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, chatbot)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrChatbotAlreadyExists
		}
		if IsForeignKeyViolationError(err) {
			return apperrors.ErrUserNotFound
		}
		return apperrors.Wrap(err, "failed to create chatbot")
	}

	return nil
}

// CreateTx creates a new chatbot within a transaction
func (r *ChatbotRepository) CreateTx(ctx context.Context, tx *Transaction, chatbot *Chatbot) error {
	if chatbot.ID == uuid.Nil {
		chatbot.ID = uuid.New()
	}
	if chatbot.CreatedAt.IsZero() {
		chatbot.CreatedAt = time.Now()
	}
	if chatbot.UpdatedAt.IsZero() {
		chatbot.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO chatbots (
			id, user_id, organization_id, name, description, system_instructions,
			model_name, temperature_param, max_tokens, use_max_tokens, save_messages, is_enabled, created_at, updated_at
		) VALUES (
			:id, :user_id, :organization_id, :name, :description, :system_instructions,
			:model_name, :temperature_param, :max_tokens, :use_max_tokens, :save_messages, :is_enabled, :created_at, :updated_at
		)
	`

	_, err := tx.NamedExecContext(ctx, query, chatbot)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrChatbotAlreadyExists
		}
		if IsForeignKeyViolationError(err) {
			return apperrors.ErrUserNotFound
		}
		return apperrors.Wrap(err, "failed to create chatbot")
	}

	return nil
}

// FindByID finds a chatbot by ID
func (r *ChatbotRepository) FindByID(ctx context.Context, id uuid.UUID) (*Chatbot, error) {
	var chatbot Chatbot
	query := `
		SELECT id, user_id, organization_id, name, description, system_instructions,
		       model_name, temperature_param, max_tokens, use_max_tokens, save_messages, is_enabled, created_at, updated_at
		FROM chatbots
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &chatbot, query, id)
	if err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrChatbotNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find chatbot by ID")
	}

	return &chatbot, nil
}

// FindByIDAndScope finds a chatbot by ID scoped to either user or organization
func (r *ChatbotRepository) FindByIDAndScope(ctx context.Context, id uuid.UUID, userID string, orgID *uuid.UUID) (*Chatbot, error) {
	var chatbot Chatbot
	query := `
		SELECT id, user_id, organization_id, name, description, system_instructions,
		       model_name, temperature_param, max_tokens, use_max_tokens, save_messages, is_enabled, created_at, updated_at
		FROM chatbots
		WHERE id = $1 AND (
			($2::uuid IS NULL AND organization_id IS NULL AND user_id = $3)
			OR (organization_id = $2::uuid)
		)
	`

	err := r.db.GetContext(ctx, &chatbot, query, id, orgID, userID)
	if err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrChatbotNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find chatbot by ID and user ID")
	}

	return &chatbot, nil
}

// FindByUserID finds all chatbots for a user
func (r *ChatbotRepository) FindByUserID(ctx context.Context, userID string) ([]*Chatbot, error) {
	var chatbots []*Chatbot
	query := `
		SELECT id, user_id, organization_id, name, description, system_instructions,
		       model_name, temperature_param, max_tokens, use_max_tokens, save_messages, is_enabled, created_at, updated_at
		FROM chatbots
		WHERE user_id = $1 AND organization_id IS NULL
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &chatbots, query, userID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find chatbots by user ID")
	}

	return chatbots, nil
}

// FindByOrgID finds all chatbots for an organization
func (r *ChatbotRepository) FindByOrgID(ctx context.Context, orgID uuid.UUID) ([]*Chatbot, error) {
	var chatbots []*Chatbot
	query := `
		SELECT id, user_id, organization_id, name, description, system_instructions,
		       model_name, temperature_param, max_tokens, use_max_tokens, save_messages, is_enabled, created_at, updated_at
		FROM chatbots
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`

	if err := r.db.SelectContext(ctx, &chatbots, query, orgID); err != nil {
		return nil, apperrors.Wrap(err, "failed to find chatbots by organization ID")
	}

	return chatbots, nil
}

// FindByUserIDWithPagination finds chatbots for a user with pagination
func (r *ChatbotRepository) FindByUserIDWithPagination(ctx context.Context, userID string, offset, limit int) ([]*Chatbot, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM chatbots WHERE user_id = $1 AND organization_id IS NULL`
	err := r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to get total chatbots count")
	}

	// Get paginated results
	var chatbots []*Chatbot
	query := `
		SELECT id, user_id, name, description, system_instructions,
		       model_name, temperature_param, max_tokens, use_max_tokens, save_messages, is_enabled, created_at, updated_at
		FROM chatbots
		WHERE user_id = $1 AND organization_id IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err = r.db.SelectContext(ctx, &chatbots, query, userID, limit, offset)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to find chatbots with pagination")
	}

	return chatbots, total, nil
}

// Update updates a chatbot
func (r *ChatbotRepository) Update(ctx context.Context, chatbot *Chatbot) error {
	chatbot.UpdatedAt = time.Now()

	query := `
		UPDATE chatbots
		SET name = :name, description = :description, system_instructions = :system_instructions,
		    model_name = :model_name, temperature_param = :temperature_param,
		    max_tokens = :max_tokens, use_max_tokens = :use_max_tokens,
		    save_messages = :save_messages, is_enabled = :is_enabled, updated_at = :updated_at
		WHERE id = :id AND (
			(:organization_id IS NULL AND organization_id IS NULL AND user_id = :user_id)
			OR organization_id = CAST(:organization_id AS uuid)
		)
	`

	result, err := r.db.NamedExecContext(ctx, query, chatbot)
	if err != nil {
		return apperrors.Wrap(err, "failed to update chatbot")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrChatbotNotFound
	}

	return nil
}

// UpdateTx updates a chatbot within a transaction
func (r *ChatbotRepository) UpdateTx(ctx context.Context, tx *Transaction, chatbot *Chatbot) error {
	chatbot.UpdatedAt = time.Now()

	query := `
		UPDATE chatbots
		SET name = :name, description = :description, system_instructions = :system_instructions,
		    model_name = :model_name, temperature_param = :temperature_param,
		    max_tokens = :max_tokens, use_max_tokens = :use_max_tokens,
		    save_messages = :save_messages, is_enabled = :is_enabled, updated_at = :updated_at
		WHERE id = :id AND (
			(:organization_id IS NULL AND organization_id IS NULL AND user_id = :user_id)
			OR organization_id = CAST(:organization_id AS uuid)
		)
	`

	result, err := tx.NamedExecContext(ctx, query, chatbot)
	if err != nil {
		return apperrors.Wrap(err, "failed to update chatbot")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrChatbotNotFound
	}

	return nil
}

// Delete deletes a chatbot with scope awareness
func (r *ChatbotRepository) Delete(ctx context.Context, id uuid.UUID, userID string, orgID *uuid.UUID) error {
	query := `
		DELETE FROM chatbots
		WHERE id = $1 AND (
			($2::uuid IS NULL AND organization_id IS NULL AND user_id = $3)
			OR organization_id = $2::uuid
		)
	`

	result, err := r.db.ExecContext(ctx, query, id, orgID, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete chatbot")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrChatbotNotFound
	}

	return nil
}

// DeleteTx deletes a chatbot within a transaction
func (r *ChatbotRepository) DeleteTx(ctx context.Context, tx *Transaction, id uuid.UUID, userID string, orgID *uuid.UUID) error {
	query := `
		DELETE FROM chatbots
		WHERE id = $1 AND (
			($2::uuid IS NULL AND organization_id IS NULL AND user_id = $3)
			OR organization_id = $2::uuid
		)
	`

	result, err := tx.ExecContext(ctx, query, id, orgID, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete chatbot")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrChatbotNotFound
	}

	return nil
}

// CheckOwnership checks if a user owns a chatbot
func (r *ChatbotRepository) CheckOwnership(ctx context.Context, id uuid.UUID, userID string, orgID *uuid.UUID) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM chatbots
			WHERE id = $1 AND (
				($2::uuid IS NULL AND organization_id IS NULL AND user_id = $3)
				OR organization_id = $2
			)
		)
	`

	err := r.db.GetContext(ctx, &exists, query, id, orgID, userID)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to check chatbot ownership")
	}

	return exists, nil
}

// TransferToOrganization moves a personal chatbot into an organization.
func (r *ChatbotRepository) TransferToOrganization(ctx context.Context, id uuid.UUID, userID string, orgID uuid.UUID) error {
	query := `
		UPDATE chatbots
		SET organization_id = $1, updated_at = $2
		WHERE id = $3 AND organization_id IS NULL AND user_id = $4
	`

	result, err := r.db.ExecContext(ctx, query, orgID, time.Now(), id, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to transfer chatbot to organization")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to read rows affected")
	}
	if rows == 0 {
		return apperrors.ErrUnauthorizedChatbotAccess
	}

	return nil
}

// UpdateBasicInfo updates basic chatbot information
func (r *ChatbotRepository) UpdateBasicInfo(ctx context.Context, id uuid.UUID, userID string, orgID *uuid.UUID, name, description string) error {
	query := `
		UPDATE chatbots
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4 AND (
			($5::uuid IS NULL AND organization_id IS NULL AND user_id = $6)
			OR organization_id = $5
		)
	`

	result, err := r.db.ExecContext(ctx, query, name, description, time.Now(), id, orgID, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to update chatbot basic info")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrChatbotNotFound
	}

	return nil
}

// UpdateSystemInstructions updates chatbot system instructions
func (r *ChatbotRepository) UpdateSystemInstructions(ctx context.Context, id uuid.UUID, userID string, orgID *uuid.UUID, instructions string) error {
	query := `
		UPDATE chatbots
		SET system_instructions = $1, updated_at = $2
		WHERE id = $3 AND (
			($4::uuid IS NULL AND organization_id IS NULL AND user_id = $5)
			OR organization_id = $4
		)
	`

	result, err := r.db.ExecContext(ctx, query, instructions, time.Now(), id, orgID, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to update chatbot system instructions")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrChatbotNotFound
	}

	return nil
}

// UpdateModelSettings updates chatbot model settings
func (r *ChatbotRepository) UpdateModelSettings(ctx context.Context, id uuid.UUID, userID string, orgID *uuid.UUID, modelName string, temperature float64, maxTokens int) error {
	query := `
		UPDATE chatbots
		SET model_name = $1, temperature_param = $2, max_tokens = $3, updated_at = $4
		WHERE id = $5 AND (
			($6::uuid IS NULL AND organization_id IS NULL AND user_id = $7)
			OR organization_id = $6
		)
	`

	result, err := r.db.ExecContext(ctx, query, modelName, temperature, maxTokens, time.Now(), id, orgID, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to update chatbot model settings")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.ErrChatbotNotFound
	}

	return nil
}
