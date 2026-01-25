package summary

// EnsureV4Format sets the v4 format hint while preserving schemaVersion v3.
func EnsureV4Format(config map[string]any) map[string]any {
	if config == nil {
		config = map[string]any{}
	}
	if _, ok := config["format"]; !ok {
		config["format"] = "v4"
	}
	return config
}
