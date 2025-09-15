package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ChatMessageRepository handles database operations for chat messages.
type ChatMessageRepository struct {
	db *Database
}

// NewChatMessageRepository creates a new ChatMessageRepository.
func NewChatMessageRepository(db *Database) *ChatMessageRepository {
	return &ChatMessageRepository{db: db}
}

// Create saves a new chat message to the database.
func (r *ChatMessageRepository) Create(ctx context.Context, message *ChatMessage) error {
	query := `INSERT INTO chat_messages (id, chatbot_id, session_id, role, content, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, message.ID, message.ChatbotID, message.SessionID, message.Role, message.Content, message.CreatedAt)
	return err
}

// FindLastBySessionID retrieves the most recent messages for a given session ID.
func (r *ChatMessageRepository) FindLastBySessionID(ctx context.Context, sessionID uuid.UUID, limit int) ([]*ChatMessage, error) {
	var messages []*ChatMessage
	query := `SELECT id, chatbot_id, session_id, role, content, created_at
		 FROM chat_messages
		 WHERE session_id = $1
		 ORDER BY created_at ASC
		 LIMIT $2`

	err := r.db.SelectContext(ctx, &messages, query, sessionID, limit)
	if err != nil {
		return nil, err
	}

	// Reverse the slice to have messages in chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// FindRecentBySessionID retrieves the most recent messages for a given session ID.
func (r *ChatMessageRepository) FindRecentBySessionID(ctx context.Context, sessionID uuid.UUID, limit int) ([]*ChatMessage, error) {
	var messages []*ChatMessage
	query := `SELECT id, chatbot_id, session_id, role, content, created_at
		 FROM chat_messages
		 WHERE session_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`

	err := r.db.SelectContext(ctx, &messages, query, sessionID, limit)
	if err != nil {
		return nil, err
	}

	// Reverse the slice to have messages in chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// FindAllBySessionID retrieves all messages for a given session ID ordered chronologically.
func (r *ChatMessageRepository) FindAllBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*ChatMessage, error) {
	var messages []*ChatMessage
	query := `SELECT id, chatbot_id, session_id, role, content, created_at
         FROM chat_messages
         WHERE session_id = $1
         ORDER BY created_at ASC`

	err := r.db.SelectContext(ctx, &messages, query, sessionID)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// CountAssistantMessagesByChatbotID returns the number of assistant messages for a chatbot.
func (r *ChatMessageRepository) CountAssistantMessagesByChatbotID(ctx context.Context, chatbotID uuid.UUID) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM chat_messages WHERE chatbot_id = $1 AND role = 'assistant'`
	err := r.db.GetContext(ctx, &count, query, chatbotID)
	return count, err
}

// CountAssistantMessagesByChatbotIDs returns assistant message counts for multiple chatbots.
func (r *ChatMessageRepository) CountAssistantMessagesByChatbotIDs(ctx context.Context, chatbotIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	counts := make(map[uuid.UUID]int64, len(chatbotIDs))
	if len(chatbotIDs) == 0 {
		return counts, nil
	}

	ids := make([]string, len(chatbotIDs))
	for i, id := range chatbotIDs {
		ids[i] = id.String()
	}

	type row struct {
		ChatbotID uuid.UUID `db:"chatbot_id"`
		Count     int64     `db:"count"`
	}

	query := `
		SELECT chatbot_id, COUNT(*) AS count
		FROM chat_messages
		WHERE chatbot_id = ANY($1::uuid[])
		  AND role = 'assistant'
		GROUP BY chatbot_id
	`

	var results []row
	if err := r.db.SelectContext(ctx, &results, query, pq.Array(ids)); err != nil {
		return nil, err
	}

	for _, result := range results {
		counts[result.ChatbotID] = result.Count
	}

	return counts, nil
}
