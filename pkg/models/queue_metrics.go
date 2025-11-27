package models

// CrawlQueueMetrics captures basic JetStream queue health.
type CrawlQueueMetrics struct {
	Stream struct {
		Pending uint64 `json:"pending"`
		Bytes   uint64 `json:"bytes"`
		Subjects int   `json:"subjects"`
	} `json:"stream"`
	Consumer struct {
		NumPending   uint64 `json:"num_pending"`
		NumAckPending uint64 `json:"num_ack_pending"`
		NumWaiting   int    `json:"num_waiting"`
		NumRedelivered uint64 `json:"num_redelivered"`
	} `json:"consumer"`
}
