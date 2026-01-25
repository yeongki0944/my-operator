package spec

import (
	"fmt"
	"strings"

	"github.com/yeongki/my-operator/pkg/slo/common/promkey"
)

// MetricRef identifies a metric input to an SLI.
// v3: simplest form uses a canonical Prometheus "text key" string.
// Example: controller_runtime_reconcile_total{result="success"}
type MetricRef struct {
	Key   string
	Alias string // optional
}

func UnsafePromKey(key string) MetricRef { return MetricRef{Key: key} }

type Labels map[string]string

func PromMetric(name string, labels Labels) MetricRef {
	return MetricRef{Key: promkey.Format(name, map[string]string(labels))}
}

// TODO spec_v4 참고 향후 통합될 예정임.
type ComputeMode string

const (
	ComputeSingle ComputeMode = "single" // use start snapshot only
	ComputeDelta  ComputeMode = "delta"  // end - start
)

// ComputeSpec describes how to compute the SLI.
type ComputeSpec struct {
	Mode ComputeMode
}

type Level string

const (
	LevelWarn Level = "warn"
	LevelFail Level = "fail"
)

type Op string

const (
	OpLE Op = "<="
	OpGE Op = ">="
	OpLT Op = "<"
	OpGT Op = ">"
	OpEQ Op = "=="
)

func (o *Op) UnmarshalText(text []byte) error {
	op, ok := NormalizeOp(string(text))
	if !ok {
		return fmt.Errorf("invalid op: %q", string(text))
	}
	*o = op
	return nil
}

// Rule is a tiny evaluation rule for v3.
type Rule struct {
	Metric string  // usually "value" for v3
	Op     Op      // OpLE/OpGE/...
	Target float64 // threshold
	Level  Level   // LevelWarn | LevelFail
}

type JudgeSpec struct {
	Rules []Rule
}

// SLISpec is a declarative SLI definition.
// It is intentionally small in v3.
type SLISpec struct {
	ID          string
	Title       string
	Unit        string
	Kind        string // "delta_counter" | "gauge" | "derived" (v3 minimal)
	Description string

	Inputs  []MetricRef
	Compute ComputeSpec
	Judge   *JudgeSpec
}

func NormalizeOp(s string) (Op, bool) {
	t := strings.TrimSpace(strings.ToLower(s))

	switch t {
	case "<=", "=<":
		return OpLE, true
	case ">=", "=>":
		return OpGE, true
	case "<":
		return OpLT, true
	case ">":
		return OpGT, true
	case "==", "=":
		return OpEQ, true

	// 선택: 사람 친화 별칭
	case "le", "lte":
		return OpLE, true
	case "ge", "gte":
		return OpGE, true
	case "lt":
		return OpLT, true
	case "gt":
		return OpGT, true
	case "eq":
		return OpEQ, true
	// TODO(v4): normalize Unicode operators (≤ ≥), 유니코드로 사용할 수 있어서 향후 업데트할때 넣자.
	default:
		return "", false
	}
}
