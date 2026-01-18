package spec

// MetricRef identifies a metric input to an SLI.
// v1: simplest form uses a canonical Prometheus "text key" string.
// Example: controller_runtime_reconcile_total{result="success"}
type MetricRef struct {
	Key   string
	Alias string // optional
}

func PromKey(key string) MetricRef { return MetricRef{Key: key} }

const (
	ComputeSingle = "single" // use start snapshot only
	ComputeDelta  = "delta"  // end - start
)

// ComputeSpec describes how to compute the SLI.
type ComputeSpec struct {
	Mode string
}

// Rule is a tiny evaluation rule for v1.
type Rule struct {
	Metric string  // usually "value" for v1
	Op     string  // "<=", ">=", "<", ">", "=="
	Target float64 // threshold
	Level  string  // "warn" | "fail"
}

type JudgeSpec struct {
	Rules []Rule
}

// SLISpec is a declarative SLI definition.
// It is intentionally small in v1.
type SLISpec struct {
	ID          string
	Title       string
	Unit        string
	Kind        string // "delta_counter" | "gauge" | "derived" (v1 minimal)
	Description string

	Inputs  []MetricRef
	Compute ComputeSpec
	Judge   *JudgeSpec
}
