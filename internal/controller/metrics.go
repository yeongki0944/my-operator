package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ReconcileDurationSeconds: Reconcile 작업 소요 시간 히스토그램 (0.1s ~ 30s)
	ReconcileDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "joboperator_reconcile_duration_seconds",
			Help:    "JobOperator reconcile latency in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 3.0, 4.0, 5.0, 10.0, 20.0, 30.0},
		},
		[]string{"name", "namespace", "result"},
	)

	// ReconcileTotal: Reconcile 총 시도 횟수
	ReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "joboperator_reconcile_total",
			Help: "Total number of JobOperator reconcile attempts",
		},
		[]string{"name", "namespace", "result"},
	)

	// ReconcileErrors: Reconcile 에러 발생 횟수
	ReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "joboperator_reconcile_errors_total",
			Help: "Total number of JobOperator reconcile errors",
		},
		[]string{"name", "namespace", "error_type"},
	)
)

func init() {
	// 정의한 메트릭을 전역 레지스트리에 등록
	metrics.Registry.MustRegister(
		ReconcileDurationSeconds,
		ReconcileTotal,
		ReconcileErrors,
	)
}