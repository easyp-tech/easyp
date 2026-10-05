// Package schemagen writes the JSON Schemas used by v1 configuration validation.
package schemagen

import (
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

const DefaultOutDir = "schemas"

// Options controls where v1 schema artifacts are written.
type Options struct {
	OutDir string
}

// Run creates versioned schemas and latest aliases for the three YAML files.
func Run(opts Options) error {
	dir := opts.OutDir
	if dir == "" {
		dir = DefaultOutDir
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}
	for _, name := range []string{"easyp", "easyp.gen", "protobuf.lock"} {
		data, err := v1.SchemaJSON(name)
		if err != nil {
			return fmt.Errorf("SchemaJSON: %w", err)
		}
		versioned, latest := schemaNames(name)
		for _, filename := range []string{versioned, latest} {
			if err := os.WriteFile(filepath.Join(dir, filename), data, 0o644); err != nil {
				return fmt.Errorf("WriteFile: %w", err)
			}
		}
	}
	return nil
}

func schemaNames(name string) (string, string) {
	return name + "-v1.schema.json", name + ".schema.json"
}
