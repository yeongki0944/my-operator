package slo

// Labels 일단 5개로 고정, 추후 필요시 확장 가능, 문서 참고
type Labels struct {
	Suite     string `json:"suite"`
	TestCase  string `json:"test_case"`
	Namespace string `json:"namespace"`
	RunID     string `json:"run_id"`
	Result    string `json:"result"` // success|fail|skip
}
