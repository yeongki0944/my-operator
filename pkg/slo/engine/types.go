package engine

import (
	"time"

	"github.com/yeongki/my-operator/pkg/slo/spec"
)

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
	Format        string
	EvidencePaths map[string]string
}

type ExecuteRequest struct {
	Config  RunConfig
	Specs   []spec.SLISpec // core input: 직접 주입
	OutPath string
	// 호환성/편의용: 레지스트리를 쓰는 호출자를 위해 남길 수 있음, 일단 주석처리함.
	// SLIIDs  []string
}
