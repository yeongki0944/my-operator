package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/yeongki/my-operator/pkg/slo/fetch"
	"github.com/yeongki/my-operator/pkg/slo/spec"
	"github.com/yeongki/my-operator/pkg/slo/summary"
)

// Logger keeps pkg/slo independent from klog/logr.
type Logger interface {
	Logf(format string, args ...any)
}

type Engine struct {
	fetcher fetch.MetricsFetcher
	reg     *spec.Registry
	writer  summary.Writer
	logf    func(string, ...any)
}

func New(fetcher fetch.MetricsFetcher, writer summary.Writer, reg *spec.Registry, l Logger) *Engine {
	logf := func(string, ...any) {}
	if l != nil {
		logf = l.Logf
	}
	return &Engine{fetcher: fetcher, writer: writer, reg: reg, logf: logf}
}

func (e *Engine) Execute(ctx context.Context, req ExecuteRequest) (*summary.Summary, error) {
	cfg := req.Config
	if cfg.StartedAt.IsZero() || cfg.FinishedAt.IsZero() {
		return nil, fmt.Errorf("StartedAt/FinishedAt must be set")
	}

	// Fetch snapshots
	start, err := e.fetcher.Fetch(ctx, cfg.StartedAt)
	if err != nil {
		// philosophy: "measurement failure is not test failure" → return a Summary with warnings
		s := e.emptySummary(cfg, []string{fmt.Sprintf("fetch(start) failed: %v", err)})
		_ = e.writer.Write(req.OutPath, *s)
		return s, nil
	}
	end, err := e.fetcher.Fetch(ctx, cfg.FinishedAt)
	if err != nil {
		s := e.emptySummary(cfg, []string{fmt.Sprintf("fetch(end) failed: %v", err)})
		_ = e.writer.Write(req.OutPath, *s)
		return s, nil
	}

	sum := summary.Summary{
		SchemaVersion: "slo.v1",
		GeneratedAt:   time.Now(),
		Config: summary.RunConfig{
			RunID:      cfg.RunID,
			StartedAt:  cfg.StartedAt,
			FinishedAt: cfg.FinishedAt,
			Mode: summary.RunMode{
				Location: cfg.Mode.Location,
				Trigger:  cfg.Mode.Trigger,
			},
			Tags:          cfg.Tags,
			EvidencePaths: cfg.EvidencePaths,
		},
	}

	for _, id := range req.SLIIDs {
		specItem, ok := e.reg.Get(id)
		if !ok {
			sum.Warnings = append(sum.Warnings, fmt.Sprintf("unknown sli id: %s", id))
			sum.Results = append(sum.Results, summary.SLIResult{
				ID:     id,
				Status: "skip",
				Reason: "unknown sli id",
			})
			continue
		}
		r := evalSLI(specItem, start.Values, end.Values)
		sum.Results = append(sum.Results, r)
	}

	if err := e.writer.Write(req.OutPath, sum); err != nil {
		return nil, err
	}
	return &sum, nil
}

func (e *Engine) emptySummary(cfg RunConfig, warnings []string) *summary.Summary {
	return &summary.Summary{
		SchemaVersion: "slo.v1",
		GeneratedAt:   time.Now(),
		Config: summary.RunConfig{
			RunID:         cfg.RunID,
			StartedAt:     cfg.StartedAt,
			FinishedAt:    cfg.FinishedAt,
			Mode:          summary.RunMode{Location: cfg.Mode.Location, Trigger: cfg.Mode.Trigger},
			Tags:          cfg.Tags,
			EvidencePaths: cfg.EvidencePaths,
		},
		Results:  []summary.SLIResult{},
		Warnings: warnings,
	}
}

func evalSLI(s spec.SLISpec, start, end map[string]float64) summary.SLIResult {
	res := summary.SLIResult{
		ID:          s.ID,
		Title:       s.Title,
		Unit:        s.Unit,
		Kind:        s.Kind,
		Description: s.Description,
		Status:      "ok",
	}

	used := make([]string, 0, len(s.Inputs))
	missing := make([]string, 0)

	// v1: one-input SLI recommended. If multiple inputs exist, we sum them.
	var valStart, valEnd float64
	for _, in := range s.Inputs {
		used = append(used, in.Key)
		a, okA := start[in.Key]
		b, okB := end[in.Key]
		if !okA || !okB {
			missing = append(missing, in.Key)
			continue
		}
		valStart += a
		valEnd += b
	}
	res.InputsUsed = used
	res.InputsMissing = missing

	if len(missing) > 0 {
		res.Status = "skip"
		res.Reason = "missing input metrics"
		return res
	}

	var value float64
	switch s.Compute.Mode {
	case spec.ComputeSingle:
		value = valStart
	case spec.ComputeDelta:
		value = valEnd - valStart
	default:
		res.Status = "skip"
		res.Reason = "unknown compute mode"
		return res
	}
	res.Value = &value

	// Apply judge rules (optional)
	if s.Judge != nil {
		status, reason := judge(value, s.Judge.Rules)
		if status != "" {
			res.Status = status
			res.Reason = reason
		}
	}

	return res
}

func judge(v float64, rules []spec.Rule) (status, reason string) {
	// v1: fail dominates warn
	var warn string
	for _, r := range rules {
		if compare(v, r.Op, r.Target) {
			if r.Level == "fail" {
				return "fail", fmt.Sprintf("rule fail: value %s %v", r.Op, r.Target)
			}
			if r.Level == "warn" {
				warn = fmt.Sprintf("rule warn: value %s %v", r.Op, r.Target)
			}
		}
	}
	if warn != "" {
		return "warn", warn
	}
	return "", ""
}

func compare(v float64, op string, target float64) bool {
	switch op {
	case "<=":
		return v <= target
	case ">=":
		return v >= target
	case "<":
		return v < target
	case ">":
		return v > target
	case "==":
		return v == target
	default:
		return false
	}
}
