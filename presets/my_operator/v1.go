package my_operator

import (
	"github.com/yeongki/my-operator/pkg/slo/spec"
	"github.com/yeongki/my-operator/presets/controller_runtime"
)

func RegisterV1(reg *spec.Registry) {
	// baseline
	controller_runtime.RegisterV1(reg)

	// override queue labels (example)
	reg.MustRegister(WorkqueueAddsDeltaMyOperator())
	reg.MustRegister(WorkqueueRetriesDeltaMyOperator())

	// operator specific SLIs (examples)
	reg.MustRegister(CRCreatedDelta())
}

func WorkqueueAddsDeltaMyOperator() spec.SLISpec {
	return spec.SLISpec{
		ID:          "my_operator.workqueue_adds_delta",
		Title:       "Workqueue adds delta (my-operator)",
		Unit:        "count",
		Kind:        "delta_counter",
		Description: "Override queue name for my-operator controller.",
		Inputs: []spec.MetricRef{
			spec.PromKey(`workqueue_adds_total{name="my-operator-controller"}`),
		},
		Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
	}
}

func WorkqueueRetriesDeltaMyOperator() spec.SLISpec {
	return spec.SLISpec{
		ID:    "my_operator.workqueue_retries_delta",
		Title: "Workqueue retries delta (my-operator)",
		Unit:  "count",
		Kind:  "delta_counter",
		Inputs: []spec.MetricRef{
			spec.PromKey(`workqueue_retries_total{name="my-operator-controller"}`),
		},
		Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
	}
}

// Example: Replace metric name/labels with your real operator metric.
func CRCreatedDelta() spec.SLISpec {
	return spec.SLISpec{
		ID:          "my_operator.cr_created_delta",
		Title:       "CR created delta",
		Unit:        "count",
		Kind:        "delta_counter",
		Description: "How many CRs created during the window (churn signal).",
		Inputs: []spec.MetricRef{
			spec.PromKey(`my_operator_cr_created_total{kind="SloJob"}`),
		},
		Compute: spec.ComputeSpec{Mode: spec.ComputeDelta},
	}
}
