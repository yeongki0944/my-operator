package summary

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Writer writes a Summary artifact to a destination.
type Writer interface {
	Write(path string, s Summary) error
}

type JSONFileWriter struct{}

func NewJSONFileWriter() *JSONFileWriter { return &JSONFileWriter{} }

// Write sync=true for atomic writer durability (fsync)
func (w *JSONFileWriter) Write(path string, s Summary) error {
	if path == "" {
		// skip (no output path configured)
		return nil
	}
	return writeJSONAtomic(path, s, 0o644, 0o755, true)
}

// writeJSONAtomic writes JSON to a temp file in the same directory and then renames it.
// - Atomic replace is provided by os.Rename (same filesystem).
// - If doSync is true, it fsyncs the temp file before close for stronger durability.
func writeJSONAtomic(path string, s Summary, fileMode, dirMode os.FileMode, doSync bool) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return err
	}

	f, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()

	success := false
	defer func() {
		if !success {
			_ = f.Close()
			_ = os.Remove(tmp)
		}
	}()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return err
	}

	if doSync {
		if err := f.Sync(); err != nil {
			return err
		}
	}

	if err := f.Close(); err != nil {
		return err
	}

	if err := os.Chmod(tmp, fileMode); err != nil {
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		return err
	}

	success = true
	return nil
}
