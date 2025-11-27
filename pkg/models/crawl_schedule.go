package models

import (
	"time"

	"github.com/google/uuid"
)

// CrawlScheduleRequest defines the payload for creating or updating a crawl schedule.
type CrawlScheduleRequest struct {
	URL      string `json:"url" example:"https://docs.example.com"`
	CronExpr string `json:"cron_expr" example:"0 3 * * *"`          // standard 5-field cron
	Timezone string `json:"timezone" example:"America/New_York"`    // IANA timezone
	Enabled  bool   `json:"enabled" example:"true"`                 // whether the schedule is active
}

// CrawlScheduleResponse represents a saved schedule.
type CrawlScheduleResponse struct {
	ID                     uuid.UUID  `json:"id"`
	URL                    string     `json:"url"`
	CronExpr               string     `json:"cron_expr"`
	Timezone               string     `json:"timezone"`
	Enabled                bool       `json:"enabled"`
	LastRunAt              *time.Time `json:"last_run_at,omitempty"`
	NextRunAt              *time.Time `json:"next_run_at,omitempty"`
	LastStatus             *string    `json:"last_status,omitempty"`
	LastError              *string    `json:"last_error,omitempty"`
	ChatbotID              *uuid.UUID `json:"chatbot_id,omitempty"`
	SharedKnowledgeBaseID  *uuid.UUID `json:"shared_knowledge_base_id,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// CrawlScheduleListResponse wraps multiple schedules.
type CrawlScheduleListResponse struct {
	Schedules []CrawlScheduleResponse `json:"schedules"`
}
