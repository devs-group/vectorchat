package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/robfig/cron/v3"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/pkg/jobs"
	"github.com/yourusername/vectorchat/pkg/models"
)

// CrawlScheduleService manages persistence and validation of crawl schedules for both chatbot and shared knowledge bases.
type CrawlScheduleService struct {
	repo       *db.CrawlScheduleRepository
	kbService  *KnowledgeBaseService
	timeNowUTC func() time.Time
	js         nats.JetStreamContext
}

func NewCrawlScheduleService(repo *db.CrawlScheduleRepository, kbService *KnowledgeBaseService, js nats.JetStreamContext) *CrawlScheduleService {
	return &CrawlScheduleService{
		repo:       repo,
		kbService:  kbService,
		timeNowUTC: func() time.Time { return time.Now().UTC() },
		js:         js,
	}
}

// Upsert creates or updates a schedule for the given target and URL.
func (s *CrawlScheduleService) Upsert(ctx context.Context, target KnowledgeBaseTarget, req *models.CrawlScheduleRequest) (*models.CrawlScheduleResponse, error) {
	if err := target.validate(); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "request body is required")
	}
	root := strings.TrimSpace(req.URL)
	if root == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "url is required")
	}
	if _, err := s.kbService.ParseURL(root); err != nil {
		return nil, apperrors.Wrap(err, "invalid url")
	}

	cronExpr := strings.TrimSpace(req.CronExpr)
	if cronExpr == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "cron expression is required")
	}
	if err := validateCron(cronExpr); err != nil {
		return nil, err
	}
	tz := strings.TrimSpace(req.Timezone)
	if tz == "" {
		tz = "UTC"
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return nil, apperrors.Wrap(err, "invalid timezone")
	}

	schedule := &db.CrawlSchedule{
		ChatbotID:             target.ChatbotID,
		SharedKnowledgeBaseID: target.SharedKnowledgeBaseID,
		RootURL:               root,
		CronExpr:              cronExpr,
		Timezone:              tz,
		Enabled:               req.Enabled,
	}

	if err := s.repo.Upsert(ctx, schedule); err != nil {
		return nil, err
	}

	// Immediately enqueue an initial crawl through the queue if JetStream is configured.
	if s.js != nil {
		if err := s.enqueueJob(ctx, schedule); err != nil {
			// log only; don't fail the API to avoid blocking users
			return toCrawlScheduleResponse(schedule), nil
		}
	}

	return toCrawlScheduleResponse(schedule), nil
}

// EnqueueOnce pushes a one-off crawl job (no schedule persisted).
func (s *CrawlScheduleService) EnqueueOnce(ctx context.Context, target KnowledgeBaseTarget, url string) error {
	if err := target.validate(); err != nil {
		return err
	}
	root := strings.TrimSpace(url)
	if root == "" {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "url is required")
	}
	if _, err := s.kbService.ParseURL(root); err != nil {
		return apperrors.Wrap(err, "invalid url")
	}
	if s.js == nil {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "queue unavailable")
	}
	payload := jobs.CrawlJobPayload{
		JobID:                 uuid.New(),
		ScheduleID:            uuid.Nil,
		RootURL:               root,
		RequestedAt:           s.timeNowUTC(),
		ChatbotID:             target.ChatbotID,
		SharedKnowledgeBaseID: target.SharedKnowledgeBaseID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = s.js.Publish(jobs.CrawlSubject, body)
	return err
}

func (s *CrawlScheduleService) enqueueJob(ctx context.Context, sched *db.CrawlSchedule) error {
	payload := jobs.CrawlJobPayload{
		JobID:                 uuid.New(),
		ScheduleID:            sched.ID,
		RootURL:               sched.RootURL,
		RequestedAt:           s.timeNowUTC(),
		ChatbotID:             sched.ChatbotID,
		SharedKnowledgeBaseID: sched.SharedKnowledgeBaseID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = s.js.Publish(jobs.CrawlSubject, body)
	if err == nil {
		slog.Info("crawl job enqueued", "schedule_id", sched.ID, "url", sched.RootURL)
	}
	return err
}

// validateCron ensures the expression is acceptable by robfig/cron and not empty.
func validateCron(expr string) error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	if _, err := parser.Parse(expr); err != nil {
		return apperrors.Wrap(err, "invalid cron expression")
	}
	return nil
}

// List returns all schedules for the given target.
func (s *CrawlScheduleService) List(ctx context.Context, target KnowledgeBaseTarget) (*models.CrawlScheduleListResponse, error) {
	if err := target.validate(); err != nil {
		return nil, err
	}

	items, err := s.repo.ListByScope(ctx, target.ChatbotID, target.SharedKnowledgeBaseID)
	if err != nil {
		return nil, err
	}

	resp := models.CrawlScheduleListResponse{Schedules: make([]models.CrawlScheduleResponse, 0, len(items))}
	for _, it := range items {
		resp.Schedules = append(resp.Schedules, *toCrawlScheduleResponse(it))
	}
	return &resp, nil
}

// Delete removes a schedule, ensuring it belongs to the provided target.
func (s *CrawlScheduleService) Delete(ctx context.Context, target KnowledgeBaseTarget, id uuid.UUID) error {
	if err := target.validate(); err != nil {
		return err
	}
	if id == uuid.Nil {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "schedule id is required")
	}

	sched, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if (target.ChatbotID != nil && sched.ChatbotID != nil && *target.ChatbotID == *sched.ChatbotID) ||
		(target.SharedKnowledgeBaseID != nil && sched.SharedKnowledgeBaseID != nil && *target.SharedKnowledgeBaseID == *sched.SharedKnowledgeBaseID) {
		return s.repo.Delete(ctx, id)
	}

	return apperrors.ErrUnauthorizedKnowledgeBaseAccess
}

func toCrawlScheduleResponse(sched *db.CrawlSchedule) *models.CrawlScheduleResponse {
	if sched == nil {
		return nil
	}
	return &models.CrawlScheduleResponse{
		ID:                    sched.ID,
		URL:                   sched.RootURL,
		CronExpr:              sched.CronExpr,
		Timezone:              sched.Timezone,
		Enabled:               sched.Enabled,
		LastRunAt:             sched.LastRunAt,
		NextRunAt:             sched.NextRunAt,
		LastStatus:            sched.LastStatus,
		LastError:             sched.LastError,
		ChatbotID:             sched.ChatbotID,
		SharedKnowledgeBaseID: sched.SharedKnowledgeBaseID,
		CreatedAt:             sched.CreatedAt,
		UpdatedAt:             sched.UpdatedAt,
	}
}
