package devutil

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// RenderTemplateFile reads a template file at (rootDir + relPath), executes it with data,
// and returns the rendered bytes.
//
// - missing keys cause error (missingkey=error)
// - rootDir is typically the project root (e.g., repo root)
func RenderTemplateFile(rootDir, relPath string, data any) ([]byte, error) {
	if rootDir == "" {
		return nil, fmt.Errorf("rootDir is empty")
	}
	if relPath == "" {
		return nil, fmt.Errorf("relPath is empty")
	}

	path := filepath.Join(rootDir, relPath)

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(filepath.Base(relPath)).
		Option("missingkey=error").
		Parse(string(b))
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func RenderTemplateFileString(rootDir, relPath string, data any) (string, error) {
	b, err := RenderTemplateFile(rootDir, relPath, data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
