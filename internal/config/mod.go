package config

import (
	"fmt"
	"os"

	"github.com/a8m/envsubst"
	"gopkg.in/yaml.v3"
)

// ModFile is the protobuf module dependencies file (protobuf.mod).
type ModFile struct {
	// Deps is the dependencies repositories in format <repo>@<version>.
	Deps []string `json:"deps,omitempty" yaml:"deps,omitempty"`
}

// ParseModFile parses protobuf.mod content with environment variable expansion.
func ParseModFile(buf []byte) (*ModFile, error) {
	expanded, err := envsubst.String(string(buf))
	if err != nil {
		return nil, fmt.Errorf("envsubst.String: %w", err)
	}
	buf = []byte(expanded)

	mod := &ModFile{}
	err = yaml.Unmarshal(buf, &mod)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal: %w", err)
	}

	return mod, nil
}

// loadModFileDeps reads deps from protobuf.mod at path.
// Returns os.ErrNotExist when the file is missing.
func loadModFileDeps(path string) ([]string, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	mod, err := ParseModFile(buf)
	if err != nil {
		return nil, fmt.Errorf("ParseModFile: %w", err)
	}

	return mod.Deps, nil
}
