package engine

import (
	"context"

	"github.com/yeongki/my-operator/pkg/slo/spec"
	"github.com/yeongki/my-operator/pkg/slo/summary"
)

// ExecuteRequestV4 is the v4 request shape.
type ExecuteRequestV4 struct {
	Method  MeasurementMethod
	Config  RunConfig
	Specs   []spec.SLISpec
	OutPath string
}

// ExecuteV4 applies v4 defaults and delegates to the v3 engine.
func ExecuteV4(ctx context.Context, eng *Engine, req ExecuteRequestV4) (*summary.Summary, error) {
	if req.Config.Format == "" {
		req.Config.Format = "v4"
	}
	mode := MapMethodToRunModeV4(req.Method)
	req.Config.Mode = RunMode{
		Location: string(mode.Location),
		Trigger:  string(mode.Trigger),
	}
	return eng.Execute(ctx, ExecuteRequest{
		Config:  req.Config,
		Specs:   req.Specs,
		OutPath: req.OutPath,
	})
}
