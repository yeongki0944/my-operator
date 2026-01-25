package engine

// MeasurementMethod defines the v4 measurement method.
type MeasurementMethod string

const (
	InsideSnapshot MeasurementMethod = "InsideSnapshot"

	// Reserved for later phases.
	InsideAnnotation MeasurementMethod = "InsideAnnotation"
	OutsideSnapshot  MeasurementMethod = "OutsideSnapshot"
)

// RunLocation describes where the measurement runs.
type RunLocation string

// RunTrigger describes what triggers measurement capture.
type RunTrigger string

const (
	RunLocationInside  RunLocation = "inside"
	RunLocationOutside RunLocation = "outside"

	RunTriggerNone       RunTrigger = "none"
	RunTriggerAnnotation RunTrigger = "annotation"
)

// RunModeV4 holds the resolved run mode for v4.
type RunModeV4 struct {
	Location RunLocation
	Trigger  RunTrigger
}

// MapMethodToRunModeV4 converts a v4 method into a framework-owned run mode.
func MapMethodToRunModeV4(method MeasurementMethod) RunModeV4 {
	switch method {
	case InsideAnnotation:
		return RunModeV4{Location: RunLocationInside, Trigger: RunTriggerAnnotation}
	case OutsideSnapshot:
		return RunModeV4{Location: RunLocationOutside, Trigger: RunTriggerNone}
	case InsideSnapshot:
		fallthrough
	default:
		return RunModeV4{Location: RunLocationInside, Trigger: RunTriggerNone}
	}
}
