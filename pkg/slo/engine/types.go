package engine

import "time"

type RunMode struct {
	Location string // "inside" | "outside"
	Trigger  string // "none" | "annotation"
}

type RunConfig struct {
	RunID      string
	StartedAt  time.Time
	FinishedAt time.Time
	Mode       RunMode

	Tags          map[string]string
	EvidencePaths map[string]string
}

type ExecuteRequest struct {
	Config  RunConfig
	SLIIDs  []string
	OutPath string
}
