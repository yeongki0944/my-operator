package summary

import (
	"encoding/json"
	"os"
)

type Writer interface {
	Write(path string, s Summary) error
}

type JSONFileWriter struct{}

func NewJSONFileWriter() *JSONFileWriter { return &JSONFileWriter{} }

func (w *JSONFileWriter) Write(path string, s Summary) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
