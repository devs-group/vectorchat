package db

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v4/pgxpool"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// ChatbotStore handles database operations for chatbots
type ChatbotStore struct {
	pool *pgxpool.Pool
}

// NewChatbotStore creates a new chatbot store
func NewChatbotStore(pool *pgxpool.Pool) *ChatbotStore {
	return &ChatbotStore{
		pool: pool,
	}
}

// CreateChatbot creates a new chatbot in the database
func (s *ChatbotStore) CreateChatbot(ctx context.Context, chatbot *Chatbot) error {
	query := `
		INSERT INTO chatbots (
			user_id, name, description, system_prompt, 
			model_name, model_temperature, max_tokens, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW()) 
		RETURNING id, created_at, updated_at
	`
	
	return s.pool.QueryRow(
		ctx, 
		query,
		chatbot.UserID,
		chatbot.Name,
		chatbot.Description,
		chatbot.SystemInstructions,
		chatbot.ModelName,
		chatbot.TemperatureParam,
		chatbot.MaxTokens,
	).Scan(&chatbot.ID, &chatbot.CreatedAt, &chatbot.UpdatedAt)
}


// FindChatbotByIDAndUserID retrieves a chatbot by its ID and user ID
func (s *ChatbotStore) FindChatbotByIDAndUserID(ctx context.Context, id string, userID string) (*Chatbot, error) {
	query := `
		SELECT id, user_id, name, description, system_prompt, 
		       model_name, model_temperature, max_tokens, 
		       created_at, updated_at
		FROM chatbots
		WHERE id = $1 AND user_id = $2
	`
	
	var chatbot Chatbot
	err := s.pool.QueryRow(ctx, query, id, userID).Scan(
		&chatbot.ID,
		&chatbot.UserID,
		&chatbot.Name,
		&chatbot.Description,
		&chatbot.SystemInstructions,
		&chatbot.ModelName,
		&chatbot.TemperatureParam,
		&chatbot.MaxTokens,
		&chatbot.CreatedAt,
		&chatbot.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrChatbotNotFound
		}
		return nil, apperrors.Wrap(err, "failed to get chatbot")
	}
	
	return &chatbot, nil
}

// FindChatbotsByUserID retrieves all chatbots owned by a specific user
func (s *ChatbotStore) FindChatbotsByUserID(ctx context.Context, userID string) ([]Chatbot, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, name, description, system_instructions, 
			model_name, temperature_param, max_tokens, created_at, updated_at
		FROM chatbots
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, "failed to query chatbots: %v", err)
	}
	defer rows.Close()

	var chatbots []Chatbot
	for rows.Next() {
		var chatbot Chatbot
		err := rows.Scan(
			&chatbot.ID, &chatbot.UserID, &chatbot.Name, &chatbot.Description,
			&chatbot.SystemInstructions, &chatbot.ModelName, &chatbot.TemperatureParam,
			&chatbot.MaxTokens, &chatbot.CreatedAt, &chatbot.UpdatedAt,
		)
		if err != nil {
			return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, "failed to scan chatbot row: %v", err)
		}
		chatbots = append(chatbots, chatbot)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, "error iterating chatbot rows: %v", err)
	}

	return chatbots, nil
}

// UpdateChatbot updates an existing chatbot
func (s *ChatbotStore) UpdateChatbot(ctx context.Context, chatbot Chatbot) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE chatbots
		SET name = $1,
		description = $2,
		system_instructions = $3, 
		model_name = $4,
		temperature_param = $5,
		max_tokens = $6,
		updated_at = $7
		WHERE id = $8 AND user_id = $9
	`, chatbot.Name, chatbot.Description, chatbot.SystemInstructions,
		chatbot.ModelName, chatbot.TemperatureParam, chatbot.MaxTokens,
		chatbot.UpdatedAt, chatbot.ID, chatbot.UserID)

	if err != nil {
		return apperrors.Wrapf(apperrors.ErrDatabaseOperation, "failed to update chatbot: %v", err)
	}

	return nil
}

// DeleteChatbot deletes a chatbot by ID and owner
func (s *ChatbotStore) DeleteChatbot(ctx context.Context, id, userID string) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM chatbots
		WHERE id = $1 AND user_id = $2
	`, id, userID)

	if err != nil {
		return apperrors.Wrapf(apperrors.ErrDatabaseOperation, "failed to delete chatbot: %v", err)
	}

	// Check if any row was actually deleted
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return apperrors.Wrapf(apperrors.ErrChatbotNotFound, 
			"chatbot with ID %s not found or not owned by user %s", id, userID)
	}

	return nil
}

// CheckChatbotOwnership verifies if a user owns a specific chatbot
func (s *ChatbotStore) CheckChatbotOwnership(ctx context.Context, chatbotID, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM chatbots
			WHERE id = $1 AND user_id = $2
		)
	`, chatbotID, userID).Scan(&exists)

	if err != nil {
		return false, apperrors.Wrapf(apperrors.ErrDatabaseOperation, 
			"failed to check chatbot ownership: %v", err)
	}

	return exists, nil
}
