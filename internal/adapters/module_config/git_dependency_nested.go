package moduleconfig

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

type nestedGitDependency struct {
	module       v1.Module
	found        bool
	hasManifests bool
}

// readNestedGitDependencyModule selects one declared module from a Git
// checkout. Roots are returned relative to the checkout, as expected by the
// resolver and the verified module cache.
func readNestedGitDependencyModule(dir, source string) (nestedGitDependency, error) {
	var result nestedGitDependency
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("WalkDir: %w", walkErr)
		}
		if entry.IsDir() {
			if path != dir && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "easyp_vendor") {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != dependencyManifestFile || filepath.Dir(path) == dir {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("ReadFile: %s: %w", path, err)
		}
		if !v1.IsModuleManifest(raw) {
			return nil
		}
		result.hasManifests = true
		module, err := v1.ParseModule(bytes.NewReader(raw))
		if err != nil {
			return fmt.Errorf("ParseModule: %s: %w", path, err)
		}
		if module.Name != source {
			return nil
		}
		if result.found {
			return fmt.Errorf("module %s is declared more than once in %s", source, dir)
		}
		location, err := filepath.Rel(dir, filepath.Dir(path))
		if err != nil {
			return fmt.Errorf("Rel: %w", err)
		}
		for i, root := range module.Roots {
			module.Roots[i] = filepath.Join(location, root)
		}
		result.module = module
		result.found = true
		return nil
	})
	if err != nil {
		return nestedGitDependency{}, fmt.Errorf("WalkDir: %w", err)
	}
	return result, nil
}
