package e2eutil

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"

	"github.com/yeongki/my-operator/pkg/slo"
)

// 사용 예시
// logger := slo.NewLogger(utils.GinkgoLog) // nil이면 noop
// logger.Logf("hello %s", "world")

// GinkgoLogger adapts slo.Logger to GinkgoWriter.
type GinkgoLogger struct{}

func (GinkgoLogger) Logf(format string, args ...any) {
	_, _ = fmt.Fprintf(ginkgo.GinkgoWriter, format+"\n", args...)
}

// Compile-time check
var _ slo.Logger = (*GinkgoLogger)(nil)

// Ready-to-use instance
var GinkgoLog slo.Logger = GinkgoLogger{}
