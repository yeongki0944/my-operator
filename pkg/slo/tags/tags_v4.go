package tags

// AutoTagsV4Input defines the auto-tag fields for v4.
type AutoTagsV4Input struct {
	Suite     string
	TestCase  string
	Namespace string
	RunID     string
}

// AutoTagsV4 returns the v4 auto-tags map.
func AutoTagsV4(input AutoTagsV4Input) map[string]string {
	return map[string]string{
		"suite":     input.Suite,
		"test_case": input.TestCase,
		"namespace": input.Namespace,
		"run_id":    input.RunID,
	}
}

// MergeTagsV4 merges user tags over auto-tags (user overrides).
func MergeTagsV4(userTags map[string]string, autoTags map[string]string) map[string]string {
	merged := map[string]string{}
	for key, value := range autoTags {
		if value != "" {
			merged[key] = value
		}
	}
	for key, value := range userTags {
		merged[key] = value
	}
	return merged
}
