package moduleconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/easyp-tech/easyp/internal/adapters/modfile"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// ReadGitDependency adapts metadata in a checked-out Git repository to
// the roots and requirements needed by the v1 resolver. Its identity is the
// source named by the requiring module for pre-v1 repositories.
func ReadGitDependency(dir, source string) (v1.Module, error) {
	module := v1.Module{Name: source}
	manifest, err := os.ReadFile(filepath.Join(dir, "protobuf.mod"))
	if err != nil && !os.IsNotExist(err) {
		return v1.Module{}, err
	}
	if err == nil {
		if v1.IsModuleManifest(manifest) {
			parsed, err := v1.ParseModule(strings.NewReader(string(manifest)))
			if err != nil {
				return v1.Module{}, err
			}
			if parsed.Name != source {
				return v1.Module{}, fmt.Errorf("%s declares module %s, want %s", dir, parsed.Name, source)
			}
			return parsed, nil
		}
		legacy, err := modfile.Parse(manifest)
		if err != nil {
			return v1.Module{}, fmt.Errorf("legacy protobuf.mod: %w", err)
		}
		for _, raw := range legacy.Direct {
			requirement, err := parseLegacyV1Requirement(raw)
			if err != nil {
				return v1.Module{}, err
			}
			module.Requires = append(module.Requires, requirement)
		}
	}
	bufRoots, foundBuf, err := readBufDependencyRoots(dir)
	if err != nil {
		return v1.Module{}, err
	}
	legacyRoots, legacyRequires, err := readLegacyEasyPRootsAndRequires(dir)
	if err != nil {
		return v1.Module{}, err
	}
	module.Requires = append(module.Requires, legacyRequires...)
	if foundBuf {
		module.Roots = bufRoots
	} else {
		module.Roots = legacyRoots
	}
	if len(module.Roots) == 0 {
		module.Roots = []string{"."}
	}
	for _, root := range module.Roots {
		if !filepath.IsLocal(root) && root != "." {
			return v1.Module{}, fmt.Errorf("dependency %s root %q leaves the repository", source, root)
		}
	}
	return module, nil
}
