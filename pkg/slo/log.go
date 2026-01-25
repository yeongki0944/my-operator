package slo

// Logger is the minimal logging contract for pkg/slo.
type Logger interface {
	Logf(format string, args ...any)
}

type nopLogger struct{}

func (nopLogger) Logf(string, ...any) {}

// NewLogger returns a safe Logger. If l is nil, returns no-op.
func NewLogger(l Logger) Logger {
	if l == nil {
		return nopLogger{}
	}
	return l
}

// NopLogger exported singleton if you like
var NopLogger Logger = nopLogger{}
