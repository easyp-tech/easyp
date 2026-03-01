package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/mcp/easypconfig"
)

func main() {
	var (
		versionedOut string
		latestOut    string
	)

	flag.StringVar(&versionedOut, "out-versioned", "schemas/easyp-config-v1.schema.json", "path to versioned schema file")
	flag.StringVar(&latestOut, "out-latest", "schemas/easyp-config.schema.json", "path to latest schema alias file")
	flag.Parse()

	data, err := easypconfig.MarshalConfigJSONSchema()
	if err != nil {
		exitf("marshal schema: %v", err)
	}
	data = append(data, '\n')

	if err := writeFile(versionedOut, data); err != nil {
		exitf("write versioned schema: %v", err)
	}
	if err := writeFile(latestOut, data); err != nil {
		exitf("write latest schema: %v", err)
	}
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

func exitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
