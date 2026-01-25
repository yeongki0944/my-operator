package devutil

import "strings"

// GetNonEmptyLines splits output by newline and returns non-empty trimmed lines.
func GetNonEmptyLines(output string) []string {
	lines := strings.Split(output, "\n")
	res := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		res = append(res, line)
	}
	return res
}
