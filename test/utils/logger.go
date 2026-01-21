package utils

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
)

// 사용예시
//logger := utils.NewLogger(utils.GinkgoLog) // 또는 nil 넣어도 OK
//logger.Logf("hello %s", "world")
//
//logger2 := utils.NewLogger(nil) // NoopLogger로 치환됨
//logger2.Logf("this will not print")

// Logger is the minimal logging contract.
// Keep it tiny so core stays independent from klog/logr/controller-runtime/Ginkgo.
type Logger interface {
	Logf(format string, args ...any)
}

// GinkgoLogger writes logs to GinkgoWriter.
type GinkgoLogger struct{}

func (GinkgoLogger) Logf(format string, args ...any) {
	_, _ = fmt.Fprintf(GinkgoWriter, format+"\n", args...)
}

// NoopLogger is a safe default: never panics, never outputs.
type NoopLogger struct{}

func (NoopLogger) Logf(string, ...any) {}

// NewLogger returns a safe Logger.
// If l is nil, it returns a NoopLogger.
func NewLogger(l Logger) Logger {
	if l == nil {
		return NoopLogger{}
	}
	return l
}

// Compile-time checks (optional but nice).
var _ Logger = (*GinkgoLogger)(nil)
var _ Logger = (*NoopLogger)(nil)

// Ready-to-use instances.
var (
	GinkgoLog Logger = GinkgoLogger{}
	NopLog    Logger = NoopLogger{}
)
