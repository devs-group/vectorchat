package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// CrawlSchedule represents a recurring website crawl configuration bound to either
// a chatbot-specific knowledge base or a shared knowledge base.
type CrawlSchedule struct {
	ID                     uuid.UUID  `db:"id"`
	ChatbotID              *uuid.UUID `db:"chatbot_id"`
	SharedKnowledgeBaseID  *uuid.UUID `db:"shared_knowledge_base_id"`
	RootURL                string     `db:"root_url"`
	CronExpr               string     `db:"cron_expr"`
	Timezone               string     `db:"timezone"`
	Enabled                bool       `db:"enabled"`
	LastRunAt              *time.Time `db:"last_run_at"`
	NextRunAt              *time.Time `db:"next_run_at"`
	LastStatus             *string    `db:"last_status"`
	LastError              *string    `db:"last_error"`
	CreatedAt              time.Time  `db:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at"`
}

type CrawlScheduleRepository struct {
	db *Database
}

func NewCrawlScheduleRepository(db *Database) *CrawlScheduleRepository {
	return &CrawlScheduleRepository{db: db}
}

// Upsert inserts or updates a schedule for the given scope + root URL.
func (r *CrawlScheduleRepository) Upsert(ctx context.Context, s *CrawlSchedule) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now().UTC()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now

	var query string
	if s.ChatbotID != nil && (s.SharedKnowledgeBaseID == nil) {
		query = `
			INSERT INTO crawl_schedules (
				id, chatbot_id, shared_knowledge_base_id, root_url, cron_expr, timezone,
				enabled, last_run_at, next_run_at, last_status, last_error, created_at, updated_at
			) VALUES (
				:id, :chatbot_id, :shared_knowledge_base_id, :root_url, :cron_expr, :timezone,
				:enabled, :last_run_at, :next_run_at, :last_status, :last_error, :created_at, :updated_at
			)
			ON CONFLICT (chatbot_id, root_url) DO UPDATE
				SET cron_expr = EXCLUDED.cron_expr,
					timezone = EXCLUDED.timezone,
					enabled = EXCLUDED.enabled,
					next_run_at = EXCLUDED.next_run_at,
					last_status = EXCLUDED.last_status,
					last_error = EXCLUDED.last_error,
					updated_at = EXCLUDED.updated_at
			RETURNING id, chatbot_id, shared_knowledge_base_id, root_url, cron_expr, timezone, enabled,
			          last_run_at, next_run_at, last_status, last_error, created_at, updated_at
		`
	} else {
		query = `
			INSERT INTO crawl_schedules (
				id, chatbot_id, shared_knowledge_base_id, root_url, cron_expr, timezone,
				enabled, last_run_at, next_run_at, last_status, last_error, created_at, updated_at
			) VALUES (
				:id, :chatbot_id, :shared_knowledge_base_id, :root_url, :cron_expr, :timezone,
				:enabled, :last_run_at, :next_run_at, :last_status, :last_error, :created_at, :updated_at
			)
			ON CONFLICT (shared_knowledge_base_id, root_url) DO UPDATE
				SET cron_expr = EXCLUDED.cron_expr,
					timezone = EXCLUDED.timezone,
					enabled = EXCLUDED.enabled,
					next_run_at = EXCLUDED.next_run_at,
					last_status = EXCLUDED.last_status,
					last_error = EXCLUDED.last_error,
					updated_at = EXCLUDED.updated_at
			RETURNING id, chatbot_id, shared_knowledge_base_id, root_url, cron_expr, timezone, enabled,
			          last_run_at, next_run_at, last_status, last_error, created_at, updated_at
		`
	}

	var stored CrawlSchedule
	rows, err := r.db.NamedQueryContext(ctx, query, s)
	if err != nil {
		return apperrors.Wrap(err, "failed to upsert crawl schedule")
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(&stored); err != nil {
			return apperrors.Wrap(err, "failed to scan crawl schedule")
		}
		*s = stored
	}
	return nil
}

// FindByScope returns the schedule for a chatbot or shared knowledge base + URL.
func (r *CrawlScheduleRepository) FindByScope(ctx context.Context, chatbotID *uuid.UUID, sharedID *uuid.UUID, rootURL string) (*CrawlSchedule, error) {
	query := `
		SELECT * FROM crawl_schedules
		WHERE root_url = $1
		  AND (
		      (chatbot_id = $2 AND $2 IS NOT NULL)
		   OR (shared_knowledge_base_id = $3 AND $3 IS NOT NULL)
		  )
	`
	var result CrawlSchedule
	if err := r.db.GetContext(ctx, &result, query, rootURL, chatbotID, sharedID); err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find crawl schedule")
	}
	return &result, nil
}

// FindByID returns a schedule by its identifier.
func (r *CrawlScheduleRepository) FindByID(ctx context.Context, id uuid.UUID) (*CrawlSchedule, error) {
	var result CrawlSchedule
	if err := r.db.GetContext(ctx, &result, `SELECT * FROM crawl_schedules WHERE id = $1`, id); err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find crawl schedule by id")
	}
	return &result, nil
}

// ListByScope returns all schedules for the given chatbot or shared knowledge base.
func (r *CrawlScheduleRepository) ListByScope(ctx context.Context, chatbotID *uuid.UUID, sharedID *uuid.UUID) ([]*CrawlSchedule, error) {
	query := `
		SELECT * FROM crawl_schedules
		WHERE (chatbot_id = $1 AND $1 IS NOT NULL)
		   OR (shared_knowledge_base_id = $2 AND $2 IS NOT NULL)
		ORDER BY root_url
	`
	var results []*CrawlSchedule
	if err := r.db.SelectContext(ctx, &results, query, chatbotID, sharedID); err != nil {
		return nil, apperrors.Wrap(err, "failed to list crawl schedules")
	}
	return results, nil
}

// ListActive returns enabled schedules with their next run time populated.
func (r *CrawlScheduleRepository) ListActive(ctx context.Context) ([]*CrawlSchedule, error) {
	query := `
		SELECT * FROM crawl_schedules
		WHERE enabled = TRUE
	`
	var results []*CrawlSchedule
	if err := r.db.SelectContext(ctx, &results, query); err != nil {
		return nil, apperrors.Wrap(err, "failed to list active crawl schedules")
	}
	return results, nil
}

// UpdateRunInfo updates execution metadata after a run attempt.
func (r *CrawlScheduleRepository) UpdateRunInfo(ctx context.Context, id uuid.UUID, lastRun, nextRun *time.Time, status, errMsg *string) error {
	query := `
		UPDATE crawl_schedules
		SET last_run_at = $1,
		    next_run_at = $2,
		    last_status = $3,
		    last_error = $4,
		    updated_at = NOW()
		WHERE id = $5
	`
	if _, err := r.db.ExecContext(ctx, query, lastRun, nextRun, status, errMsg, id); err != nil {
		return apperrors.Wrap(err, "failed to update crawl schedule run info")
	}
	return nil
}

// Delete removes a schedule by ID.
func (r *CrawlScheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM crawl_schedules WHERE id = $1`, id); err != nil {
		return apperrors.Wrap(err, "failed to delete crawl schedule")
	}
	return nil
}
