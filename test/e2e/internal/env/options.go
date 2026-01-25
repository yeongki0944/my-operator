package env

import (
	"path/filepath"
	"time"
)

// Options is e2e-only configuration.
// Keep this independent from pkg/slo (v1 legacy).
type Options struct {
	Enabled      bool
	ArtifactsDir string
	RunID        string

	SkipCleanup            bool
	SkipCertManagerInstall bool

	TokenRequestTimeout time.Duration
}

func (o Options) Validate() Options {
	out := o
	if out.ArtifactsDir == "" {
		out.ArtifactsDir = "/tmp"
	}
	if out.TokenRequestTimeout == 0 {
		out.TokenRequestTimeout = 2 * time.Minute
	}
	return out
}

func (o Options) SummaryPath(filename string) string {
	v := o.Validate()
	if filename == "" {
		filename = "sli-summary.json"
	}
	return filepath.Join(v.ArtifactsDir, filename)
}
