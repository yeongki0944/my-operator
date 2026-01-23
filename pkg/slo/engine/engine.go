package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/yeongki/my-operator/pkg/slo"
	"github.com/yeongki/my-operator/pkg/slo/fetch"
	"github.com/yeongki/my-operator/pkg/slo/spec"
	"github.com/yeongki/my-operator/pkg/slo/summary"
)

// Logger keeps pkg/slo independent from klog/logr.
//type Logger interface {
//	Logf(format string, args ...any)
//}

type Engine struct {
	fetcher fetch.MetricsFetcher
	//Spec  registry.Registry // (옵션) 레지스트리를 쓰는 호출자를 위해 남길 수 있음, 일단 주석처리함.
	//reg     *spec.Registry
	writer summary.Writer
	logf   func(string, ...any)
}

func New(fetcher fetch.MetricsFetcher, writer summary.Writer, l slo.Logger) *Engine {
	logf := func(string, ...any) {}
	if l != nil {
		logf = l.Logf
	}
	return &Engine{fetcher: fetcher, writer: writer, logf: logf}
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
		SchemaVersion: "slo.v3",
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
			Format:        cfg.Format,
			EvidencePaths: cfg.EvidencePaths,
		},
	}

	for _, s := range req.Specs {
		// specItem, ok := e.reg.Get(id)
		// if !ok {
		// 	sum.Warnings = append(sum.Warnings, fmt.Sprintf("unknown sli id: %s", id))
		// 	sum.Results = append(sum.Results, summary.SLIResult{
		// 		ID:     id,
		// 		Status: "skip",
		// 		Reason: "unknown sli id",
		// 	})
		// 	continue
		// }
		// r := evalSLI(specItem, start.Values, end.Values)
		r := evalSLI(s, start.Values, end.Values)
		sum.Results = append(sum.Results, r)
	}

	if err := e.writer.Write(req.OutPath, sum); err != nil {
		return nil, err
	}
	return &sum, nil
}

func (e *Engine) emptySummary(cfg RunConfig, warnings []string) *summary.Summary {
	return &summary.Summary{
		SchemaVersion: "slo.v3",
		GeneratedAt:   time.Now(),
		Config: summary.RunConfig{
			RunID:         cfg.RunID,
			StartedAt:     cfg.StartedAt,
			FinishedAt:    cfg.FinishedAt,
			Mode:          summary.RunMode{Location: cfg.Mode.Location, Trigger: cfg.Mode.Trigger},
			Tags:          cfg.Tags,
			Format:        cfg.Format,
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
		Status:      summary.StatusPass,
	}

	used := make([]string, 0, len(s.Inputs))
	missing := make([]string, 0)

	// v3: one-input SLI recommended. If multiple inputs exist, we sum them.
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
		res.Status = summary.StatusSkip
		res.Reason = "missing input metrics"
		return res
	}

	var value float64
	switch s.Compute.Mode {
	case spec.ComputeSingle:
		value = valStart
	case spec.ComputeDelta:
		value = valEnd - valStart
		if value < 0 {
			// v3: counter reset suspected (process restart)
			res.Value = &value
			res.Status = summary.StatusWarn
			res.Reason = "delta < 0 (counter reset suspected)"
			// judge가 있으면 judge 결과로 덮어써버리니까,
			// 이 경우 judge를 건너뛰는 정책을 택할지 결정해야 함.
			return res // judge skip
		}
	default:
		res.Status = summary.StatusSkip
		res.Reason = "unknown compute mode"
		return res
	}
	res.Value = &value

	if s.Judge != nil {
		res.Status, res.Reason = judge(value, s.Judge.Rules)
	}

	return res
}

func judge(v float64, rules []spec.Rule) (status summary.Status, reason string) {
	// v3: fail dominates warn
	var warn string
	for _, r := range rules {
		if !compare(v, r.Op, r.Target) {
			continue
		}
		switch r.Level {
		case spec.LevelFail:
			return summary.StatusFail, fmt.Sprintf("rule fail: value %s %v", r.Op, r.Target)
		case spec.LevelWarn:
			warn = fmt.Sprintf("rule warn: value %s %v", r.Op, r.Target)
		default:
			// TODO(v4): unknown level -> warn/skip?
		}
	}
	if warn != "" {
		return summary.StatusWarn, warn
	}
	return summary.StatusPass, ""
}

func compare(v float64, op spec.Op, target float64) bool {
	switch op {
	case spec.OpLE:
		return v <= target
	case spec.OpGE:
		return v >= target
	case spec.OpLT:
		return v < target
	case spec.OpGT:
		return v > target
	case spec.OpEQ:
		return v == target
	default:
		return false
	}
}
