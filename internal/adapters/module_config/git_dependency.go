package moduleconfig

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/easyp-tech/easyp/internal/adapters/modfile"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

const dependencyManifestFile = "protobuf.mod"

// ReadGitDependency adapts metadata in a checked-out Git repository to
// the roots and requirements needed by the v1 resolver. Its identity is the
// source named by the requiring module for pre-v1 repositories.
func ReadGitDependency(dir, source string) (v1.Module, error) {
	module, isV1, err := readGitDependencyManifest(dir, source)
	if err != nil {
		return v1.Module{}, fmt.Errorf("readGitDependencyManifest: %w", err)
	}
	if isV1 {
		return module, nil
	}
	bufRoots, foundBuf, err := readBufDependencyRoots(dir)
	if err != nil {
		return v1.Module{}, fmt.Errorf("readBufDependencyRoots: %w", err)
	}
	legacyRoots, legacyRequires, err := readLegacyEasyPRootsAndRequires(dir)
	if err != nil {
		return v1.Module{}, fmt.Errorf("readLegacyEasyPRootsAndRequires: %w", err)
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

func readGitDependencyManifest(dir, source string) (v1.Module, bool, error) {
	module := v1.Module{Name: source}
	manifest, found, err := readOptionalDependencyConfig(dir, dependencyManifestFile)
	if err != nil {
		return v1.Module{}, false, fmt.Errorf("readOptionalDependencyConfig: %w", err)
	}
	if !found {
		return module, false, nil
	}
	if v1.IsModuleManifest(manifest) {
		parsed, err := v1.ParseModule(bytes.NewReader(manifest))
		if err != nil {
			return v1.Module{}, false, fmt.Errorf("ParseModule: %w", err)
		}
		if parsed.Name != source {
			return v1.Module{}, false, fmt.Errorf("%s declares module %s, want %s", dir, parsed.Name, source)
		}
		return parsed, true, nil
	}
	legacy, err := modfile.Parse(manifest)
	if err != nil {
		return v1.Module{}, false, fmt.Errorf("Parse: %s: %w", dependencyManifestFile, err)
	}
	for _, raw := range legacy.Direct {
		requirement, err := parseLegacyV1Requirement(raw)
		if err != nil {
			return v1.Module{}, false, fmt.Errorf("parseLegacyV1Requirement: %w", err)
		}
		module.Requires = append(module.Requires, requirement)
	}
	return module, false, nil
}

func readOptionalDependencyConfig(dir, filename string) ([]byte, bool, error) {
	raw, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("ReadFile: %w", err)
	}
	return raw, true, nil
}
