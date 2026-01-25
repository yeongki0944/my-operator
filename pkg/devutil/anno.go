package devutil

import "time"

const TestStartTimeAnnoKey = "test/start-time"

// SetTestStartTimeAnno sets test/start-time annotation to current UTC time (RFC3339Nano).
// Glue-layer helper: keeps core independent from k8s types (metav1.Object etc.).
func SetTestStartTimeAnno(ann map[string]string) map[string]string {
	return SetTestStartTimeAnnoAt(ann, time.Now())
}

// SetTestStartTimeAnnoAt sets test/start-time annotation using the provided time.
// Prefer this in callers that want one captured "now" reused across multiple objects.
func SetTestStartTimeAnnoAt(ann map[string]string, now time.Time) map[string]string {
	if ann == nil {
		ann = map[string]string{}
	}
	ann[TestStartTimeAnnoKey] = now.UTC().Format(time.RFC3339Nano)
	return ann
}
