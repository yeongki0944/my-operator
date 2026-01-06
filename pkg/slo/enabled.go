package slo

import "os"

// Enabled Default OFF. Only enabled when SLOLAB_ENABLED=1
func Enabled() bool {
	return os.Getenv("SLOLAB_ENABLED") == "1"
}
