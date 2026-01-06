package slo

// Logger is the minimal logging contract for pkg/slo.
// Keep it tiny so pkg/slo stays independent from klog/logr/controller-runtime.
type Logger interface {
	Logf(format string, args ...any)
}

// nopLogger is a safe default: never panics, never outputs.
type nopLogger struct{}

func (nopLogger) Logf(string, ...any) {}

// newLogf returns a safe log function.
// If l is nil, it returns a no-op func.
func newLogf(l Logger) func(string, ...any) {
	if l == nil {
		// Inline no-op avoids allocation of nopLogger in hot paths.
		return func(string, ...any) {}
	}
	return l.Logf
}
