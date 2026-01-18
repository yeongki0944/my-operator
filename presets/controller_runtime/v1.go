package controller_runtime

import "github.com/yeongki/my-operator/pkg/slo/spec"

func RegisterV1(reg *spec.Registry) {
	reg.MustRegister(ReconcileTotalDeltaSuccess())
	reg.MustRegister(ReconcileTotalDeltaError())
	reg.MustRegister(WorkqueueAddsDelta())
	reg.MustRegister(WorkqueueRetriesDelta())
}

func ReconcileTotalDeltaSuccess() spec.SLISpec {
	return spec.SLISpec{
		ID:          "controller_runtime.reconcile_total_delta.success",
		Title:       "Reconcile total delta (success)",
		Unit:        "count",
		Kind:        "delta_counter",
		Description: "Delta of successful reconciliations between start/end snapshots.",
		Inputs: []spec.MetricRef{
			spec.PromKey(`controller_runtime_reconcile_total{result="success"}`),
		},
		Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
		Judge: &spec.JudgeSpec{
			Rules: []spec.Rule{
				{Metric: "value", Op: ">", Target: 500, Level: "warn"},
				{Metric: "value", Op: ">", Target: 2000, Level: "fail"},
			},
		},
	}
}

func ReconcileTotalDeltaError() spec.SLISpec {
	return spec.SLISpec{
		ID:          "controller_runtime.reconcile_total_delta.error",
		Title:       "Reconcile total delta (error)",
		Unit:        "count",
		Kind:        "delta_counter",
		Description: "Delta of errored reconciliations between start/end snapshots.",
		Inputs: []spec.MetricRef{
			spec.PromKey(`controller_runtime_reconcile_total{result="error"}`),
		},
		Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
		Judge: &spec.JudgeSpec{
			Rules: []spec.Rule{
				{Metric: "value", Op: ">", Target: 0, Level: "warn"},
				{Metric: "value", Op: ">", Target: 10, Level: "fail"},
			},
		},
	}
}

func WorkqueueAddsDelta() spec.SLISpec {
	return spec.SLISpec{
		ID:          "controller_runtime.workqueue_adds_delta",
		Title:       "Workqueue adds delta",
		Unit:        "count",
		Kind:        "delta_counter",
		Description: "Delta of workqueue adds between start/end snapshots.",
		Inputs: []spec.MetricRef{
			spec.PromKey(`workqueue_adds_total{name="controller"}`),
		},
		Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
	}
}

func WorkqueueRetriesDelta() spec.SLISpec {
	return spec.SLISpec{
		ID:          "controller_runtime.workqueue_retries_delta",
		Title:       "Workqueue retries delta",
		Unit:        "count",
		Kind:        "delta_counter",
		Description: "Delta of workqueue retries between start/end snapshots.",
		Inputs: []spec.MetricRef{
			spec.PromKey(`workqueue_retries_total{name="controller"}`),
		},
		Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
	}
}
