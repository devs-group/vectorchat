package services

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/yourusername/vectorchat/pkg/jobs"
	"github.com/yourusername/vectorchat/pkg/models"
)

// QueueMetricsService exposes lightweight JetStream stats for the crawl queue.
type QueueMetricsService struct {
	js nats.JetStreamContext
}

func NewQueueMetricsService(js nats.JetStreamContext) *QueueMetricsService {
	return &QueueMetricsService{js: js}
}

func (s *QueueMetricsService) GetCrawlMetrics(ctx context.Context) (*models.CrawlQueueMetrics, error) {
	if s == nil || s.js == nil {
		return nil, nats.ErrInvalidConnection
	}

	si, err := s.js.StreamInfo(jobs.CrawlStream, nats.Context(ctx))
	if err != nil {
		return nil, err
	}
	ci, err := s.js.ConsumerInfo(jobs.CrawlStream, "crawler-workers", nats.Context(ctx))
	if err != nil {
		return nil, err
	}

	var m models.CrawlQueueMetrics
	m.Stream.Pending = si.State.Msgs
	m.Stream.Bytes = si.State.Bytes
	m.Stream.Subjects = len(si.Config.Subjects)
	m.Consumer.NumPending = uint64(ci.NumPending)
	m.Consumer.NumAckPending = uint64(ci.NumAckPending)
	m.Consumer.NumWaiting = ci.NumWaiting
	m.Consumer.NumRedelivered = uint64(ci.NumRedelivered)

	return &m, nil
}
