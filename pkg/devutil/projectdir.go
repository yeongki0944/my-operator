package devutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// GetProjectDir returns the project root directory.
//
// Strategy:
//  1. Walk up from current working directory and find a directory containing go.mod (preferred).
//     If not found, also accept ".git" as a fallback.
//  2. As a final fallback (legacy behavior), strip a known suffix "/test/e2e" if present.
func GetProjectDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return wd, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// 1) Prefer go.mod-based root detection.
	if root, ok := findUpwards(wd, "go.mod"); ok {
		return root, nil
	}
	// 1b) Fallback to .git directory.
	if root, ok := findUpwards(wd, ".git"); ok {
		return root, nil
	}

	// 2) Legacy fallback: if cwd includes "/test/e2e" (old scaffold layout).
	legacy := filepath.Clean(wd)
	suffix := string(filepath.Separator) + filepath.Join("test", "e2e")
	if hasPathSuffix(legacy, suffix) {
		return filepath.Clean(legacy[:len(legacy)-len(suffix)]), nil
	}

	return wd, errors.New("project root not found (no go.mod/.git upwards); returning cwd may be incorrect")
}

// findUpwards searches for marker file/directory by walking up from start directory.
func findUpwards(start string, marker string) (string, bool) {
	dir := filepath.Clean(start)
	for {
		if exists(filepath.Join(dir, marker)) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// exists checks if the given path exists.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// hasPathSuffix checks suffix on cleaned path boundaries (best-effort, avoids false positives).
func hasPathSuffix(path, suffix string) bool {
	path = filepath.Clean(path)
	suffix = filepath.Clean(suffix)
	if len(path) < len(suffix) {
		return false
	}
	return path[len(path)-len(suffix):] == suffix
}
