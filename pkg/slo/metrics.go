package slo

import (
	"github.com/prometheus/client_golang/prometheus"
)

// TODO E2EConvergenceTimeSeconds 의 경우, 일단 mil 로 할지 second 으로 할지 고민, 문서에서는 seconds 로 되어 있음.

var (
	//E2EConvergenceTimeSeconds  e2e_convergence_time_seconds histogram
	E2EConvergenceTimeSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "e2e_convergence_time_seconds",
			Help: "Time from primary CR create trigger to Ready (E2E observed).",
			// Buckets는 오늘 고정 안 해도 됨. 일단 기본/간단히.
			// Buckets: prometheus.DefBuckets,
		},
		[]string{"suite", "test_case", "namespace", "run_id", "result"},
	)

	//ReconcileTotalDelta reconcile_total_delta gauge
	ReconcileTotalDelta = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "reconcile_total_delta",
			Help: "Delta of controller_runtime_reconcile_total during E2E test window (churn proxy).",
		},
		[]string{"suite", "test_case", "namespace", "run_id", "result"},
	)
)

// Register registers SLO metrics to a registry.
// Important: Call this only when Enabled() == true to keep true No-op default.
func Register(reg prometheus.Registerer) {
	reg.MustRegister(E2EConvergenceTimeSeconds)
	reg.MustRegister(ReconcileTotalDelta)
}
