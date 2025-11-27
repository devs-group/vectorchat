package jobs

import (
	"time"

	"github.com/google/uuid"
)

const (
	// CrawlSubject is the JetStream subject used for crawl jobs.
	CrawlSubject = "crawl.schedule.triggered"
	// CrawlDLQSubject is the subject used for dead-lettered crawl jobs.
	CrawlDLQSubject = "crawl.schedule.dlq"
	// CrawlStream is the JetStream stream name backing crawl jobs.
	CrawlStream = "CrawlJobs"
)

// CrawlJobPayload defines the message enqueued by the scheduler and consumed by workers.
type CrawlJobPayload struct {
	JobID                  uuid.UUID  `json:"job_id"`
	ScheduleID             uuid.UUID  `json:"schedule_id"`
	RootURL                string     `json:"root_url"`
	RequestedAt            time.Time  `json:"requested_at"`
	TraceID                string     `json:"trace_id,omitempty"`
	ChatbotID              *uuid.UUID `json:"chatbot_id,omitempty"`
	SharedKnowledgeBaseID  *uuid.UUID `json:"shared_knowledge_base_id,omitempty"`
}
