package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// RevisionRepository handles database operations for answer revisions
type RevisionRepository struct {
	db *Database
}

// NewRevisionRepository creates a new RevisionRepository
func NewRevisionRepository(db *Database) *RevisionRepository {
	return &RevisionRepository{db: db}
}

// CreateRevision creates a new answer revision
func (r *RevisionRepository) CreateRevision(ctx context.Context, revision *AnswerRevision) error {
	query := `
		INSERT INTO answer_revisions (
			id, chatbot_id, original_message_id, question,
			original_answer, revised_answer, question_embedding,
			revision_reason, revised_by, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		revision.ID,
		revision.ChatbotID,
		revision.OriginalMessageID,
		revision.Question,
		revision.OriginalAnswer,
		revision.RevisedAnswer,
		revision.QuestionEmbedding,
		revision.RevisionReason,
		revision.RevisedBy,
		revision.IsActive,
	)

	if err != nil {
		return apperrors.Wrap(err, "failed to create answer revision")
	}

	return nil
}

// FindSimilarRevisions finds revisions with similar questions using vector similarity
func (r *RevisionRepository) FindSimilarRevisions(ctx context.Context, questionEmbedding []float32, chatbotID uuid.UUID, threshold float64, limit int) ([]*AnswerRevisionWithEmbedding, error) {
	// Using cosine similarity (1 - (embedding <=> $1)) to get similarity score
	query := `
		SELECT
			id, chatbot_id, original_message_id, question,
			original_answer, revised_answer, question_embedding,
			revision_reason, revised_by, created_at, updated_at, is_active,
			1 - (question_embedding <=> $1) as similarity
		FROM answer_revisions
		WHERE chatbot_id = $2
			AND is_active = true
			AND 1 - (question_embedding <=> $1) >= $3
		ORDER BY similarity DESC
		LIMIT $4
	`

	rows, err := r.db.QueryContext(ctx, query, pgvector.NewVector(questionEmbedding), chatbotID, threshold, limit)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find similar revisions")
	}
	defer rows.Close()

	var revisions []*AnswerRevisionWithEmbedding
	for rows.Next() {
		var rev AnswerRevision
		var similarity float64

		err := rows.Scan(
			&rev.ID,
			&rev.ChatbotID,
			&rev.OriginalMessageID,
			&rev.Question,
			&rev.OriginalAnswer,
			&rev.RevisedAnswer,
			&rev.QuestionEmbedding,
			&rev.RevisionReason,
			&rev.RevisedBy,
			&rev.CreatedAt,
			&rev.UpdatedAt,
			&rev.IsActive,
			&similarity,
		)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to scan revision")
		}

		revWithEmbedding := rev.ToAnswerRevisionWithEmbedding()
		revWithEmbedding.Similarity = similarity
		revisions = append(revisions, revWithEmbedding)
	}

	return revisions, nil
}

// GetRevisionsByChat gets all revisions for a specific chatbot
func (r *RevisionRepository) GetRevisionsByChat(ctx context.Context, chatbotID uuid.UUID, includeInactive bool) ([]*AnswerRevision, error) {
	query := `
		SELECT
			id, chatbot_id, original_message_id, question,
			original_answer, revised_answer, question_embedding,
			revision_reason, revised_by, created_at, updated_at, is_active
		FROM answer_revisions
		WHERE chatbot_id = $1
	`

	if !includeInactive {
		query += " AND is_active = true"
	}

	query += " ORDER BY created_at DESC"

	var revisions []*AnswerRevision
	err := r.db.SelectContext(ctx, &revisions, query, chatbotID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get revisions by chatbot")
	}

	return revisions, nil
}

// GetRevisionByID gets a single revision by ID
func (r *RevisionRepository) GetRevisionByID(ctx context.Context, id uuid.UUID) (*AnswerRevision, error) {
	query := `
		SELECT
			id, chatbot_id, original_message_id, question,
			original_answer, revised_answer, question_embedding,
			revision_reason, revised_by, created_at, updated_at, is_active
		FROM answer_revisions
		WHERE id = $1
	`

	var revision AnswerRevision
	err := r.db.GetContext(ctx, &revision, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.Wrap(apperrors.ErrNotFound, "revision not found")
		}
		return nil, apperrors.Wrap(err, "failed to get revision by ID")
	}

	return &revision, nil
}

// UpdateRevision updates an existing revision
func (r *RevisionRepository) UpdateRevision(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic update query
	setClauses := []string{}
	args := []interface{}{}
	argCount := 1

	for field, value := range updates {
		// Whitelist allowed fields to prevent SQL injection
		switch field {
		case "question", "revised_answer", "revision_reason", "is_active", "question_embedding":
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argCount))
			args = append(args, value)
			argCount++
		}
	}

	if len(setClauses) == 0 {
		return apperrors.New("no valid fields to update")
	}

	// Always update the updated_at timestamp
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argCount))
	args = append(args, time.Now())
	argCount++

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE answer_revisions
		SET %s
		WHERE id = $%d
	`, joinStrings(setClauses, ", "), argCount)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return apperrors.Wrap(err, "failed to update revision")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "revision not found")
	}

	return nil
}

// DeactivateRevision sets a revision as inactive
func (r *RevisionRepository) DeactivateRevision(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE answer_revisions
		SET is_active = false
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to deactivate revision")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "revision not found")
	}

	return nil
}

// GetConversationsWithMessages gets all conversations (sessions) with their messages for a chatbot
func (r *RevisionRepository) GetConversationsWithMessages(ctx context.Context, chatbotID uuid.UUID, limit int, offset int) ([]map[string]interface{}, error) {
	// Get unique sessions with their latest message
	query := `
		WITH sessions AS (
			SELECT DISTINCT
				session_id,
				MAX(created_at) as last_message_at
			FROM chat_messages
			WHERE chatbot_id = $1
			GROUP BY session_id
			ORDER BY last_message_at DESC
			LIMIT $2 OFFSET $3
		)
		SELECT
			cm.id,
			cm.chatbot_id,
			cm.session_id,
			cm.role,
			cm.content,
			cm.created_at
		FROM chat_messages cm
		INNER JOIN sessions s ON cm.session_id = s.session_id
		WHERE cm.chatbot_id = $1
		ORDER BY s.last_message_at DESC, cm.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, chatbotID, limit, offset)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get conversations")
	}
	defer rows.Close()

	// Group messages by session
	sessionMap := make(map[uuid.UUID][]map[string]interface{})
	sessionOrder := []uuid.UUID{}

	for rows.Next() {
		var msg ChatMessage
		err := rows.Scan(
			&msg.ID,
			&msg.ChatbotID,
			&msg.SessionID,
			&msg.Role,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to scan message")
		}

		// Track session order
		if _, exists := sessionMap[msg.SessionID]; !exists {
			sessionOrder = append(sessionOrder, msg.SessionID)
			sessionMap[msg.SessionID] = []map[string]interface{}{}
		}

		// Add message to session
		sessionMap[msg.SessionID] = append(sessionMap[msg.SessionID], map[string]interface{}{
			"id":         msg.ID,
			"chatbot_id": msg.ChatbotID,
			"role":       msg.Role,
			"content":    msg.Content,
			"created_at": msg.CreatedAt,
		})
	}

	// Build result maintaining order
	result := []map[string]interface{}{}
	for _, sessionID := range sessionOrder {
		messages := sessionMap[sessionID]
		if len(messages) > 0 {
			result = append(result, map[string]interface{}{
				"session_id": sessionID,
				"messages":   messages,
				"created_at": messages[0]["created_at"],
			})
		}
	}

	return result, nil
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
