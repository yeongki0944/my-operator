package harness

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"

	"github.com/yeongki/my-operator/pkg/slo/engine"
	"github.com/yeongki/my-operator/pkg/slo/fetch"
	"github.com/yeongki/my-operator/pkg/slo/fetch/promtext"
	"github.com/yeongki/my-operator/pkg/slo/spec"
	"github.com/yeongki/my-operator/pkg/slo/summary"
)

// HarnessDeps = “Ginkgo hook + RunConfig/tags/output에 필요한 것”
type HarnessDeps struct {
	ArtifactsDir string

	Suite    string
	TestCase string
	RunID    string

	Enabled bool
}

// FetchDeps = “metrics를 어떻게 가져올지(inside curl-pod)에 필요한 것”
type FetchDeps struct {
	Namespace          string
	Token              string
	MetricsServiceName string
	ServiceAccountName string
}

// CurlPodFns are injected to avoid import cycles (harness should not import test/e2e directly).
type CurlPodFns struct {
	RunCurlMetricsOnce  func(ns, token, metricsSvc, sa string) (podName string, err error)
	WaitCurlMetricsDone func(ns, podName string)
	CurlMetricsLogs     func(ns, podName string) (string, error)
	DeletePodNoWait     func(ns, podName string) error
}

// SpecsProvider provides SLI specs for this test.
// - nil allowed: treat as empty specs (engine will still write summary with empty results).
type SpecsProvider func() []spec.SLISpec

// Attach registers BeforeEach/AfterEach hooks to automatically measure SLO for each test.
// - It does NOT read env vars.
// - It does NOT know how to obtain token.
// - It relies on providers to supply per-test deps + SLI specs.
func Attach(hdepsProvider func() HarnessDeps, fdepsProvider func() FetchDeps, specsProvider SpecsProvider, fns CurlPodFns) {
	var sess *session
	var enabled bool

	BeforeEach(func() {
		hdeps := hdepsProvider()
		fdeps := fdepsProvider()

		enabled = hdeps.Enabled
		if !enabled {
			sess = nil
			return
		}

		// Auto-fill TestCase if empty.
		if strings.TrimSpace(hdeps.TestCase) == "" {
			hdeps.TestCase = CurrentSpecReport().LeafNodeText
		}

		var specs []spec.SLISpec
		if specsProvider != nil {
			specs = specsProvider()
		} else {
			specs = nil // empty: no SLI results, but summary still produced
		}
		// 일단, sess 자체가 nil 이 나올 수 없음.
		sess = newSession(hdeps, fdeps, specs, fns)
		sess.Start()
	})

	AfterEach(func() {
		if !enabled || sess == nil {
			return
		}
		if err := sess.End(context.Background()); err != nil {
			_, _ = fmt.Fprintf(GinkgoWriter, "SLO(v3): End failed (skip): %v\n", err)
		}
	})
}

type session struct {
	eng     *engine.Engine
	outPath string
	mode    engine.RunMode
	tags    map[string]string
	runID   string
	specs   []spec.SLISpec

	started time.Time
}

func newSession(hdeps HarnessDeps, fdeps FetchDeps, specs []spec.SLISpec, fns CurlPodFns) *session {
	writer := summary.Writer(noopWriter{})
	outPath := ""
	if strings.TrimSpace(hdeps.ArtifactsDir) != "" {
		filename := fmt.Sprintf(
			"sli-summary.v3.%s.%s.json",
			SanitizeFilename(hdeps.RunID),
			SanitizeFilename(hdeps.TestCase),
		)
		outPath = filepath.Join(hdeps.ArtifactsDir, filename)
		writer = summary.NewJSONFileWriter()
	}

	fetcher := curlMetricsFetcher{
		deps: fdeps,
		fns:  fns,
	}

	// v3 engine: Specs are directly injected via ExecuteRequest.
	eng := engine.New(fetcher, writer, nil)

	return &session{
		eng:     eng,
		outPath: outPath,
		mode: engine.RunMode{
			Location: "inside",
			Trigger:  "none",
		},
		tags: map[string]string{
			"suite":     hdeps.Suite,
			"test_case": hdeps.TestCase,
			"namespace": fdeps.Namespace,
			"run_id":    hdeps.RunID,
		},
		runID: hdeps.RunID,
		specs: specs,
	}
}

func (s *session) Start() {
	s.started = time.Now()
}

func (s *session) End(ctx context.Context) error {
	finished := time.Now()

	_, err := s.eng.Execute(ctx, engine.ExecuteRequest{
		Config: engine.RunConfig{
			RunID:      s.runID,
			StartedAt:  s.started,
			FinishedAt: finished,
			Mode:       s.mode,
			Tags:       s.tags,
		},
		Specs:   s.specs,
		OutPath: s.outPath,
	})
	return err
}

type noopWriter struct{}

func (noopWriter) Write(path string, s summary.Summary) error { return nil }

type curlMetricsFetcher struct {
	deps FetchDeps
	fns  CurlPodFns
}

func (f curlMetricsFetcher) Fetch(ctx context.Context, at time.Time) (fetch.Sample, error) {
	_ = ctx

	podName, err := f.fns.RunCurlMetricsOnce(
		f.deps.Namespace,
		f.deps.Token,
		f.deps.MetricsServiceName,
		f.deps.ServiceAccountName,
	)
	if err != nil {
		return fetch.Sample{}, err
	}

	f.fns.WaitCurlMetricsDone(f.deps.Namespace, podName)

	raw, err := f.fns.CurlMetricsLogs(f.deps.Namespace, podName)
	_ = f.fns.DeletePodNoWait(f.deps.Namespace, podName)
	if err != nil {
		return fetch.Sample{}, err
	}

	values, err := parsePrometheusText(raw)
	if err != nil {
		return fetch.Sample{}, err
	}

	return fetch.Sample{
		At:     at,
		Values: values,
	}, nil
}

func parsePrometheusText(raw string) (map[string]float64, error) {
	base, err := promtext.ParseTextToMap(strings.NewReader(raw))
	if err != nil {
		return nil, err
	}

	// v3 convenience: also aggregate per-metric-name (strip label set).
	out := map[string]float64{}
	for key, val := range base {
		out[key] = val
		if idx := strings.Index(key, "{"); idx > 0 {
			name := key[:idx]
			out[name] = out[name] + val
		}
	}
	return out, nil
}
