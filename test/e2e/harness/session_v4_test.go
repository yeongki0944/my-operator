package harness

import (
	"context"
	"testing"
	"time"

	"github.com/yeongki/my-operator/pkg/slo/fetch"
	"github.com/yeongki/my-operator/pkg/slo/spec"
)

type fakeFetcherV4 struct {
	samples []fetch.Sample
}

func (f *fakeFetcherV4) Fetch(_ context.Context, _ time.Time) (fetch.Sample, error) {
	sample := f.samples[0]
	f.samples = f.samples[1:]
	return sample, nil
}

func TestSessionV4BuildsSummary(t *testing.T) {
	start := time.Now().Add(-time.Minute)
	end := time.Now()
	fetcher := &fakeFetcherV4{
		samples: []fetch.Sample{
			{At: start, Values: map[string]float64{"metric": 1}},
			{At: end, Values: map[string]float64{"metric": 3}},
		},
	}

	session := NewSessionV4(SessionV4Config{
		Namespace:          "default",
		MetricsServiceName: "metrics",
		TestCase:           "case",
		Suite:              "auto-suite",
		RunID:              "run-1",
		Tags: map[string]string{
			"suite":  "user-suite",
			"run_id": "override-run",
		},
		Fetcher: fetcher,
		Specs: []spec.SLISpec{
			{
				ID:     "metric_delta",
				Inputs: []spec.MetricRef{spec.PromMetric("metric", nil)},
				Compute: spec.ComputeSpec{
					Mode: spec.ComputeDelta,
				},
			},
		},
	})

	session.Start()
	summary, err := session.End(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if summary == nil {
		t.Fatalf("expected summary, got nil")
	}
	if summary.SchemaVersion != "slo.v3" {
		t.Fatalf("expected schemaVersion slo.v3, got %q", summary.SchemaVersion)
	}
	if summary.Config.Format != "v4" {
		t.Fatalf("expected config.format v4, got %q", summary.Config.Format)
	}
	if summary.Config.Tags["suite"] != "user-suite" {
		t.Fatalf("expected user suite tag override, got %q", summary.Config.Tags["suite"])
	}
	if summary.Config.Tags["run_id"] != "override-run" {
		t.Fatalf("expected user run_id tag override, got %q", summary.Config.Tags["run_id"])
	}
}
