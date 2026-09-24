package generation

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// inheritV1GenerateOptions reads only options from ancestor generator files.
// The nearest explicit value wins; generate and plugins belong to the consumer.
func inheritV1GenerateOptions(repoRoot, configPath string, gen *v1.Generate) error {
	configDir := filepath.Dir(configPath)
	rel, err := filepath.Rel(repoRoot, configDir)
	if err != nil {
		return fmt.Errorf("Rel: %w", err)
	}
	if !filepath.IsLocal(rel) {
		return fmt.Errorf("generator %s is outside repository %s", configPath, repoRoot)
	}
	if configDir == repoRoot {
		return nil
	}
	for _, dir := range ancestorDirs(configDir, repoRoot)[1:] {
		parentPath := filepath.Join(dir, v1.GenerateFile)
		raw, found, err := readOptionalFile(parentPath)
		if err != nil {
			return fmt.Errorf("%s: %w", parentPath, err)
		}
		if !found {
			continue
		}
		parent, err := v1.ParseGenerate(bytes.NewReader(raw))
		if err != nil {
			return fmt.Errorf("%s: %w", parentPath, err)
		}
		if gen.Options.Go.PackagePrefix == nil && parent.Options.Go.PackagePrefix != nil {
			gen.Options.Go.PackagePrefix = parent.Options.Go.PackagePrefix
			gen.InheritedGoPackagePrefix = true
		}
	}
	return nil
}

func readOptionalFile(path string) ([]byte, bool, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("ReadFile: %w", err)
	}
	return raw, true, nil
}

func ancestorDirs(start, root string) []string {
	var directories []string
	for directory := start; ; directory = filepath.Dir(directory) {
		directories = append(directories, directory)
		if directory == root || directory == filepath.Dir(directory) {
			return directories
		}
	}
}
