package harness

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yeongki/my-operator/pkg/slo/engine"
	"github.com/yeongki/my-operator/pkg/slo/fetch"
	"github.com/yeongki/my-operator/pkg/slo/fetch/promtext"
	"github.com/yeongki/my-operator/pkg/slo/spec"
	"github.com/yeongki/my-operator/pkg/slo/summary"
	"github.com/yeongki/my-operator/pkg/slo/tags"
	"github.com/yeongki/my-operator/test/e2e/curlmetrics"
)

// SessionV4Config contains v4 session inputs and defaults.
type SessionV4Config struct {
	Namespace          string
	MetricsServiceName string
	TestCase           string
	Suite              string
	RunID              string
	ServiceAccountName string
	Token              string
	ArtifactsDir       string
	Tags               map[string]string
	Now                func() time.Time

	Specs   []spec.SLISpec
	Fetcher fetch.MetricsFetcher
}

// SessionV4 holds v4 runtime state.
type SessionV4 struct {
	Config SessionV4Config

	MetricsPort      int
	ServiceURLFormat string
	CurlImage        string

	ScrapeTimeout      time.Duration
	WaitPodDoneTimeout time.Duration
	LogsTimeout        time.Duration

	RunID string
	Tags  map[string]string

	Warnings []string

	specs   []spec.SLISpec
	fetcher fetch.MetricsFetcher
	writer  summary.Writer
	started time.Time
}

// NewSessionV4 builds a session with defaults applied.
func NewSessionV4(cfg SessionV4Config) *SessionV4 {
	now := cfg.Now
	if now == nil {
		now = time.Now
	}

	runID := cfg.RunID
	if strings.TrimSpace(runID) == "" {
		runID = fmt.Sprintf("local-%d", now().Unix())
	}

	autoTags := tags.AutoTagsV4(tags.AutoTagsV4Input{
		Suite:     cfg.Suite,
		TestCase:  cfg.TestCase,
		Namespace: cfg.Namespace,
		RunID:     runID,
	})

	mergedTags := tags.MergeTagsV4(cfg.Tags, autoTags)

	return &SessionV4{
		Config:             cfg,
		MetricsPort:        8443,
		ServiceURLFormat:   "https://%s.%s.svc:8443/metrics",
		CurlImage:          "curlimages/curl:latest",
		ScrapeTimeout:      2 * time.Minute,
		WaitPodDoneTimeout: 5 * time.Minute,
		LogsTimeout:        2 * time.Minute,
		RunID:              runID,
		Tags:               mergedTags,
		specs:              defaultSpecsV4(cfg.Specs),
		fetcher:            cfg.Fetcher,
		writer:             summary.NewJSONFileWriter(),
	}
}

// ShouldWriteArtifacts reports whether v4 should write summary output.
func (s *SessionV4) ShouldWriteArtifacts() bool {
	return s.Config.ArtifactsDir != ""
}

// NextSummaryPath returns a unique summary path by appending -<n> on collisions.
func (s *SessionV4) NextSummaryPath(filename string) (string, error) {
	if s.Config.ArtifactsDir == "" {
		return "", nil
	}

	base := filepath.Join(s.Config.ArtifactsDir, filename)
	path := base
	for i := 1; ; i++ {
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return path, nil
			}
			return "", err
		}
		path = fmt.Sprintf("%s-%d", base, i)
	}
}

// AddWarning records a warning message for BestEffort mode.
func (s *SessionV4) AddWarning(message string) {
	if message == "" {
		return
	}
	s.Warnings = append(s.Warnings, message)
}

// Start begins v4 measurement.
func (s *SessionV4) Start() {
	s.started = time.Now()
}

// End completes v4 measurement.
func (s *SessionV4) End(ctx context.Context) (*summary.Summary, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	finished := time.Now()

	fetcher := s.fetcher
	if fetcher == nil {
		fetcher = newCurlPodFetcherV4(s)
	}

	eng := engine.New(fetcher, s.writer, nil)
	outPath := ""
	if s.ShouldWriteArtifacts() {
		filename := fmt.Sprintf(
			"sli-summary.v3.%s.%s.json",
			SanitizeFilename(s.RunID),
			SanitizeFilename(s.Config.TestCase),
		)
		path, err := s.NextSummaryPath(filename)
		if err != nil {
			return nil, err
		}
		outPath = path
	}

	return engine.ExecuteV4(ctx, eng, engine.ExecuteRequestV4{
		Method: engine.InsideSnapshot,
		Config: engine.RunConfig{
			RunID:      s.RunID,
			StartedAt:  s.started,
			FinishedAt: finished,
			Format:     "v4",
			Tags:       s.Tags,
		},
		Specs:   s.specs,
		OutPath: outPath,
	})
}

type curlPodFetcherV4 struct {
	session *SessionV4
	pod     *curlmetrics.CurlPodV4
}

func newCurlPodFetcherV4(session *SessionV4) fetch.MetricsFetcher {
	return &curlPodFetcherV4{
		session: session,
		pod: &curlmetrics.CurlPodV4{
			Namespace:          session.Config.Namespace,
			MetricsServiceName: session.Config.MetricsServiceName,
			ServiceAccountName: session.Config.ServiceAccountName,
			Token:              session.Config.Token,
			Image:              session.CurlImage,
			ServiceURLFormat:   session.ServiceURLFormat,
		},
	}
}

func (f *curlPodFetcherV4) Fetch(ctx context.Context, at time.Time) (fetch.Sample, error) {
	podCtx, cancel := context.WithTimeout(ctx, f.session.ScrapeTimeout)
	defer cancel()

	raw, err := f.pod.Run(podCtx, f.session.WaitPodDoneTimeout, f.session.LogsTimeout)
	if err != nil {
		return fetch.Sample{}, err
	}

	values, err := parsePrometheusTextV4(raw)
	if err != nil {
		return fetch.Sample{}, err
	}

	return fetch.Sample{
		At:     at,
		Values: values,
	}, nil
}

func parsePrometheusTextV4(raw string) (map[string]float64, error) {
	base, err := promtext.ParseTextToMap(strings.NewReader(raw))
	if err != nil {
		return nil, err
	}

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

func defaultSpecsV4(specs []spec.SLISpec) []spec.SLISpec {
	if specs != nil {
		return specs
	}
	return DefaultV3Specs()
}
