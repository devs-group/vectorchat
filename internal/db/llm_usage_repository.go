package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// LLMUsageRepository persists LLM token usage records.
type LLMUsageRepository struct {
	db *Database
}

// NewLLMUsageRepository creates a repository instance.
func NewLLMUsageRepository(db *Database) *LLMUsageRepository {
	return &LLMUsageRepository{db: db}
}

// Create inserts a single usage row.
func (r *LLMUsageRepository) Create(ctx context.Context, usage *LLMUsage) error {
	if usage == nil {
		return apperrors.Wrap(apperrors.ErrInvalidUserData, "llm usage payload is nil")
	}

	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}
	if usage.CreatedAt.IsZero() {
		usage.CreatedAt = time.Now()
	}

	query := `INSERT INTO llm_usage (id, user_id, org_id, trace_id, model_alias, provider, prompt_tokens, completion_tokens, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query, usage.ID, usage.UserID, usage.OrgID, usage.TraceID, usage.ModelAlias, usage.Provider, usage.PromptTokens, usage.CompletionTokens, usage.CreatedAt)
	if err != nil {
		return apperrors.Wrap(err, "failed to insert llm usage")
	}

	return nil
}
