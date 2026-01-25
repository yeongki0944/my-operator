package env

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// LoadOptions Options holds E2E test configuration loaded from environment variables.
func LoadOptions() Options {
	return Options{
		Enabled: boolEnv("SLOLAB_ENABLED", false),

		ArtifactsDir: stringEnv("ARTIFACTS_DIR", "/tmp"),
		RunID:        stringEnv("CI_RUN_ID", ""),

		SkipCleanup:            boolEnv("E2E_SKIP_CLEANUP", false),
		SkipCertManagerInstall: boolEnv("CERT_MANAGER_INSTALL_SKIP", false),

		TokenRequestTimeout: durationEnv("TOKEN_REQUEST_TIMEOUT", 2*time.Minute),
	}
}

// --- helpers (규칙 통일: "1"/"true"/"yes"/"on" 모두 허용) ---

// stringEnv returns environment variable as string.
func stringEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

// boolEnv parses environment variable as bool.
func boolEnv(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "t", "yes", "y", "on":
		return true
	case "0", "false", "f", "no", "n", "off":
		return false
	default:
		return def
	}
}

// durationEnv parses environment variable as time.Duration. 다만, 숫자만 들어오면 초단위로 간주.
func durationEnv(key string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(n) * time.Second
	}
	return def
}
