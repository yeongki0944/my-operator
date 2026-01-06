package slo

import (
	"time"
)

type SummaryWriter interface {
	WriteSummary(summary Summary) error
}

// SummaryMetrics TODO 추가적으로 확장해감.
type SummaryMetrics struct {
	E2EConvergenceTimeSeconds *float64 `json:"e2e_convergence_time_seconds,omitempty"`
	ReconcileTotalDelta       *float64 `json:"reconcile_total_delta,omitempty"`
}

type Summary struct {
	Labels    Labels         `json:"labels"`
	CreatedAt time.Time      `json:"created_at"`
	Metrics   SummaryMetrics `json:"metrics"`
}

// WriteSummary writes summary JSON to path. Creates parent dirs.
// This should never fail the test flow by design; caller can log+ignore errors.
/*func WriteSummary(path string, s Summary) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&s); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}*/

/*
internal/artifacts/json_writer.go

type JSONFileWriter struct {
	Path string
}
// TODO 문서 WriteFileAtomically 참고.
func (w JSONFileWriter) WriteSummary(s slo.Summary) error {
	if err := os.MkdirAll(filepath.Dir(w.Path), 0o755); err != nil {
		return err
	}

	tmp := w.Path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, w.Path)
}
// 이렇게 사용할 수 있음.
w := artifacts.JSONFileWriter{Path: "artifacts/result.json"}
r := slo.NewRecorder(true, testLogger, w)
_ = r.RecordAndSave(labels, d)

*/
