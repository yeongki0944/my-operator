package slo

import "time"

type Recorder struct {
	enabled bool
	logf    func(string, ...any)
	writer  SummaryWriter // nil이면 저장 안 함
}

func NewRecorder(enabled bool, l Logger, w SummaryWriter) *Recorder {
	return &Recorder{
		enabled: enabled,
		logf:    newLogf(l),
		writer:  w,
	}
}

// TODO 일단 NewRecorder 여기서 enabled 하고, log 잡아주는데, 아래와 같은 옵션 패턴도 고민해보자.
// WithEnabled(bool)
// WithLogger(logr.Logger)

func (r *Recorder) Enabled() bool { return r.enabled }

// ObserveConvergence records a single convergence time into histogram.
// No-op if disabled.
func (r *Recorder) ObserveConvergence(l Labels, d time.Duration) {
	if !r.enabled {
		return
	}
	// Guardrail: negative durations are almost always a bug in start/end timestamps.
	if d < 0 {
		r.log("slo: negative convergence duration ignored: %v labels=%+v", d, l)
		return
	}
	// TODO 라벨에 대한 이슈가 있는데, 정확히 파악하지 못하고 있음. 특히 RunID 문제를 지속적으로 제기하고 있는데 살펴봐야함.
	// TODO 보통, WithLabelValues(l.Result) 이렇게 표현하던데, 차이점을 세부적으로 찾아봐야 함.
	E2EConvergenceTimeSeconds.WithLabelValues(
		l.Suite, l.TestCase, l.Namespace, l.RunID, l.Result,
	).Observe(d.Seconds())
}

// ObserveReconcileDelta records delta into gauge.
// No-op if disabled.
func (r *Recorder) ObserveReconcileDelta(l Labels, delta int64) {
	if !r.enabled {
		return
	}
	// guardrail: measurement failure != test failure
	if delta < 0 {
		r.log("[slo-lab] skip reconcile_total_delta: negative delta=%d", delta)
		return
	}
	// TODO 라벨에 대한 이슈가 있는데, 정확히 파악하지 못하고 있음. 특히 RunID 문제를 지속적으로 제기하고 있는데 살펴봐야함.
	// TODO 내가 생각하기에 RunID 불변이면 이상없을 듯함.
	ReconcileTotalDelta.WithLabelValues(
		l.Suite, l.TestCase, l.Namespace, l.RunID, l.Result,
	).Set(float64(delta))
}

// RecordAndSave records the convergence metric and optionally persists a summary.
// - Always calls ObserveConvergence (when enabled).
// - If writer is set, emits a JSON-friendly summary for CI artifacts.
// - Returns error only for "save" step; metric observe is best-effort.
func (r *Recorder) RecordAndSave(l Labels, d time.Duration) error {
	// 1) 메트릭 기록 (enabled/guardrail은 ObserveConvergence가 책임)
	r.ObserveConvergence(l, d)

	// 2) 저장은 옵션 (writer 없으면 끝)
	if !r.enabled || r.writer == nil {
		return nil
	}
	// ObserveConvergence와 동일한 guardrail을 유지(중복이 싫으면 ObserveConvergence가 bool 반환하도록 리팩터 가능)
	if d < 0 {
		return nil
	}

	val := d.Seconds()
	summary := Summary{
		Labels: l,
		Metrics: SummaryMetrics{
			E2EConvergenceTimeSeconds: &val,
		},
	}

	if err := r.writer.WriteSummary(summary); err != nil {
		// 저장 실패는 CI에서 중요한 신호일 수 있으니 로그는 남김
		r.logf("slo: failed to write summary: %v labels=%+v", err, l)
		return err
	}
	return nil
}

func (r *Recorder) log(format string, args ...any) {
	r.logf(format, args...)
}
