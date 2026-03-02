package schemagen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/mcp/easypconfig"
)

const (
	DefaultVersionedOut = "schemas/easyp-config-v1.schema.json"
	DefaultLatestOut    = "schemas/easyp-config.schema.json"
)

type Options struct {
	VersionedOut string
	LatestOut    string
}

func Run(opts Options) error {
	versionedOut := opts.VersionedOut
	if versionedOut == "" {
		versionedOut = DefaultVersionedOut
	}

	latestOut := opts.LatestOut
	if latestOut == "" {
		latestOut = DefaultLatestOut
	}

	data, err := easypconfig.MarshalConfigJSONSchema()
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}
	data = append(data, '\n')

	if err := writeFile(versionedOut, data); err != nil {
		return fmt.Errorf("write versioned schema: %w", err)
	}

	if err := writeFile(latestOut, data); err != nil {
		return fmt.Errorf("write latest schema: %w", err)
	}

	return nil
}

func writeFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("os.WriteFile %s: %w", path, err)
	}

	return nil
}
