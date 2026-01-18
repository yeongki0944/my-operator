package summary

import "time"

// Summary is the contract output. All measurement methods must converge to this schema.
type Summary struct {
	SchemaVersion string    `json:"schemaVersion"`
	GeneratedAt   time.Time `json:"generatedAt"`

	Config RunConfig `json:"config"`

	Results  []SLIResult `json:"results"`
	Warnings []string    `json:"warnings,omitempty"`
}

// RunConfig is embedded in the summary (so analysis tools can be method-agnostic).
type RunConfig struct {
	RunID      string            `json:"runId,omitempty"`
	StartedAt  time.Time         `json:"startedAt"`
	FinishedAt time.Time         `json:"finishedAt"`
	Mode       RunMode           `json:"mode"`
	Tags       map[string]string `json:"tags,omitempty"`

	// EvidencePaths points to raw artifacts (optional).
	EvidencePaths map[string]string `json:"evidencePaths,omitempty"`
}

type RunMode struct {
	Location string `json:"location"` // "inside" | "outside"
	Trigger  string `json:"trigger"`  // "none" | "annotation"
}

type SLIResult struct {
	ID          string `json:"id"`
	Title       string `json:"title,omitempty"`
	Unit        string `json:"unit,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Description string `json:"description,omitempty"`

	// v1: a single numeric result. Future: Fields for p50/p99 etc.
	Value  *float64           `json:"value,omitempty"`
	Fields map[string]float64 `json:"fields,omitempty"`

	Status string `json:"status"` // "ok" | "warn" | "fail" | "skip"
	Reason string `json:"reason,omitempty"`

	InputsUsed    []string `json:"inputsUsed,omitempty"`
	InputsMissing []string `json:"inputsMissing,omitempty"`
}
