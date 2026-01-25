package harness

import "github.com/yeongki/my-operator/pkg/slo/spec"

// DefaultV3Specs is kept for backward compatibility.
// It returns the baseline preset set.
func DefaultV3Specs() []spec.SLISpec {
	return BaselineV3Specs()
}

// BaselineV3Specs is the expanded, reusable preset set:
// controller-runtime + workqueue + rest-client.
func BaselineV3Specs() []spec.SLISpec {
	return []spec.SLISpec{
		// ---------------------------
		// controller-runtime reconcile
		// ---------------------------
		{
			ID:          "reconcile_total_delta",
			Title:       "reconcile total delta",
			Unit:        "count",
			Kind:        "delta_counter",
			Description: "Delta of controller_runtime_reconcile_total during the test window (all results).",
			Inputs: []spec.MetricRef{
				// name-only aggregation is supported by parsePrometheusText(out[name]+=val)
				spec.PromMetric("controller_runtime_reconcile_total", nil),
			},
			Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
		},
		{
			ID:          "reconcile_success_delta",
			Title:       "reconcile success delta",
			Unit:        "count",
			Kind:        "delta_counter",
			Description: `Delta of controller_runtime_reconcile_total{result="success"}.`,
			Inputs: []spec.MetricRef{
				spec.PromMetric("controller_runtime_reconcile_total", spec.Labels{"result": "success"}),
			},
			Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
		},
		{
			ID:          "reconcile_error_delta",
			Title:       "reconcile error delta",
			Unit:        "count",
			Kind:        "delta_counter",
			Description: `Delta of controller_runtime_reconcile_total{result="error"}.`,
			Inputs: []spec.MetricRef{
				spec.PromMetric("controller_runtime_reconcile_total", spec.Labels{"result": "error"}),
			},
			Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
			// Optional judge example: error delta should be 0
			// Judge: &spec.JudgeSpec{Rules: []spec.Rule{{Op: spec.OpGT, Target: 0, Level: spec.LevelFail}}},
		},

		// ---------------------------
		// workqueue (controller-runtime)
		// ---------------------------
		{
			ID:          "workqueue_adds_total_delta",
			Title:       "workqueue adds total delta",
			Unit:        "count",
			Kind:        "delta_counter",
			Description: "Delta of workqueue_adds_total during the test window (all queues).",
			Inputs: []spec.MetricRef{
				spec.PromMetric("workqueue_adds_total", nil),
			},
			Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
		},
		{
			ID:          "workqueue_retries_total_delta",
			Title:       "workqueue retries total delta",
			Unit:        "count",
			Kind:        "delta_counter",
			Description: "Delta of workqueue_retries_total during the test window (all queues).",
			Inputs: []spec.MetricRef{
				spec.PromMetric("workqueue_retries_total", nil),
			},
			Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
		},
		{
			ID:          "workqueue_depth_end",
			Title:       "workqueue depth at end",
			Unit:        "items",
			Kind:        "gauge",
			Description: "workqueue_depth gauge snapshot at the end time (all queues).",
			Inputs: []spec.MetricRef{
				spec.PromMetric("workqueue_depth", nil),
			},
			Compute: spec.ComputeSpec{Mode: spec.ComputeSingle}, // end-only gauge would be better in v4; v3 uses single(start)
			// NOTE(v3): ComputeSingle uses start snapshot in your engine.
			// If you want end snapshot for gauges, we should add ComputeEnd or ComputeSingleAt in v3.
		},

		// ---------------------------
		// rest-client (client-go)
		// ---------------------------
		{
			ID:          "rest_client_requests_total_delta",
			Title:       "rest client requests total delta",
			Unit:        "count",
			Kind:        "delta_counter",
			Description: "Delta of rest_client_requests_total during the test window (all codes/methods).",
			Inputs: []spec.MetricRef{
				spec.PromMetric("rest_client_requests_total", nil),
			},
			Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
		},
		{
			ID:          "rest_client_429_delta",
			Title:       "rest client 429 delta",
			Unit:        "count",
			Kind:        "delta_counter",
			Description: `Delta of rest_client_requests_total{code="429"}. Indicates API server throttling.`,
			Inputs: []spec.MetricRef{
				spec.PromMetric("rest_client_requests_total", spec.Labels{"code": "429"}),
			},
			Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
		},
		{
			ID:          "rest_client_5xx_delta",
			Title:       "rest client 5xx delta",
			Unit:        "count",
			Kind:        "delta_counter",
			Description: `Delta of rest_client_requests_total{code="5xx"}. Some client-go versions aggregate 5xx as "5xx".`,
			Inputs: []spec.MetricRef{
				spec.PromMetric("rest_client_requests_total", spec.Labels{"code": "5xx"}),
			},
			Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
		},
	}
}
