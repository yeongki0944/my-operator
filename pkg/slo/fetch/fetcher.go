package fetch

import (
	"context"
	"time"
)

// Sample is one snapshot at a point in time.
type Sample struct {
	At     time.Time
	Values map[string]float64 // metricKey -> value
}

// MetricsFetcher fetches one snapshot. Implementations decide how to obtain it.
// - outside: curl /metrics (via Pod, port-forward, HTTP)
// - inside: direct HTTP to localhost
// - trigger: could fetch from /metrics or status/log-derived metrics
type MetricsFetcher interface {
	Fetch(ctx context.Context, at time.Time) (Sample, error)
}
