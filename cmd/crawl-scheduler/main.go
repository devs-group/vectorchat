package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/yourusername/vectorchat/internal/db"
	"github.com/yourusername/vectorchat/internal/queue"
	"github.com/yourusername/vectorchat/pkg/config"
	"github.com/yourusername/vectorchat/pkg/jobs"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var cfg config.AppConfig
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("scheduler: failed to load config: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	dbConn, err := db.NewDatabase(cfg.PGConnection)
	if err != nil {
		logger.Error("failed to connect database", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	nc, err := queue.Connect(cfg.NATSURL, cfg.NATSUsername, cfg.NATSPassword)
	if err != nil {
		logger.Error("failed to connect nats", "error", err)
		os.Exit(1)
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		logger.Error("failed to init jetstream", "error", err)
		os.Exit(1)
	}
	if err := queue.EnsureStreams(js); err != nil {
		logger.Warn("failed to ensure streams; will continue and expect existing stream", "error", err, "stream", jobs.CrawlStream)
	}

	repo := db.NewCrawlScheduleRepository(dbConn)
	manager := newScheduleManager(repo, js, logger)
	if err := manager.loadAndSchedule(ctx); err != nil {
		logger.Error("failed to register schedules", "error", err)
	}

	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			manager.stopAll()
			logger.Info("scheduler shutting down")
			return
		case <-ticker.C:
			if err := manager.loadAndSchedule(ctx); err != nil {
				logger.Warn("periodic reload failed", "error", err)
			}
		}
	}
}

type scheduleManager struct {
	repo   *db.CrawlScheduleRepository
	js     nats.JetStreamContext
	logger *slog.Logger
	// keyed by timezone
	runners map[string]gocron.Scheduler
}

func newScheduleManager(repo *db.CrawlScheduleRepository, js nats.JetStreamContext, logger *slog.Logger) *scheduleManager {
	return &scheduleManager{
		repo:    repo,
		js:      js,
		logger:  logger,
		runners: make(map[string]gocron.Scheduler),
	}
}

func (m *scheduleManager) stopAll() {
	for _, sched := range m.runners {
		_ = sched.Shutdown()
	}
	m.runners = make(map[string]gocron.Scheduler)
}

func (m *scheduleManager) loadAndSchedule(ctx context.Context) error {
	schedules, err := m.repo.ListActive(ctx)
	if err != nil {
		return err
	}

	// reset schedulers
	m.stopAll()

	// group by timezone and register
	for _, s := range schedules {
		loc := time.UTC
		if s.Timezone != "" {
			if l, err := time.LoadLocation(s.Timezone); err == nil {
				loc = l
			}
		}
		tzKey := loc.String()
		sched, ok := m.runners[tzKey]
		if !ok {
			sched, err = gocron.NewScheduler(gocron.WithLocation(loc))
			if err != nil {
				return err
			}
			m.runners[tzKey] = sched
		}

		// capture copy
		scheduleCopy := *s
		_, err := sched.NewJob(
			gocron.CronJob(scheduleCopy.CronExpr, false),
			gocron.NewTask(func() {
				m.enqueue(scheduleCopy)
			}),
		)
		if err != nil {
			m.logger.Warn("failed to add job", "schedule_id", scheduleCopy.ID, "error", err)
			continue
		}
	}

	for _, sched := range m.runners {
		sched.Start()
	}

	m.logger.Info("schedules loaded", "count", len(schedules))
	return nil
}

func (m *scheduleManager) enqueue(schedule db.CrawlSchedule) {
	now := time.Now().UTC()
	payload := jobs.CrawlJobPayload{
		JobID:                 uuid.New(),
		ScheduleID:            schedule.ID,
		RootURL:               schedule.RootURL,
		RequestedAt:           now,
		ChatbotID:             schedule.ChatbotID,
		SharedKnowledgeBaseID: schedule.SharedKnowledgeBaseID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		m.logger.Error("failed to marshal payload", "schedule_id", schedule.ID, "error", err)
		return
	}

	if _, err := m.js.Publish(jobs.CrawlSubject, body); err != nil {
		msg := err.Error()
		m.repo.UpdateRunInfo(context.Background(), schedule.ID, &now, nil, strPtr("publish_error"), &msg)
		m.logger.Error("failed to publish job", "schedule_id", schedule.ID, "error", err)
		return
	}

	status := "enqueued"
	_ = m.repo.UpdateRunInfo(context.Background(), schedule.ID, &now, nil, &status, nil)
	m.logger.Info("enqueued crawl job", "schedule_id", schedule.ID, "url", schedule.RootURL, "target", targetLabel(schedule))
}

func targetLabel(s db.CrawlSchedule) string {
	if s.ChatbotID != nil {
		return "chatbot:" + s.ChatbotID.String()
	}
	if s.SharedKnowledgeBaseID != nil {
		return "shared:" + s.SharedKnowledgeBaseID.String()
	}
	return "unknown"
}

func strPtr(s string) *string { return &s }
