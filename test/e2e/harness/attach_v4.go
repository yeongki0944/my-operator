package harness

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
)

// AttachV4Config defines the minimal v4 inputs for InsideSnapshot.
type AttachV4Config struct {
	Namespace          string
	MetricsServiceName string
	TestCase           string
	Suite              string
	RunID              string
	ServiceAccountName string
	Token              string

	ArtifactsDir string
	Tags         map[string]string
}

// AttachV4 provides a v4 Ginkgo entrypoint that does not require CurlPodFns.
// It creates, starts, and ends a v4 session around each test case.
func AttachV4(cfg AttachV4Config) (*SessionV4, error) {
	if cfg.Namespace == "" {
		return nil, errors.New("v4: Namespace is required")
	}
	if cfg.MetricsServiceName == "" {
		return nil, errors.New("v4: MetricsServiceName is required")
	}

	if cfg.TestCase == "" {
		cfg.TestCase = ginkgo.CurrentSpecReport().LeafNodeText
	}

	session := NewSessionV4(SessionV4Config{
		Namespace:          cfg.Namespace,
		MetricsServiceName: cfg.MetricsServiceName,
		TestCase:           cfg.TestCase,
		Suite:              cfg.Suite,
		RunID:              cfg.RunID,
		ServiceAccountName: cfg.ServiceAccountName,
		Token:              cfg.Token,
		ArtifactsDir:       cfg.ArtifactsDir,
		Tags:               cfg.Tags,
		Now:                time.Now,
	})

	ginkgo.BeforeEach(func() {
		session.Start()
	})

	ginkgo.AfterEach(func() {
		if _, err := session.End(context.Background()); err != nil {
			_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, "SLO(v4): End failed (skip): %v\n", err)
		}
	})

	return session, nil
}
