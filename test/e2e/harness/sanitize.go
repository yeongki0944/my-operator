package harness

import "strings"

// SanitizeFilename makes a string safe-ish for filenames.
func SanitizeFilename(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "na"
	}
	r := strings.NewReplacer(
		"/", "_", "\\", "_", " ", "_", ":", "_", ";", "_",
		"\"", "_", "'", "_", "\n", "_", "\r", "_", "\t", "_",
	)
	return r.Replace(s)
}
