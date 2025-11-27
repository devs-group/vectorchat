package queue

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/yourusername/vectorchat/pkg/jobs"
)

// Connect creates a NATS connection with sensible defaults (reconnect, timeouts).
func Connect(url, user, pass string) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name("vectorchat-crawler"),
		nats.Timeout(5 * time.Second),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2 * time.Second),
	}
	if user != "" {
		opts = append(opts, nats.UserInfo(user, pass))
	}
	return nats.Connect(url, opts...)
}

// EnsureStreams declares the crawl job stream and DLQ if they do not exist.
func EnsureStreams(js nats.JetStreamContext) error {
	// main stream
	_, err := js.StreamInfo(jobs.CrawlStream)
	if err == nats.ErrStreamNotFound {
		if _, err = js.AddStream(&nats.StreamConfig{
			Name:      jobs.CrawlStream,
			Subjects:  []string{jobs.CrawlSubject},
			Retention: nats.LimitsPolicy,
			Storage:   nats.FileStorage,
			MaxBytes:  512 * 1024 * 1024, // 512MB
			MaxMsgs:   -1,
			MaxAge:    0,
		}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// DLQ
	if _, err = js.StreamInfo(jobs.CrawlStream + ".DLQ"); err == nats.ErrStreamNotFound {
		if _, err = js.AddStream(&nats.StreamConfig{
			Name:      jobs.CrawlStream + ".DLQ",
			Subjects:  []string{jobs.CrawlDLQSubject},
			Retention: nats.LimitsPolicy,
			Storage:   nats.FileStorage,
			MaxBytes:  128 * 1024 * 1024,
		}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
