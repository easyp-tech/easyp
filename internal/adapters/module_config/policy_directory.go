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

// DependencyManifestDirectory finds the actual directory of an already verified
// module. Proto import roots cannot locate policies outside those roots. It
// selects by declared identity before parsing, so unrelated manifests stay lazy.
func DependencyManifestDirectory(repositoryRoot, identity string) (string, error) {
	var matches []string
	err := filepath.WalkDir(repositoryRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != repositoryRoot && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "easyp_vendor" || entry.Name() == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != v1.ModuleFile {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("ReadFile: %w", err)
		}
		if nativeDependencyName(raw) != identity {
			return nil
		}
		module, err := v1.ParseModule(bytes.NewReader(raw))
		if err != nil {
			return fmt.Errorf("ParseModule: %s: %w", path, err)
		}
		if module.Name == identity {
			matches = append(matches, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("WalkDir: %w", err)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("module %s is declared more than once in %s", identity, repositoryRoot)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	// Pre-native Git metadata has no module manifest; its policy root is the
	// verified checkout, not one of the adapted proto import roots.
	return repositoryRoot, nil
}
