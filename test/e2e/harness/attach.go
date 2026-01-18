package harness

import (
	"context"
	"time"

	"github.com/yeongki/my-operator/pkg/slo/engine"
	"github.com/yeongki/my-operator/pkg/slo/spec"
	"github.com/yeongki/my-operator/pkg/slo/summary"
)

// Attach is test glue: it owns timing and best-effort behavior.
// It must NOT panic the test on measurement failures.
type Attach struct {
	eng *engine.Engine
	reg *spec.Registry
	out string

	mode engine.RunMode
	tags map[string]string
}

func NewAttach(eng *engine.Engine, reg *spec.Registry, outPath string, mode engine.RunMode, tags map[string]string) *Attach {
	return &Attach{eng: eng, reg: reg, out: outPath, mode: mode, tags: tags}
}

type Session struct {
	a       *Attach
	started time.Time
}

func (a *Attach) Start() *Session {
	return &Session{a: a, started: time.Now()}
}

func (s *Session) End(ctx context.Context, sliIDs []string) (*summary.Summary, error) {
	finished := time.Now()
	return s.a.eng.Execute(ctx, engine.ExecuteRequest{
		Config: engine.RunConfig{
			RunID:      "", // v1: label에 넣지 말고 meta로만 (필요하면 채움)
			StartedAt:  s.started,
			FinishedAt: finished,
			Mode:       s.a.mode,
			Tags:       s.a.tags,
		},
		SLIIDs:  sliIDs,
		OutPath: s.a.out,
	})
}
